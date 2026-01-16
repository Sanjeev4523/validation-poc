package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"validation-service/backend/config"
	"validation-service/backend/logger"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// SchemaService handles schema retrieval from local filesystem or BSR
type SchemaService struct {
	bsrOrg           string
	bsrModule        string
	basePath         string
	httpClient       *http.Client
	schemaSourceMode config.SchemaSourceMode
	bsrToken         string
}

// NewSchemaService creates a new schema service instance
func NewSchemaService(bsrOrg, bsrModule, basePath string, schemaSourceMode config.SchemaSourceMode) *SchemaService {
	bsrToken := config.GetEnv("BUF_TOKEN", "")
	if bsrToken == "" {
		logger.Warn("BUF_TOKEN is not set. BSR requests may fail for private repositories.")
	} else {
		logger.Debug("BUF_TOKEN is set (length: %d)", len(bsrToken))
	}
	logger.Debug("Initializing SchemaService with org=%s, module=%s, basePath=%s, mode=%d", bsrOrg, bsrModule, basePath, schemaSourceMode)
	return &SchemaService{
		bsrOrg:           bsrOrg,
		bsrModule:        bsrModule,
		basePath:         basePath,
		httpClient:       &http.Client{},
		schemaSourceMode: schemaSourceMode,
		bsrToken:         bsrToken,
	}
}

// GetSchema retrieves the JSON schema for a given message name
// Behavior depends on schemaSourceMode:
// - BSROnly: Fetches directly from BSR (skips local check)
// - LocalOnly: Only checks local files (never fetches from BSR)
// - LocalThenBSR: Checks local first, then falls back to BSR
func (s *SchemaService) GetSchema(messageName string) ([]byte, error) {
	logger.Debug("GetSchema called for messageName=%s, mode=%d", messageName, s.schemaSourceMode)

	// Validate message name format
	if err := s.validateMessageName(messageName); err != nil {
		logger.Debug("Message name validation failed for %s: %v", messageName, err)
		return nil, fmt.Errorf("invalid message name: %w", err)
	}
	logger.Debug("Message name validation passed for %s", messageName)

	// Handle BSROnly mode: skip local check, fetch directly from BSR
	if s.schemaSourceMode == config.BSROnly {
		logger.Debug("BSROnly mode: fetching schema directly from BSR for %s", messageName)
		schema, err := s.fetchFromBSR(messageName)
		if err != nil {
			logger.Error("Failed to fetch schema from BSR for %s: %v", messageName, err)
			return nil, fmt.Errorf("failed to fetch from BSR: %w", err)
		}
		logger.Info("Successfully fetched schema from BSR for %s (size: %d bytes)", messageName, len(schema))
		return schema, nil
	}

	// Handle LocalOnly mode: only check local files, never fetch from BSR
	if s.schemaSourceMode == config.LocalOnly {
		logger.Debug("LocalOnly mode: checking local filesystem for schema: %s", messageName)
		schema, found := s.checkLocalSchema(messageName)
		if !found {
			logger.Error("Schema not found locally for %s and LocalOnly mode is enabled", messageName)
			return nil, fmt.Errorf("schema not found locally for %s", messageName)
		}
		logger.Info("Schema found locally for %s (size: %d bytes)", messageName, len(schema))
		return schema, nil
	}

	// Handle LocalThenBSR mode (default): check local first, then fallback to BSR
	logger.Debug("LocalThenBSR mode: checking local filesystem first for schema: %s", messageName)
	schema, found := s.checkLocalSchema(messageName)
	if found {
		logger.Info("Schema found locally for %s (size: %d bytes)", messageName, len(schema))
		return schema, nil
	}
	logger.Debug("Schema not found locally for %s, attempting BSR fetch", messageName)

	// If not found locally, fetch from BSR
	logger.Debug("Fetching schema from BSR for %s", messageName)
	schema, err := s.fetchFromBSR(messageName)
	if err != nil {
		logger.Error("Failed to fetch schema from BSR for %s: %v", messageName, err)
		return nil, fmt.Errorf("failed to fetch from BSR: %w", err)
	}
	logger.Info("Successfully fetched schema from BSR for %s (size: %d bytes)", messageName, len(schema))

	return schema, nil
}

// validateMessageName validates that the message name follows package.Message format
func (s *SchemaService) validateMessageName(messageName string) error {
	if messageName == "" {
		return fmt.Errorf("message name cannot be empty")
	}

	// Must contain at least one dot (package.Message)
	if !strings.Contains(messageName, ".") {
		return fmt.Errorf("message name must be in format 'package.Message', got: %s", messageName)
	}

	// Validate format: package.Message (alphanumeric, dots, underscores)
	// Pattern: one or more valid identifiers separated by dots
	pattern := `^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)+$`
	matched, err := regexp.MatchString(pattern, messageName)
	if err != nil {
		return fmt.Errorf("error validating message name: %w", err)
	}

	if !matched {
		return fmt.Errorf("message name must match pattern 'package.Message' with valid identifiers, got: %s", messageName)
	}

	return nil
}

// checkLocalSchema checks if the schema exists locally
func (s *SchemaService) checkLocalSchema(messageName string) ([]byte, bool) {
	// Construct local file path: gen/jsonschema/{FULL_NAME}.schema.bundle.json
	schemaPath := filepath.Join(s.basePath, "gen", "jsonschema", fmt.Sprintf("%s.schema.bundle.json", messageName))
	logger.Debug("Checking local schema path: %s", schemaPath)

	data, err := os.ReadFile(schemaPath)
	if err != nil {
		logger.Debug("Local schema file not found or unreadable: %s (error: %v)", schemaPath, err)
		return nil, false
	}

	logger.Debug("Local schema file found and read successfully: %s", schemaPath)
	return data, true
}

