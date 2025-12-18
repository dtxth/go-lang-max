package test

import (
	"auth-service/internal/infrastructure/database"
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

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


// CaptureLogger captures log output for testing
type CaptureLogger struct {
	buffer *bytes.Buffer
	logs   []LogEntry
}

// LogEntry represents a captured log entry
type LogEntry struct {
	Message string
	Fields  map[string]interface{}
}

// NewCaptureLogger creates a new capture logger
func NewCaptureLogger() *CaptureLogger {
	return &CaptureLogger{
		buffer: &bytes.Buffer{},
		logs:   make([]LogEntry, 0),
	}
}

// Info captures an info log
func (l *CaptureLogger) Info(ctx context.Context, message string, fields map[string]interface{}) {
	l.logs = append(l.logs, LogEntry{
		Message: message,
		Fields:  fields,
	})
}

// Error captures an error log
func (l *CaptureLogger) Error(ctx context.Context, message string, fields map[string]interface{}) {
	l.logs = append(l.logs, LogEntry{
		Message: message,
		Fields:  fields,
	})
}

// GetLogs returns all captured logs
func (l *CaptureLogger) GetLogs() []LogEntry {
	return l.logs
}

// Reset clears all captured logs
func (l *CaptureLogger) Reset() {
	l.logs = make([]LogEntry, 0)
	l.buffer.Reset()
}

// ContainsLog checks if a log with the given message exists
func (l *CaptureLogger) ContainsLog(message string) bool {
	for _, log := range l.logs {
		if log.Message == message {
			return true
		}
	}
	return false
}

// GetLogByMessage returns the first log with the given message
func (l *CaptureLogger) GetLogByMessage(message string) (LogEntry, bool) {
	for _, log := range l.logs {
		if log.Message == message {
			return log, true
		}
	}
	return LogEntry{}, false
}

// setupAuthServiceWithLogger creates a configured AuthService with capture logger
func setupAuthServiceWithLogger(db *database.DB) (*usecase.AuthService, *CaptureLogger) {
	userRepo := repository.NewUserPostgres(db)
	refreshRepo := repository.NewRefreshPostgres(db)
	resetTokenRepo := repository.NewPasswordResetPostgres(db)
	hasher := hash.NewBcryptHasher()
	jwtManager := jwt.NewManager("test-secret", "test-secret", time.Hour, time.Hour*24)
	captureLogger := NewCaptureLogger()
	
	// Create a logger for the mock notification service
	mockLogger := logger.NewDefault()
	notificationService := notification.NewMockNotificationService(mockLogger)
	
	authService := usecase.NewAuthService(userRepo, refreshRepo, hasher, jwtManager, nil)
	authService.SetPasswordResetRepository(resetTokenRepo)
	authService.SetNotificationService(notificationService)
	authService.SetLogger(captureLogger)
	
	return authService, captureLogger
}

// TestProperty17_ComprehensiveAuditLogging tests that all password operations are logged
// **Feature: secure-password-management, Property 17: Comprehensive audit logging**
// **Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5**
func TestProperty17_ComprehensiveAuditLogging(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService, captureLogger := setupAuthServiceWithLogger(db)
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	// Test 1: User creation is logged
	properties.Property("user creation is logged with user_id and timestamp", prop.ForAll(
		func(phoneNum int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			phone := "+7" + padNumber(phoneNum, 10)
			password := "TestPassword123!"
			
			userID, err := authService.CreateUser(phone, password)
			if err != nil {
				t.Logf("CreateUser failed: %v", err)
				return false
			}
			
			// Check that user_created log exists
			if !captureLogger.ContainsLog("user_created") {
				t.Logf("Expected user_created log not found")
				return false
			}
			
			log, ok := captureLogger.GetLogByMessage("user_created")
			if !ok {
				return false
			}
			
			// Verify log contains user_id
			if log.Fields["user_id"] == nil {
				t.Logf("Log missing user_id field")
				return false
			}
			
			logUserID, ok := log.Fields["user_id"].(int64)
			if !ok {
				t.Logf("user_id field is not int64")
				return false
			}
			
			if logUserID != userID {
				t.Logf("user_id mismatch: expected %d, got %d", userID, logUserID)
				return false
			}
			
			// Verify log contains timestamp
			if log.Fields["timestamp"] == nil {
				t.Logf("Log missing timestamp field")
				return false
			}
			
			// Verify log does NOT contain password or hash
			logJSON, _ := json.Marshal(log)
			logStr := string(logJSON)
			if containsSubstring(logStr, password) {
				t.Logf("Log contains plaintext password")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999),
	))

	// Test 2: Password reset request is logged
	properties.Property("password reset request is logged with user_id and timestamp", prop.ForAll(
		func(phoneNum int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			phone := "+7" + padNumber(phoneNum, 10)
			createTestUser(t, db.GetUnderlyingDB(), phone)
			
			err := authService.RequestPasswordReset(phone)
			if err != nil {
				t.Logf("RequestPasswordReset failed: %v", err)
				return false
			}
			
			// Check that password_reset_requested log exists
			if !captureLogger.ContainsLog("password_reset_requested") {
				t.Logf("Expected password_reset_requested log not found")
				return false
			}
			
			log, ok := captureLogger.GetLogByMessage("password_reset_requested")
			if !ok {
				return false
			}
			
			// Verify log contains user_id and timestamp
			if log.Fields["user_id"] == nil || log.Fields["timestamp"] == nil {
				t.Logf("Log missing required fields")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999),
	))
	
	// Test 3: Password reset completion is logged
	properties.Property("password reset completion is logged with user_id and timestamp", prop.ForAll(
		func(phoneNum int, passwordSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			phone := "+7" + padNumber(phoneNum, 10)
			newPassword := "NewPass" + padNumber(passwordSeed, 6) + "!Aa"
			createTestUser(t, db.GetUnderlyingDB(), phone)
			
			// Request reset
			err := authService.RequestPasswordReset(phone)
			if err != nil {
				t.Logf("RequestPasswordReset failed: %v", err)
				return false
			}
			
			// Get token
			var token string
			err = db.QueryRow("SELECT token FROM password_reset_tokens ORDER BY created_at DESC LIMIT 1").Scan(&token)
			if err != nil {
				t.Logf("Failed to get token: %v", err)
				return false
			}
			
			captureLogger.Reset() // Clear previous logs
			
			// Reset password
			err = authService.ResetPassword(token, newPassword)
			if err != nil {
				t.Logf("ResetPassword failed: %v", err)
				return false
			}
			
			// Check that password_reset_completed log exists
			if !captureLogger.ContainsLog("password_reset_completed") {
				t.Logf("Expected password_reset_completed log not found")
				return false
			}
			
			log, ok := captureLogger.GetLogByMessage("password_reset_completed")
			if !ok {
				return false
			}
			
			// Verify log contains user_id and timestamp
			if log.Fields["user_id"] == nil || log.Fields["timestamp"] == nil {
				t.Logf("Log missing required fields")
				return false
			}
			
			// Verify log does NOT contain password
			logJSON, _ := json.Marshal(log)
			logStr := string(logJSON)
			if containsSubstring(logStr, newPassword) {
				t.Logf("Log contains plaintext password")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999),
		gen.IntRange(100000, 999999),
	))

	// Test 4: Password change is logged
	properties.Property("password change is logged with user_id and timestamp", prop.ForAll(
		func(phoneNum int, passwordSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			phone := "+7" + padNumber(phoneNum, 10)
			newPassword := "NewPass" + padNumber(passwordSeed, 6) + "!Aa"
			currentPassword := "TestPassword123!"
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			
			err := authService.ChangePassword(userID, currentPassword, newPassword)
			if err != nil {
				t.Logf("ChangePassword failed: %v", err)
				return false
			}
			
			// Check that password_changed log exists
			if !captureLogger.ContainsLog("password_changed") {
				t.Logf("Expected password_changed log not found")
				return false
			}
			
			log, ok := captureLogger.GetLogByMessage("password_changed")
			if !ok {
				return false
			}
			
			// Verify log contains user_id and timestamp
			if log.Fields["user_id"] == nil || log.Fields["timestamp"] == nil {
				t.Logf("Log missing required fields")
				return false
			}
			
			// Verify log does NOT contain passwords
			logJSON, _ := json.Marshal(log)
			logStr := string(logJSON)
			if containsSubstring(logStr, currentPassword) || containsSubstring(logStr, newPassword) {
				t.Logf("Log contains plaintext password")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999),
		gen.IntRange(100000, 999999),
	))
	
	// Test 5: Token expiration is logged
	properties.Property("token expiration is logged when expired token is used", prop.ForAll(
		func(phoneNum int, passwordSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			phone := "+7" + padNumber(phoneNum, 10)
			newPassword := "NewPass" + padNumber(passwordSeed, 6) + "!Aa"
			userID := createTestUser(t, db.GetUnderlyingDB(), phone)
			
			// Create an expired token directly in database
			expiredToken := "expired-token-" + padNumber(phoneNum, 10)
			_, err := db.Exec(
				"INSERT INTO password_reset_tokens (user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4)",
				userID, expiredToken, time.Now().Add(-1*time.Hour), time.Now().Add(-2*time.Hour),
			)
			if err != nil {
				t.Logf("Failed to create expired token: %v", err)
				return false
			}
			
			// Try to use expired token
			err = authService.ResetPassword(expiredToken, newPassword)
			if err == nil {
				t.Logf("Expected error for expired token")
				return false
			}
			
			// Check that password_reset_token_expired log exists
			if !captureLogger.ContainsLog("password_reset_token_expired") {
				t.Logf("Expected password_reset_token_expired log not found")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999),
		gen.IntRange(100000, 999999),
	))

	// Test 6: Token use is logged
	properties.Property("token use is logged when already-used token is attempted", prop.ForAll(
		func(phoneNum int, password1Seed int, password2Seed int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			phone := "+7" + padNumber(phoneNum, 10)
			newPassword1 := "NewPass" + padNumber(password1Seed, 6) + "!Aa"
			newPassword2 := "NewPass" + padNumber(password2Seed, 6) + "!Bb"
			createTestUser(t, db.GetUnderlyingDB(), phone)
			
			// Request reset
			err := authService.RequestPasswordReset(phone)
			if err != nil {
				t.Logf("RequestPasswordReset failed: %v", err)
				return false
			}
			
			// Get token
			var token string
			err = db.QueryRow("SELECT token FROM password_reset_tokens ORDER BY created_at DESC LIMIT 1").Scan(&token)
			if err != nil {
				t.Logf("Failed to get token: %v", err)
				return false
			}
			
			// Use token once
			err = authService.ResetPassword(token, newPassword1)
			if err != nil {
				t.Logf("First ResetPassword failed: %v", err)
				return false
			}
			
			captureLogger.Reset() // Clear previous logs
			
			// Try to use token again
			err = authService.ResetPassword(token, newPassword2)
			if err == nil {
				t.Logf("Expected error for used token")
				return false
			}
			
			// Check that password_reset_token_already_used log exists
			if !captureLogger.ContainsLog("password_reset_token_already_used") {
				t.Logf("Expected password_reset_token_already_used log not found")
				return false
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999),
		gen.IntRange(100000, 999999),
		gen.IntRange(100000, 999999),
	))
	
	properties.TestingRun(t)
}

// containsSubstring checks if a string contains a substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestProperty5_NoPlaintextPasswordLeakage tests that passwords never appear in logs
// **Feature: secure-password-management, Property 5: No plaintext password leakage**
// **Validates: Requirements 1.5, 2.4, 5.5**
func TestProperty5_NoPlaintextPasswordLeakage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService, captureLogger := setupAuthServiceWithLogger(db)
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	properties.Property("passwords never appear in logs during any operation", prop.ForAll(
		func(phoneNum int, passwordSeed1 int, passwordSeed2 int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			phone := "+7" + padNumber(phoneNum, 10)
			initialPassword := "InitPass" + padNumber(passwordSeed1, 6) + "!Aa"
			newPassword := "NewPass" + padNumber(passwordSeed2, 6) + "!Bb"
			
			// Operation 1: Create user
			userID, err := authService.CreateUser(phone, initialPassword)
			if err != nil {
				t.Logf("CreateUser failed: %v", err)
				return false
			}
			
			// Check logs don't contain initial password
			allLogs := captureLogger.GetLogs()
			for _, log := range allLogs {
				logJSON, _ := json.Marshal(log)
				logStr := string(logJSON)
				if containsSubstring(logStr, initialPassword) {
					t.Logf("Initial password found in logs after CreateUser")
					return false
				}
			}
			
			// Operation 2: Request password reset
			err = authService.RequestPasswordReset(phone)
			if err != nil {
				t.Logf("RequestPasswordReset failed: %v", err)
				return false
			}
			
			// Check logs still don't contain password
			allLogs = captureLogger.GetLogs()
			for _, log := range allLogs {
				logJSON, _ := json.Marshal(log)
				logStr := string(logJSON)
				if containsSubstring(logStr, initialPassword) {
					t.Logf("Initial password found in logs after RequestPasswordReset")
					return false
				}
			}
			
			// Get token
			var token string
			err = db.QueryRow("SELECT token FROM password_reset_tokens ORDER BY created_at DESC LIMIT 1").Scan(&token)
			if err != nil {
				t.Logf("Failed to get token: %v", err)
				return false
			}
			
			// Operation 3: Reset password
			err = authService.ResetPassword(token, newPassword)
			if err != nil {
				t.Logf("ResetPassword failed: %v", err)
				return false
			}
			
			// Check logs don't contain either password
			allLogs = captureLogger.GetLogs()
			for _, log := range allLogs {
				logJSON, _ := json.Marshal(log)
				logStr := string(logJSON)
				if containsSubstring(logStr, initialPassword) {
					t.Logf("Initial password found in logs after ResetPassword")
					return false
				}
				if containsSubstring(logStr, newPassword) {
					t.Logf("New password found in logs after ResetPassword")
					return false
				}
			}
			
			// Operation 4: Change password
			anotherPassword := "Another" + padNumber(passwordSeed1+passwordSeed2, 6) + "!Cc"
			err = authService.ChangePassword(userID, newPassword, anotherPassword)
			if err != nil {
				t.Logf("ChangePassword failed: %v", err)
				return false
			}
			
			// Check logs don't contain any password
			allLogs = captureLogger.GetLogs()
			for _, log := range allLogs {
				logJSON, _ := json.Marshal(log)
				logStr := string(logJSON)
				if containsSubstring(logStr, initialPassword) {
					t.Logf("Initial password found in logs after ChangePassword")
					return false
				}
				if containsSubstring(logStr, newPassword) {
					t.Logf("New password found in logs after ChangePassword")
					return false
				}
				if containsSubstring(logStr, anotherPassword) {
					t.Logf("Another password found in logs after ChangePassword")
					return false
				}
			}
			
			// Also check that password hashes are not in logs
			var hash string
			err = db.QueryRow("SELECT password_hash FROM users WHERE id=$1", userID).Scan(&hash)
			if err != nil {
				t.Logf("Failed to get password hash: %v", err)
				return false
			}
			
			for _, log := range allLogs {
				logJSON, _ := json.Marshal(log)
				logStr := string(logJSON)
				if containsSubstring(logStr, hash) {
					t.Logf("Password hash found in logs")
					return false
				}
			}
			
			return true
		},
		gen.IntRange(1000000000, 9999999999),
		gen.IntRange(100000, 999999),
		gen.IntRange(100000, 999999),
	))
	
	properties.TestingRun(t)
}
