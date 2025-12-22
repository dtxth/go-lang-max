package test

import (
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/database"
	"auth-service/internal/infrastructure/max"
	"auth-service/internal/infrastructure/repository"
	"auth-service/internal/usecase"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	_ "github.com/lib/pq"
)

// MaxCaptureLogger captures log output for MAX authentication testing
type MaxCaptureLogger struct {
	buffer *bytes.Buffer
	logs   []MaxLogEntry
}

// MaxLogEntry represents a captured log entry for MAX authentication
type MaxLogEntry struct {
	Message string
	Fields  map[string]interface{}
	IsError bool
}

// NewMaxCaptureLogger creates a new capture logger for MAX authentication
func NewMaxCaptureLogger() *MaxCaptureLogger {
	return &MaxCaptureLogger{
		buffer: &bytes.Buffer{},
		logs:   make([]MaxLogEntry, 0),
	}
}

// Info captures an info log
func (l *MaxCaptureLogger) Info(ctx context.Context, message string, fields map[string]interface{}) {
	l.logs = append(l.logs, MaxLogEntry{
		Message: message,
		Fields:  fields,
		IsError: false,
	})
}

// Error captures an error log
func (l *MaxCaptureLogger) Error(ctx context.Context, message string, fields map[string]interface{}) {
	l.logs = append(l.logs, MaxLogEntry{
		Message: message,
		Fields:  fields,
		IsError: true,
	})
}

// GetLogs returns all captured logs
func (l *MaxCaptureLogger) GetLogs() []MaxLogEntry {
	return l.logs
}

// Reset clears all captured logs
func (l *MaxCaptureLogger) Reset() {
	l.logs = make([]MaxLogEntry, 0)
	l.buffer.Reset()
}

// ContainsLog checks if a log with the given message exists
func (l *MaxCaptureLogger) ContainsLog(message string) bool {
	for _, log := range l.logs {
		if log.Message == message {
			return true
		}
	}
	return false
}

// GetLogByMessage returns the first log with the given message
func (l *MaxCaptureLogger) GetLogByMessage(message string) (MaxLogEntry, bool) {
	for _, log := range l.logs {
		if log.Message == message {
			return log, true
		}
	}
	return MaxLogEntry{}, false
}

// GetErrorLogs returns all error logs
func (l *MaxCaptureLogger) GetErrorLogs() []MaxLogEntry {
	var errorLogs []MaxLogEntry
	for _, log := range l.logs {
		if log.IsError {
			errorLogs = append(errorLogs, log)
		}
	}
	return errorLogs
}

// GetInfoLogs returns all info logs
func (l *MaxCaptureLogger) GetInfoLogs() []MaxLogEntry {
	var infoLogs []MaxLogEntry
	for _, log := range l.logs {
		if !log.IsError {
			infoLogs = append(infoLogs, log)
		}
	}
	return infoLogs
}

// setupMaxAuthServiceWithLogger creates a configured AuthService with capture logger for MAX authentication
func setupMaxAuthServiceWithLogger(db *database.DB) (*usecase.AuthService, *MaxCaptureLogger) {
	userRepo := repository.NewUserPostgres(db)
	refreshRepo := repository.NewRefreshPostgres(db)
	hasher := &mockHasher{}
	jwtManager := &mockJWTManager{}
	maxAuthValidator := max.NewAuthValidator()
	captureLogger := NewMaxCaptureLogger()
	
	authService := usecase.NewAuthService(userRepo, refreshRepo, hasher, jwtManager, nil)
	authService.SetMaxAuthValidator(maxAuthValidator)
	authService.SetMaxBotToken("test_bot_token")
	authService.SetLogger(captureLogger)
	
	return authService, captureLogger
}

