package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"validation-service/backend/handler"
	"validation-service/backend/logger"
	"validation-service/backend/service"

	"buf.build/go/protovalidate"
)

// startTestServer starts a test server on an available port and returns the base URL
func startTestServer(t *testing.T) string {
	// Initialize logger
	logger.Init()

	// Create validator instance
	validator, err := protovalidate.New()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Get base path
	basePath, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("Failed to get base path: %v", err)
	}

	// Initialize services
	_ = service.NewSchemaService("sanjeev-personal", "validation", basePath) // Not used in tests but needed for initialization
	validationService := service.NewValidationService(validator)
	validationHandler := handler.NewValidationHandler(validationService)

	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/validate-proto", validationHandler.ValidateProto)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to be ready
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get(baseURL + "/api/v1/validate-proto")
		if err == nil {
			resp.Body.Close()
			// Server is ready
			break
		}
		if i == maxAttempts-1 {
			t.Fatalf("Server failed to start on %s after %d attempts", baseURL, maxAttempts)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Cleanup function
	t.Cleanup(func() {
		server.Close()
	})

	return baseURL
}

// validateProtoRequest represents the request payload for validation API
type validateProtoRequest struct {
	SchemaName string          `json:"schemaName"`
	Payload    json.RawMessage `json:"payload"`
}

// validateProtoResponse represents the response from validation API
type validateProtoResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Friendly  string `json:"friendly"`
		Technical string `json:"technical"`
	} `json:"errors"`
}

// callValidateAPI makes a POST request to the validate-proto endpoint
func callValidateAPI(t *testing.T, baseURL string, schemaName string, payload interface{}) (*validateProtoResponse, int, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal payload: %w", err)
	}

	reqBody := validateProtoRequest{
		SchemaName: schemaName,
		Payload:    payloadBytes,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/v1/validate-proto", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var result validateProtoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// If response is not JSON (e.g., error message), read as text
		bodyBytes := make([]byte, 1024)
		n, _ := resp.Body.Read(bodyBytes)
		return nil, resp.StatusCode, fmt.Errorf("failed to decode response (status %d): %s", resp.StatusCode, string(bodyBytes[:n]))
	}

	return &result, resp.StatusCode, nil
}

func TestMain(m *testing.M) {
	// Set environment variable to reduce log noise during tests
	os.Setenv("LOG_LEVEL", "ERROR")
	code := m.Run()
	os.Exit(code)
}
