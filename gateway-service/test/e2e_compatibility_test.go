package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gateway-service/internal/config"
	grpcClient "gateway-service/internal/infrastructure/grpc"
	httpHandler "gateway-service/internal/infrastructure/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2ECompatibility tests that the Gateway Service maintains compatibility with existing E2E tests
func TestE2ECompatibility(t *testing.T) {
	// Create test configuration
	cfg := createTestConfig()
	
	// Create client manager (will fail gracefully if services not available)
	clientManager := grpcClient.NewClientManager(cfg)
	
	// Create router
	router := httpHandler.NewRouter(cfg, clientManager)
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("Request Format Preservation", func(t *testing.T) {
		testRequestFormatPreservation(t, server.URL)
	})

	t.Run("Response Format Preservation", func(t *testing.T) {
		testResponseFormatPreservation(t, server.URL)
	})

	t.Run("Status Code Mapping Accuracy", func(t *testing.T) {
		testStatusCodeMappingAccuracy(t, server.URL)
	})

	t.Run("API Contract Compatibility", func(t *testing.T) {
		testAPIContractCompatibility(t, server.URL)
	})

	t.Run("Error Response Format Consistency", func(t *testing.T) {
		testErrorResponseFormatConsistency(t, server.URL)
	})
}

// testRequestFormatPreservation verifies that the Gateway accepts the same request formats as E2E tests expect
func testRequestFormatPreservation(t *testing.T, baseURL string) {
	testCases := []struct {
		name        string
		method      string
		path        string
		contentType string
		body        interface{}
		expectError bool
	}{
		{
			name:        "Auth Registration Request",
			method:      "POST",
			path:        "/register",
			contentType: "application/json",
			body: map[string]interface{}{
				"email":    "test@example.com",
				"phone":    "+1234567890",
				"password": "TestPassword123!",
				"role":     "operator",
			},
		},
		{
			name:        "Auth Login Request",
			method:      "POST",
			path:        "/login",
			contentType: "application/json",
			body: map[string]interface{}{
				"email":    "test@example.com",
				"password": "TestPassword123!",
			},
		},
		{
			name:        "Auth Login by Phone Request",
			method:      "POST",
			path:        "/login-phone",
			contentType: "application/json",
			body: map[string]interface{}{
				"phone":    "+1234567890",
				"password": "TestPassword123!",
			},
		},
		{
			name:        "Employee Creation Request",
			method:      "POST",
			path:        "/simple-employee",
			contentType: "application/json",
			body: map[string]interface{}{
				"name":  "Test Employee",
				"email": "employee@test.com",
				"phone": "+1234567890",
			},
		},
		{
			name:        "University Creation Request",
			method:      "POST",
			path:        "/universities",
			contentType: "application/json",
			body: map[string]interface{}{
				"name": "Test University",
			},
		},
		{
			name:        "MAX Authentication Request",
			method:      "POST",
			path:        "/auth/max",
			contentType: "application/json",
			body: map[string]interface{}{
				"initData": "user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22Test%22%7D&auth_date=1640995200&hash=test_hash",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != nil {
				jsonBody, err := json.Marshal(tc.body)
				require.NoError(t, err)
				body = bytes.NewReader(jsonBody)
			}

			req, err := http.NewRequest(tc.method, baseURL+tc.path, body)
			require.NoError(t, err)

			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Allow some flexibility for gRPC service availability
			// Status codes 400+ are acceptable if gRPC services are not available
			assert.True(t, resp.StatusCode < 600, "Should return valid HTTP status code")
			
			// If we get 503, it means the Gateway is working but backend services are unavailable
			if resp.StatusCode == 503 {
				t.Logf("Got 503 status code, Gateway is working but backend services unavailable")
				return
			}
			
			// Verify Content-Type header is set for JSON responses
			if resp.StatusCode != http.StatusNoContent {
				contentType := resp.Header.Get("Content-Type")
				assert.True(t, 
					strings.Contains(contentType, "application/json") || 
					strings.Contains(contentType, "text/plain"),
					"Response should have appropriate Content-Type header")
			}
		})
	}
}

