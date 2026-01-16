package service

import (
	"fmt"
	"validation-service/backend/logger"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

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
// Returns success status, array of error messages, and any processing error
func (s *ValidationService) ValidateProto(schemaName string, jsonPayload []byte) (bool, []string, error) {
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
		var errors []string
		if validationErr, ok := err.(*protovalidate.ValidationError); ok {
			// protovalidate.ValidationError contains detailed error information
			errors = s.collectValidationErrors(validationErr)
		} else {
			// Fallback to simple error message
			errors = []string{err.Error()}
		}

		logger.Info("Validation failed for schemaName=%s with %d error(s)", schemaName, len(errors))
		return false, errors, nil
	}

	logger.Info("Validation succeeded for schemaName=%s", schemaName)
	return true, []string{}, nil
}

// collectValidationErrors extracts error messages from a ValidationError
func (s *ValidationService) collectValidationErrors(err *protovalidate.ValidationError) []string {
	var errors []string

	// Add the main violation message
	if err.Violations != nil {
		for _, violation := range err.Violations {
			// Access fields through the Proto field
			proto := violation.Proto
			if proto == nil {
				// Fallback to String() method
				errors = append(errors, violation.String())
				continue
			}

			// Get field path and message from the proto
			fieldPath := protovalidate.FieldPathString(proto.GetField())
			message := proto.GetMessage()

			if fieldPath != "" {
				errors = append(errors, fmt.Sprintf("field '%s': %s", fieldPath, message))
			} else if message != "" {
				errors = append(errors, message)
			} else {
				// Fallback to String() method
				errors = append(errors, violation.String())
			}
		}
	}

	// If no violations found, use the error message itself
	if len(errors) == 0 {
		errors = []string{err.Error()}
	}

	return errors
}
