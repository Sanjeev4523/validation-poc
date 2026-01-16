package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"validation-service/backend/config"
	"validation-service/backend/logger"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// ValidationError represents a validation error with both friendly and technical messages
type ValidationError struct {
	Friendly  string `json:"friendly"`  // Human-readable message
	Technical string `json:"technical"` // Original technical error
}

// ValidationService handles proto validation using dynamic messages
type ValidationService struct {
	validator        protovalidate.Validator
	schemaSourceMode config.SchemaSourceMode
	bsrOrg           string
	bsrModule        string
	bsrToken         string
	httpClient       *http.Client
}

// NewValidationService creates a new validation service instance
func NewValidationService(validator protovalidate.Validator, schemaSourceMode config.SchemaSourceMode, bsrOrg, bsrModule, bsrToken string) *ValidationService {
	logger.Debug("Initializing ValidationService with mode=%d, org=%s, module=%s", schemaSourceMode, bsrOrg, bsrModule)
	return &ValidationService{
		validator:        validator,
		schemaSourceMode: schemaSourceMode,
		bsrOrg:           bsrOrg,
		bsrModule:        bsrModule,
		bsrToken:         bsrToken,
		httpClient:       &http.Client{},
	}
}

// GetFileDescriptorSetRequest represents the request body for BSR Reflection API
type GetFileDescriptorSetRequest struct {
	Module  string   `json:"module"`
	Version string   `json:"version,omitempty"`
	Symbols []string `json:"symbols,omitempty"`
}

// GetFileDescriptorSetResponse represents the response from BSR Reflection API
// The fileDescriptorSet field is a JSON object that needs to be unmarshaled separately
type GetFileDescriptorSetResponse struct {
	FileDescriptorSet json.RawMessage `json:"fileDescriptorSet"`
	Version           string          `json:"version,omitempty"`
}

// fetchDescriptorFromBSR fetches the FileDescriptorSet from BSR using the Reflection API
// and returns a *protoregistry.Files
// schemaName is the fully qualified message name (e.g., "proto.Task") to include in symbols
func (s *ValidationService) fetchDescriptorFromBSR(schemaName string) (*protoregistry.Files, error) {
	// Build module name in format: buf.build/{org}/{module}
	moduleName := fmt.Sprintf("buf.build/%s/%s", s.bsrOrg, s.bsrModule)

	// Get version from environment variable, default to "latest"
	version := config.GetEnv("BSR_VERSION", "main")

	// Build request body with symbols (fully qualified message name)
	requestBody := GetFileDescriptorSetRequest{
		Module:  moduleName,
		Version: version,
		Symbols: []string{schemaName}, // Include the fully qualified message name
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("Failed to marshal request body: %v", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build BSR Reflection API URL
	url := "https://buf.build/buf.reflect.v1beta1.FileDescriptorSetService/GetFileDescriptorSet"

	// Log URL and request body in debug mode
	logger.Debug("BSR Reflection API URL: %s", url)
	logger.Debug("BSR Reflection API Request Body: %s", string(jsonBody))
	logger.Debug("Fetching descriptor from BSR Reflection API: module=%s, version=%s, symbols=%v", moduleName, version, requestBody.Symbols)

	// Create HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Error("Failed to create HTTP request for URL %s: %v", url, err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if s.bsrToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.bsrToken))
		logger.Debug("Added Bearer token to BSR request")
	}

	// Execute the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("HTTP POST request failed for URL %s: %v", url, err)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	logger.Debug("BSR HTTP response status: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode == http.StatusNotFound {
		logger.Debug("Descriptor not found in BSR (404)")
		return nil, fmt.Errorf("descriptor not found in BSR")
	}

	if resp.StatusCode != http.StatusOK {
		// Try to read error body for better error messages
		errorBody, _ := io.ReadAll(resp.Body)
		logger.Error("BSR returned unexpected status code %d: %s", resp.StatusCode, string(errorBody))
		return nil, fmt.Errorf("BSR returned status code %d", resp.StatusCode)
	}

	// Read JSON response
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read BSR response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Debug("Successfully read BSR response body (size: %d bytes)", len(data))

	// Parse JSON response
	var apiResponse GetFileDescriptorSetResponse
	if err := json.Unmarshal(data, &apiResponse); err != nil {
		logger.Error("Failed to unmarshal JSON response: %v", err)
		return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	if len(apiResponse.FileDescriptorSet) == 0 {
		logger.Error("FileDescriptorSet is empty in API response")
		return nil, fmt.Errorf("FileDescriptorSet is empty in API response")
	}

	// Unmarshal FileDescriptorSet from JSON using protojson
	var fds descriptorpb.FileDescriptorSet
	unmarshalOpts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	if err := unmarshalOpts.Unmarshal(apiResponse.FileDescriptorSet, &fds); err != nil {
		logger.Error("Failed to unmarshal FileDescriptorSet from JSON: %v", err)
		return nil, fmt.Errorf("failed to unmarshal FileDescriptorSet: %w", err)
	}

	// Convert FileDescriptorSet to *protoregistry.Files
	files, err := protodesc.NewFiles(&fds)
	if err != nil {
		logger.Error("Failed to create Files from FileDescriptorSet: %v", err)
		return nil, fmt.Errorf("failed to create Files: %w", err)
	}

	logger.Debug("Successfully created Files from BSR descriptor (version: %s)", apiResponse.Version)
	return files, nil
}

// findMessageDescriptor finds a message descriptor by fully qualified name
// It tries the provided files first, then falls back to GlobalFiles
func (s *ValidationService) findMessageDescriptor(schemaName string, files *protoregistry.Files) (protoreflect.MessageDescriptor, error) {
	fullName := protoreflect.FullName(schemaName)

	// Try to find in provided files first
	if files != nil {
		desc, err := files.FindDescriptorByName(fullName)
		if err == nil {
			if md, ok := desc.(protoreflect.MessageDescriptor); ok {
				logger.Debug("Found message descriptor in provided files: %s", schemaName)
				return md, nil
			}
		}
	}

	// Fallback to GlobalFiles
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(fullName)
	if err != nil {
		return nil, err
	}

	md, ok := desc.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("schema name %s does not refer to a message", schemaName)
	}

	logger.Debug("Found message descriptor in GlobalFiles: %s", schemaName)
	return md, nil
}

