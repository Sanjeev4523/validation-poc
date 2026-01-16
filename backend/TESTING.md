# Integration Testing Guide

This guide explains how to run the integration tests for the validation API.

## Overview

The integration tests start an actual HTTP server and make real API calls to test the validation endpoints. This ensures that the entire stack (HTTP handler → validation service → protovalidate) works correctly.

## Running Tests

### Run All Integration Tests

```bash
cd backend
go test -v
```

### Run Specific Test Suite

```bash
# Run Task validation tests
go test -run TestTaskValidationAPI -v

# Run UpdateTask validation tests
go test -run TestUpdateTaskValidationAPI -v
```

### Run with Verbose Output

```bash
go test -v
```

### Run Specific Test Case

```bash
# Run a specific test case by name
go test -run TestTaskValidationAPI/valid_task_with_all_fields -v
```

## Test Structure

### Test Files

- `server_test.go` - Contains helper functions to start the test server and make API calls
- `integration_task_test.go` - Contains tests for `proto.Task` and `proto.UpdateTask` message types

### How It Works

1. **Server Setup**: Each test starts a fresh HTTP server on an available port
2. **API Calls**: Tests make POST requests to `/api/v1/validate-proto` endpoint
3. **Validation**: Tests verify the response structure and validation results
4. **Cleanup**: Server is automatically shut down after each test

### Example Test Flow

```go
func TestTaskValidationAPI(t *testing.T) {
    baseURL := startTestServer(t)  // Start server
    
    // Make API call
    result, statusCode, err := callValidateAPI(t, baseURL, "proto.Task", payload)
    
    // Assert results
    if result.Success != expected {
        t.Errorf("Validation failed")
    }
}
```

## Adding New Tests

To add tests for other proto message types:

1. Create a new test file: `integration_<message_type>_test.go`
2. Use the same pattern as `integration_task_test.go`
3. Import the helper functions from `server_test.go`

Example:

```go
package main

import (
    "net/http"
    "testing"
)

func TestSimpleUserValidationAPI(t *testing.T) {
    baseURL := startTestServer(t)
    
    tests := []struct {
        name        string
        schemaName  string
        payload     interface{}
        wantSuccess bool
        wantErrors  int
    }{
        // Your test cases here
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, statusCode, err := callValidateAPI(t, baseURL, tt.schemaName, tt.payload)
            // Assertions...
        })
    }
}
```

## Test Coverage

Current test coverage includes:

- ✅ `proto.Task` - All validation scenarios
- ✅ `proto.UpdateTask` - Including CEL constraint validation

## Troubleshooting

### Port Already in Use

If you see "port already in use" errors, the test server automatically finds an available port, so this shouldn't happen. If it does, check for other running instances.

### Server Not Starting

If the server fails to start:
- Check that all dependencies are installed: `go mod download`
- Verify proto files are generated: `make proto`
- Check logs for initialization errors

### Tests Failing

If tests fail:
- Ensure the server is running correctly
- Check that proto descriptors are properly registered
- Verify the API endpoint is correct: `/api/v1/validate-proto`

## Next Steps

To add more test coverage, create additional integration test files for:
- Simple validation messages (`SimpleUser`, `Product`, etc.)
- Complex validation messages (`ComplexOrder`, `EmployeeProfile`, etc.)
- CEL validation messages (`PaymentInfo`, `DateRange`, etc.)