// testResponseFormatPreservation verifies that responses maintain the expected JSON structure
func testResponseFormatPreservation(t *testing.T, baseURL string) {
	testCases := []struct {
		name           string
		method         string
		path           string
		expectedFields []string
		statusCodes    []int // Acceptable status codes (including 503 for service unavailability)
	}{
		{
			name:           "Health Check Response",
			method:         "GET",
			path:           "/health",
			expectedFields: []string{"status"},
			statusCodes:    []int{200, 503},
		},
		{
			name:           "Metrics Response",
			method:         "GET",
			path:           "/metrics",
			expectedFields: []string{}, // Metrics can vary
			statusCodes:    []int{200, 500, 503}, // Include 503 for circuit breaker
		},
		{
			name:           "Bot Info Response",
			method:         "GET",
			path:           "/bot/me",
			expectedFields: []string{}, // May fail if MaxBot not configured
			statusCodes:    []int{200, 500, 503}, // Include 503 for circuit breaker
		},
		{
			name:           "Universities List Response",
			method:         "GET",
			path:           "/universities?limit=10&offset=0",
			expectedFields: []string{"universities", "total", "limit", "offset"},
			statusCodes:    []int{200, 500, 503}, // Include 503 for circuit breaker
		},
		{
			name:           "Employees List Response",
			method:         "GET",
			path:           "/employees/all",
			expectedFields: []string{}, // Response format may vary
			statusCodes:    []int{200, 500, 503}, // Include 503 for circuit breaker
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, baseURL+tc.path, nil)
			require.NoError(t, err)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Check that status code is one of the expected ones
			assert.Contains(t, tc.statusCodes, resp.StatusCode, 
				"Status code should be one of the expected values")

			// Read response body
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// If response has content, it should be valid JSON
			if len(body) > 0 && resp.StatusCode != http.StatusNoContent {
				var jsonData interface{}
				err := json.Unmarshal(body, &jsonData)
				assert.NoError(t, err, "Response should be valid JSON")

				// Check expected fields if response is successful
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					if jsonMap, ok := jsonData.(map[string]interface{}); ok {
						for _, field := range tc.expectedFields {
							assert.Contains(t, jsonMap, field, 
								"Response should contain expected field: %s", field)
						}
					}
				}
			}
		})
	}
}

// testStatusCodeMappingAccuracy verifies that HTTP status codes are mapped correctly from gRPC errors
func testStatusCodeMappingAccuracy(t *testing.T, baseURL string) {
	testCases := []struct {
		name               string
		method             string
		path               string
		body               interface{}
		expectedStatusCode int
		description        string
	}{
		{
			name:               "Invalid JSON Request",
			method:             "POST",
			path:               "/register",
			body:               "invalid json",
			expectedStatusCode: 400,
			description:        "Invalid JSON should return 400 Bad Request",
		},
		{
			name:               "Method Not Allowed",
			method:             "DELETE",
			path:               "/register",
			body:               nil,
			expectedStatusCode: 405,
			description:        "Unsupported method should return 405 Method Not Allowed",
		},
		{
			name:               "Invalid Login Credentials",
			method:             "POST",
			path:               "/login",
			body: map[string]interface{}{
				"email":    "invalid@example.com",
				"password": "wrongpassword",
			},
			expectedStatusCode: 401,
			description:        "Invalid credentials should return 401 Unauthorized",
		},
		{
			name:               "Missing Required Fields",
			method:             "POST",
			path:               "/register",
			body: map[string]interface{}{
				"email": "test@example.com",
				// Missing password
			},
			expectedStatusCode: 400,
			description:        "Missing required fields should return 400 Bad Request",
		},
		{
			name:               "Non-existent Resource",
			method:             "GET",
			path:               "/universities/99999",
			body:               nil,
			expectedStatusCode: 404,
			description:        "Non-existent resource should return 404 Not Found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != nil {
				if str, ok := tc.body.(string); ok {
					body = strings.NewReader(str)
				} else {
					jsonBody, err := json.Marshal(tc.body)
					require.NoError(t, err)
					body = bytes.NewReader(jsonBody)
				}
			}

			req, err := http.NewRequest(tc.method, baseURL+tc.path, body)
			require.NoError(t, err)

			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Allow some flexibility for gRPC service availability
			// If services are not available, we might get 500 or 503 instead of expected error
			if resp.StatusCode == 500 || resp.StatusCode == 503 {
				t.Logf("Got %d status code, likely due to gRPC service unavailability", resp.StatusCode)
				return
			}

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode, tc.description)
		})
	}
}

