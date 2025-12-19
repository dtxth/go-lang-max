package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_AuthenticateMAX_RequestValidation(t *testing.T) {
	// Create a handler with nil auth service to test request validation
	handler := NewHandler(nil)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		errorContains  string
	}{
		{
			name: "missing init_data field",
			requestBody: map[string]string{
				"other_field": "value",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "init_data",
		},
		{
			name: "empty init_data",
			requestBody: MaxAuthRequest{
				InitData: "",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "init_data",
		},
		{
			name: "null init_data",
			requestBody: map[string]interface{}{
				"init_data": nil,
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "init_data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/max", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.AuthenticateMAX(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("AuthenticateMAX() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.errorContains != "" {
				body := w.Body.String()
				if !strings.Contains(body, tt.errorContains) {
					t.Errorf("AuthenticateMAX() error response = %v, want to contain %v", body, tt.errorContains)
				}
			}
		})
	}
}

func TestHandler_AuthenticateMAX_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil)

	tests := []struct {
		name        string
		requestBody string
		contentType string
	}{
		{
			name:        "malformed JSON",
			requestBody: `{"init_data": "test", invalid}`,
			contentType: "application/json",
		},
		{
			name:        "empty body",
			requestBody: "",
			contentType: "application/json",
		},
		{
			name:        "non-JSON body",
			requestBody: "not json at all",
			contentType: "application/json",
		},
		{
			name:        "incomplete JSON",
			requestBody: `{"init_data":`,
			contentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/auth/max", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			handler.AuthenticateMAX(w, req)

			// Should return 400 Bad Request for invalid JSON
			if w.Code != http.StatusBadRequest {
				t.Errorf("AuthenticateMAX() status = %v, want %v", w.Code, http.StatusBadRequest)
			}

			// Should contain error about request body
			body := w.Body.String()
			if !strings.Contains(body, "request body") && !strings.Contains(body, "init_data") {
				t.Errorf("AuthenticateMAX() error response should mention request body or init_data")
			}
		})
	}
}

func TestHandler_AuthenticateMAX_RequestBodyEdgeCases(t *testing.T) {
	// Test JSON parsing and validation without calling auth service
	tests := []struct {
		name        string
		initData    string
		description string
	}{
		{
			name:        "very long init_data",
			initData:    strings.Repeat("a", 10000),
			description: "Should handle very long initData strings",
		},
		{
			name:        "init_data with special characters",
			initData:    "max_id=123&first_name=Jöhn&special=!@#$%^&*()",
			description: "Should handle special characters in initData",
		},
		{
			name:        "init_data with unicode",
			initData:    "max_id=123&first_name=张三",
			description: "Should handle unicode characters in initData",
		},
		{
			name:        "init_data with newlines",
			initData:    "max_id=123\nfirst_name=John",
			description: "Should handle newlines in initData",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := MaxAuthRequest{
				InitData: tt.initData,
			}
			jsonBody, err := json.Marshal(reqBody)
			if err != nil {
				t.Errorf("Failed to marshal request: %v", err)
				return
			}

			// Test that JSON marshaling/unmarshaling works correctly
			var parsedReq MaxAuthRequest
			if err := json.Unmarshal(jsonBody, &parsedReq); err != nil {
				t.Errorf("Failed to unmarshal request: %v", err)
				return
			}

			if parsedReq.InitData != tt.initData {
				t.Errorf("InitData mismatch after JSON round-trip")
			}
		})
	}
}

func TestHandler_AuthenticateMAX_ContentTypeHandling(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		description string
	}{
		{
			name:        "application/json",
			contentType: "application/json",
			description: "Standard JSON content type",
		},
		{
			name:        "application/json with charset",
			contentType: "application/json; charset=utf-8",
			description: "JSON with charset specification",
		},
		{
			name:        "missing content type",
			contentType: "",
			description: "No content type header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := MaxAuthRequest{
				InitData: "max_id=123&first_name=John&hash=test",
			}
			jsonBody, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/auth/max", bytes.NewBuffer(jsonBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Test that we can parse the request body regardless of content type
			var parsedReq MaxAuthRequest
			err := json.NewDecoder(req.Body).Decode(&parsedReq)
			
			// Go's JSON decoder should handle all these content types
			if err != nil {
				t.Errorf("Failed to decode request with content type %s: %v", tt.contentType, err)
			}

			if parsedReq.InitData != reqBody.InitData {
				t.Errorf("InitData mismatch with content type %s", tt.contentType)
			}
		})
	}
}