// TestProperty7_ComprehensiveAuditLogging tests comprehensive audit logging for MAX authentication
// **Feature: max-miniapp-auth, Property 7: Comprehensive audit logging**
// **Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5**
func TestProperty7_ComprehensiveAuditLogging(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())
	
	authService, captureLogger := setupMaxAuthServiceWithLogger(db)
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)
	
	// Test 1: Successful authentication is logged with user identifiers (Requirement 5.1)
	properties.Property("successful MAX authentication is logged with user identifiers", prop.ForAll(
		func(maxID int64, usernameSeed int, firstNameSeed int, lastNameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			username := "user" + padNumber(usernameSeed, 6)
			firstName := "First" + padNumber(firstNameSeed, 6)
			lastName := "Last" + padNumber(lastNameSeed, 6)
			
			// Create valid initData
			initData := fmt.Sprintf("user={\"id\":%d,\"username\":\"%s\",\"first_name\":\"%s\",\"last_name\":\"%s\"}&auth_date=%d", 
				maxID, username, firstName, lastName, time.Now().Unix())
			
			// Add valid hash
			initDataWithHash := addValidHash(initData, "test_bot_token")
			
			_, err := authService.AuthenticateMAX(initDataWithHash)
			if err != nil {
				t.Logf("AuthenticateMAX failed: %v", err)
				return false
			}
			
			// Check that successful authentication is logged
			if !captureLogger.ContainsLog("max_authentication_successful") {
				t.Logf("Expected max_authentication_successful log not found")
				return false
			}
			
			log, ok := captureLogger.GetLogByMessage("max_authentication_successful")
			if !ok {
				return false
			}
			
			// Verify log contains user_id, max_id, username, and timestamp
			if log.Fields["user_id"] == nil {
				t.Logf("Log missing user_id field")
				return false
			}
			
			if log.Fields["max_id"] == nil {
				t.Logf("Log missing max_id field")
				return false
			}
			
			logMaxID, ok := log.Fields["max_id"].(int64)
			if !ok {
				t.Logf("max_id field is not int64")
				return false
			}
			
			if logMaxID != maxID {
				t.Logf("max_id mismatch: expected %d, got %d", maxID, logMaxID)
				return false
			}
			
			if log.Fields["username"] == nil {
				t.Logf("Log missing username field")
				return false
			}
			
			logUsername, ok := log.Fields["username"].(string)
			if !ok {
				t.Logf("username field is not string")
				return false
			}
			
			if logUsername != username {
				t.Logf("username mismatch: expected %s, got %s", username, logUsername)
				return false
			}
			
			if log.Fields["timestamp"] == nil {
				t.Logf("Log missing timestamp field")
				return false
			}
			
			return true
		},
		gen.Int64Range(1000000, 9999999999),
		gen.IntRange(100000, 999999),
		gen.IntRange(100000, 999999),
		gen.IntRange(100000, 999999),
	))
	
	// Test 2: Hash verification failures are logged without sensitive data (Requirement 5.2)
	properties.Property("hash verification failures are logged without sensitive data", prop.ForAll(
		func(maxID int64, usernameSeed int, firstNameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			username := "user" + padNumber(usernameSeed, 6)
			firstName := "First" + padNumber(firstNameSeed, 6)
			
			// Create initData with invalid hash
			initData := fmt.Sprintf("user={\"id\":%d,\"username\":\"%s\",\"first_name\":\"%s\"}&auth_date=%d&hash=invalid_hash", 
				maxID, username, firstName, time.Now().Unix())
			
			_, err := authService.AuthenticateMAX(initData)
			if err == nil {
				t.Logf("Expected authentication to fail with invalid hash")
				return false
			}
			
			// Check that hash verification failure is logged
			if !captureLogger.ContainsLog("max_auth_validation_failed") {
				t.Logf("Expected max_auth_validation_failed log not found")
				return false
			}
			
			log, ok := captureLogger.GetLogByMessage("max_auth_validation_failed")
			if !ok {
				return false
			}
			
			// Verify log contains error message and timestamp
			if log.Fields["error"] == nil {
				t.Logf("Log missing error field")
				return false
			}
			
			if log.Fields["timestamp"] == nil {
				t.Logf("Log missing timestamp field")
				return false
			}
			
			// Verify log does NOT contain sensitive data (bot token, hash values)
			logJSON, _ := json.Marshal(log)
			logStr := string(logJSON)
			
			if strings.Contains(logStr, "test_bot_token") {
				t.Logf("Log contains bot token")
				return false
			}
			
			if strings.Contains(logStr, "invalid_hash") {
				t.Logf("Log contains hash value")
				return false
			}
			
			// Verify it's logged as an error
			if !log.IsError {
				t.Logf("Hash verification failure should be logged as error")
				return false
			}
			
			return true
		},
		gen.Int64Range(1000000, 9999999999),
		gen.IntRange(100000, 999999),
		gen.IntRange(100000, 999999),
	))
	
	// Test 3: Database errors are logged with debugging context (Requirement 5.3)
	properties.Property("database errors are logged with debugging context", prop.ForAll(
		func(maxID int64, usernameSeed int, firstNameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			username := "user" + padNumber(usernameSeed, 6)
			firstName := "First" + padNumber(firstNameSeed, 6)
			
			// Create existing user to force update path
			existingUser := &domain.User{
				MaxID:    &maxID,
				Username: &username,
				Name:     &firstName,
				Role:     domain.RoleOperator,
			}
			
			userRepo := repository.NewUserPostgres(db)
			err := userRepo.Create(existingUser)
			if err != nil {
				t.Logf("Failed to create existing user: %v", err)
				return false
			}
			
			// Close database connection to force database error
			db.Close()
			
			// Create valid initData
			initData := fmt.Sprintf("user={\"id\":%d,\"username\":\"%s\",\"first_name\":\"%s\"}&auth_date=%d", 
				maxID, username, firstName, time.Now().Unix())
			
			// Add valid hash
			initDataWithHash := addValidHash(initData, "test_bot_token")
			
			_, err = authService.AuthenticateMAX(initDataWithHash)
			if err == nil {
				t.Logf("Expected authentication to fail with database error")
				return false
			}
			
			// Check that database error is logged
			errorLogs := captureLogger.GetErrorLogs()
			if len(errorLogs) == 0 {
				t.Logf("Expected database error log not found")
				return false
			}
			
			// Look for database-related error logs
			foundDatabaseError := false
			for _, log := range errorLogs {
				if strings.Contains(log.Message, "update_failed") || 
				   strings.Contains(log.Message, "creation_failed") {
					
					// Verify log contains debugging context
					if log.Fields["error"] == nil {
						t.Logf("Database error log missing error field")
						return false
					}
					
					if log.Fields["max_id"] == nil {
						t.Logf("Database error log missing max_id field")
						return false
					}
					
					if log.Fields["timestamp"] == nil {
						t.Logf("Database error log missing timestamp field")
						return false
					}
					
					foundDatabaseError = true
					break
				}
			}
			
			if !foundDatabaseError {
				t.Logf("No database error log found with proper context")
				return false
			}
			
			return true
		},
		gen.Int64Range(1000000, 9999999999),
		gen.IntRange(100000, 999999),
		gen.IntRange(100000, 999999),
	))
	
	// Test 4: JWT generation failures are logged appropriately (Requirement 5.4)
	properties.Property("JWT generation failures are logged appropriately", prop.ForAll(
		func(maxID int64, usernameSeed int, firstNameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			username := "user" + padNumber(usernameSeed, 6)
			firstName := "First" + padNumber(firstNameSeed, 6)
			
			// Setup auth service with failing JWT manager
			userRepo := repository.NewUserPostgres(db)
			refreshRepo := repository.NewRefreshPostgres(db)
			hasher := &mockHasher{}
			failingJWTManager := &failingMockJWTManager{} // This will always fail
			maxAuthValidator := max.NewAuthValidator()
			
			authService := usecase.NewAuthService(userRepo, refreshRepo, hasher, failingJWTManager, nil)
			authService.SetMaxAuthValidator(maxAuthValidator)
			authService.SetMaxBotToken("test_bot_token")
			authService.SetLogger(captureLogger)
			
			// Create valid initData
			initData := fmt.Sprintf("user={\"id\":%d,\"username\":\"%s\",\"first_name\":\"%s\"}&auth_date=%d", 
				maxID, username, firstName, time.Now().Unix())
			
			// Add valid hash
			initDataWithHash := addValidHash(initData, "test_bot_token")
			
			_, err := authService.AuthenticateMAX(initDataWithHash)
			if err == nil {
				t.Logf("Expected authentication to fail with JWT generation error")
				return false
			}
			
			// Check that JWT generation failure is logged
			if !captureLogger.ContainsLog("max_jwt_generation_failed") {
				t.Logf("Expected max_jwt_generation_failed log not found")
				return false
			}
			
			log, ok := captureLogger.GetLogByMessage("max_jwt_generation_failed")
			if !ok {
				return false
			}
			
			// Verify log contains user_id, max_id, error, and timestamp
			if log.Fields["user_id"] == nil {
				t.Logf("JWT error log missing user_id field")
				return false
			}
			
			if log.Fields["max_id"] == nil {
				t.Logf("JWT error log missing max_id field")
				return false
			}
			
			if log.Fields["error"] == nil {
				t.Logf("JWT error log missing error field")
				return false
			}
			
			if log.Fields["timestamp"] == nil {
				t.Logf("JWT error log missing timestamp field")
				return false
			}
			
			// Verify it's logged as an error
			if !log.IsError {
				t.Logf("JWT generation failure should be logged as error")
				return false
			}
			
			return true
		},
		gen.Int64Range(1000000, 9999999999),
		gen.IntRange(100000, 999999),
		gen.IntRange(100000, 999999),
	))
	
	// Test 5: User creation and updates are logged with identifiers (Requirement 5.5)
	properties.Property("user creation and updates are logged with identifiers", prop.ForAll(
		func(maxID int64, usernameSeed int, firstNameSeed int, lastNameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())
			captureLogger.Reset()
			
			username := "user" + padNumber(usernameSeed, 6)
			firstName := "First" + padNumber(firstNameSeed, 6)
			lastName := "Last" + padNumber(lastNameSeed, 6)
			
			// Create valid initData
			initData := fmt.Sprintf("user={\"id\":%d,\"username\":\"%s\",\"first_name\":\"%s\",\"last_name\":\"%s\"}&auth_date=%d", 
				maxID, username, firstName, lastName, time.Now().Unix())
			
			// Add valid hash
			initDataWithHash := addValidHash(initData, "test_bot_token")
			
			// First authentication - should create user
			_, err := authService.AuthenticateMAX(initDataWithHash)
			if err != nil {
				t.Logf("First AuthenticateMAX failed: %v", err)
				return false
			}
			
			// Check that user creation is logged
			if !captureLogger.ContainsLog("max_user_created") {
				t.Logf("Expected max_user_created log not found")
				return false
			}
			
			createLog, ok := captureLogger.GetLogByMessage("max_user_created")
			if !ok {
				return false
			}
			
			// Verify creation log contains required fields
			if createLog.Fields["user_id"] == nil || createLog.Fields["max_id"] == nil || 
			   createLog.Fields["username"] == nil || createLog.Fields["timestamp"] == nil {
				t.Logf("User creation log missing required fields")
				return false
			}
			
			captureLogger.Reset()
			
			// Second authentication with updated data - should update user
			updatedUsername := "updated" + padNumber(usernameSeed, 6)
			updatedInitData := fmt.Sprintf("user={\"id\":%d,\"username\":\"%s\",\"first_name\":\"%s\",\"last_name\":\"%s\"}&auth_date=%d", 
				maxID, updatedUsername, firstName, lastName, time.Now().Unix())
			
			updatedInitDataWithHash := addValidHash(updatedInitData, "test_bot_token")
			
			_, err = authService.AuthenticateMAX(updatedInitDataWithHash)
			if err != nil {
				t.Logf("Second AuthenticateMAX failed: %v", err)
				return false
			}
			
			// Check that user update is logged
			if !captureLogger.ContainsLog("max_user_updated") {
				t.Logf("Expected max_user_updated log not found")
				return false
			}
			
			updateLog, ok := captureLogger.GetLogByMessage("max_user_updated")
			if !ok {
				return false
			}
			
			// Verify update log contains required fields
			if updateLog.Fields["user_id"] == nil || updateLog.Fields["max_id"] == nil || 
			   updateLog.Fields["username"] == nil || updateLog.Fields["timestamp"] == nil {
				t.Logf("User update log missing required fields")
				return false
			}
			
			// Verify updated username is logged
			logUsername, ok := updateLog.Fields["username"].(string)
			if !ok || logUsername != updatedUsername {
				t.Logf("Updated username not properly logged")
				return false
			}
			
			return true
		},
		gen.Int64Range(1000000, 9999999999),
		gen.IntRange(100000, 999999),
		gen.IntRange(100000, 999999),
		gen.IntRange(100000, 999999),
	))
	
	properties.TestingRun(t)
}

