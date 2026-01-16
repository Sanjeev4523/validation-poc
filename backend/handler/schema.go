package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"validation-service/backend/logger"
	"validation-service/backend/service"
)

// SchemaHandler handles HTTP requests for schema retrieval
type SchemaHandler struct {
	schemaService *service.SchemaService
}

// NewSchemaHandler creates a new schema handler
func NewSchemaHandler(schemaService *service.SchemaService) *SchemaHandler {
	return &SchemaHandler{
		schemaService: schemaService,
	}
}

// GetSchema handles GET /api/v1/schema/{messageName}
func (h *SchemaHandler) GetSchema(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Received request: method=%s, path=%s, remote=%s", r.Method, r.URL.Path, r.RemoteAddr)

	// Only allow GET method
	if r.Method != http.MethodGet {
		logger.Debug("Method not allowed: %s (expected GET)", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract message name from URL path
	// Expected format: /api/v1/schema/{messageName}
	path := r.URL.Path
	prefix := "/api/v1/schema/"

	if !strings.HasPrefix(path, prefix) {
		logger.Debug("Invalid path format: %s (expected prefix: %s)", path, prefix)
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}

	messageName := strings.TrimPrefix(path, prefix)
	if messageName == "" {
		logger.Debug("Empty message name in request path: %s", path)
		http.Error(w, "Message name is required", http.StatusBadRequest)
		return
	}

	// URL decode the message name in case it contains encoded characters
	decodedName, err := url.QueryUnescape(messageName)
	if err != nil {
		logger.Debug("Failed to URL decode message name '%s': %v", messageName, err)
		http.Error(w, "Invalid message name encoding", http.StatusBadRequest)
		return
	}
	if decodedName != messageName {
		logger.Debug("URL decoded message name: '%s' -> '%s'", messageName, decodedName)
	}
	messageName = decodedName

	logger.Info("Processing schema request for messageName=%s", messageName)

	// Get schema from service
	schemaData, err := h.schemaService.GetSchema(messageName)
	if err != nil {
		logger.Debug("Schema retrieval failed for messageName=%s: %v", messageName, err)
		h.handleError(w, err)
		return
	}

	// Parse JSON to validate it's valid JSON
	var schemaJSON interface{}
	if err := json.Unmarshal(schemaData, &schemaJSON); err != nil {
		logger.Warn("Schema from BSR is not valid JSON for messageName=%s: %v", messageName, err)
		// Still return the data, but log the warning
	} else {
		logger.Debug("Schema JSON validation passed for messageName=%s", messageName)
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write response
	bytesWritten, err := w.Write(schemaData)
	if err != nil {
		logger.Error("Failed to write response for messageName=%s: %v", messageName, err)
		return
	}
	logger.Info("Successfully returned schema for messageName=%s (size: %d bytes, written: %d bytes)", messageName, len(schemaData), bytesWritten)
}

// ListProtoFiles handles GET /api/v1/proto-files
func (h *SchemaHandler) ListProtoFiles(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Received request: method=%s, path=%s, remote=%s", r.Method, r.URL.Path, r.RemoteAddr)

	// Only allow GET method
	if r.Method != http.MethodGet {
		logger.Debug("Method not allowed: %s (expected GET)", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logger.Info("Processing list proto files request")

	// Get proto files list from service
	protoFiles, err := h.schemaService.ListProtoFiles()
	if err != nil {
		logger.Error("Failed to list proto files: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send JSON response
	if err := json.NewEncoder(w).Encode(protoFiles); err != nil {
		logger.Error("Failed to encode proto files response: %v", err)
		return
	}

	logger.Info("Successfully returned %d proto file(s)", len(protoFiles))
}

// handleError handles errors and returns appropriate HTTP status codes
func (h *SchemaHandler) handleError(w http.ResponseWriter, err error) {
	errorMsg := err.Error()

	// Determine status code based on error message
	switch {
	case strings.Contains(errorMsg, "invalid message name") || strings.Contains(errorMsg, "cannot be empty"):
		logger.Debug("Returning 400 Bad Request: %s", errorMsg)
		http.Error(w, errorMsg, http.StatusBadRequest)
	case strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "not found in BSR"):
		logger.Debug("Returning 404 Not Found: %s", errorMsg)
		http.Error(w, errorMsg, http.StatusNotFound)
	default:
		logger.Error("Internal server error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
