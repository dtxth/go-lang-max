package main

import (
	"e2e-tests/utils"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGatewayCompatibility tests that the Gateway Service maintains API compatibility
// with the original microservice endpoints for E2E test compatibility
func TestGatewayCompatibility(t *testing.T) {
	// Use Gateway Service configuration
	gatewayConfig := utils.ServiceConfig{
		BaseURL: "http://localhost:8080", // Gateway Service port
		Timeout: 30 * time.Second,
	}
	client := utils.NewTestClient(gatewayConfig)

	// Wait for Gateway Service to be available
	err := utils.WaitForService(gatewayConfig.BaseURL, 10)
	require.NoError(t, err, "Gateway service should be available")

	t.Run("Auth Service Compatibility", func(t *testing.T) {
		testAuthServiceCompatibility(t, client)
	})

	t.Run("Employee Service Compatibility", func(t *testing.T) {
		testEmployeeServiceCompatibility(t, client)
	})

	t.Run("Chat Service Compatibility", func(t *testing.T) {
		testChatServiceCompatibility(t, client)
	})

	t.Run("Structure Service Compatibility", func(t *testing.T) {
		testStructureServiceCompatibility(t, client)
	})
}

func testAuthServiceCompatibility(t *testing.T, client *utils.TestClient) {
	t.Run("Health Check Response Format", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/health")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
	})

	t.Run("Metrics Response Format", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/metrics")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var metrics map[string]interface{}
		err = json.Unmarshal(resp.Body(), &metrics)
		require.NoError(t, err)
		assert.Contains(t, metrics, "user_creations")
	})

	t.Run("Registration Response Format", func(t *testing.T) {
		testUser := utils.GenerateTestUser()
		
		resp, err := client.GetClient().R().
			SetBody(testUser).
			Post("/register")
		
		require.NoError(t, err)
		
		// Test successful registration response format
		if resp.StatusCode() == 200 {
			var user map[string]interface{}
			err = json.Unmarshal(resp.Body(), &user)
			require.NoError(t, err)
			assert.Contains(t, user, "email")
			assert.Contains(t, user, "phone")
			assert.Equal(t, testUser.Email, user["email"])
			assert.Equal(t, testUser.Phone, user["phone"])
		}
	})

	t.Run("Login Response Format", func(t *testing.T) {
		testUser := utils.GenerateTestUser()
		
		// Register user first
		_, err := client.GetClient().R().
			SetBody(testUser).
			Post("/register")
		require.NoError(t, err)
		
		loginData := map[string]string{
			"email":    testUser.Email,
			"password": testUser.Password,
		}
		
		resp, err := client.GetClient().R().
			SetBody(loginData).
			Post("/login")
		
		require.NoError(t, err)
		
		// Test successful login response format
		if resp.StatusCode() == 200 {
			var tokens map[string]interface{}
			err = json.Unmarshal(resp.Body(), &tokens)
			require.NoError(t, err)
			assert.Contains(t, tokens, "access_token")
			assert.Contains(t, tokens, "refresh_token")
		}
	})

	t.Run("Error Response Format", func(t *testing.T) {
		invalidUser := map[string]string{
			"email": "invalid-email",
		}
		
		resp, err := client.GetClient().R().
			SetBody(invalidUser).
			Post("/register")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
		
		// Verify error response has expected structure
		var errorResp map[string]interface{}
		err = json.Unmarshal(resp.Body(), &errorResp)
		require.NoError(t, err)
		// Error response should contain error information
		assert.True(t, len(errorResp) > 0)
	})
}

