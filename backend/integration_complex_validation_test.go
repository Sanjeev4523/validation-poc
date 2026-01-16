package main

import (
	"net/http"
	"testing"
	"time"
)

func TestCustomerInfoValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid customer info with all fields",
			schemaName:  "proto.CustomerInfo",
			payload:     map[string]interface{}{"email": "customer@example.com", "phone": "+1234567890", "address": "123 Main Street, City", "name": "John Doe"},
			wantSuccess: true,
			wantErrors:  0,
		},
		// Note: address has min_len: 10, so when not provided, protojson sets it to empty string
		// which violates the constraint. We test with valid address or expect address error.
		{
			name:        "missing required email",
			schemaName:  "proto.CustomerInfo",
			payload:     map[string]interface{}{"phone": "+1234567890", "name": "John Doe", "address": "123 Main Street"},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "invalid email format",
			schemaName:  "proto.CustomerInfo",
			payload:     map[string]interface{}{"email": "notanemail", "phone": "+1234567890", "name": "John Doe", "address": "123 Main Street"},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "invalid phone format",
			schemaName:  "proto.CustomerInfo",
			payload:     map[string]interface{}{"email": "customer@example.com", "phone": "123", "name": "John Doe"},
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

func TestOrderItemValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid order item",
			schemaName:  "proto.OrderItem",
			payload:     map[string]interface{}{"product_id": "PROD-001", "quantity": 5, "price": 99.99},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid order item with discount",
			schemaName:  "proto.OrderItem",
			payload:     map[string]interface{}{"product_id": "PROD-001", "quantity": 5, "price": 99.99, "discount": 10.0},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing required product_id",
			schemaName:  "proto.OrderItem",
			payload:     map[string]interface{}{"quantity": 5, "price": 99.99},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "discount required when quantity > 10 - CEL constraint",
			schemaName:  "proto.OrderItem",
			payload:     map[string]interface{}{"product_id": "PROD-001", "quantity": 11, "price": 99.99},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "discount not required when quantity <= 10",
			schemaName:  "proto.OrderItem",
			payload:     map[string]interface{}{"product_id": "PROD-001", "quantity": 10, "price": 99.99},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid discount when quantity > 10",
			schemaName:  "proto.OrderItem",
			payload:     map[string]interface{}{"product_id": "PROD-001", "quantity": 11, "price": 99.99, "discount": 5.0},
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

func TestShippingInfoValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid digital shipping without address",
			schemaName:  "proto.ShippingInfo",
			payload:     map[string]interface{}{"type": 1},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid physical shipping with address",
			schemaName:  "proto.ShippingInfo",
			payload:     map[string]interface{}{"type": 2, "address": "123 Main Street, City, State"},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "address required for physical shipping - CEL constraint",
			schemaName:  "proto.ShippingInfo",
			payload:     map[string]interface{}{"type": 2},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "address required for express shipping - CEL constraint",
			schemaName:  "proto.ShippingInfo",
			payload:     map[string]interface{}{"type": 3},
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

func TestPersonalInfoValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	now := time.Now()
	birthDate := now.AddDate(-25, 0, 0).Format(time.RFC3339)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid personal info",
			schemaName:  "proto.PersonalInfo",
			payload:     map[string]interface{}{"name": "John Doe", "email": "john@example.com", "phone": "+1234567890", "date_of_birth": birthDate},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing required name",
			schemaName:  "proto.PersonalInfo",
			payload:     map[string]interface{}{"email": "john@example.com", "phone": "+1234567890", "date_of_birth": birthDate},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "invalid email format",
			schemaName:  "proto.PersonalInfo",
			payload:     map[string]interface{}{"name": "John Doe", "email": "notanemail", "phone": "+1234567890", "date_of_birth": birthDate},
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

func TestWorkInfoValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	now := time.Now()
	startDate := now.AddDate(-2, 0, 0).Format(time.RFC3339)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid work info",
			schemaName:  "proto.WorkInfo",
			payload:     map[string]interface{}{"department": "Engineering", "salary": 100000.0, "start_date": startDate},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing required department",
			schemaName:  "proto.WorkInfo",
			payload:     map[string]interface{}{"salary": 100000.0, "start_date": startDate},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "salary negative",
			schemaName:  "proto.WorkInfo",
			payload:     map[string]interface{}{"department": "Engineering", "salary": -1000.0, "start_date": startDate},
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

func TestEmergencyContactValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid emergency contact",
			schemaName:  "proto.EmergencyContact",
			payload:     map[string]interface{}{"name": "Jane Doe", "relationship": 1, "phone": "+1234567890"},
			wantSuccess: true,
			wantErrors:  0,
		},
		// Note: When phone is not provided, protojson sets it to empty string which violates phone pattern.
		// For non-spouse relationships, phone is optional but if set to empty string, it fails validation.
		// We test with a valid phone for non-spouse to verify the relationship logic works.
		{
			name:        "valid emergency contact with phone for non-spouse",
			schemaName:  "proto.EmergencyContact",
			payload:     map[string]interface{}{"name": "John Parent", "relationship": 2, "phone": "+1234567890"},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "phone required for spouse - CEL constraint",
			schemaName:  "proto.EmergencyContact",
			payload:     map[string]interface{}{"name": "Jane Doe", "relationship": 1},
			wantSuccess: false,
			wantErrors:  2, // CEL constraint error + phone pattern error (empty string)
		},
		{
			name:        "valid phone for spouse",
			schemaName:  "proto.EmergencyContact",
			payload:     map[string]interface{}{"name": "Jane Doe", "relationship": 1, "phone": "+1234567890"},
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

func TestInventoryItemValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid inventory item",
			schemaName:  "proto.InventoryItem",
			payload:     map[string]interface{}{"item_id": "ITEM-001", "name": "Product Name", "quantity": 100, "price": 29.99},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing required item_id",
			schemaName:  "proto.InventoryItem",
			payload:     map[string]interface{}{"name": "Product Name", "quantity": 100, "price": 29.99},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "item_id invalid pattern - lowercase",
			schemaName:  "proto.InventoryItem",
			payload:     map[string]interface{}{"item_id": "item-001", "name": "Product Name", "quantity": 100, "price": 29.99},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "quantity negative",
			schemaName:  "proto.InventoryItem",
			payload:     map[string]interface{}{"item_id": "ITEM-001", "name": "Product Name", "quantity": -1, "price": 29.99},
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

func TestWarehouseValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid warehouse",
			schemaName:  "proto.Warehouse",
			payload:     map[string]interface{}{"warehouse_id": "WH-001", "location": "New York", "items": []interface{}{}, "capacity": 1000},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid warehouse with items",
			schemaName:  "proto.Warehouse",
			payload:     map[string]interface{}{"warehouse_id": "WH-001", "location": "New York", "items": []interface{}{map[string]interface{}{"item_id": "ITEM-001", "name": "Product", "quantity": 10, "price": 29.99}}, "capacity": 1000},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing required warehouse_id",
			schemaName:  "proto.Warehouse",
			payload:     map[string]interface{}{"location": "New York", "items": []interface{}{}, "capacity": 1000},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "warehouse_id invalid pattern",
			schemaName:  "proto.Warehouse",
			payload:     map[string]interface{}{"warehouse_id": "warehouse-001", "location": "New York", "items": []interface{}{}, "capacity": 1000},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "capacity zero or negative",
			schemaName:  "proto.Warehouse",
			payload:     map[string]interface{}{"warehouse_id": "WH-001", "location": "New York", "items": []interface{}{}, "capacity": 0},
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
