package test

import (
	"auth-service/internal/domain"
	httpInfra "auth-service/internal/infrastructure/http"
	"auth-service/internal/infrastructure/max"
	"auth-service/internal/usecase"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty6_APIContractCompliance tests API contract compliance
// **Feature: max-miniapp-auth, Property 6: API contract compliance**
// **Validates: Requirements 4.2, 4.3, 4.4, 4.5**
func TestProperty6_APIContractCompliance(t *testing.T) {
	// Setup mock dependencies
	userRepo := &mockUserRepository{}
	refreshRepo := &mockRefreshRepository{}
	hasher := &mockHasher{}
	jwtManager := &mockJWTManager{}
	maxValidator := max.NewAuthValidator()

	authService := usecase.NewAuthService(userRepo, refreshRepo, hasher, jwtManager, nil)
	authService.SetMaxAuthValidator(maxValidator)
	authService.SetMaxBotToken("test_bot_token_12345")

	handler := httpInfra.NewHandler(authService)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Test 1: Valid JSON payloads with correct initData should return success with tokens
	properties.Property("valid JSON payloads with correct initData should return success with tokens", prop.ForAll(
		func(maxID int64, firstNameSeed int, usernameSeed int) bool {
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
			username := "user" + strconv.Itoa(usernameSeed%1000000)

			// Create valid initData with correct hash
			params := fmt.Sprintf("max_id=%d&first_name=%s&username=%s",
				maxID, firstName, username)
			initData := createInitDataWithCorrectHash(params, "test_bot_token_12345")

			// Create request payload
			requestBody := map[string]string{
				"init_data": initData,
			}
			jsonBody, _ := json.Marshal(requestBody)

			// Make request
			req := httptest.NewRequest("POST", "/auth/max", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.AuthenticateMAX(w, req)

			// Should return 200 OK
			if w.Code != 200 {
				t.Logf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
				return false
			}

			// Response should contain access_token and refresh_token
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Logf("Failed to parse response JSON: %v", err)
				return false
			}

			accessToken, hasAccessToken := response["access_token"].(string)
			refreshToken, hasRefreshToken := response["refresh_token"].(string)

			if !hasAccessToken || accessToken == "" {
				t.Logf("Response missing or empty access_token")
				return false
			}

			if !hasRefreshToken || refreshToken == "" {
				t.Logf("Response missing or empty refresh_token")
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
		gen.IntRange(100000, 999999),    // username seed
	))

	// Test 2: Invalid JSON payloads should return validation errors
	properties.Property("invalid JSON payloads should return validation errors", prop.ForAll(
		func(invalidJSONSeed int) bool {
			// Create various types of invalid JSON
			invalidJSONs := []string{
				"invalid json",
				"{invalid}",
				"null",
				"[]",
				"{\"wrong_field\": \"value\"}",
			}

			invalidJSON := invalidJSONs[invalidJSONSeed%len(invalidJSONs)]

			// Make request with invalid JSON
			req := httptest.NewRequest("POST", "/auth/max", strings.NewReader(invalidJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.AuthenticateMAX(w, req)

			// Should return 400 Bad Request
			if w.Code != 400 {
				t.Logf("Expected status 400 for invalid JSON, got %d. JSON: %s, Response: %s", 
					w.Code, invalidJSON, w.Body.String())
				return false
			}

			// Response should contain error information
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Logf("Failed to parse error response JSON: %v", err)
				return false
			}

			// Should have error field
			if _, hasError := response["error"]; !hasError {
				t.Logf("Error response missing error field")
				return false
			}

			return true
		},
		gen.IntRange(0, 4), // invalidJSON seed
	))

	// Test 3: Missing init_data field should return validation error
	properties.Property("missing init_data field should return validation error", prop.ForAll(
		func(otherFieldSeed int) bool {
			// Create request without init_data field
			requestBody := map[string]string{
				"other_field": "value" + strconv.Itoa(otherFieldSeed%1000000),
			}
			jsonBody, _ := json.Marshal(requestBody)

			// Make request
			req := httptest.NewRequest("POST", "/auth/max", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.AuthenticateMAX(w, req)

			// Should return 400 Bad Request
			if w.Code != 400 {
				t.Logf("Expected status 400 for missing init_data, got %d. Response: %s", 
					w.Code, w.Body.String())
				return false
			}

			// Response should indicate missing field
			responseBody := w.Body.String()
			if !strings.Contains(strings.ToLower(responseBody), "init_data") {
				t.Logf("Error response should mention init_data field. Response: %s", responseBody)
				return false
			}

			return true
		},
		gen.IntRange(100000, 999999), // otherField seed
	))

	// Test 4: Empty init_data field should return validation error
	properties.Property("empty init_data field should return validation error", prop.ForAll(
		func() bool {
			// Create request with empty init_data
			requestBody := map[string]string{
				"init_data": "",
			}
			jsonBody, _ := json.Marshal(requestBody)

			// Make request
			req := httptest.NewRequest("POST", "/auth/max", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.AuthenticateMAX(w, req)

			// Should return 400 Bad Request
			if w.Code != 400 {
				t.Logf("Expected status 400 for empty init_data, got %d. Response: %s", 
					w.Code, w.Body.String())
				return false
			}

			return true
		},
	))

	// Test 5: Invalid initData should return authentication failure
	properties.Property("invalid initData should return authentication failure", prop.ForAll(
		func(invalidDataSeed int) bool {
			// Create various types of invalid initData
			invalidInitDatas := []string{
				"invalid_init_data",
				"max_id=123&hash=invalid_hash",
				"malformed_query_string",
				"max_id=abc&first_name=test&hash=123",
			}

			invalidInitData := invalidInitDatas[invalidDataSeed%len(invalidInitDatas)]

			// Create request with invalid initData
			requestBody := map[string]string{
				"init_data": invalidInitData,
			}
			jsonBody, _ := json.Marshal(requestBody)

			// Make request
			req := httptest.NewRequest("POST", "/auth/max", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.AuthenticateMAX(w, req)

			// Should return 401 Unauthorized or 400 Bad Request
			if w.Code != 401 && w.Code != 400 {
				t.Logf("Expected status 401 or 400 for invalid initData, got %d. InitData: %s, Response: %s", 
					w.Code, invalidInitData, w.Body.String())
				return false
			}

			// Response should contain error information
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Logf("Failed to parse error response JSON: %v", err)
				return false
			}

			// Should have error field
			if _, hasError := response["error"]; !hasError {
				t.Logf("Error response missing error field")
				return false
			}

			return true
		},
		gen.IntRange(0, 3), // invalidData seed
	))

	// Test 6: Response content type should always be application/json
	properties.Property("response content type should always be application/json", prop.ForAll(
		func(maxID int64, firstNameSeed int, isValidRequest bool) bool {
			var req *http.Request

			if isValidRequest {
				// Create valid request
				firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
				params := fmt.Sprintf("max_id=%d&first_name=%s", maxID, firstName)
				initData := createInitDataWithCorrectHash(params, "test_bot_token_12345")

				requestBody := map[string]string{
					"init_data": initData,
				}
				jsonBody, _ := json.Marshal(requestBody)
				req = httptest.NewRequest("POST", "/auth/max", bytes.NewReader(jsonBody))
			} else {
				// Create invalid request
				req = httptest.NewRequest("POST", "/auth/max", strings.NewReader("invalid json"))
			}

			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.AuthenticateMAX(w, req)

			// Content-Type should always be application/json
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Logf("Expected Content-Type application/json, got %s", contentType)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
		gen.Bool(),                      // isValidRequest
	))

	properties.TestingRun(t)
}


// Mock user repository for API contract testing
type mockUserRepository struct {
	users map[int64]*domain.User
}

func (m *mockUserRepository) Create(user *domain.User) error {
	if m.users == nil {
		m.users = make(map[int64]*domain.User)
	}
	user.ID = int64(len(m.users) + 1)
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) GetByID(id int64) (*domain.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserRepository) GetByEmail(email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserRepository) GetByPhone(phone string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Phone == phone {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserRepository) GetByMaxID(maxID int64) (*domain.User, error) {
	for _, user := range m.users {
		if user.MaxID != nil && *user.MaxID == maxID {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserRepository) Update(user *domain.User) error {
	if _, exists := m.users[user.ID]; exists {
		m.users[user.ID] = user
		return nil
	}
	return fmt.Errorf("user not found")
}

// Mock refresh repository for API contract testing
type mockRefreshRepository struct {
	tokens map[string]bool
}

func (m *mockRefreshRepository) Save(jti string, userID int64, expiresAt time.Time) error {
	if m.tokens == nil {
		m.tokens = make(map[string]bool)
	}
	m.tokens[jti] = true
	return nil
}

func (m *mockRefreshRepository) IsValid(jti string) (bool, error) {
	valid, exists := m.tokens[jti]
	return exists && valid, nil
}

func (m *mockRefreshRepository) Revoke(jti string) error {
	delete(m.tokens, jti)
	return nil
}

func (m *mockRefreshRepository) RevokeAllForUser(userID int64) error {
	// For simplicity, just clear all tokens in mock
	m.tokens = make(map[string]bool)
	return nil
}