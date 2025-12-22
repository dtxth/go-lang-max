package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"auth-service/internal/domain"
)

// Mock implementations for testing
type mockMaxAuthValidator struct {
	validateFunc func(initData string, botToken string) (*domain.MaxUserData, error)
}

func (m *mockMaxAuthValidator) ValidateInitData(initData string, botToken string) (*domain.MaxUserData, error) {
	if m.validateFunc != nil {
		return m.validateFunc(initData, botToken)
	}
	return nil, errors.New("not implemented")
}

type mockUserRepository struct {
	users       map[int64]*domain.User
	usersByMaxID map[int64]*domain.User
	createFunc  func(user *domain.User) error
	updateFunc  func(user *domain.User) error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:       make(map[int64]*domain.User),
		usersByMaxID: make(map[int64]*domain.User),
	}
}

func (m *mockUserRepository) Create(user *domain.User) error {
	if m.createFunc != nil {
		return m.createFunc(user)
	}
	
	// Simulate auto-increment ID
	user.ID = int64(len(m.users) + 1)
	m.users[user.ID] = user
	
	if user.MaxID != nil {
		m.usersByMaxID[*user.MaxID] = user
	}
	
	return nil
}

func (m *mockUserRepository) Update(user *domain.User) error {
	if m.updateFunc != nil {
		return m.updateFunc(user)
	}
	
	if _, exists := m.users[user.ID]; !exists {
		return errors.New("user not found")
	}
	
	m.users[user.ID] = user
	if user.MaxID != nil {
		m.usersByMaxID[*user.MaxID] = user
	}
	
	return nil
}

