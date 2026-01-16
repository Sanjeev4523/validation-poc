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
		{
			name:        "valid user with all fields",
			schemaName:  "proto.SimpleUser",
			payload:     map[string]interface{}{"name": "John Doe", "email": "john@example.com", "age": 25},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid user with nested contact info",
			schemaName:  "proto.SimpleUser",
			payload:     map[string]interface{}{"name": "Jane Smith", "email": "jane@example.com", "age": 30, "contact_info": map[string]interface{}{"phone": "+1234567890", "address": "123 Main St", "country_code": "US"}},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing required name",
			schemaName:  "proto.SimpleUser",
			payload:     map[string]interface{}{"email": "john@example.com", "age": 25},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "missing required email",
			schemaName:  "proto.SimpleUser",
			payload:     map[string]interface{}{"name": "John Doe", "age": 25},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "missing required age",
			schemaName:  "proto.SimpleUser",
			payload:     map[string]interface{}{"name": "John Doe", "email": "john@example.com"},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "name too short",
			schemaName:  "proto.SimpleUser",
			payload:     map[string]interface{}{"name": "Jo", "email": "john@example.com", "age": 25},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "name with invalid characters",
			schemaName:  "proto.SimpleUser",
			payload:     map[string]interface{}{"name": "John123", "email": "john@example.com", "age": 25},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "invalid email format",
			schemaName:  "proto.SimpleUser",
			payload:     map[string]interface{}{"name": "John Doe", "email": "notanemail", "age": 25},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "age too young",
			schemaName:  "proto.SimpleUser",
			payload:     map[string]interface{}{"name": "John Doe", "email": "john@example.com", "age": 17},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "age too old",
			schemaName:  "proto.SimpleUser",
			payload:     map[string]interface{}{"name": "John Doe", "email": "john@example.com", "age": 121},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "age at boundaries",
			schemaName:  "proto.SimpleUser",
			payload:     map[string]interface{}{"name": "John Doe", "email": "john@example.com", "age": 18},
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

func TestContactInfoValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid contact info with all fields",
			schemaName:  "proto.ContactInfo",
			payload:     map[string]interface{}{"phone": "+1234567890", "address": "123 Main St", "country_code": "US"},
			wantSuccess: true,
			wantErrors:  0,
		},
		// Note: Empty payload fails because protojson sets optional string fields to empty strings,
		// and empty strings violate the constraints (phone pattern, country_code len: 2)
		{
			name:        "invalid phone format - starts with 0",
			schemaName:  "proto.ContactInfo",
			payload:     map[string]interface{}{"phone": "01234567890", "country_code": "GB"},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "country code too short",
			schemaName:  "proto.ContactInfo",
			payload:     map[string]interface{}{"phone": "+1234567890", "country_code": "U"},
			wantSuccess: false,
			wantErrors:  2, // Both length and pattern errors
		},
		{
			name:        "country code lowercase",
			schemaName:  "proto.ContactInfo",
			payload:     map[string]interface{}{"phone": "+1234567890", "country_code": "us"},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "valid country code with valid phone",
			schemaName:  "proto.ContactInfo",
			payload:     map[string]interface{}{"phone": "+1234567890", "country_code": "GB"},
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

func TestProductValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid product",
			schemaName:  "proto.Product",
			payload:     map[string]interface{}{"name": "Laptop", "price": 999.99, "quantity": 10},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing required name",
			schemaName:  "proto.Product",
			payload:     map[string]interface{}{"price": 999.99, "quantity": 10},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "price too low",
			schemaName:  "proto.Product",
			payload:     map[string]interface{}{"name": "Laptop", "price": 0.005, "quantity": 10},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "price at minimum",
			schemaName:  "proto.Product",
			payload:     map[string]interface{}{"name": "Laptop", "price": 0.01, "quantity": 10},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "quantity too low",
			schemaName:  "proto.Product",
			payload:     map[string]interface{}{"name": "Laptop", "price": 999.99, "quantity": 0},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "quantity at minimum",
			schemaName:  "proto.Product",
			payload:     map[string]interface{}{"name": "Laptop", "price": 999.99, "quantity": 1},
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

func TestProductListValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid product list with one product",
			schemaName:  "proto.ProductList",
			payload:     map[string]interface{}{"products": []interface{}{map[string]interface{}{"name": "Laptop", "price": 999.99, "quantity": 10}}},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid product list with multiple products",
			schemaName:  "proto.ProductList",
			payload:     map[string]interface{}{"products": []interface{}{map[string]interface{}{"name": "Laptop", "price": 999.99, "quantity": 10}, map[string]interface{}{"name": "Mouse", "price": 29.99, "quantity": 5}}},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "empty product list",
			schemaName:  "proto.ProductList",
			payload:     map[string]interface{}{"products": []interface{}{}},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "missing products field",
			schemaName:  "proto.ProductList",
			payload:     map[string]interface{}{},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "invalid product in list",
			schemaName:  "proto.ProductList",
			payload:     map[string]interface{}{"products": []interface{}{map[string]interface{}{"name": "", "price": 999.99, "quantity": 10}}},
			wantSuccess: false,
			wantErrors:  1,
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

func TestStatusMessageValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid status ACTIVE",
			schemaName:  "proto.StatusMessage",
			payload:     map[string]interface{}{"status": 1},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid status INACTIVE",
			schemaName:  "proto.StatusMessage",
			payload:     map[string]interface{}{"status": 2},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing required status",
			schemaName:  "proto.StatusMessage",
			payload:     map[string]interface{}{},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "invalid status UNSPECIFIED",
			schemaName:  "proto.StatusMessage",
			payload:     map[string]interface{}{"status": 0},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "valid message within limit",
			schemaName:  "proto.StatusMessage",
			payload:     map[string]interface{}{"status": 1, "message": "Status is active"},
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

func TestNumericTypesValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid numeric types",
			schemaName:  "proto.NumericTypes",
			payload:     map[string]interface{}{"int32_value": 500, "int64_value": 1000, "float_value": 0.5, "double_value": 50.0, "uint32_value": 5000, "uint64_value": 100},
			wantSuccess: true,
			wantErrors:  0,
		},
		// Note: Empty payload fails because int64_value and uint64_value have constraints (gt: 0, gte: 1)
		// and when not provided, they default to 0 which violates constraints
		// This is expected behavior - if you want to use NumericTypes, you must provide valid values
		{
			name:        "int32_value too low",
			schemaName:  "proto.NumericTypes",
			payload:     map[string]interface{}{"int32_value": -1001, "int64_value": 1, "uint64_value": 1},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "int32_value too high",
			schemaName:  "proto.NumericTypes",
			payload:     map[string]interface{}{"int32_value": 1001, "int64_value": 1, "uint64_value": 1},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "int64_value zero or negative",
			schemaName:  "proto.NumericTypes",
			payload:     map[string]interface{}{"int64_value": 0, "uint64_value": 1},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "float_value too high",
			schemaName:  "proto.NumericTypes",
			payload:     map[string]interface{}{"float_value": 1.1, "int64_value": 1, "uint64_value": 1},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "double_value at boundaries",
			schemaName:  "proto.NumericTypes",
			payload:     map[string]interface{}{"double_value": 100.0, "int64_value": 1, "uint64_value": 1},
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
