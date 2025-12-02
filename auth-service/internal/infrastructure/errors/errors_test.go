package errors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidationError(t *testing.T) {
	err := ValidationError("invalid input")
	
	if err.Code != ErrCodeValidation {
		t.Errorf("Expected code %s, got %s", ErrCodeValidation, err.Code)
	}
	
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
	
	if err.Message != "invalid input" {
		t.Errorf("Expected message 'invalid input', got '%s'", err.Message)
	}
}

func TestInvalidPhoneError(t *testing.T) {
	phone := "123"
	err := InvalidPhoneError(phone)
	
	if err.Code != ErrCodeInvalidPhone {
		t.Errorf("Expected code %s, got %s", ErrCodeInvalidPhone, err.Code)
	}
	
	if err.Details["phone"] != phone {
		t.Errorf("Expected phone detail '%s', got '%v'", phone, err.Details["phone"])
	}
	
	if err.Details["expected"] == nil {
		t.Error("Expected 'expected' detail to be set")
	}
}

func TestMissingFieldError(t *testing.T) {
	field := "email"
	err := MissingFieldError(field)
	
	if err.Code != ErrCodeMissingField {
		t.Errorf("Expected code %s, got %s", ErrCodeMissingField, err.Code)
	}
	
	if err.Details["field"] != field {
		t.Errorf("Expected field detail '%s', got '%v'", field, err.Details["field"])
	}
}

func TestNotFoundError(t *testing.T) {
	resource := "user"
	err := NotFoundError(resource)
	
	if err.Code != ErrCodeNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeNotFound, err.Code)
	}
	
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, err.StatusCode)
	}
	
	if err.Details["resource"] != resource {
		t.Errorf("Expected resource detail '%s', got '%v'", resource, err.Details["resource"])
	}
}

func TestAlreadyExistsError(t *testing.T) {
	resource := "user"
	identifier := "email"
	err := AlreadyExistsError(resource, identifier)
	
	if err.Code != ErrCodeAlreadyExists {
		t.Errorf("Expected code %s, got %s", ErrCodeAlreadyExists, err.Code)
	}
	
	if err.StatusCode != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, err.StatusCode)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	err := ValidationError("test error").WithDetails("field", "test")
	requestID := "test-request-id"
	
	WriteError(w, err, requestID)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
	
	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response.Error.Code != ErrCodeValidation {
		t.Errorf("Expected code %s, got %s", ErrCodeValidation, response.Error.Code)
	}
	
	if response.Error.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", response.Error.Message)
	}
	
	if response.Error.Details["field"] != "test" {
		t.Errorf("Expected field detail 'test', got '%v'", response.Error.Details["field"])
	}
}

func TestErrorWithError(t *testing.T) {
	underlyingErr := http.ErrBodyNotAllowed
	err := InternalError("test", underlyingErr)
	
	if err.Err != underlyingErr {
		t.Errorf("Expected underlying error to be set")
	}
	
	errorString := err.Error()
	if errorString == "" {
		t.Error("Expected non-empty error string")
	}
}

func TestGRPCError(t *testing.T) {
	service := "AuthService"
	method := "ValidateToken"
	underlyingErr := http.ErrBodyNotAllowed
	
	err := GRPCError(service, method, underlyingErr)
	
	if err.Code != ErrCodeGRPCError {
		t.Errorf("Expected code %s, got %s", ErrCodeGRPCError, err.Code)
	}
	
	if err.StatusCode != http.StatusBadGateway {
		t.Errorf("Expected status %d, got %d", http.StatusBadGateway, err.StatusCode)
	}
	
	if err.Details["service"] != service {
		t.Errorf("Expected service detail '%s', got '%v'", service, err.Details["service"])
	}
	
	if err.Details["method"] != method {
		t.Errorf("Expected method detail '%s', got '%v'", method, err.Details["method"])
	}
}

func TestDatabaseError(t *testing.T) {
	operation := "insert"
	underlyingErr := http.ErrBodyNotAllowed
	
	err := DatabaseError(operation, underlyingErr)
	
	if err.Code != ErrCodeDatabase {
		t.Errorf("Expected code %s, got %s", ErrCodeDatabase, err.Code)
	}
	
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, err.StatusCode)
	}
	
	if err.Details["operation"] != operation {
		t.Errorf("Expected operation detail '%s', got '%v'", operation, err.Details["operation"])
	}
}
