package main

import (
	"net/http"
	"testing"
)

func TestConditionalOrderValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid standard order without express fee",
			schemaName:  "proto.ConditionalOrder",
			payload:     map[string]interface{}{"order_type": 1},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid express order with express fee",
			schemaName:  "proto.ConditionalOrder",
			payload:     map[string]interface{}{"order_type": 2, "express_fee": 10.0},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "express fee required for express order - CEL constraint",
			schemaName:  "proto.ConditionalOrder",
			payload:     map[string]interface{}{"order_type": 2},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "express fee zero or negative",
			schemaName:  "proto.ConditionalOrder",
			payload:     map[string]interface{}{"order_type": 2, "express_fee": 0.0},
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

func TestAgeRestrictedProductValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid product with user age >= min age",
			schemaName:  "proto.AgeRestrictedProduct",
			payload:     map[string]interface{}{"product_name": "Alcohol", "min_age": 21, "user_age": 25},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid product with user age equal to min age",
			schemaName:  "proto.AgeRestrictedProduct",
			payload:     map[string]interface{}{"product_name": "Alcohol", "min_age": 21, "user_age": 21},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "user age less than min age - CEL constraint",
			schemaName:  "proto.AgeRestrictedProduct",
			payload:     map[string]interface{}{"product_name": "Alcohol", "min_age": 21, "user_age": 18},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "min_age too low",
			schemaName:  "proto.AgeRestrictedProduct",
			payload:     map[string]interface{}{"product_name": "Alcohol", "min_age": 17, "user_age": 18},
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

func TestDiscountCouponValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid coupon with min purchase <= 100",
			schemaName:  "proto.DiscountCoupon",
			payload:     map[string]interface{}{"coupon_code": "SAVE10", "min_purchase": 50.0},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid coupon with discount when min purchase > 100",
			schemaName:  "proto.DiscountCoupon",
			payload:     map[string]interface{}{"coupon_code": "SAVE20", "discount_percent": 20.0, "min_purchase": 150.0},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "discount required when min purchase > 100 - CEL constraint",
			schemaName:  "proto.DiscountCoupon",
			payload:     map[string]interface{}{"coupon_code": "SAVE20", "min_purchase": 150.0},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "coupon_code invalid pattern - lowercase",
			schemaName:  "proto.DiscountCoupon",
			payload:     map[string]interface{}{"coupon_code": "save10", "min_purchase": 50.0},
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

func TestPaymentInfoValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid credit card payment",
			schemaName:  "proto.PaymentInfo",
			payload:     map[string]interface{}{"payment_method": 1, "card_number": "1234567890123456"},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "valid paypal payment",
			schemaName:  "proto.PaymentInfo",
			payload:     map[string]interface{}{"payment_method": 3, "paypal_email": "user@example.com"},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "card number required for credit card - CEL constraint",
			schemaName:  "proto.PaymentInfo",
			payload:     map[string]interface{}{"payment_method": 1},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "paypal email required for paypal - CEL constraint",
			schemaName:  "proto.PaymentInfo",
			payload:     map[string]interface{}{"payment_method": 3},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "bank account required for bank transfer - CEL constraint",
			schemaName:  "proto.PaymentInfo",
			payload:     map[string]interface{}{"payment_method": 4},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "card number invalid format - too short",
			schemaName:  "proto.PaymentInfo",
			payload:     map[string]interface{}{"payment_method": 1, "card_number": "123456789012"},
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

func TestDateRangeValidationAPI(t *testing.T) {
	baseURL := startTestServer(t)

	tests := []struct {
		name        string
		schemaName  string
		payload     interface{}
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid date range",
			schemaName:  "proto.DateRange",
			payload:     map[string]interface{}{"start_date": 1609459200, "end_date": 1612137600},
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing required start_date",
			schemaName:  "proto.DateRange",
			payload:     map[string]interface{}{"end_date": 1612137600},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "end_date must be greater than start_date - CEL constraint",
			schemaName:  "proto.DateRange",
			payload:     map[string]interface{}{"start_date": 1612137600, "end_date": 1609459200},
			wantSuccess: false,
			wantErrors:  1,
		},
		{
			name:        "end_date equal to start_date - CEL constraint",
			schemaName:  "proto.DateRange",
			payload:     map[string]interface{}{"start_date": 1609459200, "end_date": 1609459200},
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