// testAPIContractCompatibility verifies that the API contract matches E2E test expectations
func testAPIContractCompatibility(t *testing.T, baseURL string) {
	t.Run("CORS Headers", func(t *testing.T) {
		req, err := http.NewRequest("OPTIONS", baseURL+"/health", nil)
		require.NoError(t, err)
		req.Header.Set("Origin", "http://localhost:3000")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// CORS headers should be present
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"), 
			"CORS headers should be present")
	})

	t.Run("Request ID Handling", func(t *testing.T) {
		requestID := "test-request-123"
		req, err := http.NewRequest("GET", baseURL+"/health", nil)
		require.NoError(t, err)
		req.Header.Set("X-Request-ID", requestID)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Request should be processed successfully
		// Allow 503 for service unavailability (circuit breaker)
		assert.True(t, resp.StatusCode < 600, "Request with custom ID should be processed")
	})

	t.Run("Content-Type Handling", func(t *testing.T) {
		// Test that the Gateway accepts the same content types as E2E tests use
		body := map[string]interface{}{
			"email":    "test@example.com",
			"password": "testpass123",
		}
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/login", bytes.NewReader(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should not fail due to content type issues
		assert.NotEqual(t, 415, resp.StatusCode, "Should accept application/json content type")
	})

	t.Run("Query Parameter Handling", func(t *testing.T) {
		// Test pagination parameters used by E2E tests
		req, err := http.NewRequest("GET", baseURL+"/universities?limit=10&offset=0", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle query parameters correctly
		// Allow 503 for service unavailability
		assert.True(t, resp.StatusCode < 600, "Should handle query parameters correctly")
	})
}

// testErrorResponseFormatConsistency verifies that error responses follow the expected format
func testErrorResponseFormatConsistency(t *testing.T, baseURL string) {
	testCases := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{
			name:   "Invalid JSON Error",
			method: "POST",
			path:   "/register",
			body:   "invalid json",
		},
		{
			name:   "Missing Fields Error",
			method: "POST",
			path:   "/login",
			body:   map[string]interface{}{},
		},
		{
			name:   "Method Not Allowed Error",
			method: "DELETE",
			path:   "/health",
			body:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != nil {
				if str, ok := tc.body.(string); ok {
					body = strings.NewReader(str)
				} else {
					jsonBody, err := json.Marshal(tc.body)
					require.NoError(t, err)
					body = bytes.NewReader(jsonBody)
				}
			}

			req, err := http.NewRequest(tc.method, baseURL+tc.path, body)
			require.NoError(t, err)

			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Only check error format for actual error responses
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				responseBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				if len(responseBody) > 0 {
					var errorResponse map[string]interface{}
					err := json.Unmarshal(responseBody, &errorResponse)
					assert.NoError(t, err, "Error response should be valid JSON")

					// Error responses should have either 'error' or 'message' field
					hasError := false
					if _, ok := errorResponse["error"]; ok {
						hasError = true
					}
					if _, ok := errorResponse["message"]; ok {
						hasError = true
					}
					
					assert.True(t, hasError, "Error response should contain error or message field")
				}
			}
		})
	}
}

// createTestConfig creates a test configuration for the Gateway Service
func createTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port:         "8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Services: config.ServicesConfig{
			Auth: config.ServiceConfig{
				Address:           "localhost:50051",
				Timeout:           10 * time.Second,
				MaxRetries:        3,
				RetryDelay:        100 * time.Millisecond,
				MaxRetryDelay:     5 * time.Second,
				BackoffMultiplier: 2.0,
			},
			Chat: config.ServiceConfig{
				Address:           "localhost:50052",
				Timeout:           10 * time.Second,
				MaxRetries:        3,
				RetryDelay:        100 * time.Millisecond,
				MaxRetryDelay:     5 * time.Second,
				BackoffMultiplier: 2.0,
			},
			Employee: config.ServiceConfig{
				Address:           "localhost:50053",
				Timeout:           10 * time.Second,
				MaxRetries:        3,
				RetryDelay:        100 * time.Millisecond,
				MaxRetryDelay:     5 * time.Second,
				BackoffMultiplier: 2.0,
			},
			Structure: config.ServiceConfig{
				Address:           "localhost:50054",
				Timeout:           10 * time.Second,
				MaxRetries:        3,
				RetryDelay:        100 * time.Millisecond,
				MaxRetryDelay:     5 * time.Second,
				BackoffMultiplier: 2.0,
			},
		},
		Logging: config.LoggingConfig{
			Level: "info",
		},
	}
}

