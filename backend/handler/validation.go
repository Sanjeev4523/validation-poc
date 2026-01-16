package handler

import (
	"encoding/json"
	"net/http"
	"validation-service/backend/logger"
	"validation-service/backend/service"
)

// ValidationHandler handles HTTP requests for proto validation
type ValidationHandler struct {
	validationService *service.ValidationService
}

// NewValidationHandler creates a new validation handler
func NewValidationHandler(validationService *service.ValidationService) *ValidationHandler {
	return &ValidationHandler{
		validationService: validationService,
	}
}

// ValidateProtoRequest represents the request payload
type ValidateProtoRequest struct {
	SchemaName string          `json:"schemaName"`
	Payload    json.RawMessage `json:"payload"`
	Commit     string          `json:"commit,omitempty"` // Optional commit ID, defaults to "main"
}

// ValidateProtoResponse represents the response payload
type ValidateProtoResponse struct {
	Success bool                      `json:"success"`
	Errors  []service.ValidationError `json:"errors"`
}

// ValidateProto handles POST /api/v1/validate-proto
func (h *ValidationHandler) ValidateProto(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Received validation request: method=%s, path=%s, remote=%s", r.Method, r.URL.Path, r.RemoteAddr)

	// Only allow POST method
	if r.Method != http.MethodPost {
		logger.Debug("Method not allowed: %s (expected POST)", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req ValidateProtoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Debug("Failed to decode request body: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SchemaName == "" {
		logger.Debug("Missing required field: schemaName")
		http.Error(w, "schemaName is required", http.StatusBadRequest)
		return
	}

	if len(req.Payload) == 0 {
		logger.Debug("Missing required field: payload")
		http.Error(w, "payload is required", http.StatusBadRequest)
		return
	}

	// Set default commit to "main" if not provided
	commit := req.Commit
	if commit == "" {
		commit = "main"
	}

	logger.Info("Processing validation request for schemaName=%s, commit=%s", req.SchemaName, commit)

	// Call validation service
	success, errors, err := h.validationService.ValidateProto(req.SchemaName, req.Payload, commit)
	if err != nil {
		logger.Debug("Validation service error for schemaName=%s: %v", req.SchemaName, err)
		// Check if it's a client error (unknown schema, invalid JSON, etc.)
		if isClientError(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Server error
		logger.Error("Internal server error during validation: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Build response
	response := ValidateProtoResponse{
		Success: success,
		Errors:  errors,
	}

	// Write response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response: %v", err)
		return
	}

	if success {
		logger.Info("Validation succeeded for schemaName=%s", req.SchemaName)
	} else {
		logger.Info("Validation failed for schemaName=%s with %d error(s)", req.SchemaName, len(errors))
	}
}

// isClientError determines if an error is a client error (400) vs server error (500)
func isClientError(err error) bool {
	// Client errors: unknown schema, invalid JSON, etc.
	// Most validation service errors are client errors
	return true
}
