package grpc

import (
	"auth-service/api/proto"
	"auth-service/internal/domain"
	"auth-service/internal/usecase"
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Mock repositories and services for testing
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
	user, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepository) GetByPhone(phone string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Phone == phone {
			return user, nil
		}
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

func (m *mockUserRepository) Update(user *domain.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return errors.New("user not found")
	}
	m.users[user.ID] = user
	return nil
}

type mockPasswordResetRepository struct {
	tokens map[string]*domain.PasswordResetToken
}

func (m *mockPasswordResetRepository) Create(token *domain.PasswordResetToken) error {
	if m.tokens == nil {
		m.tokens = make(map[string]*domain.PasswordResetToken)
	}
	m.tokens[token.Token] = token
	return nil
}

func (m *mockPasswordResetRepository) GetByToken(token string) (*domain.PasswordResetToken, error) {
	t, ok := m.tokens[token]
	if !ok {
		return nil, errors.New("token not found")
	}
	return t, nil
}

func (m *mockPasswordResetRepository) Invalidate(token string) error {
	if _, ok := m.tokens[token]; !ok {
		return errors.New("token not found")
	}
	return nil
}

func (m *mockPasswordResetRepository) DeleteExpired() error {
	return nil
}

type mockRefreshTokenRepository struct{}

func (m *mockRefreshTokenRepository) Save(jti string, userID int64, expiresAt time.Time) error {
	return nil
}

func (m *mockRefreshTokenRepository) IsValid(jti string) (bool, error) {
	return true, nil
}

func (m *mockRefreshTokenRepository) Revoke(jti string) error {
	return nil
}

func (m *mockRefreshTokenRepository) RevokeAllForUser(userID int64) error {
	return nil
}

type mockUserRoleRepository struct{}

func (m *mockUserRoleRepository) Create(userRole *domain.UserRole) error {
	return nil
}

func (m *mockUserRoleRepository) GetByUserID(userID int64) ([]*domain.UserRoleWithDetails, error) {
	return nil, nil
}

func (m *mockUserRoleRepository) Delete(id int64) error {
	return nil
}

func (m *mockUserRoleRepository) DeleteByUserID(userID int64) error {
	return nil
}

func (m *mockUserRoleRepository) GetByUserIDAndRole(userID int64, roleName string) (*domain.UserRoleWithDetails, error) {
	return nil, nil
}

func (m *mockUserRoleRepository) GetRoleByName(name string) (*domain.Role, error) {
	return nil, nil
}

type mockPasswordHasher struct{}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (m *mockPasswordHasher) Compare(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

type mockJWTManager struct{}

func (m *mockJWTManager) GenerateTokens(userID int64, identifier, role string) (*domain.TokensWithJTI, error) {
	return &domain.TokensWithJTI{
		TokenPair: domain.TokenPair{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
		},
		RefreshJTI: "mock-jti",
	}, nil
}

func (m *mockJWTManager) GenerateTokensWithContext(userID int64, identifier, role string, ctx *domain.TokenContext) (*domain.TokensWithJTI, error) {
	return m.GenerateTokens(userID, identifier, role)
}

func (m *mockJWTManager) VerifyAccessToken(token string) (int64, string, string, error) {
	return 1, "+79991234567", "operator", nil
}

func (m *mockJWTManager) VerifyAccessTokenWithContext(token string) (int64, string, string, *domain.TokenContext, error) {
	return 1, "+79991234567", "operator", nil, nil
}

func (m *mockJWTManager) VerifyRefreshToken(token string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"jti":    "mock-jti",
		"userID": int64(1),
		"email":  "+79991234567",
		"role":   "operator",
	}, nil
}

func (m *mockJWTManager) RefreshTTL() time.Duration {
	return 24 * time.Hour
}

type mockNotificationService struct {
	sendPasswordCalled bool
	sendResetCalled    bool
	shouldFail         bool
}

func (m *mockNotificationService) SendPasswordNotification(ctx context.Context, phone, password string) error {
	m.sendPasswordCalled = true
	if m.shouldFail {
		return errors.New("notification failed")
	}
	return nil
}

func (m *mockNotificationService) SendResetTokenNotification(ctx context.Context, phone, token string) error {
	m.sendResetCalled = true
	if m.shouldFail {
		return errors.New("notification failed")
	}
	return nil
}

// Helper function to create a test auth service
func createTestAuthService(userRepo *mockUserRepository, resetRepo *mockPasswordResetRepository, notifService *mockNotificationService) *usecase.AuthService {
	refreshRepo := &mockRefreshTokenRepository{}
	userRoleRepo := &mockUserRoleRepository{}
	hasher := &mockPasswordHasher{}
	jwtManager := &mockJWTManager{}

	authService := usecase.NewAuthService(
		userRepo,
		refreshRepo,
		hasher,
		jwtManager,
		userRoleRepo,
	)
	authService.SetPasswordResetRepository(resetRepo)
	authService.SetNotificationService(notifService)

	return authService
}

// Test RequestPasswordReset handler
func TestRequestPasswordReset_Success(t *testing.T) {
	userRepo := &mockUserRepository{
		users: map[int64]*domain.User{
			1: {
				ID:       1,
				Phone:    "+79991234567",
				Password: "hashed_password",
			},
		},
	}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.RequestPasswordResetRequest{
		Phone: "+79991234567",
	}

	resp, err := handler.RequestPasswordReset(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got success=false with error: %s", resp.Error)
	}

	if !notifService.sendResetCalled {
		t.Error("Expected notification service to be called")
	}
}