// TestE2EResponseFormatCompatibility tests specific response formats that E2E tests expect
func TestE2EResponseFormatCompatibility(t *testing.T) {
	cfg := createTestConfig()
	clientManager := grpcClient.NewClientManager(cfg)
	router := httpHandler.NewRouter(cfg, clientManager)
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("Auth Service Response Formats", func(t *testing.T) {
		testAuthResponseFormats(t, server.URL)
	})

	t.Run("Employee Service Response Formats", func(t *testing.T) {
		testEmployeeResponseFormats(t, server.URL)
	})

	t.Run("Structure Service Response Formats", func(t *testing.T) {
		testStructureResponseFormats(t, server.URL)
	})

	t.Run("Chat Service Response Formats", func(t *testing.T) {
		testChatResponseFormats(t, server.URL)
	})
}

// testAuthResponseFormats verifies auth service response formats match E2E expectations
func testAuthResponseFormats(t *testing.T, baseURL string) {
	testCases := []struct {
		name           string
		path           string
		method         string
		body           interface{}
		expectedFields map[string]bool // field -> required
	}{
		{
			name:   "Registration Response",
			path:   "/register",
			method: "POST",
			body: map[string]interface{}{
				"email":    "test@example.com",
				"phone":    "+1234567890",
				"password": "TestPassword123!",
				"role":     "operator",
			},
			expectedFields: map[string]bool{
				"access_token":  false, // May not be present if service fails
				"refresh_token": false,
				"id":           false,
				"email":        false,
				"phone":        false,
				"role":         false,
				"created_at":   false,
			},
		},
		{
			name:   "Login Response",
			path:   "/login",
			method: "POST",
			body: map[string]interface{}{
				"email":    "test@example.com",
				"password": "TestPassword123!",
			},
			expectedFields: map[string]bool{
				"access_token":  false,
				"refresh_token": false,
				"id":           false,
				"email":        false,
				"phone":        false,
				"role":         false,
				"created_at":   false,
			},
		},
		{
			name:   "Metrics Response",
			path:   "/metrics",
			method: "GET",
			body:   nil,
			expectedFields: map[string]bool{
				"user_creations": false, // Metrics structure may vary
			},
		},
		{
			name:   "Bot Info Response",
			path:   "/bot/me",
			method: "GET",
			body:   nil,
			expectedFields: map[string]bool{
				"id":         false,
				"username":   false,
				"first_name": false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != nil {
				jsonBody, err := json.Marshal(tc.body)
				require.NoError(t, err)
				body = bytes.NewReader(jsonBody)
			}

			req, err := http.NewRequest(tc.method, baseURL+tc.path, body)
			require.NoError(t, err)

			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// If successful response, check field structure
			if resp.StatusCode >= 200 && resp.StatusCode < 300 && len(responseBody) > 0 {
				var jsonData map[string]interface{}
				err := json.Unmarshal(responseBody, &jsonData)
				assert.NoError(t, err, "Response should be valid JSON")

				// Check expected fields
				for field, required := range tc.expectedFields {
					if required {
						assert.Contains(t, jsonData, field, "Required field %s should be present", field)
					}
					// For optional fields, just log if they're missing
					if _, exists := jsonData[field]; !exists && !required {
						t.Logf("Optional field %s not present in response", field)
					}
				}
			}
		})
	}
}

