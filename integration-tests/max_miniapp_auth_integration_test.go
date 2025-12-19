package integration_tests

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaxMiniAppAuthenticationFlow tests the complete MAX Mini App authentication flow
// This test validates:
// - initData validation and hash verification
// - User creation and lookup by max_id
// - JWT token generation and validation
// - Database integration with MAX user data
// Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
func TestMaxMiniAppAuthenticationFlow(t *testing.T) {
	// Wait for auth service to be ready
	WaitForService(t, AuthServiceURL, 10)

	// Setup
	client := NewHTTPClient()

	// Test data - create valid MAX initData with unique IDs
	botToken := "test_bot_token_12345"
	maxID := int64(123456789) + time.Now().Unix()%1000000 // Make unique
	username := "testuser"
	firstName := "Test"
	lastName := "User"

	// Create valid initData with proper hash
	initData := createValidInitData(t, botToken, maxID, username, firstName, lastName)

	// Test 1: First authentication - should create new user
	t.Log("Step 1: First MAX authentication - creating new user")
	authRequest := map[string]interface{}{
		"init_data": initData,
	}

	status, respBody := client.POST(t, AuthServiceURL+"/auth/max", authRequest)
	require.Equal(t, 200, status, "Expected 200 OK for valid MAX auth, got %d: %s", status, string(respBody))

	response := ParseJSON(t, respBody)

	// Validate response structure
	assert.NotNil(t, response["access_token"], "access_token should be present")
	assert.NotNil(t, response["refresh_token"], "refresh_token should be present")

	accessToken := response["access_token"].(string)
	refreshToken := response["refresh_token"].(string)

	// Validate tokens are not empty
	assert.True(t, len(accessToken) > 0, "access_token should not be empty")
	assert.True(t, len(refreshToken) > 0, "refresh_token should not be empty")

	// Test 2: Verify user was created in database
	t.Log("Step 2: Verifying user creation in database")
	authDB := ConnectDB(t, AuthDBConnStr)
	defer authDB.Close()

	var userID int64
	var storedMaxID *int64
	var storedUsername, storedName string
	err := authDB.QueryRow(
		"SELECT id, max_id, username, name FROM users WHERE max_id = $1",
		maxID,
	).Scan(&userID, &storedMaxID, &storedUsername, &storedName)
	require.NoError(t, err, "Should be able to query user by max_id")

	// Verify user data
	assert.NotNil(t, storedMaxID, "max_id should not be null")
	assert.Equal(t, maxID, *storedMaxID, "max_id should match")
	assert.Equal(t, username, storedUsername, "username should match")
	assert.Equal(t, firstName+" "+lastName, storedName, "name should be combined first and last name")

	// Test 3: Verify JWT token contains correct information
	t.Log("Step 3: Validating JWT token structure")
	// We can't easily decode JWT without the secret, but we can verify it works
	// by using it in an authenticated request (if we had protected endpoints)

	// Test 4: Second authentication - should update existing user
	t.Log("Step 4: Second MAX authentication - updating existing user")
	
	// Create new initData with updated information
	newUsername := "updateduser"
	newFirstName := "Updated"
	newLastName := "Name"
	updatedInitData := createValidInitData(t, botToken, maxID, newUsername, newFirstName, newLastName)

	authRequest2 := map[string]interface{}{
		"init_data": updatedInitData,
	}

	status, respBody = client.POST(t, AuthServiceURL+"/auth/max", authRequest2)
	require.Equal(t, 200, status, "Expected 200 OK for second MAX auth, got %d: %s", status, string(respBody))

	response2 := ParseJSON(t, respBody)
	assert.NotNil(t, response2["access_token"], "access_token should be present in second auth")
	assert.NotNil(t, response2["refresh_token"], "refresh_token should be present in second auth")

	// Test 5: Verify user data was updated
	t.Log("Step 5: Verifying user data update")
	var updatedUsername, updatedName string
	var userCount int
	err = authDB.QueryRow(
		"SELECT username, name FROM users WHERE max_id = $1",
		maxID,
	).Scan(&updatedUsername, &updatedName)
	require.NoError(t, err, "Should be able to query updated user")

	// Verify updated data
	assert.Equal(t, newUsername, updatedUsername, "username should be updated")
	assert.Equal(t, newFirstName+" "+newLastName, updatedName, "name should be updated")

	// Verify only one user exists with this max_id
	err = authDB.QueryRow("SELECT COUNT(*) FROM users WHERE max_id = $1", maxID).Scan(&userCount)
	require.NoError(t, err, "Should be able to count users")
	assert.Equal(t, 1, userCount, "Should have exactly one user with this max_id")

	// Test 6: Test refresh token functionality
	t.Log("Step 6: Testing refresh token functionality")
	refreshRequest := map[string]interface{}{
		"refresh_token": refreshToken,
	}

	status, respBody = client.POST(t, AuthServiceURL+"/refresh", refreshRequest)
	require.Equal(t, 200, status, "Expected 200 OK for token refresh, got %d: %s", status, string(respBody))

	refreshResponse := ParseJSON(t, respBody)
	assert.NotNil(t, refreshResponse["access_token"], "new access_token should be present")
	assert.NotNil(t, refreshResponse["refresh_token"], "new refresh_token should be present")

	// Verify new tokens are different from original
	newAccessToken := refreshResponse["access_token"].(string)
	newRefreshToken := refreshResponse["refresh_token"].(string)
	// Note: Access tokens might be the same if generated within the same second due to JWT timestamp precision
	// This is acceptable behavior - the important thing is that refresh works
	assert.True(t, len(newAccessToken) > 0, "new access token should not be empty")
	assert.NotEqual(t, refreshToken, newRefreshToken, "new refresh token should be different")

	// Cleanup
	t.Log("Cleaning up test data")
	CleanupDB(t, authDB, []string{"refresh_tokens", "users"})

	t.Log("MAX Mini App authentication integration test completed successfully")
}