// fetchFromBSR fetches the schema directly from BSR via HTTP
func (s *SchemaService) fetchFromBSR(messageName string) ([]byte, error) {
	url := s.buildBSRURL(messageName)
	logger.Debug("Fetching from BSR URL: %s", url)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Failed to create HTTP request for URL %s: %v", url, err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Bearer token authentication if available
	if s.bsrToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.bsrToken))
		logger.Debug("Added Bearer token to BSR request")
	}

	// Execute the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("HTTP GET request failed for URL %s: %v", url, err)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	logger.Debug("BSR HTTP response status: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode == http.StatusNotFound {
		logger.Debug("Schema not found in BSR (404) for messageName=%s", messageName)
		return nil, fmt.Errorf("schema not found in BSR")
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("BSR returned unexpected status code %d for messageName=%s", resp.StatusCode, messageName)
		return nil, fmt.Errorf("BSR returned status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read BSR response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Debug("Successfully read BSR response body (size: %d bytes)", len(data))
	return data, nil
}

// buildBSRURL constructs the BSR URL for fetching the schema
func (s *SchemaService) buildBSRURL(messageName string) string {
	// URL format: https://buf.build/gen/archive/{org}/{module}/bufbuild/protoschema-jsonschema/raw/latest/{FULL_NAME}.schema.bundle.json
	url := fmt.Sprintf(
		"https://buf.build/gen/archive/%s/%s/bufbuild/protoschema-jsonschema/raw/latest/%s.schema.bundle.json",
		s.bsrOrg,
		s.bsrModule,
		messageName,
	)
	logger.Debug("Built BSR URL for messageName=%s: %s", messageName, url)
	return url
}

// ProtoFile represents a proto message file with metadata
type ProtoFile struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	FullyQualifiedName string `json:"fullyQualifiedName"`
}

// ListProtoFiles enumerates all available proto message types from the protobuf registry
func (s *SchemaService) ListProtoFiles() ([]ProtoFile, error) {
	logger.Debug("ListProtoFiles called")

	var protoFiles []ProtoFile
	seenMessages := make(map[string]bool) // Track seen messages to avoid duplicates

	// Recursively walk nested message descriptors
	var walkMessages func(md protoreflect.MessageDescriptor)
	walkMessages = func(md protoreflect.MessageDescriptor) {
		fullyQualifiedName := string(md.FullName())

		// Only include messages in the "proto" namespace
		if !strings.HasPrefix(fullyQualifiedName, "proto.") {
			logger.Debug("Skipping message not in proto namespace: %s", fullyQualifiedName)
			// Still process nested messages in case they're in proto namespace
			nested := md.Messages()
			for i := 0; i < nested.Len(); i++ {
				walkMessages(nested.Get(i))
			}
			return
		}

		// Skip if we've already seen this message
		if seenMessages[fullyQualifiedName] {
			return
		}
		seenMessages[fullyQualifiedName] = true

		// Get the message name (last part of the fully qualified name)
		name := string(md.Name())

		// Extract description - try to get from source locations if available
		description := ""

		// Try to get description from the file descriptor's source locations
		// Note: Source code info may not always be available in compiled descriptors
		if parent := md.Parent(); parent != nil {
			if fd, ok := parent.(protoreflect.FileDescriptor); ok {
				loc := fd.SourceLocations().ByDescriptor(md)
				leadingComments := loc.LeadingComments
				if leadingComments != "" {
					// Clean up the comment - remove extra whitespace and newlines
					description = strings.TrimSpace(leadingComments)
					// Remove leading comment markers if present
					lines := strings.Split(description, "\n")
					var cleanedLines []string
					for _, line := range lines {
						line = strings.TrimSpace(line)
						line = strings.TrimPrefix(line, "//")
						line = strings.TrimPrefix(line, "/*")
						line = strings.TrimSuffix(line, "*/")
						line = strings.TrimSpace(line)
						if line != "" {
							cleanedLines = append(cleanedLines, line)
						}
					}
					description = strings.Join(cleanedLines, " ")
				}
			}
		}

		// If no description found, use a default based on the message name
		if description == "" {
			description = fmt.Sprintf("%s message", s.formatMessageName(name))
		}

		protoFiles = append(protoFiles, ProtoFile{
			Name:               s.formatMessageName(name),
			Description:        description,
			FullyQualifiedName: fullyQualifiedName,
		})

		logger.Debug("Found proto message: %s (%s)", fullyQualifiedName, name)

		// Recursively process nested messages
		nested := md.Messages()
		for i := 0; i < nested.Len(); i++ {
			walkMessages(nested.Get(i))
		}
	}

	// Iterate through all file descriptors in the global registry
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		// For each file, get all top-level messages
		msgs := fd.Messages()
		for i := 0; i < msgs.Len(); i++ {
			walkMessages(msgs.Get(i))
		}
		return true // Continue iteration
	})

	logger.Info("ListProtoFiles found %d proto message(s)", len(protoFiles))
	return protoFiles, nil
}

// formatMessageName converts CamelCase to a more readable format
func (s *SchemaService) formatMessageName(name string) string {
	// Simple conversion: insert spaces before capital letters (except the first one)
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}
	return result.String()
}