// failingMockJWTManager always fails to generate tokens
type failingMockJWTManager struct{}

func (m *failingMockJWTManager) GenerateTokens(userID int64, identifier, role string) (*domain.TokensWithJTI, error) {
	return nil, fmt.Errorf("JWT generation failed")
}

func (m *failingMockJWTManager) GenerateTokensWithContext(userID int64, identifier, role string, ctx *domain.TokenContext) (*domain.TokensWithJTI, error) {
	return nil, fmt.Errorf("JWT generation failed")
}

func (m *failingMockJWTManager) VerifyAccessToken(token string) (int64, string, string, error) {
	return 0, "", "", fmt.Errorf("token verification failed")
}

func (m *failingMockJWTManager) VerifyAccessTokenWithContext(token string) (int64, string, string, *domain.TokenContext, error) {
	return 0, "", "", nil, fmt.Errorf("token verification failed")
}

func (m *failingMockJWTManager) VerifyRefreshToken(token string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("refresh token verification failed")
}

func (m *failingMockJWTManager) RefreshTTL() time.Duration {
	return time.Hour * 24
}

// Helper function to create initData with correct hash using MAX Mini App algorithm
func addValidHash(params, botToken string) string {
	// Parse parameters to sort them alphabetically
	values, _ := url.ParseQuery(params)
	
	var sortedParams []string
	for key := range values {
		value := values.Get(key)
		sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", key, value))
	}
	sort.Strings(sortedParams)

	// Create data string for hash calculation (newline separated)
	dataCheckString := strings.Join(sortedParams, "\n")

	// Calculate hash using HMAC-SHA256 with SHA256(botToken) as secret
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))

	// Return initData with hash parameter
	return params + "&hash=" + hash
}