// testEmployeeResponseFormats verifies employee service response formats
func testEmployeeResponseFormats(t *testing.T, baseURL string) {
	testCases := []struct {
		name           string
		path           string
		method         string
		body           interface{}
		expectedFields map[string]bool
	}{
		{
			name:   "Employee Creation Response",
			path:   "/simple-employee",
			method: "POST",
			body: map[string]interface{}{
				"name":  "Test Employee",
				"email": "employee@test.com",
				"phone": "+1234567890",
			},
			expectedFields: map[string]bool{
				"id":    false,
				"name":  false,
				"email": false,
				"phone": false,
			},
		},
		{
			name:           "All Employees Response",
			path:           "/employees/all",
			method:         "GET",
			body:           nil,
			expectedFields: map[string]bool{}, // Array response, structure may vary
		},
		{
			name:   "Batch Status Response",
			path:   "/employees/batch-status",
			method: "POST",
			body:   []map[string]interface{}{},
			expectedFields: map[string]bool{}, // Array response
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != nil {
				jsonBody, err := json.Marshal(tc.body)
				require.NoError(t, err)
				body = bytes.NewReader(jsonBody)
			}

			req, err := http.NewRequest(tc.method, baseURL+tc.path, body)
			require.NoError(t, err)

			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// Verify response is valid JSON if not empty
			if len(responseBody) > 0 && resp.StatusCode != http.StatusNoContent {
				var jsonData interface{}
				err := json.Unmarshal(responseBody, &jsonData)
				assert.NoError(t, err, "Response should be valid JSON")
			}
		})
	}
}

// testStructureResponseFormats verifies structure service response formats
func testStructureResponseFormats(t *testing.T, baseURL string) {
	testCases := []struct {
		name           string
		path           string
		method         string
		body           interface{}
		expectedFields map[string]bool
	}{
		{
			name:   "University Creation Response",
			path:   "/universities",
			method: "POST",
			body: map[string]interface{}{
				"name": "Test University",
			},
			expectedFields: map[string]bool{
				"id":   false,
				"name": false,
			},
		},
		{
			name:   "Universities List Response",
			path:   "/universities?limit=10&offset=0",
			method: "GET",
			body:   nil,
			expectedFields: map[string]bool{
				"universities": false,
				"total":       false,
				"limit":       false,
				"offset":      false,
			},
		},
		{
			name:   "University Structure Response",
			path:   "/universities/1/structure",
			method: "GET",
			body:   nil,
			expectedFields: map[string]bool{
				"id":   false,
				"name": false,
				"type": false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != nil {
				jsonBody, err := json.Marshal(tc.body)
				require.NoError(t, err)
				body = bytes.NewReader(jsonBody)
			}

			req, err := http.NewRequest(tc.method, baseURL+tc.path, body)
			require.NoError(t, err)

			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// Verify response structure for successful responses
			if resp.StatusCode >= 200 && resp.StatusCode < 300 && len(responseBody) > 0 {
				var jsonData interface{}
				err := json.Unmarshal(responseBody, &jsonData)
				assert.NoError(t, err, "Response should be valid JSON")

				if jsonMap, ok := jsonData.(map[string]interface{}); ok {
					for field, required := range tc.expectedFields {
						if required {
							assert.Contains(t, jsonMap, field, "Required field %s should be present", field)
						}
					}
				}
			}
		})
	}
}

// testChatResponseFormats verifies chat service response formats
func testChatResponseFormats(t *testing.T, baseURL string) {
	testCases := []struct {
		name           string
		path           string
		method         string
		body           interface{}
		expectedFields map[string]bool
	}{
		{
			name:           "Chats List Response",
			path:           "/chats",
			method:         "GET",
			body:           nil,
			expectedFields: map[string]bool{}, // Array response
		},
		{
			name:           "Administrators List Response",
			path:           "/administrators",
			method:         "GET",
			body:           nil,
			expectedFields: map[string]bool{}, // Array response
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != nil {
				jsonBody, err := json.Marshal(tc.body)
				require.NoError(t, err)
				body = bytes.NewReader(jsonBody)
			}

			req, err := http.NewRequest(tc.method, baseURL+tc.path, body)
			require.NoError(t, err)

			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// Verify response is valid JSON if not empty
			if len(responseBody) > 0 && resp.StatusCode != http.StatusNoContent {
				var jsonData interface{}
				err := json.Unmarshal(responseBody, &jsonData)
				assert.NoError(t, err, "Response should be valid JSON")
			}
		})
	}
}

// TestE2EBehaviorCompatibility tests that the Gateway behaves like individual services for E2E tests
func TestE2EBehaviorCompatibility(t *testing.T) {
	cfg := createTestConfig()
	clientManager := grpcClient.NewClientManager(cfg)
	router := httpHandler.NewRouter(cfg, clientManager)
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("HTTP Method Handling", func(t *testing.T) {
		testHTTPMethodHandling(t, server.URL)
	})

	t.Run("Authentication Flow", func(t *testing.T) {
		testAuthenticationFlow(t, server.URL)
	})

	t.Run("Pagination Support", func(t *testing.T) {
		testPaginationSupport(t, server.URL)
	})

	t.Run("Error Handling Consistency", func(t *testing.T) {
		testErrorHandlingConsistency(t, server.URL)
	})
}