func TestRequestPasswordReset_EmptyPhone(t *testing.T) {
	userRepo := &mockUserRepository{}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.RequestPasswordResetRequest{
		Phone: "",
	}

	resp, err := handler.RequestPasswordReset(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false for empty phone")
	}

	if resp.Error != "phone number is required" {
		t.Errorf("Expected error 'phone number is required', got '%s'", resp.Error)
	}
}

func TestRequestPasswordReset_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepository{}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.RequestPasswordResetRequest{
		Phone: "+79991234567",
	}

	resp, err := handler.RequestPasswordReset(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false for non-existent user")
	}

	if resp.Error == "" {
		t.Error("Expected error message for non-existent user")
	}
}

// Test ResetPassword handler
func TestResetPassword_Success(t *testing.T) {
	userRepo := &mockUserRepository{
		users: map[int64]*domain.User{
			1: {
				ID:       1,
				Phone:    "+79991234567",
				Password: "$2a$10$oldhashedpassword",
			},
		},
	}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	// First create a reset token
	_ = authService.RequestPasswordReset("+79991234567")

	// Get the token that was created
	var token string
	for t := range resetRepo.tokens {
		token = t
		break
	}

	req := &proto.ResetPasswordRequest{
		Token:       token,
		NewPassword: "NewSecurePass123!",
	}

	resp, err := handler.ResetPassword(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got success=false with error: %s", resp.Error)
	}
}

func TestResetPassword_EmptyToken(t *testing.T) {
	userRepo := &mockUserRepository{}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.ResetPasswordRequest{
		Token:       "",
		NewPassword: "NewSecurePass123!",
	}

	resp, err := handler.ResetPassword(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false for empty token")
	}

	if resp.Error != "reset token is required" {
		t.Errorf("Expected error 'reset token is required', got '%s'", resp.Error)
	}
}

func TestResetPassword_EmptyPassword(t *testing.T) {
	userRepo := &mockUserRepository{}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.ResetPasswordRequest{
		Token:       "some-token",
		NewPassword: "",
	}

	resp, err := handler.ResetPassword(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false for empty password")
	}

	if resp.Error != "new password is required" {
		t.Errorf("Expected error 'new password is required', got '%s'", resp.Error)
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	userRepo := &mockUserRepository{}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.ResetPasswordRequest{
		Token:       "invalid-token",
		NewPassword: "NewSecurePass123!",
	}

	resp, err := handler.ResetPassword(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false for invalid token")
	}

	if resp.Error == "" {
		t.Error("Expected error message for invalid token")
	}
}

// Test ChangePassword handler
func TestChangePassword_Success(t *testing.T) {
	// Create a user with a known password
	// Generate the hash for "password"
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	
	userRepo := &mockUserRepository{
		users: map[int64]*domain.User{
			1: {
				ID:       1,
				Phone:    "+79991234567",
				Password: string(hash),
			},
		},
	}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.ChangePasswordRequest{
		UserId:          1,
		CurrentPassword: "password",
		NewPassword:     "NewSecurePass123!",
	}

	resp, err := handler.ChangePassword(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got success=false with error: %s", resp.Error)
	}
}

func TestChangePassword_EmptyUserID(t *testing.T) {
	userRepo := &mockUserRepository{}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.ChangePasswordRequest{
		UserId:          0,
		CurrentPassword: "password",
		NewPassword:     "NewSecurePass123!",
	}

	resp, err := handler.ChangePassword(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false for empty user ID")
	}

	if resp.Error != "user ID is required" {
		t.Errorf("Expected error 'user ID is required', got '%s'", resp.Error)
	}
}

func TestChangePassword_EmptyCurrentPassword(t *testing.T) {
	userRepo := &mockUserRepository{}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.ChangePasswordRequest{
		UserId:          1,
		CurrentPassword: "",
		NewPassword:     "NewSecurePass123!",
	}

	resp, err := handler.ChangePassword(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false for empty current password")
	}

	if resp.Error != "current password is required" {
		t.Errorf("Expected error 'current password is required', got '%s'", resp.Error)
	}
}

func TestChangePassword_EmptyNewPassword(t *testing.T) {
	userRepo := &mockUserRepository{}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.ChangePasswordRequest{
		UserId:          1,
		CurrentPassword: "password",
		NewPassword:     "",
	}

	resp, err := handler.ChangePassword(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false for empty new password")
	}

	if resp.Error != "new password is required" {
		t.Errorf("Expected error 'new password is required', got '%s'", resp.Error)
	}
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	// Generate the hash for "password"
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	
	userRepo := &mockUserRepository{
		users: map[int64]*domain.User{
			1: {
				ID:       1,
				Phone:    "+79991234567",
				Password: string(hash),
			},
		},
	}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.ChangePasswordRequest{
		UserId:          1,
		CurrentPassword: "wrongpassword",
		NewPassword:     "NewSecurePass123!",
	}

	resp, err := handler.ChangePassword(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false for wrong current password")
	}

	if resp.Error == "" {
		t.Error("Expected error message for wrong current password")
	}
}

func TestChangePassword_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepository{}
	resetRepo := &mockPasswordResetRepository{}
	notifService := &mockNotificationService{}

	authService := createTestAuthService(userRepo, resetRepo, notifService)
	handler := NewAuthHandler(authService)

	req := &proto.ChangePasswordRequest{
		UserId:          999,
		CurrentPassword: "password",
		NewPassword:     "NewSecurePass123!",
	}

	resp, err := handler.ChangePassword(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false for non-existent user")
	}

	if resp.Error == "" {
		t.Error("Expected error message for non-existent user")
	}
}
