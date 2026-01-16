package main

import (
	"net/http"
	"testing"
)

func TestTaskValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid task with all fields",
			schemaName:  "proto.Task",
			payload:     map[string]interface{}{"name": "Complete project", "description": "Finish the validation service", "timestamp": "2024-01-15T10:30:00Z", "status": 1},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid task with minimal required fields",
			schemaName:  "proto.Task",
			payload:     map[string]interface{}{"name": "Task", "timestamp": "2024-01-15T10:30:00Z", "status": 2},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing required name",
			schemaName:  "proto.Task",
			payload:     map[string]interface{}{"timestamp": "2024-01-15T10:30:00Z", "status": 1},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "missing required timestamp",
			schemaName:  "proto.Task",
			payload:     map[string]interface{}{"name": "Task", "status": 1},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "missing required status",
			schemaName:  "proto.Task",
			payload:     map[string]interface{}{"name": "Task", "timestamp": "2024-01-15T10:30:00Z"},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "empty name",
			schemaName:  "proto.Task",
			payload:     map[string]interface{}{"name": "", "timestamp": "2024-01-15T10:30:00Z", "status": 1},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "invalid status UNSPECIFIED",
			schemaName:  "proto.Task",
			payload:     map[string]interface{}{"name": "Task", "timestamp": "2024-01-15T10:30:00Z", "status": 0},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "valid status OPEN",
			schemaName:  "proto.Task",
			payload:     map[string]interface{}{"name": "Task", "timestamp": "2024-01-15T10:30:00Z", "status": 1},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid status COMPLETED",
			schemaName:  "proto.Task",
			payload:     map[string]interface{}{"name": "Task", "timestamp": "2024-01-15T10:30:00Z", "status": 4},
			wantSuccess: true,
			wantErrors:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, statusCode, err := callValidateAPI(t, baseURL, tt.schemaName, tt.payload)

			if err != nil {
				if statusCode == http.StatusBadRequest {
					// Expected for some test cases
					if tt.wantSuccess {
						t.Errorf("Unexpected error: %v", err)
					}
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

func TestUpdateTaskValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid update with status and comment",
			schemaName:  "proto.UpdateTask",
			payload:     map[string]interface{}{"status": 1, "comment": "Updated status"},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid update with status only",
			schemaName:  "proto.UpdateTask",
			payload:     map[string]interface{}{"status": 2},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing required status",
			schemaName:  "proto.UpdateTask",
			payload:     map[string]interface{}{"comment": "Some comment"},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "missing comment when status is BLOCKED - CEL constraint",
			schemaName:  "proto.UpdateTask",
			payload:     map[string]interface{}{"status": 3},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "valid comment when status is BLOCKED",
			schemaName:  "proto.UpdateTask",
			payload:     map[string]interface{}{"status": 3, "comment": "Blocked due to dependency"},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "comment not required for non-BLOCKED status",
			schemaName:  "proto.UpdateTask",
			payload:     map[string]interface{}{"status": 1},
			wantSuccess: true,
			wantErrors:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, statusCode, err := callValidateAPI(t, baseURL, tt.schemaName, tt.payload)

			if err != nil {
				if statusCode == http.StatusBadRequest {
					if tt.wantSuccess {
						t.Errorf("Unexpected error: %v", err)
					}
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