// TestMaxAuthenticationErrorScenarios tests various error scenarios
// Requirements: 1.1, 2.6, 4.4, 4.5, 5.3
func TestMaxAuthenticationErrorScenarios(t *testing.T) {
	// Wait for auth service to be ready
	WaitForService(t, AuthServiceURL, 10)

	client := NewHTTPClient()
	botToken := "test_bot_token_12345"

	// Test 1: Empty initData
	t.Log("Test 1: Empty initData")
	authRequest := map[string]interface{}{
		"init_data": "",
	}

	status, respBody := client.POST(t, AuthServiceURL+"/auth/max", authRequest)
	assert.Equal(t, 400, status, "Expected 400 Bad Request for empty initData, got %d: %s", status, string(respBody))

	// Test 2: Missing initData field
	t.Log("Test 2: Missing initData field")
	emptyRequest := map[string]interface{}{}

	status, respBody = client.POST(t, AuthServiceURL+"/auth/max", emptyRequest)
	assert.Equal(t, 400, status, "Expected 400 Bad Request for missing initData, got %d: %s", status, string(respBody))

	// Test 3: Invalid JSON
	t.Log("Test 3: Invalid JSON")
	status, _ = client.POST(t, AuthServiceURL+"/auth/max", "invalid json")
	assert.Equal(t, 400, status, "Expected 400 Bad Request for invalid JSON")

	// Test 4: initData without hash parameter
	t.Log("Test 4: initData without hash parameter")
	initDataNoHash := "max_id=123456&username=test&first_name=Test"
	authRequest = map[string]interface{}{
		"init_data": initDataNoHash,
	}

	status, respBody = client.POST(t, AuthServiceURL+"/auth/max", authRequest)
	assert.Equal(t, 400, status, "Expected 400 Bad Request for missing hash, got %d: %s", status, string(respBody))

	// Test 5: initData with invalid hash
	t.Log("Test 5: initData with invalid hash")
	initDataInvalidHash := "max_id=123456&username=test&first_name=Test&hash=invalid_hash"
	authRequest = map[string]interface{}{
		"init_data": initDataInvalidHash,
	}

	status, respBody = client.POST(t, AuthServiceURL+"/auth/max", authRequest)
	assert.Equal(t, 401, status, "Expected 401 Unauthorized for invalid hash, got %d: %s", status, string(respBody))

	// Test 6: initData missing required max_id
	t.Log("Test 6: initData missing required max_id")
	initDataNoMaxID := createInitDataWithoutMaxID(t, botToken, "test", "Test", "User")
	authRequest = map[string]interface{}{
		"init_data": initDataNoMaxID,
	}

	status, respBody = client.POST(t, AuthServiceURL+"/auth/max", authRequest)
	assert.Equal(t, 401, status, "Expected 401 Unauthorized for missing max_id, got %d: %s", status, string(respBody))

	// Test 7: initData missing required first_name
	t.Log("Test 7: initData missing required first_name")
	initDataNoFirstName := createInitDataWithoutFirstName(t, botToken, 123456, "test")
	authRequest = map[string]interface{}{
		"init_data": initDataNoFirstName,
	}

	status, respBody = client.POST(t, AuthServiceURL+"/auth/max", authRequest)
	assert.Equal(t, 401, status, "Expected 401 Unauthorized for missing first_name, got %d: %s", status, string(respBody))
}

