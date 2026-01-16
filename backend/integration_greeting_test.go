package main

import (
	"net/http"
	"testing"
)

func TestHelloRequestValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid hello request",
			schemaName:  "proto.HelloRequest",
			payload:     map[string]interface{}{"name": "John"},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "name too short",
			schemaName:  "proto.HelloRequest",
			payload:     map[string]interface{}{"name": "Jo"},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "empty name",
			schemaName:  "proto.HelloRequest",
			payload:     map[string]interface{}{"name": ""},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "name at minimum length",
			schemaName:  "proto.HelloRequest",
			payload:     map[string]interface{}{"name": "Jon"},
			wantSuccess: true,
			wantErrors:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, statusCode, err := callValidateAPI(t, baseURL, tt.schemaName, tt.payload)

			if err != nil {
				if statusCode == http.StatusBadRequest && !tt.wantSuccess {
					return
				}
				t.Fatalf("API call failed: %v", err)
			}

			if statusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", statusCode)
				return
			}

			if result.Success != tt.wantSuccess {
				t.Errorf("Expected success=%v, got success=%v. Errors: %v", tt.wantSuccess, result.Success, result.Errors)
			}

			if tt.wantErrors > 0 && len(result.Errors) != tt.wantErrors {
				t.Errorf("Expected %d validation errors, got %d. Errors: %v", tt.wantErrors, len(result.Errors), result.Errors)
			}
		})
	}
}

func TestHelloResponseValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid hello response",
			schemaName:  "proto.HelloResponse",
			payload:     map[string]interface{}{"message": "Hello, World!"},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid hello response with empty message",
			schemaName:  "proto.HelloResponse",
			payload:     map[string]interface{}{"message": ""},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid hello response without message",
			schemaName:  "proto.HelloResponse",
			payload:     map[string]interface{}{},
			wantSuccess: true,
			wantErrors:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, statusCode, err := callValidateAPI(t, baseURL, tt.schemaName, tt.payload)

			if err != nil {
				if statusCode == http.StatusBadRequest && !tt.wantSuccess {
					return
				}
				t.Fatalf("API call failed: %v", err)
			}

			if statusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", statusCode)
				return
			}

			if result.Success != tt.wantSuccess {
				t.Errorf("Expected success=%v, got success=%v. Errors: %v", tt.wantSuccess, result.Success, result.Errors)
			}

			if tt.wantErrors > 0 && len(result.Errors) != tt.wantErrors {
				t.Errorf("Expected %d validation errors, got %d. Errors: %v", tt.wantErrors, len(result.Errors), result.Errors)
			}
		})
	}
}