func (m *mockUserRepository) GetByID(id int64) (*domain.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetByMaxID(maxID int64) (*domain.User, error) {
	if user, exists := m.usersByMaxID[maxID]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetByEmail(email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetByPhone(phone string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Phone == phone {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

type mockRefreshTokenRepository struct {
	tokens   map[string]int64
	saveFunc func(jti string, userID int64, expiresAt time.Time) error
}

func newMockRefreshTokenRepository() *mockRefreshTokenRepository {
	return &mockRefreshTokenRepository{
		tokens: make(map[string]int64),
	}
}

func (m *mockRefreshTokenRepository) Save(jti string, userID int64, expiresAt time.Time) error {
	if m.saveFunc != nil {
		return m.saveFunc(jti, userID, expiresAt)
	}
	m.tokens[jti] = userID
	return nil
}

func (m *mockRefreshTokenRepository) IsValid(jti string) (bool, error) {
	_, exists := m.tokens[jti]
	return exists, nil
}

func (m *mockRefreshTokenRepository) Revoke(jti string) error {
	delete(m.tokens, jti)
	return nil
}

func (m *mockRefreshTokenRepository) RevokeAllForUser(userID int64) error {
	for jti, uid := range m.tokens {
		if uid == userID {
			delete(m.tokens, jti)
		}
	}
	return nil
}

type mockJWTManager struct {
	generateFunc func(userID int64, identifier, role string) (*domain.TokensWithJTI, error)
}

func (m *mockJWTManager) GenerateTokens(userID int64, identifier, role string) (*domain.TokensWithJTI, error) {
	if m.generateFunc != nil {
		return m.generateFunc(userID, identifier, role)
	}
	return &domain.TokensWithJTI{
		TokenPair: domain.TokenPair{
			AccessToken:  fmt.Sprintf("access_token_%d", userID),
			RefreshToken: fmt.Sprintf("refresh_token_%d", userID),
		},
		RefreshJTI: fmt.Sprintf("jti_%d", userID),
	}, nil
}

func (m *mockJWTManager) GenerateTokensWithContext(userID int64, identifier, role string, ctx *domain.TokenContext) (*domain.TokensWithJTI, error) {
	return m.GenerateTokens(userID, identifier, role)
}

func (m *mockJWTManager) VerifyAccessToken(token string) (int64, string, string, error) {
	return 0, "", "", errors.New("not implemented")
}

func (m *mockJWTManager) VerifyAccessTokenWithContext(token string) (int64, string, string, *domain.TokenContext, error) {
	return 0, "", "", nil, errors.New("not implemented")
}

func (m *mockJWTManager) VerifyRefreshToken(token string) (map[string]interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockJWTManager) RefreshTTL() time.Duration {
	return 24 * time.Hour
}

type mockLogger struct {
	infoLogs  []map[string]interface{}
	errorLogs []map[string]interface{}
}

func (m *mockLogger) Info(ctx context.Context, message string, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["message"] = message
	for k, v := range fields {
		entry[k] = v
	}
	m.infoLogs = append(m.infoLogs, entry)
}

func (m *mockLogger) Error(ctx context.Context, message string, fields map[string]interface{}) {
	entry := make(map[string]interface{})
	entry["message"] = message
	for k, v := range fields {
		entry[k] = v
	}
	m.errorLogs = append(m.errorLogs, entry)
}

func TestAuthService_AuthenticateMAX_Success_NewUser(t *testing.T) {
	// Setup mocks
	userRepo := newMockUserRepository()
	refreshRepo := newMockRefreshTokenRepository()
	jwtManager := &mockJWTManager{}
	logger := &mockLogger{}
	
	validator := &mockMaxAuthValidator{
		validateFunc: func(initData string, botToken string) (*domain.MaxUserData, error) {
			return &domain.MaxUserData{
				MaxID:     123,
				Username:  "johndoe",
				FirstName: "John",
				LastName:  "Doe",
			}, nil
		},
	}

	// Create auth service
	authService := NewAuthService(userRepo, refreshRepo, nil, jwtManager, nil)
	authService.SetMaxAuthValidator(validator)
	authService.SetMaxBotToken("test_token")
	authService.SetLogger(logger)

	// Test authentication
	result, err := authService.AuthenticateMAX("valid_init_data")

	// Assertions
	if err != nil {
		t.Errorf("AuthenticateMAX() unexpected error = %v", err)
		return
	}

	if result == nil {
		t.Errorf("AuthenticateMAX() result is nil")
		return
	}

	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Errorf("AuthenticateMAX() missing tokens")
	}

	// Verify user was created
	if len(userRepo.users) != 1 {
		t.Errorf("AuthenticateMAX() expected 1 user, got %d", len(userRepo.users))
	}

	// Verify user data
	var createdUser *domain.User
	for _, user := range userRepo.users {
		createdUser = user
		break
	}

	if createdUser.MaxID == nil || *createdUser.MaxID != 123 {
		t.Errorf("AuthenticateMAX() user MaxID = %v, want 123", createdUser.MaxID)
	}
	if createdUser.Username == nil || *createdUser.Username != "johndoe" {
		t.Errorf("AuthenticateMAX() user Username = %v, want johndoe", createdUser.Username)
	}
	if createdUser.Name == nil || *createdUser.Name != "John Doe" {
		t.Errorf("AuthenticateMAX() user Name = %v, want 'John Doe'", createdUser.Name)
	}
	if createdUser.Role != domain.RoleOperator {
		t.Errorf("AuthenticateMAX() user Role = %v, want %v", createdUser.Role, domain.RoleOperator)
	}

	// Verify logging
	if len(logger.infoLogs) == 0 {
		t.Errorf("AuthenticateMAX() expected info logs")
	}
}

func TestAuthService_AuthenticateMAX_Success_ExistingUser(t *testing.T) {
	// Setup mocks
	userRepo := newMockUserRepository()
	refreshRepo := newMockRefreshTokenRepository()
	jwtManager := &mockJWTManager{}
	logger := &mockLogger{}

	// Create existing user
	maxID := int64(123)
	username := "oldusername"
	name := "Old Name"
	existingUser := &domain.User{
		ID:       1,
		MaxID:    &maxID,
		Username: &username,
		Name:     &name,
		Role:     domain.RoleOperator,
	}
	userRepo.users[1] = existingUser
	userRepo.usersByMaxID[123] = existingUser
	
	validator := &mockMaxAuthValidator{
		validateFunc: func(initData string, botToken string) (*domain.MaxUserData, error) {
			return &domain.MaxUserData{
				MaxID:     123,
				Username:  "newusername",
				FirstName: "New",
				LastName:  "Name",
			}, nil
		},
	}

	// Create auth service
	authService := NewAuthService(userRepo, refreshRepo, nil, jwtManager, nil)
	authService.SetMaxAuthValidator(validator)
	authService.SetMaxBotToken("test_token")
	authService.SetLogger(logger)

	// Test authentication
	result, err := authService.AuthenticateMAX("valid_init_data")

	// Assertions
	if err != nil {
		t.Errorf("AuthenticateMAX() unexpected error = %v", err)
		return
	}

	if result == nil {
		t.Errorf("AuthenticateMAX() result is nil")
		return
	}

	// Verify user was updated, not created
	if len(userRepo.users) != 1 {
		t.Errorf("AuthenticateMAX() expected 1 user, got %d", len(userRepo.users))
	}

	// Verify user data was updated
	updatedUser := userRepo.users[1]
	if updatedUser.Username == nil || *updatedUser.Username != "newusername" {
		t.Errorf("AuthenticateMAX() user Username = %v, want newusername", updatedUser.Username)
	}
	if updatedUser.Name == nil || *updatedUser.Name != "New Name" {
		t.Errorf("AuthenticateMAX() user Name = %v, want 'New Name'", updatedUser.Name)
	}

	// Verify logging includes update message
	found := false
	for _, log := range logger.infoLogs {
		if log["message"] == "max_user_updated" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AuthenticateMAX() expected user update log")
	}
}

func TestAuthService_AuthenticateMAX_ValidationFailure(t *testing.T) {
	// Setup mocks
	userRepo := newMockUserRepository()
	refreshRepo := newMockRefreshTokenRepository()
	jwtManager := &mockJWTManager{}
	logger := &mockLogger{}
	
	validator := &mockMaxAuthValidator{
		validateFunc: func(initData string, botToken string) (*domain.MaxUserData, error) {
			return nil, errors.New("hash verification failed")
		},
	}

	// Create auth service
	authService := NewAuthService(userRepo, refreshRepo, nil, jwtManager, nil)
	authService.SetMaxAuthValidator(validator)
	authService.SetMaxBotToken("test_token")
	authService.SetLogger(logger)

	// Test authentication
	result, err := authService.AuthenticateMAX("invalid_init_data")

	// Assertions
	if err == nil {
		t.Errorf("AuthenticateMAX() expected error but got none")
		return
	}

	if result != nil {
		t.Errorf("AuthenticateMAX() expected nil result but got %v", result)
	}

	// Verify error type
	if !strings.Contains(err.Error(), "Invalid authentication data") {
		t.Errorf("AuthenticateMAX() error = %v, want error containing 'Invalid authentication data'", err)
	}

	// Verify no user was created
	if len(userRepo.users) != 0 {
		t.Errorf("AuthenticateMAX() expected 0 users, got %d", len(userRepo.users))
	}

	// Verify error logging
	if len(logger.errorLogs) == 0 {
		t.Errorf("AuthenticateMAX() expected error logs")
	}
}

func TestAuthService_AuthenticateMAX_MissingValidator(t *testing.T) {
	// Setup mocks without validator
	userRepo := newMockUserRepository()
	refreshRepo := newMockRefreshTokenRepository()
	jwtManager := &mockJWTManager{}

	// Create auth service without validator
	authService := NewAuthService(userRepo, refreshRepo, nil, jwtManager, nil)
	authService.SetMaxBotToken("test_token")

	// Test authentication
	result, err := authService.AuthenticateMAX("valid_init_data")

	// Assertions
	if err == nil {
		t.Errorf("AuthenticateMAX() expected error but got none")
		return
	}

	if result != nil {
		t.Errorf("AuthenticateMAX() expected nil result but got %v", result)
	}

	if !strings.Contains(err.Error(), "MAX auth validator not configured") {
		t.Errorf("AuthenticateMAX() error = %v, want error containing 'MAX auth validator not configured'", err)
	}
}

func TestAuthService_AuthenticateMAX_MissingBotToken(t *testing.T) {
	// Setup mocks without bot token
	userRepo := newMockUserRepository()
	refreshRepo := newMockRefreshTokenRepository()
	jwtManager := &mockJWTManager{}
	validator := &mockMaxAuthValidator{}

	// Create auth service without bot token
	authService := NewAuthService(userRepo, refreshRepo, nil, jwtManager, nil)
	authService.SetMaxAuthValidator(validator)

	// Test authentication
	result, err := authService.AuthenticateMAX("valid_init_data")

	// Assertions
	if err == nil {
		t.Errorf("AuthenticateMAX() expected error but got none")
		return
	}

	if result != nil {
		t.Errorf("AuthenticateMAX() expected nil result but got %v", result)
	}

	if !strings.Contains(err.Error(), "MAX bot token not configured") {
		t.Errorf("AuthenticateMAX() error = %v, want error containing 'MAX bot token not configured'", err)
	}
}

func TestAuthService_AuthenticateMAX_DatabaseError(t *testing.T) {
	// Setup mocks
	userRepo := newMockUserRepository()
	userRepo.createFunc = func(user *domain.User) error {
		return errors.New("database connection failed")
	}
	
	refreshRepo := newMockRefreshTokenRepository()
	jwtManager := &mockJWTManager{}
	logger := &mockLogger{}
	
	validator := &mockMaxAuthValidator{
		validateFunc: func(initData string, botToken string) (*domain.MaxUserData, error) {
			return &domain.MaxUserData{
				MaxID:     123,
				Username:  "johndoe",
				FirstName: "John",
				LastName:  "Doe",
			}, nil
		},
	}

	// Create auth service
	authService := NewAuthService(userRepo, refreshRepo, nil, jwtManager, nil)
	authService.SetMaxAuthValidator(validator)
	authService.SetMaxBotToken("test_token")
	authService.SetLogger(logger)

	// Test authentication
	result, err := authService.AuthenticateMAX("valid_init_data")

	// Assertions
	if err == nil {
		t.Errorf("AuthenticateMAX() expected error but got none")
		return
	}

	if result != nil {
		t.Errorf("AuthenticateMAX() expected nil result but got %v", result)
	}

	if !strings.Contains(err.Error(), "failed to create user") {
		t.Errorf("AuthenticateMAX() error = %v, want error containing 'failed to create user'", err)
	}

	// Verify error logging
	found := false
	for _, log := range logger.errorLogs {
		if log["message"] == "max_user_creation_failed" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AuthenticateMAX() expected user creation failure log")
	}
}

func TestAuthService_AuthenticateMAX_JWTGenerationError(t *testing.T) {
	// Setup mocks
	userRepo := newMockUserRepository()
	refreshRepo := newMockRefreshTokenRepository()
	jwtManager := &mockJWTManager{
		generateFunc: func(userID int64, identifier, role string) (*domain.TokensWithJTI, error) {
			return nil, errors.New("JWT generation failed")
		},
	}
	logger := &mockLogger{}
	
	validator := &mockMaxAuthValidator{
		validateFunc: func(initData string, botToken string) (*domain.MaxUserData, error) {
			return &domain.MaxUserData{
				MaxID:     123,
				Username:  "johndoe",
				FirstName: "John",
				LastName:  "Doe",
			}, nil
		},
	}

	// Create auth service
	authService := NewAuthService(userRepo, refreshRepo, nil, jwtManager, nil)
	authService.SetMaxAuthValidator(validator)
	authService.SetMaxBotToken("test_token")
	authService.SetLogger(logger)

	// Test authentication
	result, err := authService.AuthenticateMAX("valid_init_data")

	// Assertions
	if err == nil {
		t.Errorf("AuthenticateMAX() expected error but got none")
		return
	}

	if result != nil {
		t.Errorf("AuthenticateMAX() expected nil result but got %v", result)
	}

	if !strings.Contains(err.Error(), "failed to generate tokens") {
		t.Errorf("AuthenticateMAX() error = %v, want error containing 'failed to generate tokens'", err)
	}

	// Verify error logging
	found := false
	for _, log := range logger.errorLogs {
		if log["message"] == "max_jwt_generation_failed" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AuthenticateMAX() expected JWT generation failure log")
	}
}

func TestAuthService_AuthenticateMAX_RefreshTokenSaveError(t *testing.T) {
	// Setup mocks
	userRepo := newMockUserRepository()
	refreshRepo := newMockRefreshTokenRepository()
	refreshRepo.saveFunc = func(jti string, userID int64, expiresAt time.Time) error {
		return errors.New("refresh token save failed")
	}
	
	jwtManager := &mockJWTManager{}
	logger := &mockLogger{}
	
	validator := &mockMaxAuthValidator{
		validateFunc: func(initData string, botToken string) (*domain.MaxUserData, error) {
			return &domain.MaxUserData{
				MaxID:     123,
				Username:  "johndoe",
				FirstName: "John",
				LastName:  "Doe",
			}, nil
		},
	}

	// Create auth service
	authService := NewAuthService(userRepo, refreshRepo, nil, jwtManager, nil)
	authService.SetMaxAuthValidator(validator)
	authService.SetMaxBotToken("test_token")
	authService.SetLogger(logger)

	// Test authentication
	result, err := authService.AuthenticateMAX("valid_init_data")

	// Assertions
	if err == nil {
		t.Errorf("AuthenticateMAX() expected error but got none")
		return
	}

	if result != nil {
		t.Errorf("AuthenticateMAX() expected nil result but got %v", result)
	}

	if !strings.Contains(err.Error(), "failed to save refresh token") {
		t.Errorf("AuthenticateMAX() error = %v, want error containing 'failed to save refresh token'", err)
	}

	// Verify error logging
	found := false
	for _, log := range logger.errorLogs {
		if log["message"] == "max_refresh_token_save_failed" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AuthenticateMAX() expected refresh token save failure log")
	}
}

func TestAuthService_AuthenticateMAX_ErrorMapping(t *testing.T) {
	tests := []struct {
		name           string
		validationError string
		expectedError  string
	}{
		{
			name:           "hash verification failure",
			validationError: "hash verification failed",
			expectedError:  "Invalid authentication data",
		},
		{
			name:           "missing hash parameter",
			validationError: "hash parameter is missing",
			expectedError:  "Invalid initData format",
		},
		{
			name:           "parse failure",
			validationError: "failed to parse initData",
			expectedError:  "Invalid initData format",
		},
		{
			name:           "empty initData",
			validationError: "initData cannot be empty",
			expectedError:  "Invalid initData format",
		},
		{
			name:           "other validation error",
			validationError: "some other error",
			expectedError:  "Authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := newMockUserRepository()
			refreshRepo := newMockRefreshTokenRepository()
			jwtManager := &mockJWTManager{}
			
			validator := &mockMaxAuthValidator{
				validateFunc: func(initData string, botToken string) (*domain.MaxUserData, error) {
					return nil, errors.New(tt.validationError)
				},
			}

			// Create auth service
			authService := NewAuthService(userRepo, refreshRepo, nil, jwtManager, nil)
			authService.SetMaxAuthValidator(validator)
			authService.SetMaxBotToken("test_token")

			// Test authentication
			_, err := authService.AuthenticateMAX("invalid_init_data")

			// Assertions
			if err == nil {
				t.Errorf("AuthenticateMAX() expected error but got none")
				return
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("AuthenticateMAX() error = %v, want error containing '%v'", err, tt.expectedError)
			}
		})
	}
}