package test

import (
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/max"
	"auth-service/internal/infrastructure/repository"
	"auth-service/internal/usecase"
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	_ "github.com/lib/pq"
)

// TestProperty5_TokenGenerationReliability tests that JWT tokens are always generated for successful authentication
// **Feature: max-miniapp-auth, Property 5: Token generation reliability**
// **Validates: Requirements 1.5**
func TestProperty5_TokenGenerationReliability(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())

	userRepo := repository.NewUserPostgres(db)
	refreshRepo := repository.NewRefreshPostgres(db)
	hasher := &mockHasher{}
	jwtManager := &mockJWTManager{}
	maxAuthValidator := max.NewAuthValidator()

	authService := usecase.NewAuthService(userRepo, refreshRepo, hasher, jwtManager, nil)
	authService.SetMaxAuthValidator(maxAuthValidator)
	authService.SetMaxBotToken("test_bot_token")

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Test 1: Successful authentication always generates both access and refresh tokens
	properties.Property("successful authentication always generates both access and refresh tokens", prop.ForAll(
		func(maxID int64, usernameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			username := "user" + padNumber(usernameSeed, 6)

			// Create valid initData for authentication
			initData := createValidInitDataForTokenGeneration(maxID, username, "FirstName", "LastName", "test_bot_token")

			// Authenticate using MAX
			result, err := authService.AuthenticateMAX(initData)
			if err != nil {
				t.Logf("Authentication failed: %v", err)
				return false
			}

			// Verify both tokens are generated and non-empty
			if result.AccessToken == "" {
				t.Logf("Access token is empty")
				return false
			}

			if result.RefreshToken == "" {
				t.Logf("Refresh token is empty")
				return false
			}

			if result.RefreshJTI == "" {
				t.Logf("Refresh JTI is empty")
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),  // max_id
		gen.IntRange(100000, 999999),  // username seed
	))

	// Test 2: Token generation is consistent for the same user
	properties.Property("token generation is consistent for the same user", prop.ForAll(
		func(maxID int64, usernameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			username := "user" + padNumber(usernameSeed, 6)

			// Create valid initData for authentication
			initData := createValidInitDataForTokenGeneration(maxID, username, "FirstName", "LastName", "test_bot_token")

			// Authenticate multiple times with the same user
			result1, err := authService.AuthenticateMAX(initData)
			if err != nil {
				t.Logf("First authentication failed: %v", err)
				return false
			}

			result2, err := authService.AuthenticateMAX(initData)
			if err != nil {
				t.Logf("Second authentication failed: %v", err)
				return false
			}

			// Tokens should be generated both times (though they may be different)
			if result1.AccessToken == "" || result1.RefreshToken == "" || result1.RefreshJTI == "" {
				t.Logf("First authentication didn't generate all tokens")
				return false
			}

			if result2.AccessToken == "" || result2.RefreshToken == "" || result2.RefreshJTI == "" {
				t.Logf("Second authentication didn't generate all tokens")
				return false
			}

			// JTIs should be different (new tokens each time)
			if result1.RefreshJTI == result2.RefreshJTI {
				t.Logf("Refresh JTIs should be different for new authentications")
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),  // max_id
		gen.IntRange(100000, 999999),  // username seed
	))

	// Test 3: Token generation works for new users (user creation scenario)
	properties.Property("token generation works for new users", prop.ForAll(
		func(maxID int64, usernameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			username := "newuser" + padNumber(usernameSeed, 6)

			// Ensure no user exists with this MAX ID
			_, err := userRepo.GetByMaxID(maxID)
			if err == nil {
				// User already exists, skip this test case
				return true
			}

			// Create valid initData for new user
			initData := createValidInitDataForTokenGeneration(maxID, username, "NewFirstName", "NewLastName", "test_bot_token")

			// Authenticate (should create new user and generate tokens)
			result, err := authService.AuthenticateMAX(initData)
			if err != nil {
				t.Logf("Authentication for new user failed: %v", err)
				return false
			}

			// Verify tokens are generated for new user
			if result.AccessToken == "" {
				t.Logf("Access token not generated for new user")
				return false
			}

			if result.RefreshToken == "" {
				t.Logf("Refresh token not generated for new user")
				return false
			}

			if result.RefreshJTI == "" {
				t.Logf("Refresh JTI not generated for new user")
				return false
			}

			// Verify user was actually created
			newUser, err := userRepo.GetByMaxID(maxID)
			if err != nil {
				t.Logf("New user was not created in database: %v", err)
				return false
			}

			if newUser.MaxID == nil || *newUser.MaxID != maxID {
				t.Logf("New user has incorrect MAX ID")
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),  // max_id
		gen.IntRange(100000, 999999),  // username seed
	))

	// Test 4: Token generation works for existing users (user update scenario)
	properties.Property("token generation works for existing users", prop.ForAll(
		func(phoneNum int, maxID int64, usernameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			phone := "+7" + padNumber(phoneNum, 10)
			username := "existinguser" + padNumber(usernameSeed, 6)

			// Create existing user with MAX ID
			oldUsername := "oldusername"
			oldName := "OldName"
			existingUser := &domain.User{
				Phone:    phone,
				Email:    "",
				Password: "hashedpassword",
				Role:     domain.RoleOperator,
				MaxID:    &maxID,
				Username: &oldUsername,
				Name:     &oldName,
			}

			err := userRepo.Create(existingUser)
			if err != nil {
				t.Logf("Failed to create existing user: %v", err)
				return false
			}

			// Create valid initData for existing user with updated info
			initData := createValidInitDataForTokenGeneration(maxID, username, "UpdatedFirstName", "UpdatedLastName", "test_bot_token")

			// Authenticate (should update existing user and generate tokens)
			result, err := authService.AuthenticateMAX(initData)
			if err != nil {
				t.Logf("Authentication for existing user failed: %v", err)
				return false
			}

			// Verify tokens are generated for existing user
			if result.AccessToken == "" {
				t.Logf("Access token not generated for existing user")
				return false
			}

			if result.RefreshToken == "" {
				t.Logf("Refresh token not generated for existing user")
				return false
			}

			if result.RefreshJTI == "" {
				t.Logf("Refresh JTI not generated for existing user")
				return false
			}

			// Verify user was updated (not a new user created)
			updatedUser, err := userRepo.GetByMaxID(maxID)
			if err != nil {
				t.Logf("Failed to retrieve updated user: %v", err)
				return false
			}

			if updatedUser.ID != existingUser.ID {
				t.Logf("New user created instead of updating existing user")
				return false
			}

			// Verify user data was updated
			expectedName := "UpdatedFirstName UpdatedLastName"
			if updatedUser.Name == nil || *updatedUser.Name != expectedName {
				t.Logf("User name not updated: expected %s, got %v", expectedName, updatedUser.Name)
				return false
			}

			return true
		},
		gen.IntRange(1000000000, 9999999999), // phone number
		gen.Int64Range(1, 999999999),         // max_id
		gen.IntRange(100000, 999999),         // username seed
	))

	// Test 5: Token generation fails gracefully with invalid initData
	properties.Property("token generation fails gracefully with invalid initData", prop.ForAll(
		func(maxID int64, usernameSeed int, invalidHashSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			username := "user" + padNumber(usernameSeed, 6)
			invalidHash := "invalid_hash_" + padNumber(invalidHashSeed, 10)

			// Create initData with invalid hash
			invalidInitData := fmt.Sprintf("max_id=%d&username=%s&first_name=FirstName&last_name=LastName&hash=%s",
				maxID, username, invalidHash)

			// Authentication should fail
			result, err := authService.AuthenticateMAX(invalidInitData)
			if err == nil {
				t.Logf("Authentication should have failed with invalid initData")
				return false
			}

			// No tokens should be returned
			if result != nil {
				t.Logf("Result should be nil for failed authentication")
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),         // max_id
		gen.IntRange(100000, 999999),         // username seed
		gen.IntRange(1000000000, 9999999999), // invalid hash seed
	))

	// Test 6: Token generation includes refresh token JTI in database
	properties.Property("token generation includes refresh token JTI in database", prop.ForAll(
		func(maxID int64, usernameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			username := "user" + padNumber(usernameSeed, 6)

			// Create valid initData for authentication
			initData := createValidInitDataForTokenGeneration(maxID, username, "FirstName", "LastName", "test_bot_token")

			// Authenticate using MAX
			result, err := authService.AuthenticateMAX(initData)
			if err != nil {
				t.Logf("Authentication failed: %v", err)
				return false
			}

			// Verify refresh JTI is stored in database
			var storedJTI string
			err = db.QueryRow("SELECT jti FROM refresh_tokens ORDER BY created_at DESC LIMIT 1").Scan(&storedJTI)
			if err != nil {
				t.Logf("Failed to retrieve refresh JTI from database: %v", err)
				return false
			}

			if storedJTI != result.RefreshJTI {
				t.Logf("Stored JTI doesn't match returned JTI: expected %s, got %s", result.RefreshJTI, storedJTI)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),  // max_id
		gen.IntRange(100000, 999999),  // username seed
	))

	properties.TestingRun(t)
}

// Helper function to create valid initData for token generation tests
func createValidInitDataForTokenGeneration(maxID int64, username, firstName, lastName, botToken string) string {
	params := fmt.Sprintf("max_id=%d&username=%s&first_name=%s&last_name=%s",
		maxID, username, firstName, lastName)
	return createInitDataWithCorrectHash(params, botToken)
}