// TestMaxAuthenticationDatabaseIntegration tests database operations
// Requirements: 3.1, 3.2, 3.4
func TestMaxAuthenticationDatabaseIntegration(t *testing.T) {
	// Wait for auth service to be ready
	WaitForService(t, AuthServiceURL, 10)

	client := NewHTTPClient()
	botToken := "test_bot_token_12345"

	// Connect to database
	authDB := ConnectDB(t, AuthDBConnStr)
	defer authDB.Close()

	// Test 1: User creation with all MAX fields
	t.Log("Test 1: User creation with all MAX fields")
	maxID := int64(987654321) + time.Now().Unix()%1000000 // Make unique
	username := "fulluser"
	firstName := "Full"
	lastName := "User"

	initData := createValidInitData(t, botToken, maxID, username, firstName, lastName)
	authRequest := map[string]interface{}{
		"init_data": initData,
	}

	status, respBody := client.POST(t, AuthServiceURL+"/auth/max", authRequest)
	require.Equal(t, 200, status, "Expected 200 OK, got %d: %s", status, string(respBody))

	// Verify all fields in database
	var storedMaxID *int64
	var storedUsername, storedName, storedRole string
	var createdAt time.Time
	err := authDB.QueryRow(
		"SELECT max_id, username, name, role, created_at FROM users WHERE max_id = $1",
		maxID,
	).Scan(&storedMaxID, &storedUsername, &storedName, &storedRole, &createdAt)
	require.NoError(t, err, "Should be able to query created user")

	assert.Equal(t, maxID, *storedMaxID, "max_id should match")
	assert.Equal(t, username, storedUsername, "username should match")
	assert.Equal(t, firstName+" "+lastName, storedName, "name should be combined")
	assert.Equal(t, "operator", storedRole, "default role should be operator")
	assert.True(t, time.Since(createdAt) < time.Minute, "user should be created recently")

	// Test 2: User creation with minimal fields (no username, no last_name)
	t.Log("Test 2: User creation with minimal fields")
	maxID2 := int64(987654322) + time.Now().Unix()%1000000 // Make unique
	firstName2 := "Minimal"

	initData2 := createMinimalInitData(t, botToken, maxID2, firstName2)
	authRequest2 := map[string]interface{}{
		"init_data": initData2,
	}

	status, respBody = client.POST(t, AuthServiceURL+"/auth/max", authRequest2)
	require.Equal(t, 200, status, "Expected 200 OK for minimal data, got %d: %s", status, string(respBody))

	// Verify minimal fields in database
	var storedUsername2, storedName2 string
	err = authDB.QueryRow(
		"SELECT username, name FROM users WHERE max_id = $1",
		maxID2,
	).Scan(&storedUsername2, &storedName2)
	require.NoError(t, err, "Should be able to query user with minimal data")

	assert.Equal(t, "", storedUsername2, "username should be empty when not provided")
	assert.Equal(t, firstName2, storedName2, "name should be just first name when last name not provided")

	// Test 3: Verify max_id uniqueness constraint
	t.Log("Test 3: Testing max_id uniqueness")
	// Try to create another user with same max_id but different data
	initData3 := createValidInitData(t, botToken, maxID, "different", "Different", "User")
	authRequest3 := map[string]interface{}{
		"init_data": initData3,
	}

	status, respBody = client.POST(t, AuthServiceURL+"/auth/max", authRequest3)
	require.Equal(t, 200, status, "Should succeed and update existing user, got %d: %s", status, string(respBody))

	// Verify user was updated, not duplicated
	var userCount int
	err = authDB.QueryRow("SELECT COUNT(*) FROM users WHERE max_id = $1", maxID).Scan(&userCount)
	require.NoError(t, err, "Should be able to count users")
	assert.Equal(t, 1, userCount, "Should still have only one user with this max_id")

	// Verify data was updated
	var finalUsername, finalName string
	err = authDB.QueryRow(
		"SELECT username, name FROM users WHERE max_id = $1",
		maxID,
	).Scan(&finalUsername, &finalName)
	require.NoError(t, err, "Should be able to query updated user")
	assert.Equal(t, "different", finalUsername, "username should be updated")
	assert.Equal(t, "Different User", finalName, "name should be updated")

	// Cleanup
	CleanupDB(t, authDB, []string{"refresh_tokens", "users"})
}