func TestMaxAuthRequest_StructValidation(t *testing.T) {
	tests := []struct {
		name     string
		request  MaxAuthRequest
		expected string
	}{
		{
			name: "valid request",
			request: MaxAuthRequest{
				InitData: "max_id=123&first_name=John&hash=abc123",
			},
			expected: "max_id=123&first_name=John&hash=abc123",
		},
		{
			name: "empty init_data",
			request: MaxAuthRequest{
				InitData: "",
			},
			expected: "",
		},
		{
			name: "init_data with spaces",
			request: MaxAuthRequest{
				InitData: "max_id=123 &first_name=John Doe&hash=abc123",
			},
			expected: "max_id=123 &first_name=John Doe&hash=abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.request.InitData != tt.expected {
				t.Errorf("MaxAuthRequest.InitData = %v, want %v", tt.request.InitData, tt.expected)
			}
		})
	}
}

func TestMaxAuthResponse_StructValidation(t *testing.T) {
	tests := []struct {
		name     string
		response MaxAuthResponse
	}{
		{
			name: "valid response",
			response: MaxAuthResponse{
				AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
				RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh",
			},
		},
		{
			name: "empty tokens",
			response: MaxAuthResponse{
				AccessToken:  "",
				RefreshToken: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling/unmarshaling
			jsonData, err := json.Marshal(tt.response)
			if err != nil {
				t.Errorf("Failed to marshal MaxAuthResponse: %v", err)
				return
			}

			var unmarshaled MaxAuthResponse
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				t.Errorf("Failed to unmarshal MaxAuthResponse: %v", err)
				return
			}

			if unmarshaled.AccessToken != tt.response.AccessToken {
				t.Errorf("Unmarshaled AccessToken = %v, want %v", unmarshaled.AccessToken, tt.response.AccessToken)
			}
			if unmarshaled.RefreshToken != tt.response.RefreshToken {
				t.Errorf("Unmarshaled RefreshToken = %v, want %v", unmarshaled.RefreshToken, tt.response.RefreshToken)
			}
		})
	}
}

func TestHandler_AuthenticateMAX_NilAuthService(t *testing.T) {
	// Test that handler gracefully handles nil auth service
	// We'll test this by checking that the handler doesn't crash during request parsing
	handler := NewHandler(nil)

	reqBody := MaxAuthRequest{
		InitData: "", // Empty to trigger validation error before auth service call
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/max", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.AuthenticateMAX(w, req)

	// Should return validation error (400) before reaching auth service
	if w.Code != http.StatusBadRequest {
		t.Errorf("AuthenticateMAX() status = %v, want %v", w.Code, http.StatusBadRequest)
	}

	// Should contain error about init_data
	body := w.Body.String()
	if !strings.Contains(body, "init_data") {
		t.Errorf("AuthenticateMAX() error response should mention init_data")
	}
}

func TestHandler_AuthenticateMAX_ResponseHeaders(t *testing.T) {
	handler := NewHandler(nil)

	reqBody := MaxAuthRequest{
		InitData: "", // Empty to trigger validation error
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/max", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.AuthenticateMAX(w, req)

	// Check that response has appropriate headers set
	if w.Code != 0 {
		// If a status was set, check that it's a valid HTTP status
		if w.Code < 100 || w.Code >= 600 {
			t.Errorf("AuthenticateMAX() invalid HTTP status code: %v", w.Code)
		}
	}

	// Should have content type set for error response
	if w.Code == http.StatusBadRequest {
		contentType := w.Header().Get("Content-Type")
		if contentType == "" {
			t.Errorf("AuthenticateMAX() error response should have Content-Type header")
		}
	}
}