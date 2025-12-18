package test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"auth-service/internal/infrastructure/database"
	"auth-service/internal/infrastructure/hash"
	"auth-service/internal/infrastructure/jwt"
	"auth-service/internal/infrastructure/logger"
	"auth-service/internal/infrastructure/notification"
	"auth-service/internal/infrastructure/repository"
	"auth-service/internal/usecase"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	_ "github.com/lib/pq"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *database.DB {
	// Use test database connection
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=auth_test sslmode=disable"
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
	}
	
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping test database: %v", err)
	}
	
	return database.NewDBFromConnection(sqlDB, nil)
}

// cleanupTestData removes test data from database
func cleanupTestData(db *sql.DB) {
	db.Exec("DELETE FROM password_reset_tokens")
	db.Exec("DELETE FROM refresh_tokens")
	db.Exec("DELETE FROM user_roles")
	db.Exec("DELETE FROM users")
}

// createTestUser creates a test user in the database
func createTestUser(t *testing.T, db *sql.DB, phone string) int64 {
	hasher := hash.NewBcryptHasher()
	password, _ := hasher.Hash("TestPassword123!")
	
	var userID int64
	err := db.QueryRow(
		"INSERT INTO users (phone, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id",
		phone, "", password, "operator",
	).Scan(&userID)
	
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	return userID
}

// setupAuthService creates a configured AuthService for testing
func setupAuthService(db *database.DB) *usecase.AuthService {
	userRepo := repository.NewUserPostgres(db)
	refreshRepo := repository.NewRefreshPostgres(db)
	resetTokenRepo := repository.NewPasswordResetPostgres(db)
	hasher := hash.NewBcryptHasher()
	jwtManager := jwt.NewManager("test-secret", "test-secret", time.Hour, time.Hour*24)
	log := logger.NewDefault()
	notificationService := notification.NewMockNotificationService(log)
	
	authService := usecase.NewAuthService(userRepo, refreshRepo, hasher, jwtManager, nil)
	authService.SetPasswordResetRepository(resetTokenRepo)
	authService.SetNotificationService(notificationService)
	
	return authService
}