// ValidateProto validates a JSON payload against a protobuf message definition
// Returns success status, array of validation errors, and any processing error
func (s *ValidationService) ValidateProto(schemaName string, jsonPayload []byte) (bool, []ValidationError, error) {
	logger.Debug("ValidateProto called for schemaName=%s, mode=%d", schemaName, s.schemaSourceMode)

	var md protoreflect.MessageDescriptor
	var err error

	// Step 1: Find message descriptor based on mode
	if s.schemaSourceMode == config.BSROnly {
		// BSROnly: Always fetch from BSR
		logger.Debug("BSROnly mode: fetching descriptor from BSR for %s", schemaName)
		files, err := s.fetchDescriptorFromBSR(schemaName)
		if err != nil {
			logger.Debug("Failed to fetch descriptor from BSR for schemaName=%s: %v", schemaName, err)
			return false, nil, fmt.Errorf("failed to fetch descriptor from BSR: %w", err)
		}
		md, err = s.findMessageDescriptor(schemaName, files)
		if err != nil {
			logger.Debug("Failed to find descriptor in BSR files for schemaName=%s: %v", schemaName, err)
			return false, nil, fmt.Errorf("unknown schema name: %s", schemaName)
		}
	} else if s.schemaSourceMode == config.LocalOnly {
		// LocalOnly: Only use GlobalFiles
		logger.Debug("LocalOnly mode: checking GlobalFiles for %s", schemaName)
		md, err = s.findMessageDescriptor(schemaName, nil)
		if err != nil {
			logger.Debug("Failed to find descriptor in GlobalFiles for schemaName=%s: %v", schemaName, err)
			return false, nil, fmt.Errorf("unknown schema name: %s", schemaName)
		}
	} else {
		// LocalThenBSR: Try local first, then fallback to BSR
		logger.Debug("LocalThenBSR mode: checking GlobalFiles first for %s", schemaName)
		md, err = s.findMessageDescriptor(schemaName, nil)
		if err != nil {
			logger.Debug("Not found in GlobalFiles, fetching from BSR for schemaName=%s", schemaName)
			// Fallback to BSR
			files, bsrErr := s.fetchDescriptorFromBSR(schemaName)
			if bsrErr != nil {
				logger.Debug("Failed to fetch descriptor from BSR for schemaName=%s: %v", schemaName, bsrErr)
				return false, nil, fmt.Errorf("unknown schema name: %s (local and BSR lookup failed)", schemaName)
			}
			md, err = s.findMessageDescriptor(schemaName, files)
			if err != nil {
				logger.Debug("Failed to find descriptor in BSR files for schemaName=%s: %v", schemaName, err)
				return false, nil, fmt.Errorf("unknown schema name: %s", schemaName)
			}
		}
	}

	// Step 2: Create dynamic message
	msg := dynamicpb.NewMessage(md)
	logger.Debug("Created dynamic message for schemaName=%s", schemaName)

	// Step 3: Unmarshal JSON to dynamic message
	unmarshalOpts := protojson.UnmarshalOptions{
		DiscardUnknown: true, // Ignore unknown fields
	}
	if err := unmarshalOpts.Unmarshal(jsonPayload, msg); err != nil {
		logger.Debug("Failed to unmarshal JSON for schemaName=%s: %v", schemaName, err)
		return false, nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	logger.Debug("Successfully unmarshaled JSON for schemaName=%s", schemaName)

	// Step 4: Validate using protovalidate
	if err := s.validator.Validate(msg); err != nil {
		logger.Debug("Validation failed for schemaName=%s: %v", schemaName, err)

		// Step 5: Collect validation errors
		var errors []ValidationError
		if validationErr, ok := err.(*protovalidate.ValidationError); ok {
			// protovalidate.ValidationError contains detailed error information
			errors = s.collectValidationErrors(validationErr)
		} else {
			// Fallback to simple error message
			technical := err.Error()
			errors = []ValidationError{
				{
					Friendly:  s.makeFriendlyError(technical),
					Technical: technical,
				},
			}
		}

		logger.Info("Validation failed for schemaName=%s with %d error(s)", schemaName, len(errors))
		return false, errors, nil
	}

	logger.Info("Validation succeeded for schemaName=%s", schemaName)
	return true, []ValidationError{}, nil
}

// collectValidationErrors extracts error messages from a ValidationError and formats them
func (s *ValidationService) collectValidationErrors(err *protovalidate.ValidationError) []ValidationError {
	var errors []ValidationError

	// Add the main violation message
	if err.Violations != nil {
		for _, violation := range err.Violations {
			// Access fields through the Proto field
			proto := violation.Proto
			technical := violation.String()
			var friendly string

			if proto == nil {
				// Fallback: use technical error as friendly
				friendly = s.makeFriendlyError(technical)
			} else {
				// Get field path and message from the proto
				fieldPath := protovalidate.FieldPathString(proto.GetField())
				message := proto.GetMessage()

				if message != "" {
					// Use the message from proto definition (this is the friendly message)
					if fieldPath != "" {
						friendly = fmt.Sprintf("field '%s': %s", fieldPath, message)
					} else {
						friendly = message
					}
				} else if fieldPath != "" {
					// No message, but we have a field path
					friendly = s.makeFriendlyError(technical)
				} else {
					// Fallback: try to make friendly from technical
					friendly = s.makeFriendlyError(technical)
				}
			}

			errors = append(errors, ValidationError{
				Friendly:  friendly,
				Technical: technical,
			})
		}
	}

	// If no violations found, use the error message itself
	if len(errors) == 0 {
		technical := err.Error()
		errors = []ValidationError{
			{
				Friendly:  s.makeFriendlyError(technical),
				Technical: technical,
			},
		}
	}

	return errors
}

// makeFriendlyError attempts to create a human-friendly error message from a technical error
func (s *ValidationService) makeFriendlyError(technical string) string {
	// Check if it's a CEL compilation error
	if strings.Contains(technical, "compilation error") {
		// Try to extract constraint ID and create a friendly message
		constraintID := s.extractConstraintID(technical)
		if constraintID != "" {
			// Try to create a readable message from constraint ID
			friendly := s.constraintIDToFriendly(constraintID)
			if friendly != "" {
				return friendly
			}
		}
		// If we can't parse it, try to clean up the error message
		return s.cleanupTechnicalError(technical)
	}

	// For other errors, try to clean them up
	return s.cleanupTechnicalError(technical)
}

// extractConstraintID extracts the constraint ID from a CEL compilation error
func (s *ValidationService) extractConstraintID(errorMsg string) string {
	// Pattern: "compilation error: failed to compile expression <constraint_id>:"
	re := regexp.MustCompile(`compilation error:.*expression\s+([a-zA-Z_][a-zA-Z0-9_]*):`)
	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// constraintIDToFriendly converts a constraint ID to a friendly message
func (s *ValidationService) constraintIDToFriendly(constraintID string) string {
	// Map known constraint IDs to friendly messages
	// This can be extended with more mappings
	constraintMap := map[string]string{
		"comment_required_if_blocked": "comment is required when status is TASK_STATUS_BLOCKED",
	}

	if friendly, ok := constraintMap[constraintID]; ok {
		return friendly
	}

	// Try to create a friendly message from the constraint ID
	// Convert snake_case to readable text
	readable := strings.ReplaceAll(constraintID, "_", " ")
	readable = strings.ToLower(readable)
	return fmt.Sprintf("Validation failed: %s", readable)
}

// cleanupTechnicalError attempts to clean up technical error messages
func (s *ValidationService) cleanupTechnicalError(technical string) string {
	// Remove common technical prefixes and suffixes
	cleaned := technical

	// Remove "ERROR:" prefix if present
	cleaned = regexp.MustCompile(`(?i)^ERROR:\s*`).ReplaceAllString(cleaned, "")

	// Remove line number references like "<input>:1:16:"
	cleaned = regexp.MustCompile(`<input>:\d+:\d+:\s*`).ReplaceAllString(cleaned, "")

	// Remove "(in container '')" suffix
	cleaned = regexp.MustCompile(`\s*\(in container '[^']*'\)`).ReplaceAllString(cleaned, "")

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	// If cleaned is empty or same as original, return original
	if cleaned == "" || cleaned == technical {
		return technical
	}

	return cleaned
}