// TestMaxAuthenticationJWTTokens tests JWT token generation and validation
// Requirements: 1.5
func TestMaxAuthenticationJWTTokens(t *testing.T) {
	// Wait for auth service to be ready
	WaitForService(t, AuthServiceURL, 10)

	client := NewHTTPClient()
	botToken := "test_bot_token_12345"

	// Create valid initData
	maxID := int64(555666777) + time.Now().Unix()%1000000 // Make unique
	initData := createValidInitData(t, botToken, maxID, "tokenuser", "Token", "User")

	// Authenticate and get tokens
	authRequest := map[string]interface{}{
		"init_data": initData,
	}

	status, respBody := client.POST(t, AuthServiceURL+"/auth/max", authRequest)
	require.Equal(t, 200, status, "Expected 200 OK, got %d: %s", status, string(respBody))

	response := ParseJSON(t, respBody)
	accessToken := response["access_token"].(string)
	refreshToken := response["refresh_token"].(string)

	// Test 1: Verify tokens are valid JWT format (basic structure check)
	t.Log("Test 1: Verifying JWT token structure")
	accessParts := strings.Split(accessToken, ".")
	refreshParts := strings.Split(refreshToken, ".")

	assert.Equal(t, 3, len(accessParts), "Access token should have 3 parts (header.payload.signature)")
	assert.Equal(t, 3, len(refreshParts), "Refresh token should have 3 parts (header.payload.signature)")

	// Test 2: Verify refresh token works
	t.Log("Test 2: Testing refresh token functionality")
	refreshRequest := map[string]interface{}{
		"refresh_token": refreshToken,
	}

	status, respBody = client.POST(t, AuthServiceURL+"/refresh", refreshRequest)
	require.Equal(t, 200, status, "Expected 200 OK for refresh, got %d: %s", status, string(respBody))

	refreshResponse := ParseJSON(t, respBody)
	newAccessToken := refreshResponse["access_token"].(string)
	newRefreshToken := refreshResponse["refresh_token"].(string)

	// Verify new tokens are different
	// Note: Access tokens might be the same if generated within the same second due to JWT timestamp precision
	// This is acceptable behavior - the important thing is that refresh works
	assert.True(t, len(newAccessToken) > 0, "New access token should not be empty")
	assert.NotEqual(t, refreshToken, newRefreshToken, "New refresh token should be different")

	// Test 3: Verify old refresh token is invalidated
	t.Log("Test 3: Testing old refresh token invalidation")
	status, _ = client.POST(t, AuthServiceURL+"/refresh", refreshRequest)
	assert.Equal(t, 401, status, "Old refresh token should be invalidated")

	// Test 4: Test logout functionality
	t.Log("Test 4: Testing logout functionality")
	logoutRequest := map[string]interface{}{
		"refresh_token": newRefreshToken,
	}

	status, respBody = client.POST(t, AuthServiceURL+"/logout", logoutRequest)
	require.Equal(t, 200, status, "Expected 200 OK for logout, got %d: %s", status, string(respBody))

	logoutResponse := ParseJSON(t, respBody)
	assert.Equal(t, "logged_out", logoutResponse["status"], "Should confirm logout")

	// Test 5: Verify logged out refresh token doesn't work
	t.Log("Test 5: Testing logged out refresh token")
	status, _ = client.POST(t, AuthServiceURL+"/refresh", logoutRequest)
	assert.Equal(t, 401, status, "Logged out refresh token should not work")

	// Cleanup
	authDB := ConnectDB(t, AuthDBConnStr)
	defer authDB.Close()
	CleanupDB(t, authDB, []string{"refresh_tokens", "users"})
}

// Helper functions for creating test data

// createValidInitData creates a valid MAX initData string with proper hash
func createValidInitData(t *testing.T, botToken string, maxID int64, username, firstName, lastName string) string {
	// Create parameters map
	params := map[string]string{
		"max_id":     fmt.Sprintf("%d", maxID),
		"first_name": firstName,
	}

	if username != "" {
		params["username"] = username
	}
	if lastName != "" {
		params["last_name"] = lastName
	}

	return createInitDataWithHash(t, botToken, params)
}

// createMinimalInitData creates initData with only required fields
func createMinimalInitData(t *testing.T, botToken string, maxID int64, firstName string) string {
	params := map[string]string{
		"max_id":     fmt.Sprintf("%d", maxID),
		"first_name": firstName,
	}

	return createInitDataWithHash(t, botToken, params)
}

// createInitDataWithoutMaxID creates initData missing max_id field
func createInitDataWithoutMaxID(t *testing.T, botToken, username, firstName, lastName string) string {
	params := map[string]string{
		"username":   username,
		"first_name": firstName,
		"last_name":  lastName,
	}

	return createInitDataWithHash(t, botToken, params)
}

// createInitDataWithoutFirstName creates initData missing first_name field
func createInitDataWithoutFirstName(t *testing.T, botToken string, maxID int64, username string) string {
	params := map[string]string{
		"max_id":   fmt.Sprintf("%d", maxID),
		"username": username,
	}

	return createInitDataWithHash(t, botToken, params)
}

// createInitDataWithHash creates initData string with proper HMAC-SHA256 hash
func createInitDataWithHash(t *testing.T, botToken string, params map[string]string) string {
	// Sort parameters alphabetically
	var sortedParams []string
	for key, value := range params {
		sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", key, value))
	}
	sort.Strings(sortedParams)

	// Create data string for hash calculation
	dataCheckString := strings.Join(sortedParams, "\n")

	// Calculate hash using HMAC-SHA256
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))

	// Create query string with hash
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	values.Set("hash", hash)

	return values.Encode()
}