// TestProperty9_ResetTokenUniquenessAndExpiration tests that reset tokens are unique with 15min expiration
// **Feature: secure-password-management, Property 9: Reset token uniqueness and expiration**
// **Validates: Requirements 3.1, 3.2**
func TestProperty9_ResetTokenUniquenessAndExpiration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService := setupAuthService(db)
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	properties.Property("reset tokens are unique with 15min expiration", prop.ForAll(
		func(phoneNum int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number
			phone := "+7" + padNumber(phoneNum, 10)
			
			// Create test user
			createTestUser(t, db.GetUnderlyingDB(), phone)
			
			// Request password reset twice
			err1 := authService.RequestPasswordReset(phone)
			if err1 != nil {
				t.Logf("First reset request failed: %v", err1)
				return false
			}
			
			time.Sleep(10 * time.Millisecond) // Small delay to ensure different tokens
			
			err2 := authService.RequestPasswordReset(phone)
			if err2 != nil {
				t.Logf("Second reset request failed: %v", err2)
				return false
			}
			
			// Get all tokens for this user
			rows, err := db.Query("SELECT token, expires_at, created_at FROM password_reset_tokens ORDER BY created_at")
			if err != nil {
				t.Logf("Failed to query tokens: %v", err)
				return false
			}
			defer rows.Close()
			
			tokens := []struct {
				token     string
				expiresAt time.Time
				createdAt time.Time
			}{}
			
			for rows.Next() {
				var token string
				var expiresAt, createdAt time.Time
				if err := rows.Scan(&token, &expiresAt, &createdAt); err != nil {
					t.Logf("Failed to scan token: %v", err)
					return false
				}
				tokens = append(tokens, struct {
					token     string
					expiresAt time.Time
					createdAt time.Time
				}{token, expiresAt, createdAt})
			}
			
			if len(tokens) < 2 {
				t.Logf("Expected at least 2 tokens, got %d", len(tokens))
				return false
			}
			
			// Check tokens are different
			if tokens[0].token == tokens[1].token {
				t.Logf("Tokens are not unique: %s == %s", tokens[0].token, tokens[1].token)
				return false
			}
			
			// Check expiration is 15 minutes from creation
			for i, token := range tokens {
				expectedExpiry := token.createdAt.Add(15 * time.Minute)
				diff := token.expiresAt.Sub(expectedExpiry).Abs()
				
				// Allow 1 second tolerance for timing differences
				if diff > time.Second {
					t.Logf("Token %d expiration mismatch: expected %v, got %v (diff: %v)",
						i, expectedExpiry, token.expiresAt, diff)
					return false
				}
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
	))
	
	properties.TestingRun(t)
}

// padNumber pads a number with leading zeros to reach the desired length
func padNumber(num int, length int) string {
	s := ""
	for i := 0; i < length; i++ {
		s = string(rune('0'+(num%10))) + s
		num /= 10
	}
	return s
}

// TrackingNotificationService is a notification service that tracks calls for testing
type TrackingNotificationService struct {
	logger                      *logger.Logger
	resetTokenNotificationCalls []struct {
		phone string
		token string
	}
}

// NewTrackingNotificationService creates a new tracking notification service
func NewTrackingNotificationService(log *logger.Logger) *TrackingNotificationService {
	return &TrackingNotificationService{
		logger: log,
		resetTokenNotificationCalls: make([]struct {
			phone string
			token string
		}, 0),
	}
}

// SendPasswordNotification logs that a password notification would be sent
func (s *TrackingNotificationService) SendPasswordNotification(ctx context.Context, phone, password string) error {
	return nil
}

// SendResetTokenNotification tracks that a reset token notification was sent
func (s *TrackingNotificationService) SendResetTokenNotification(ctx context.Context, phone, token string) error {
	s.resetTokenNotificationCalls = append(s.resetTokenNotificationCalls, struct {
		phone string
		token string
	}{phone, token})
	return nil
}

// GetResetTokenNotificationCount returns the number of reset token notifications sent
func (s *TrackingNotificationService) GetResetTokenNotificationCount() int {
	return len(s.resetTokenNotificationCalls)
}

// GetLastResetTokenNotification returns the last reset token notification sent
func (s *TrackingNotificationService) GetLastResetTokenNotification() (phone string, token string, ok bool) {
	if len(s.resetTokenNotificationCalls) == 0 {
		return "", "", false
	}
	last := s.resetTokenNotificationCalls[len(s.resetTokenNotificationCalls)-1]
	return last.phone, last.token, true
}

// Reset clears all tracked calls
func (s *TrackingNotificationService) Reset() {
	s.resetTokenNotificationCalls = make([]struct {
		phone string
		token string
	}, 0)
}

// setupAuthServiceWithTracking creates a configured AuthService with tracking notification service
func setupAuthServiceWithTracking(db *database.DB) (*usecase.AuthService, *TrackingNotificationService) {
	userRepo := repository.NewUserPostgres(db)
	refreshRepo := repository.NewRefreshPostgres(db)
	resetTokenRepo := repository.NewPasswordResetPostgres(db)
	hasher := hash.NewBcryptHasher()
	jwtManager := jwt.NewManager("test-secret", "test-secret", time.Hour, time.Hour*24)
	log := logger.NewDefault()
	trackingService := NewTrackingNotificationService(log)
	
	authService := usecase.NewAuthService(userRepo, refreshRepo, hasher, jwtManager, nil)
	authService.SetPasswordResetRepository(resetTokenRepo)
	authService.SetNotificationService(trackingService)
	
	return authService, trackingService
}

// TestProperty10_ResetTokenDelivery tests that reset tokens are sent to the user's phone
// **Feature: secure-password-management, Property 10: Reset token delivery**
// **Validates: Requirements 3.3**
func TestProperty10_ResetTokenDelivery(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService, trackingService := setupAuthServiceWithTracking(db)
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	properties.Property("reset tokens are delivered to user's phone via notification service", prop.ForAll(
		func(phoneNum int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			trackingService.Reset()
			
			// Generate phone number
			phone := "+7" + padNumber(phoneNum, 10)
			
			// Create test user
			createTestUser(t, db.GetUnderlyingDB(), phone)
			
			// Request password reset
			err := authService.RequestPasswordReset(phone)
			if err != nil {
				t.Logf("Reset request failed: %v", err)
				return false
			}
			
			// Verify notification service was called
			notificationCount := trackingService.GetResetTokenNotificationCount()
			if notificationCount != 1 {
				t.Logf("Expected 1 notification call, got %d", notificationCount)
				return false
			}
			
			// Verify notification was sent to correct phone
			sentPhone, sentToken, ok := trackingService.GetLastResetTokenNotification()
			if !ok {
				t.Logf("No notification was sent")
				return false
			}
			
			if sentPhone != phone {
				t.Logf("Notification sent to wrong phone: expected %s, got %s", phone, sentPhone)
				return false
			}
			
			// Verify the token sent in notification matches the token in database
			var dbToken string
			err = db.QueryRow("SELECT token FROM password_reset_tokens ORDER BY created_at DESC LIMIT 1").Scan(&dbToken)
			if err != nil {
				t.Logf("Failed to query token from database: %v", err)
				return false
			}
			
			if sentToken != dbToken {
				t.Logf("Token in notification doesn't match database token")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
	))
	
	properties.TestingRun(t)
}

// TestProperty11_ValidResetTokenUpdatesPassword tests that valid reset tokens update passwords
// **Feature: secure-password-management, Property 11: Valid reset token updates password**
// **Validates: Requirements 3.4**
func TestProperty11_ValidResetTokenUpdatesPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService := setupAuthService(db)
	hasher := hash.NewBcryptHasher()
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	properties.Property("valid reset token updates password", prop.ForAll(
		func(phoneNum int, newPasswordSeed int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number and new password
			phone := "+7" + padNumber(phoneNum, 10)
			newPassword := "NewPass" + padNumber(newPasswordSeed, 6) + "!Aa"
			
			// Create test user
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			
			// Get original password hash
			var originalHash string
			err := db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&originalHash)
			if err != nil {
				t.Logf("Failed to get original password: %v", err)
				return false
			}
			
			// Request password reset
			err = authService.RequestPasswordReset(phone)
			if err != nil {
				t.Logf("Reset request failed: %v", err)
				return false
			}
			
			// Get the token from database
			var token string
			err = db.QueryRow("SELECT token FROM password_reset_tokens ORDER BY created_at DESC LIMIT 1").Scan(&token)
			if err != nil {
				t.Logf("Failed to get token: %v", err)
				return false
			}
			
			// Reset password with valid token
			err = authService.ResetPassword(token, newPassword)
			if err != nil {
				t.Logf("Reset password failed: %v", err)
				return false
			}
			
			// Get new password hash
			var newHash string
			err = db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&newHash)
			if err != nil {
				t.Logf("Failed to get new password: %v", err)
				return false
			}
			
			// Verify password was changed
			if originalHash == newHash {
				t.Logf("Password hash was not changed")
				return false
			}
			
			// Verify new password works
			if !hasher.Compare(newPassword, newHash) {
				t.Logf("New password does not match hash")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
		gen.IntRange(100000, 999999),         // Generate 6-digit password seeds
	))
	
	properties.TestingRun(t)
}

// TestProperty12_TokenInvalidationAfterUse tests that tokens cannot be reused
// **Feature: secure-password-management, Property 12: Token invalidation after use**
// **Validates: Requirements 3.5**
func TestProperty12_TokenInvalidationAfterUse(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService := setupAuthService(db)
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	properties.Property("tokens cannot be reused after use", prop.ForAll(
		func(phoneNum int, password1Seed int, password2Seed int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number and passwords
			phone := "+7" + padNumber(phoneNum, 10)
			newPassword1 := "NewPass" + padNumber(password1Seed, 6) + "!Aa"
			newPassword2 := "NewPass" + padNumber(password2Seed, 6) + "!Bb"
			
			// Create test user
			createTestUser(t, db.GetUnderlyingDB(), phone)
			
			// Request password reset
			err := authService.RequestPasswordReset(phone)
			if err != nil {
				t.Logf("Reset request failed: %v", err)
				return false
			}
			
			// Get the token from database
			var token string
			err = db.QueryRow("SELECT token FROM password_reset_tokens ORDER BY created_at DESC LIMIT 1").Scan(&token)
			if err != nil {
				t.Logf("Failed to get token: %v", err)
				return false
			}
			
			// Use token once - should succeed
			err = authService.ResetPassword(token, newPassword1)
			if err != nil {
				t.Logf("First reset password failed: %v", err)
				return false
			}
			
			// Try to use token again - should fail
			err = authService.ResetPassword(token, newPassword2)
			if err == nil {
				t.Logf("Second reset password should have failed but succeeded")
				return false
			}
			
			// Verify error is about token being used
			expectedErrors := []string{"used", "invalid", "expired"}
			errorStr := err.Error()
			foundExpectedError := false
			for _, expected := range expectedErrors {
				if contains(errorStr, expected) {
					foundExpectedError = true
					break
				}
			}
			
			if !foundExpectedError {
				t.Logf("Unexpected error message: %v", err)
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
		gen.IntRange(100000, 999999),         // Generate 6-digit password seeds
		gen.IntRange(100000, 999999),         // Generate 6-digit password seeds
	))
	
	properties.TestingRun(t)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// TestProperty13_InvalidTokenRejection tests that invalid or expired tokens are rejected
// **Feature: secure-password-management, Property 13: Invalid token rejection**
// **Validates: Requirements 3.6**
func TestProperty13_InvalidTokenRejection(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService := setupAuthService(db)
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	properties.Property("invalid or expired tokens are rejected", prop.ForAll(
		func(phoneNum int, invalidTokenSeed int, passwordSeed int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number, invalid token, and password
			phone := "+7" + padNumber(phoneNum, 10)
			invalidToken := "InvalidToken" + padNumber(invalidTokenSeed, 10)
			newPassword := "NewPass" + padNumber(passwordSeed, 6) + "!Aa"
			
			// Create test user
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			
			// Get original password hash
			var originalHash string
			err := db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&originalHash)
			if err != nil {
				t.Logf("Failed to get original password: %v", err)
				return false
			}
			
			// Try to reset password with invalid token - should fail
			err = authService.ResetPassword(invalidToken, newPassword)
			if err == nil {
				t.Logf("Reset password with invalid token should have failed but succeeded")
				return false
			}
			
			// Verify password was NOT changed
			var currentHash string
			err = db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&currentHash)
			if err != nil {
				t.Logf("Failed to get current password: %v", err)
				return false
			}
			
			if originalHash != currentHash {
				t.Logf("Password was changed despite invalid token")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit token seeds
		gen.IntRange(100000, 999999),         // Generate 6-digit password seeds
	))
	
	properties.TestingRun(t)
}

// TestProperty14_CurrentPasswordVerificationRequired tests that incorrect current password fails
// **Feature: secure-password-management, Property 14: Current password verification required**
// **Validates: Requirements 4.2**
func TestProperty14_CurrentPasswordVerificationRequired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService := setupAuthService(db)
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	properties.Property("password change fails with incorrect current password", prop.ForAll(
		func(phoneNum int, wrongPasswordSeed int, newPasswordSeed int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number and passwords
			phone := "+7" + padNumber(phoneNum, 10)
			wrongPassword := "WrongPass" + padNumber(wrongPasswordSeed, 6) + "!Aa"
			newPassword := "NewPass" + padNumber(newPasswordSeed, 6) + "!Bb"
			
			// Create test user with known password
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			
			// Get original password hash
			var originalHash string
			err := db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&originalHash)
			if err != nil {
				t.Logf("Failed to get original password: %v", err)
				return false
			}
			
			// Try to change password with wrong current password - should fail
			err = authService.ChangePassword(userID, wrongPassword, newPassword)
			if err == nil {
				t.Logf("Password change with wrong current password should have failed but succeeded")
				return false
			}
			
			// Verify password was NOT changed
			var currentHash string
			err = db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&currentHash)
			if err != nil {
				t.Logf("Failed to get current password: %v", err)
				return false
			}
			
			if originalHash != currentHash {
				t.Logf("Password was changed despite wrong current password")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
		gen.IntRange(100000, 999999),         // Generate 6-digit wrong password seeds
		gen.IntRange(100000, 999999),         // Generate 6-digit new password seeds
	))
	
	properties.TestingRun(t)
}

// TestProperty15_NewPasswordValidation tests that new passwords are validated
// **Feature: secure-password-management, Property 15: New password validation**
// **Validates: Requirements 4.3**
func TestProperty15_NewPasswordValidation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService := setupAuthService(db)
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	// Test 1: Password too short (less than 12 characters)
	properties.Property("password change rejects passwords shorter than 12 characters", prop.ForAll(
		func(phoneNum int, shortPasswordLength int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number and short password (less than 12 characters)
			phone := "+7" + padNumber(phoneNum, 10)
			// Generate password with length between 1 and 11 characters
			shortPassword := "Pass" + padNumber(shortPasswordLength, shortPasswordLength%8+1)
			
			// Create test user with known password
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			correctPassword := "TestPassword123!"
			
			// Get original password hash
			var originalHash string
			err := db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&originalHash)
			if err != nil {
				t.Logf("Failed to get original password: %v", err)
				return false
			}
			
			// Try to change password with short password - should fail
			err = authService.ChangePassword(userID, correctPassword, shortPassword)
			if err == nil {
				t.Logf("Password change with short password should have failed but succeeded")
				return false
			}
			
			// Verify error message mentions password requirements
			errorStr := err.Error()
			if !contains(errorStr, "12") && !contains(errorStr, "character") && !contains(errorStr, "length") {
				t.Logf("Error message should mention password length requirement: %v", err)
				return false
			}
			
			// Verify password was NOT changed
			var currentHash string
			err = db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&currentHash)
			if err != nil {
				t.Logf("Failed to get current password: %v", err)
				return false
			}
			
			if originalHash != currentHash {
				t.Logf("Password was changed despite invalid new password")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
		gen.IntRange(1, 11),                  // Generate short password lengths (1-11)
	))
	
	// Test 2: Password missing uppercase letter
	properties.Property("password change rejects passwords without uppercase letters", prop.ForAll(
		func(phoneNum int, passwordSeed int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number and password without uppercase
			phone := "+7" + padNumber(phoneNum, 10)
			noUppercasePassword := "newpassword" + padNumber(passwordSeed, 4) + "!1"
			
			// Create test user with known password
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			correctPassword := "TestPassword123!"
			
			// Get original password hash
			var originalHash string
			err := db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&originalHash)
			if err != nil {
				t.Logf("Failed to get original password: %v", err)
				return false
			}
			
			// Try to change password without uppercase - should fail
			err = authService.ChangePassword(userID, correctPassword, noUppercasePassword)
			if err == nil {
				t.Logf("Password change without uppercase should have failed but succeeded")
				return false
			}
			
			// Verify password was NOT changed
			var currentHash string
			err = db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&currentHash)
			if err != nil {
				t.Logf("Failed to get current password: %v", err)
				return false
			}
			
			if originalHash != currentHash {
				t.Logf("Password was changed despite missing uppercase")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
		gen.IntRange(1000, 9999),             // Generate 4-digit password seeds
	))
	
	// Test 3: Password missing lowercase letter
	properties.Property("password change rejects passwords without lowercase letters", prop.ForAll(
		func(phoneNum int, passwordSeed int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number and password without lowercase
			phone := "+7" + padNumber(phoneNum, 10)
			noLowercasePassword := "NEWPASSWORD" + padNumber(passwordSeed, 4) + "!1"
			
			// Create test user with known password
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			correctPassword := "TestPassword123!"
			
			// Get original password hash
			var originalHash string
			err := db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&originalHash)
			if err != nil {
				t.Logf("Failed to get original password: %v", err)
				return false
			}
			
			// Try to change password without lowercase - should fail
			err = authService.ChangePassword(userID, correctPassword, noLowercasePassword)
			if err == nil {
				t.Logf("Password change without lowercase should have failed but succeeded")
				return false
			}
			
			// Verify password was NOT changed
			var currentHash string
			err = db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&currentHash)
			if err != nil {
				t.Logf("Failed to get current password: %v", err)
				return false
			}
			
			if originalHash != currentHash {
				t.Logf("Password was changed despite missing lowercase")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
		gen.IntRange(1000, 9999),             // Generate 4-digit password seeds
	))
	
	// Test 4: Password missing digit
	properties.Property("password change rejects passwords without digits", prop.ForAll(
		func(phoneNum int, passwordSeed int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number and password without digits
			phone := "+7" + padNumber(phoneNum, 10)
			noDigitPassword := "NewPassword!@#$"
			
			// Create test user with known password
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			correctPassword := "TestPassword123!"
			
			// Get original password hash
			var originalHash string
			err := db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&originalHash)
			if err != nil {
				t.Logf("Failed to get original password: %v", err)
				return false
			}
			
			// Try to change password without digits - should fail
			err = authService.ChangePassword(userID, correctPassword, noDigitPassword)
			if err == nil {
				t.Logf("Password change without digits should have failed but succeeded")
				return false
			}
			
			// Verify password was NOT changed
			var currentHash string
			err = db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&currentHash)
			if err != nil {
				t.Logf("Failed to get current password: %v", err)
				return false
			}
			
			if originalHash != currentHash {
				t.Logf("Password was changed despite missing digits")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
		gen.IntRange(1000, 9999),             // Generate 4-digit password seeds (unused but kept for consistency)
	))
	
	// Test 5: Password missing special character
	properties.Property("password change rejects passwords without special characters", prop.ForAll(
		func(phoneNum int, passwordSeed int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number and password without special characters
			phone := "+7" + padNumber(phoneNum, 10)
			noSpecialPassword := "NewPassword" + padNumber(passwordSeed, 4)
			
			// Create test user with known password
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			correctPassword := "TestPassword123!"
			
			// Get original password hash
			var originalHash string
			err := db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&originalHash)
			if err != nil {
				t.Logf("Failed to get original password: %v", err)
				return false
			}
			
			// Try to change password without special characters - should fail
			err = authService.ChangePassword(userID, correctPassword, noSpecialPassword)
			if err == nil {
				t.Logf("Password change without special characters should have failed but succeeded")
				return false
			}
			
			// Verify password was NOT changed
			var currentHash string
			err = db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&currentHash)
			if err != nil {
				t.Logf("Failed to get current password: %v", err)
				return false
			}
			
			if originalHash != currentHash {
				t.Logf("Password was changed despite missing special characters")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
		gen.IntRange(1000, 9999),             // Generate 4-digit password seeds
	))
	
	properties.TestingRun(t)
}

// TestProperty4_BcryptHashing tests that passwords are hashed with bcrypt
// **Feature: secure-password-management, Property 4: Bcrypt hashing**
// **Validates: Requirements 1.3, 4.4**
func TestProperty4_BcryptHashing(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService := setupAuthService(db)
	hasher := hash.NewBcryptHasher()
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	properties.Property("passwords are hashed with bcrypt and verifiable", prop.ForAll(
		func(phoneNum int, newPasswordSeed int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number and new password
			phone := "+7" + padNumber(phoneNum, 10)
			newPassword := "NewPass" + padNumber(newPasswordSeed, 6) + "!Aa"
			correctPassword := "TestPassword123!"
			
			// Create test user
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			
			// Change password
			err := authService.ChangePassword(userID, correctPassword, newPassword)
			if err != nil {
				t.Logf("Password change failed: %v", err)
				return false
			}
			
			// Get stored password hash
			var storedHash string
			err = db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&storedHash)
			if err != nil {
				t.Logf("Failed to get password hash: %v", err)
				return false
			}
			
			// Verify it's a bcrypt hash (starts with $2a$, $2b$, or $2y$)
			if len(storedHash) < 4 || storedHash[0] != '$' || storedHash[1] != '2' {
				t.Logf("Password hash doesn't appear to be bcrypt: %s", storedHash[:min(10, len(storedHash))])
				return false
			}
			
			// Verify original password matches the hash
			if !hasher.Compare(newPassword, storedHash) {
				t.Logf("New password doesn't verify against stored hash")
				return false
			}
			
			// Verify wrong password doesn't match
			wrongPassword := "WrongPassword123!"
			if hasher.Compare(wrongPassword, storedHash) {
				t.Logf("Wrong password incorrectly verified against hash")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
		gen.IntRange(100000, 999999),         // Generate 6-digit password seeds
	))
	
	properties.TestingRun(t)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestProperty16_RefreshTokenInvalidationOnPasswordChange tests that refresh tokens are invalidated
// **Feature: secure-password-management, Property 16: Refresh token invalidation on password change**
// **Validates: Requirements 4.5**
func TestProperty16_RefreshTokenInvalidationOnPasswordChange(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService := setupAuthService(db)
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	properties.Property("refresh tokens are invalidated on password change", prop.ForAll(
		func(phoneNum int, newPasswordSeed int) bool {
			// Clean up before each test
			cleanupTestData(db.GetUnderlyingDB())
			
			// Generate phone number and new password
			phone := "+7" + padNumber(phoneNum, 10)
			newPassword := "NewPass" + padNumber(newPasswordSeed, 6) + "!Aa"
			correctPassword := "TestPassword123!"
			
			// Create test user
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			
			// Create some refresh tokens for the user
			token1JTI := "jti-" + padNumber(phoneNum, 10) + "-1"
			token2JTI := "jti-" + padNumber(phoneNum, 10) + "-2"
			expiresAt := time.Now().Add(24 * time.Hour)
			
			_, err := db.Exec(
				"INSERT INTO refresh_tokens (jti, user_id, expires_at, revoked) VALUES ($1, $2, $3, false)",
				token1JTI, userID, expiresAt,
			)
			if err != nil {
				t.Logf("Failed to create refresh token 1: %v", err)
				return false
			}
			
			_, err = db.Exec(
				"INSERT INTO refresh_tokens (jti, user_id, expires_at, revoked) VALUES ($1, $2, $3, false)",
				token2JTI, userID, expiresAt,
			)
			if err != nil {
				t.Logf("Failed to create refresh token 2: %v", err)
				return false
			}
			
			// Verify tokens are not revoked initially
			var revokedCount int
			err = db.QueryRow(
				"SELECT COUNT(*) FROM refresh_tokens WHERE user_id=$1 AND revoked=true",
				userID,
			).Scan(&revokedCount)
			if err != nil {
				t.Logf("Failed to count revoked tokens: %v", err)
				return false
			}
			
			if revokedCount != 0 {
				t.Logf("Expected 0 revoked tokens initially, got %d", revokedCount)
				return false
			}
			
			// Change password
			err = authService.ChangePassword(userID, correctPassword, newPassword)
			if err != nil {
				t.Logf("Password change failed: %v", err)
				return false
			}
			
			// Verify all tokens are now revoked
			err = db.QueryRow(
				"SELECT COUNT(*) FROM refresh_tokens WHERE user_id=$1 AND revoked=true",
				userID,
			).Scan(&revokedCount)
			if err != nil {
				t.Logf("Failed to count revoked tokens after password change: %v", err)
				return false
			}
			
			if revokedCount != 2 {
				t.Logf("Expected 2 revoked tokens after password change, got %d", revokedCount)
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999), // Generate 10-digit phone numbers
		gen.IntRange(100000, 999999),         // Generate 6-digit password seeds
	))
	
	properties.TestingRun(t)
}