func testEmployeeServiceCompatibility(t *testing.T, client *utils.TestClient) {
	t.Run("Create Employee Response Format", func(t *testing.T) {
		testEmployee := utils.GenerateTestEmployee()
		
		resp, err := client.GetClient().R().
			SetBody(testEmployee).
			Post("/simple-employee")
		
		require.NoError(t, err)
		
		// Test successful creation response format
		if resp.StatusCode() == 201 {
			var employee map[string]interface{}
			err = json.Unmarshal(resp.Body(), &employee)
			require.NoError(t, err)
			assert.Contains(t, employee, "name")
			assert.Contains(t, employee, "email")
			assert.Contains(t, employee, "phone")
			assert.Equal(t, testEmployee.Name, employee["name"])
			assert.Equal(t, testEmployee.Email, employee["email"])
			assert.Equal(t, testEmployee.Phone, employee["phone"])
		}
	})

	t.Run("Get All Employees Response Format", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/employees/all")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var employees []interface{}
		err = json.Unmarshal(resp.Body(), &employees)
		require.NoError(t, err)
		// Should return an array (even if empty)
		assert.IsType(t, []interface{}{}, employees)
	})

	t.Run("Batch Status Response Format", func(t *testing.T) {
		statusData := []map[string]interface{}{
			{
				"employee_id": "test-id",
			},
		}
		
		resp, err := client.GetClient().R().
			SetBody(statusData).
			Post("/employees/batch-status")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var statuses []interface{}
		err = json.Unmarshal(resp.Body(), &statuses)
		require.NoError(t, err)
		assert.IsType(t, []interface{}{}, statuses)
	})

	t.Run("Invalid Employee Error Response", func(t *testing.T) {
		invalidEmployee := map[string]interface{}{
			"name": "",
		}
		
		resp, err := client.GetClient().R().
			SetBody(invalidEmployee).
			Post("/simple-employee")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})
}

func testChatServiceCompatibility(t *testing.T, client *utils.TestClient) {
	// First get auth token for chat service tests
	testUser := utils.GenerateTestUser()
	
	// Register user
	_, err := client.GetClient().R().
		SetBody(testUser).
		Post("/register")
	require.NoError(t, err)
	
	// Login to get token
	loginResp, err := client.GetClient().R().
		SetBody(map[string]string{
			"email":    testUser.Email,
			"password": testUser.Password,
		}).
		Post("/login")
	require.NoError(t, err)
	
	if loginResp.StatusCode() == 200 {
		var tokens map[string]interface{}
		err = json.Unmarshal(loginResp.Body(), &tokens)
		require.NoError(t, err)
		
		if accessToken, ok := tokens["access_token"].(string); ok {
			client.SetAuthToken(accessToken)
		}
	}

	t.Run("Health Check Response Format", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/health")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
	})

	t.Run("Get Chats Response Format", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/chats")
		
		require.NoError(t, err)
		// Should return 200 or appropriate status code
		assert.True(t, resp.StatusCode() >= 200 && resp.StatusCode() < 500)
		
		if resp.StatusCode() == 200 {
			var chats []interface{}
			err = json.Unmarshal(resp.Body(), &chats)
			require.NoError(t, err)
			assert.IsType(t, []interface{}{}, chats)
		}
	})

	t.Run("Unauthorized Access Status Code", func(t *testing.T) {
		// Clear auth token
		client.ClearAuth()
		
		resp, err := client.GetClient().R().Get("/chats")
		
		require.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode())
	})
}

func testStructureServiceCompatibility(t *testing.T, client *utils.TestClient) {
	t.Run("Create University Response Format", func(t *testing.T) {
		testUniversity := utils.GenerateTestUniversity()
		
		resp, err := client.GetClient().R().
			SetBody(testUniversity).
			Post("/universities")
		
		require.NoError(t, err)
		
		// Test successful creation response format
		if resp.StatusCode() == 201 {
			var university map[string]interface{}
			err = json.Unmarshal(resp.Body(), &university)
			require.NoError(t, err)
			assert.Contains(t, university, "name")
			assert.Contains(t, university, "id")
			assert.Equal(t, testUniversity.Name, university["name"])
		}
	})

	t.Run("Get Universities Response Format", func(t *testing.T) {
		resp, err := client.GetClient().R().
			SetQueryParam("limit", "10").
			SetQueryParam("offset", "0").
			Get("/universities")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var response map[string]interface{}
		err = json.Unmarshal(resp.Body(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "universities")
		assert.Contains(t, response, "total")
		assert.Contains(t, response, "limit")
		assert.Contains(t, response, "offset")
	})

	t.Run("Get Department Managers Response Format", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/departments/managers")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var managers []interface{}
		err = json.Unmarshal(resp.Body(), &managers)
		require.NoError(t, err)
		assert.IsType(t, []interface{}{}, managers)
	})

	t.Run("Invalid University ID Error Response", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/universities/invalid")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})

	t.Run("Non-existent University Error Response", func(t *testing.T) {
		resp, err := client.GetClient().R().Get("/universities/99999")
		
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode())
	})
}

