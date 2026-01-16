package service

import (
	"fmt"
	"regexp"
	"strings"
	"validation-service/backend/logger"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

// ValidationError represents a validation error with both friendly and technical messages
type ValidationError struct {
	Friendly  string `json:"friendly"`  // Human-readable message
	Technical string `json:"technical"` // Original technical error
}

// ValidationService handles proto validation using dynamic messages
type ValidationService struct {
	validator protovalidate.Validator
}

// NewValidationService creates a new validation service instance
func NewValidationService(validator protovalidate.Validator) *ValidationService {
	logger.Debug("Initializing ValidationService")
	return &ValidationService{
		validator: validator,
	}
}

// ValidateProto validates a JSON payload against a protobuf message definition
// Returns success status, array of validation errors, and any processing error
func (s *ValidationService) ValidateProto(schemaName string, jsonPayload []byte) (bool, []ValidationError, error) {
	logger.Debug("ValidateProto called for schemaName=%s", schemaName)

	// Step 1: Find message descriptor in global registry
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(schemaName))
	if err != nil {
		logger.Debug("Failed to find descriptor for schemaName=%s: %v", schemaName, err)
		return false, nil, fmt.Errorf("unknown schema name: %s", schemaName)
	}

	// Step 2: Cast to MessageDescriptor
	md, ok := desc.(protoreflect.MessageDescriptor)
	if !ok {
		logger.Debug("Descriptor is not a message descriptor for schemaName=%s", schemaName)
		return false, nil, fmt.Errorf("schema name %s does not refer to a message", schemaName)
	}

	// Step 3: Create dynamic message
	msg := dynamicpb.NewMessage(md)
	logger.Debug("Created dynamic message for schemaName=%s", schemaName)

	// Step 4: Unmarshal JSON to dynamic message
	unmarshalOpts := protojson.UnmarshalOptions{
		DiscardUnknown: true, // Ignore unknown fields
	}
	if err := unmarshalOpts.Unmarshal(jsonPayload, msg); err != nil {
		logger.Debug("Failed to unmarshal JSON for schemaName=%s: %v", schemaName, err)
		return false, nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	logger.Debug("Successfully unmarshaled JSON for schemaName=%s", schemaName)

	// Step 5: Validate using protovalidate
	if err := s.validator.Validate(msg); err != nil {
		logger.Debug("Validation failed for schemaName=%s: %v", schemaName, err)

		// Step 6: Collect validation errors
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
