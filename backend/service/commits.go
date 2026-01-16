package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"validation-service/backend/logger"
)

// CommitsService handles fetching commit history from Buf registry
type CommitsService struct {
	bsrOrg     string
	bsrModule  string
	bsrToken   string
	httpClient *http.Client
}

// NewCommitsService creates a new commits service instance
func NewCommitsService(bsrOrg, bsrModule, bsrToken string) *CommitsService {
	logger.Debug("Initializing CommitsService with org=%s, module=%s", bsrOrg, bsrModule)
	return &CommitsService{
		bsrOrg:     bsrOrg,
		bsrModule:  bsrModule,
		bsrToken:   bsrToken,
		httpClient: &http.Client{},
	}
}

// ListLabelHistoryRequest represents the request body for Buf LabelService ListLabelHistory API
type ListLabelHistoryRequest struct {
	PageSize  int32                   `json:"pageSize,omitempty"`
	LabelRef  *LabelRef               `json:"labelRef,omitempty"`
	Order     string                  `json:"order,omitempty"`
	PageToken string                  `json:"pageToken,omitempty"`
}

// LabelRef represents the label reference in the request
type LabelRef struct {
	Name *LabelName `json:"name,omitempty"`
}

// LabelName represents the label name with owner, module, and label
type LabelName struct {
	Owner  string `json:"owner,omitempty"`
	Module string `json:"module,omitempty"`
	Label  string `json:"label,omitempty"`
}

// ListLabelHistoryResponse represents the response from Buf LabelService ListLabelHistory API
type ListLabelHistoryResponse struct {
	NextPageToken string                 `json:"nextPageToken,omitempty"`
	Values        []LabelHistoryValue    `json:"values,omitempty"`
}

// LabelHistoryValue represents a single commit in the label history
type LabelHistoryValue struct {
	Commit          *Commit          `json:"commit,omitempty"`
	CommitCheckState *CommitCheckState `json:"commitCheckState,omitempty"`
}

// Commit represents commit information
type Commit struct {
	ID             string  `json:"id,omitempty"`
	CreateTime     string  `json:"createTime,omitempty"`
	OwnerID        string  `json:"ownerId,omitempty"`
	ModuleID       string  `json:"moduleId,omitempty"`
	Digest         *Digest `json:"digest,omitempty"`
	CreatedByUserID string `json:"createdByUserId,omitempty"`
}

// Digest represents the commit digest
type Digest struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// CommitCheckState represents the commit check state
type CommitCheckState struct {
	Status    string `json:"status,omitempty"`
	UpdateTime string `json:"updateTime,omitempty"`
}

// ListCommits fetches commit history from Buf registry for the specified label
func (s *CommitsService) ListCommits(pageSize int, label string, pageToken string) (*ListLabelHistoryResponse, error) {
	// Build request body
	requestBody := ListLabelHistoryRequest{
		PageSize: int32(pageSize),
		LabelRef: &LabelRef{
			Name: &LabelName{
				Owner:  s.bsrOrg,
				Module: s.bsrModule,
				Label:  label,
			},
		},
		Order: "ORDER_DESC",
	}

	// Add pageToken if provided
	if pageToken != "" {
		requestBody.PageToken = pageToken
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("Failed to marshal request body: %v", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build Buf LabelService API URL
	url := "https://buf.build/buf.registry.module.v1beta1.LabelService/ListLabelHistory"

	// Log request details in debug mode
	logger.Debug("Buf LabelService API URL: %s", url)
	logger.Debug("Buf LabelService Request Body: %s", string(jsonBody))
	logger.Debug("Fetching commits from Buf: org=%s, module=%s, label=%s, pageSize=%d", s.bsrOrg, s.bsrModule, label, pageSize)

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
		logger.Debug("Added Bearer token to Buf request")
	}

	// Execute the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("HTTP POST request failed for URL %s: %v", url, err)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	logger.Debug("Buf HTTP response status: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode == http.StatusUnauthorized {
		logger.Error("Unauthorized access to Buf API (401) - check BUF_TOKEN")
		return nil, fmt.Errorf("unauthorized: invalid or missing BUF_TOKEN")
	}

	if resp.StatusCode == http.StatusNotFound {
		logger.Debug("Label not found in Buf (404) for label=%s", label)
		return nil, fmt.Errorf("label not found: %s", label)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to read error body for better error messages
		errorBody, _ := io.ReadAll(resp.Body)
		logger.Error("Buf returned unexpected status code %d: %s", resp.StatusCode, string(errorBody))
		return nil, fmt.Errorf("Buf API returned status code %d", resp.StatusCode)
	}

	// Read JSON response
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read Buf response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Debug("Successfully read Buf response body (size: %d bytes)", len(data))

	// Parse JSON response
	var apiResponse ListLabelHistoryResponse
	if err := json.Unmarshal(data, &apiResponse); err != nil {
		logger.Error("Failed to unmarshal JSON response: %v", err)
		return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	logger.Info("Successfully fetched %d commit(s) from Buf for label=%s", len(apiResponse.Values), label)
	return &apiResponse, nil
}