// TestStatusCodeMapping tests that gRPC errors are properly mapped to HTTP status codes
func TestStatusCodeMapping(t *testing.T) {
	gatewayConfig := utils.ServiceConfig{
		BaseURL: "http://localhost:8080",
		Timeout: 30 * time.Second,
	}
	client := utils.NewTestClient(gatewayConfig)

	// Wait for Gateway Service to be available
	err := utils.WaitForService(gatewayConfig.BaseURL, 10)
	require.NoError(t, err, "Gateway service should be available")

	t.Run("Bad Request (400)", func(t *testing.T) {
		// Test invalid data that should return 400
		invalidData := map[string]interface{}{
			"invalid_field": "invalid_value",
		}
		
		resp, err := client.GetClient().R().
			SetBody(invalidData).
			Post("/register")
		
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode())
	})

	t.Run("Unauthorized (401)", func(t *testing.T) {
		// Test unauthorized access
		resp, err := client.GetClient().R().
			SetBody(map[string]string{
				"email":    "nonexistent@example.com",
				"password": "wrongpassword",
			}).
			Post("/login")
		
		require.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode())
	})

	t.Run("Not Found (404)", func(t *testing.T) {
		// Test non-existent resource
		resp, err := client.GetClient().R().Get("/universities/99999")
		
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode())
	})

	t.Run("Method Not Allowed (405)", func(t *testing.T) {
		// Test unsupported HTTP method
		resp, err := client.GetClient().R().Delete("/employees/all")
		
		require.NoError(t, err)
		assert.Equal(t, 405, resp.StatusCode())
	})
}

// TestRequestResponseFormatPreservation tests that request/response formats are preserved
func TestRequestResponseFormatPreservation(t *testing.T) {
	gatewayConfig := utils.ServiceConfig{
		BaseURL: "http://localhost:8080",
		Timeout: 30 * time.Second,
	}
	client := utils.NewTestClient(gatewayConfig)

	// Wait for Gateway Service to be available
	err := utils.WaitForService(gatewayConfig.BaseURL, 10)
	require.NoError(t, err, "Gateway service should be available")

	t.Run("JSON Content-Type Headers", func(t *testing.T) {
		testUser := utils.GenerateTestUser()
		
		resp, err := client.GetClient().R().
			SetHeader("Content-Type", "application/json").
			SetBody(testUser).
			Post("/register")
		
		require.NoError(t, err)
		
		// Response should have JSON content type
		contentType := resp.Header().Get("Content-Type")
		assert.Contains(t, contentType, "application/json")
	})

	t.Run("Query Parameters Handling", func(t *testing.T) {
		resp, err := client.GetClient().R().
			SetQueryParam("limit", "10").
			SetQueryParam("offset", "0").
			Get("/universities")
		
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
		
		var response map[string]interface{}
		err = json.Unmarshal(resp.Body(), &response)
		require.NoError(t, err)
		
		// Verify pagination parameters are handled correctly
		assert.Contains(t, response, "limit")
		assert.Contains(t, response, "offset")
	})

	t.Run("Path Parameters Handling", func(t *testing.T) {
		// Create a university first
		testUniversity := utils.GenerateTestUniversity()
		
		createResp, err := client.GetClient().R().
			SetBody(testUniversity).
			Post("/universities")
		
		require.NoError(t, err)
		
		if createResp.StatusCode() == 201 {
			var university map[string]interface{}
			err = json.Unmarshal(createResp.Body(), &university)
			require.NoError(t, err)
			
			universityID := int(university["id"].(float64))
			
			// Test path parameter handling
			resp, err := client.GetClient().R().
				Get(fmt.Sprintf("/universities/%d", universityID))
			
			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode())
			
			var retrievedUniversity map[string]interface{}
			err = json.Unmarshal(resp.Body(), &retrievedUniversity)
			require.NoError(t, err)
			assert.Equal(t, float64(universityID), retrievedUniversity["id"])
		}
	})
}