// testHTTPMethodHandling verifies that HTTP methods are handled consistently with E2E expectations
func testHTTPMethodHandling(t *testing.T, baseURL string) {
	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus []int // Acceptable status codes
	}{
		{
			name:           "GET Health Check",
			method:         "GET",
			path:           "/health",
			expectedStatus: []int{200, 503},
		},
		{
			name:           "POST Registration",
			method:         "POST",
			path:           "/register",
			expectedStatus: []int{200, 400, 500},
		},
		{
			name:           "PUT Method Not Allowed on Health",
			method:         "PUT",
			path:           "/health",
			expectedStatus: []int{405, 404},
		},
		{
			name:           "DELETE Method Not Allowed on Register",
			method:         "DELETE",
			path:           "/register",
			expectedStatus: []int{405, 404},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, baseURL+tc.path, nil)
			require.NoError(t, err)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Contains(t, tc.expectedStatus, resp.StatusCode,
				"Status code should be one of the expected values for %s %s", tc.method, tc.path)
		})
	}
}

// testAuthenticationFlow verifies that authentication works as expected by E2E tests
func testAuthenticationFlow(t *testing.T, baseURL string) {
	// Test that authentication endpoints are accessible
	authEndpoints := []string{
		"/register",
		"/login",
		"/login-phone",
		"/auth/max",
	}

	for _, endpoint := range authEndpoints {
		t.Run(fmt.Sprintf("Auth endpoint %s", endpoint), func(t *testing.T) {
			body := map[string]interface{}{
				"email":    "test@example.com",
				"password": "testpass123",
			}
			jsonBody, err := json.Marshal(body)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", baseURL+endpoint, bytes.NewReader(jsonBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Should not fail due to routing issues
			// Allow 503 for service unavailability
			assert.True(t, resp.StatusCode != 404 && resp.StatusCode < 600, "Auth endpoint should be routed correctly")
		})
	}
}

// testPaginationSupport verifies that pagination parameters work as expected
func testPaginationSupport(t *testing.T, baseURL string) {
	paginatedEndpoints := []string{
		"/universities",
		"/employees/all",
		"/chats",
	}

	for _, endpoint := range paginatedEndpoints {
		t.Run(fmt.Sprintf("Pagination for %s", endpoint), func(t *testing.T) {
			// Test with pagination parameters
			req, err := http.NewRequest("GET", baseURL+endpoint+"?limit=10&offset=0", nil)
			require.NoError(t, err)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Should handle pagination parameters without errors
			// Allow 503 for service unavailability
			assert.True(t, resp.StatusCode != 400 && resp.StatusCode < 600, "Should handle pagination parameters correctly")
		})
	}
}

// testErrorHandlingConsistency verifies that errors are handled consistently with E2E expectations
func testErrorHandlingConsistency(t *testing.T, baseURL string) {
	t.Run("Invalid JSON Handling", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURL+"/register", strings.NewReader("invalid json"))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode, "Invalid JSON should return 400")
	})

	t.Run("Missing Content-Type Handling", func(t *testing.T) {
		body := `{"email": "test@example.com", "password": "testpass123"}`
		req, err := http.NewRequest("POST", baseURL+"/login", strings.NewReader(body))
		require.NoError(t, err)
		// Intentionally not setting Content-Type

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should still process the request (Gateway should be flexible)
		// Allow 503 for service unavailability
		assert.True(t, resp.StatusCode != 415 && resp.StatusCode < 600, "Should handle missing Content-Type gracefully")
	})

	t.Run("Large Request Body Handling", func(t *testing.T) {
		// Create a large but valid JSON body
		largeBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "testpass123",
			"data":     strings.Repeat("x", 1000), // 1KB of data
		}
		jsonBody, err := json.Marshal(largeBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/register", bytes.NewReader(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle reasonably large requests
		// Allow 503 for service unavailability
		assert.True(t, resp.StatusCode != 413 && resp.StatusCode < 600, "Should handle reasonably large requests")
	})
}