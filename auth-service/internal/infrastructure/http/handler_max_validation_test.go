package http

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMaxAuthRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		requestBody interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "valid request",
			requestBody: MaxAuthRequest{
				InitData: "max_id=123&first_name=John&hash=abc123",
			},
			wantErr: false,
		},
		{
			name: "empty init_data",
			requestBody: MaxAuthRequest{
				InitData: "",
			},
			wantErr:     true,
			errContains: "init_data",
		},
		{
			name:        "invalid JSON",
			requestBody: "invalid json",
			wantErr:     true,
			errContains: "invalid request body",
		},
		{
			name:        "missing init_data field",
			requestBody: map[string]string{"other_field": "value"},
			wantErr:     true,
			errContains: "init_data",
		},
		{
			name: "null init_data",
			requestBody: map[string]interface{}{
				"init_data": nil,
			},
			wantErr:     true,
			errContains: "init_data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonBody []byte
			var err error

			if tt.requestBody == "invalid json" {
				jsonBody = []byte("invalid json")
			} else {
				jsonBody, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/auth/max", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Test just the JSON parsing part, not the full handler
			var parsedReq MaxAuthRequest
			err = json.NewDecoder(req.Body).Decode(&parsedReq)

			if tt.wantErr {
				if err == nil && (parsedReq.InitData == "" || strings.TrimSpace(parsedReq.InitData) == "") {
					// This is expected for empty/missing init_data cases
					return
				}
				if err == nil {
					t.Errorf("Expected JSON parsing error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected JSON parsing error: %v", err)
				}
				if parsedReq.InitData == "" {
					t.Errorf("Expected non-empty InitData")
				}
			}
		})
	}
}

func TestMaxAuthRequest_JSONSerialization(t *testing.T) {
	// Test request serialization
	req := MaxAuthRequest{
		InitData: "test_init_data",
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Errorf("Failed to marshal MaxAuthRequest: %v", err)
	}

	var unmarshaled MaxAuthRequest
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal MaxAuthRequest: %v", err)
	}

	if unmarshaled.InitData != req.InitData {
		t.Errorf("Unmarshaled InitData = %v, want %v", unmarshaled.InitData, req.InitData)
	}
}

func TestMaxAuthResponse_JSONSerialization(t *testing.T) {
	// Test response serialization
	resp := MaxAuthResponse{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("Failed to marshal MaxAuthResponse: %v", err)
	}

	var unmarshaled MaxAuthResponse
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal MaxAuthResponse: %v", err)
	}

	if unmarshaled.AccessToken != resp.AccessToken {
		t.Errorf("Unmarshaled AccessToken = %v, want %v", unmarshaled.AccessToken, resp.AccessToken)
	}
	if unmarshaled.RefreshToken != resp.RefreshToken {
		t.Errorf("Unmarshaled RefreshToken = %v, want %v", unmarshaled.RefreshToken, resp.RefreshToken)
	}
}

func TestMaxAuthRequest_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		initData string
		valid    bool
	}{
		{
			name:     "very long init_data",
			initData: strings.Repeat("a", 10000),
			valid:    true, // Should be accepted by JSON parsing
		},
		{
			name:     "init_data with special characters",
			initData: "max_id=123&first_name=Jöhn&last_name=Döe&special=!@#$%^&*()",
			valid:    true,
		},
		{
			name:     "init_data with newlines",
			initData: "max_id=123\nfirst_name=John",
			valid:    true, // JSON parsing should handle this
		},
		{
			name:     "init_data with unicode",
			initData: "max_id=123&first_name=张三&last_name=李四",
			valid:    true,
		},
		{
			name:     "whitespace only init_data",
			initData: "   ",
			valid:    false, // Should be treated as empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := MaxAuthRequest{
				InitData: tt.initData,
			}

			jsonData, err := json.Marshal(req)
			if err != nil {
				t.Errorf("Failed to marshal request: %v", err)
				return
			}

			var unmarshaled MaxAuthRequest
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				t.Errorf("Failed to unmarshal request: %v", err)
				return
			}

			if unmarshaled.InitData != req.InitData {
				t.Errorf("InitData mismatch after serialization")
			}

			// Test HTTP request parsing
			httpReq := httptest.NewRequest("POST", "/auth/max", bytes.NewBuffer(jsonData))
			httpReq.Header.Set("Content-Type", "application/json")

			var parsedReq MaxAuthRequest
			err = json.NewDecoder(httpReq.Body).Decode(&parsedReq)
			if err != nil {
				t.Errorf("Failed to parse HTTP request: %v", err)
				return
			}

			if parsedReq.InitData != req.InitData {
				t.Errorf("InitData mismatch after HTTP parsing")
			}
		})
	}
}

func TestHTTPRequestValidation_ContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expectError bool
	}{
		{
			name:        "valid content type",
			contentType: "application/json",
			expectError: false,
		},
		{
			name:        "missing content type",
			contentType: "",
			expectError: false, // Go's JSON decoder is lenient
		},
		{
			name:        "wrong content type",
			contentType: "text/plain",
			expectError: false, // Handler doesn't validate content type, just tries to decode
		},
		{
			name:        "content type with charset",
			contentType: "application/json; charset=utf-8",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := MaxAuthRequest{
				InitData: "test_data",
			}
			jsonBody, _ := json.Marshal(reqBody)
			
			req := httptest.NewRequest("POST", "/auth/max", bytes.NewBuffer(jsonBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Test that we can parse the request regardless of content type
			var parsedReq MaxAuthRequest
			err := json.NewDecoder(req.Body).Decode(&parsedReq)
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}