package test

import (
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/max"
	"auth-service/internal/infrastructure/repository"
	"auth-service/internal/usecase"
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	_ "github.com/lib/pq"
)

// TestProperty3_UserLookupAndCreationBehavior tests user lookup and creation behavior
// **Feature: max-miniapp-auth, Property 3: User lookup and creation behavior**
// **Validates: Requirements 1.3, 1.4**
func TestProperty3_UserLookupAndCreationBehavior(t *testing.T) {
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

	// Test 1: If user exists with max_id, lookup should return existing user
	properties.Property("if user exists with max_id, lookup should return existing user", prop.ForAll(
		func(phoneNum int, maxID int64, usernameSeed int, nameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			phone := "+7" + padNumber(phoneNum, 10)
			username := "user" + padNumber(usernameSeed, 6)
			name := "Name" + padNumber(nameSeed, 6)

			// Create existing user with MAX ID
			existingUser := &domain.User{
				Phone:    phone,
				Email:    "",
				Password: "hashedpassword",
				Role:     domain.RoleOperator,
				MaxID:    &maxID,
				Username: &username,
				Name:     &name,
			}

			err := userRepo.Create(existingUser)
			if err != nil {
				t.Logf("Failed to create existing user: %v", err)
				return false
			}

			// Create valid initData for this user
			initData := createValidInitDataForUserLookup(maxID, username, "FirstName", "LastName", "test_bot_token")

			// Authenticate using MAX
			result, err := authService.AuthenticateMAX(initData)
			if err != nil {
				t.Logf("Authentication failed: %v", err)
				return false
			}

			// Verify tokens were generated
			if result.AccessToken == "" || result.RefreshToken == "" {
				t.Logf("Tokens not generated properly")
				return false
			}

			// Verify the existing user was found and updated (not a new user created)
			updatedUser, err := userRepo.GetByMaxID(maxID)
			if err != nil {
				t.Logf("Failed to retrieve user after authentication: %v", err)
				return false
			}

			// Should be the same user ID (not a new user)
			if updatedUser.ID != existingUser.ID {
				t.Logf("New user created instead of using existing: expected ID %d, got %d", existingUser.ID, updatedUser.ID)
				return false
			}

			// User data should be updated with current MAX data
			expectedName := "FirstName LastName"
			if updatedUser.Name == nil || *updatedUser.Name != expectedName {
				t.Logf("User name not updated: expected %s, got %v", expectedName, updatedUser.Name)
				return false
			}

			return true
		},
		gen.IntRange(1000000000, 9999999999), // phone number
		gen.Int64Range(1, 999999999),         // max_id
		gen.IntRange(100000, 999999),         // username seed
		gen.IntRange(100000, 999999),         // name seed
	))

	// Test 2: If user doesn't exist with max_id, new user should be created
	properties.Property("if user doesn't exist with max_id, new user should be created", prop.ForAll(
		func(maxID int64, usernameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			username := "user" + padNumber(usernameSeed, 6)

			// Ensure no user exists with this MAX ID
			_, err := userRepo.GetByMaxID(maxID)
			if err == nil {
				// User already exists, skip this test case
				return true
			}

			// Create valid initData for new user
			initData := createValidInitDataForUserLookup(maxID, username, "NewFirstName", "NewLastName", "test_bot_token")

			// Authenticate using MAX
			result, err := authService.AuthenticateMAX(initData)
			if err != nil {
				t.Logf("Authentication failed: %v", err)
				return false
			}

			// Verify tokens were generated
			if result.AccessToken == "" || result.RefreshToken == "" {
				t.Logf("Tokens not generated properly")
				return false
			}

			// Verify new user was created
			newUser, err := userRepo.GetByMaxID(maxID)
			if err != nil {
				t.Logf("Failed to retrieve newly created user: %v", err)
				return false
			}

			// Verify user data is correct
			if newUser.MaxID == nil || *newUser.MaxID != maxID {
				t.Logf("MaxID not set correctly: expected %d, got %v", maxID, newUser.MaxID)
				return false
			}

			if newUser.Username == nil || *newUser.Username != username {
				t.Logf("Username not set correctly: expected %s, got %v", username, newUser.Username)
				return false
			}

			expectedName := "NewFirstName NewLastName"
			if newUser.Name == nil || *newUser.Name != expectedName {
				t.Logf("Name not set correctly: expected %s, got %v", expectedName, newUser.Name)
				return false
			}

			if newUser.Role != domain.RoleOperator {
				t.Logf("Default role not set correctly: expected %s, got %s", domain.RoleOperator, newUser.Role)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),  // max_id
		gen.IntRange(100000, 999999),  // username seed
	))

	// Test 3: User creation with missing optional fields (username) should work
	properties.Property("user creation with missing optional fields should work", prop.ForAll(
		func(maxID int64) bool {
			cleanupTestData(db.GetUnderlyingDB())

			// Create valid initData without username (optional field)
			initData := createValidInitDataWithoutUsernameForUserLookup(maxID, "FirstName", "LastName", "test_bot_token")

			// Authenticate using MAX
			result, err := authService.AuthenticateMAX(initData)
			if err != nil {
				t.Logf("Authentication failed: %v", err)
				return false
			}

			// Verify tokens were generated
			if result.AccessToken == "" || result.RefreshToken == "" {
				t.Logf("Tokens not generated properly")
				return false
			}

			// Verify new user was created with empty username
			newUser, err := userRepo.GetByMaxID(maxID)
			if err != nil {
				t.Logf("Failed to retrieve newly created user: %v", err)
				return false
			}

			// Username should be empty (optional field)
			if newUser.Username == nil || *newUser.Username != "" {
				t.Logf("Username should be empty for missing optional field, got %v", newUser.Username)
				return false
			}

			// Name should still be set correctly
			expectedName := "FirstName LastName"
			if newUser.Name == nil || *newUser.Name != expectedName {
				t.Logf("Name not set correctly: expected %s, got %v", expectedName, newUser.Name)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999), // max_id
	))

	// Test 4: User creation with only first name should work
	properties.Property("user creation with only first name should work", prop.ForAll(
		func(maxID int64) bool {
			cleanupTestData(db.GetUnderlyingDB())

			// Create valid initData with only first name (last name is optional)
			initData := createValidInitDataWithoutLastNameForUserLookup(maxID, "user123", "OnlyFirstName", "test_bot_token")

			// Authenticate using MAX
			result, err := authService.AuthenticateMAX(initData)
			if err != nil {
				t.Logf("Authentication failed: %v", err)
				return false
			}

			// Verify tokens were generated
			if result.AccessToken == "" || result.RefreshToken == "" {
				t.Logf("Tokens not generated properly")
				return false
			}

			// Verify new user was created with only first name
			newUser, err := userRepo.GetByMaxID(maxID)
			if err != nil {
				t.Logf("Failed to retrieve newly created user: %v", err)
				return false
			}

			// Name should be just the first name (no last name)
			expectedName := "OnlyFirstName"
			if newUser.Name == nil || *newUser.Name != expectedName {
				t.Logf("Name not set correctly: expected %s, got %v", expectedName, newUser.Name)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999), // max_id
	))

	properties.TestingRun(t)
}

// Helper function to create valid initData with all fields (specific to this test)
func createValidInitDataForUserLookup(maxID int64, username, firstName, lastName, botToken string) string {
	params := fmt.Sprintf("max_id=%d&username=%s&first_name=%s&last_name=%s",
		maxID, username, firstName, lastName)
	return createInitDataWithCorrectHash(params, botToken)
}

// Helper function to create valid initData without username (specific to this test)
func createValidInitDataWithoutUsernameForUserLookup(maxID int64, firstName, lastName, botToken string) string {
	params := fmt.Sprintf("max_id=%d&first_name=%s&last_name=%s",
		maxID, firstName, lastName)
	return createInitDataWithCorrectHash(params, botToken)
}

// Helper function to create valid initData without last name (specific to this test)
func createValidInitDataWithoutLastNameForUserLookup(maxID int64, username, firstName, botToken string) string {
	params := fmt.Sprintf("max_id=%d&username=%s&first_name=%s",
		maxID, username, firstName)
	return createInitDataWithCorrectHash(params, botToken)
}

// Mock implementations for testing
type mockHasher struct{}

func (m *mockHasher) Hash(s string) (string, error) {
	return "hashed_" + s, nil
}

func (m *mockHasher) Compare(s, hashed string) bool {
	return "hashed_"+s == hashed
}

type mockJWTManager struct{}

func (m *mockJWTManager) GenerateTokens(userID int64, identifier, role string) (*domain.TokensWithJTI, error) {
	return &domain.TokensWithJTI{
		TokenPair: domain.TokenPair{
			AccessToken:  fmt.Sprintf("access_token_%d_%s", userID, identifier),
			RefreshToken: fmt.Sprintf("refresh_token_%d_%s", userID, identifier),
		},
		RefreshJTI: fmt.Sprintf("jti_%d_%s", userID, identifier),
	}, nil
}

func (m *mockJWTManager) GenerateTokensWithContext(userID int64, identifier, role string, ctx *domain.TokenContext) (*domain.TokensWithJTI, error) {
	return m.GenerateTokens(userID, identifier, role)
}

func (m *mockJWTManager) VerifyAccessToken(token string) (int64, string, string, error) {
	return 1, "test@example.com", "operator", nil
}

func (m *mockJWTManager) VerifyAccessTokenWithContext(token string) (int64, string, string, *domain.TokenContext, error) {
	userID, identifier, role, err := m.VerifyAccessToken(token)
	return userID, identifier, role, nil, err
}

func (m *mockJWTManager) VerifyRefreshToken(token string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"jti": "test_jti",
		"sub": "1",
	}, nil
}

func (m *mockJWTManager) RefreshTTL() time.Duration {
	return 24 * time.Hour
}