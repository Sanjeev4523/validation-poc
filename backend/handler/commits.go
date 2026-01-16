package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"validation-service/backend/logger"
	"validation-service/backend/service"
)

// CommitsHandler handles HTTP requests for commit history retrieval
type CommitsHandler struct {
	commitsService *service.CommitsService
}

// NewCommitsHandler creates a new commits handler
func NewCommitsHandler(commitsService *service.CommitsService) *CommitsHandler {
	return &CommitsHandler{
		commitsService: commitsService,
	}
}

// GetCommits handles GET /api/v1/commits
func (h *CommitsHandler) GetCommits(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Received request: method=%s, path=%s, remote=%s", r.Method, r.URL.Path, r.RemoteAddr)

	// Only allow GET method
	if r.Method != http.MethodGet {
		logger.Debug("Method not allowed: %s (expected GET)", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Parse pageSize (optional, default: 26)
	pageSize := 26
	if pageSizeStr := query.Get("pageSize"); pageSizeStr != "" {
		parsed, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			logger.Debug("Invalid pageSize parameter: %s", pageSizeStr)
			http.Error(w, "Invalid pageSize: must be a positive integer", http.StatusBadRequest)
			return
		}
		if parsed <= 0 {
			logger.Debug("Invalid pageSize parameter: %d (must be positive)", parsed)
			http.Error(w, "Invalid pageSize: must be a positive integer", http.StatusBadRequest)
			return
		}
		pageSize = parsed
	}

	// Parse label (optional, default: "main")
	label := query.Get("label")
	if label == "" {
		label = "main"
	}

	// Parse pageToken (optional)
	pageToken := query.Get("pageToken")

	logger.Info("Processing commits request: label=%s, pageSize=%d, pageToken=%s", label, pageSize, pageToken)

	// Get commits from service
	commitsResponse, err := h.commitsService.ListCommits(pageSize, label, pageToken)
	if err != nil {
		logger.Debug("Commits retrieval failed: %v", err)
		h.handleError(w, err)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send JSON response
	if err := json.NewEncoder(w).Encode(commitsResponse); err != nil {
		logger.Error("Failed to encode commits response: %v", err)
		return
	}

	logger.Info("Successfully returned %d commit(s)", len(commitsResponse.Values))
}

// handleError handles errors and returns appropriate HTTP status codes
func (h *CommitsHandler) handleError(w http.ResponseWriter, err error) {
	errorMsg := err.Error()

	// Determine status code based on error message
	switch {
	case strings.Contains(errorMsg, "unauthorized") || strings.Contains(errorMsg, "BUF_TOKEN"):
		logger.Debug("Returning 401 Unauthorized: %s", errorMsg)
		http.Error(w, errorMsg, http.StatusUnauthorized)
	case strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "label not found"):
		logger.Debug("Returning 404 Not Found: %s", errorMsg)
		http.Error(w, errorMsg, http.StatusNotFound)
	case strings.Contains(errorMsg, "Invalid") || strings.Contains(errorMsg, "invalid"):
		logger.Debug("Returning 400 Bad Request: %s", errorMsg)
		http.Error(w, errorMsg, http.StatusBadRequest)
	default:
		logger.Error("Internal server error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
