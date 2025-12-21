package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"auth-service/internal/domain"
	appErrors "auth-service/internal/infrastructure/errors"
	"auth-service/internal/infrastructure/metrics"
)

type AuthService struct {
    repo                   domain.UserRepository
    refreshRepo            domain.RefreshTokenRepository
    hasher                 domain.PasswordHasher
    jwtManager             domain.JWTManager
    userRoleRepo           domain.UserRoleRepository
    resetTokenRepo         domain.PasswordResetRepository
    notificationService    domain.NotificationService
    maxBotClient           domain.MaxBotClient
    maxAuthValidator       domain.MaxAuthValidator
    maxBotToken            string
    logger                 Logger
    metrics                *metrics.Metrics
    minPasswordLength      int
    resetTokenExpiration   time.Duration
}

// Logger interface for audit logging
type Logger interface {
    Info(ctx context.Context, message string, fields map[string]interface{})
    Error(ctx context.Context, message string, fields map[string]interface{})
}

func NewAuthService(repo domain.UserRepository, refreshRepo domain.RefreshTokenRepository, hasher domain.PasswordHasher, jwtManager domain.JWTManager, userRoleRepo domain.UserRoleRepository) *AuthService {
    return &AuthService{
        repo:                 repo,
        refreshRepo:          refreshRepo,
        hasher:               hasher,
        jwtManager:           jwtManager,
        userRoleRepo:         userRoleRepo,
        minPasswordLength:    12, // Default value
        resetTokenExpiration: 15 * time.Minute, // Default value
    }
}

// SetMaxBotClient sets the MaxBot client
func (s *AuthService) SetMaxBotClient(client domain.MaxBotClient) {
    s.maxBotClient = client
}

// SetMaxAuthValidator sets the MAX auth validator
func (s *AuthService) SetMaxAuthValidator(validator domain.MaxAuthValidator) {
    s.maxAuthValidator = validator
}

// SetMaxBotToken sets the MAX bot token
func (s *AuthService) SetMaxBotToken(token string) {
    s.maxBotToken = token
}

// SetPasswordConfig sets the password configuration
func (s *AuthService) SetPasswordConfig(minLength int, resetTokenExpiration time.Duration) {
    s.minPasswordLength = minLength
    s.resetTokenExpiration = resetTokenExpiration
}

// SetLogger sets the logger for audit logging
func (s *AuthService) SetLogger(logger Logger) {
    s.logger = logger
}

// SetPasswordResetRepository sets the password reset repository
func (s *AuthService) SetPasswordResetRepository(repo domain.PasswordResetRepository) {
    s.resetTokenRepo = repo
}

// SetNotificationService sets the notification service
func (s *AuthService) SetNotificationService(service domain.NotificationService) {
    s.notificationService = service
}

// SetMetrics sets the metrics collector
func (s *AuthService) SetMetrics(m *metrics.Metrics) {
    s.metrics = m
}

// GetMetrics returns the metrics collector
func (s *AuthService) GetMetrics() *metrics.Metrics {
    return s.metrics
}

func (s *AuthService) Register(email, password, role string) (*domain.User, error) {
    hashed, err := s.hasher.Hash(password)
    if err != nil {
        return nil, err
    }

    // Валидация роли
    if role != "" && role != domain.RoleSuperAdmin && role != domain.RoleCurator && role != domain.RoleOperator {
        return nil, errors.New("invalid role")
    }

    user := &domain.User{Email: email, Password: hashed, Role: role}
    if err := s.repo.Create(user); err != nil {
        return nil, domain.ErrUserExists
    }
    return user, nil
}

func (s *AuthService) RegisterByPhone(phone, password, role string) (*domain.User, error) {
    hashed, err := s.hasher.Hash(password)
    if err != nil {
        return nil, err
    }

    // Валидация роли
    if role != "" && role != domain.RoleSuperAdmin && role != domain.RoleCurator && role != domain.RoleOperator {
        return nil, errors.New("invalid role")
    }

    user := &domain.User{Phone: phone, Password: hashed, Role: role}
    if err := s.repo.Create(user); err != nil {
        return nil, domain.ErrUserExists
    }
    return user, nil
}

// CreateUser создает нового пользователя без роли (роль назначается отдельно через AssignRole)
func (s *AuthService) CreateUser(phone, password string) (int64, error) {

    
    // Проверяем, не существует ли уже пользователь с таким телефоном
    existingUser, err := s.repo.GetByPhone(phone)
    if err == nil && existingUser != nil && existingUser.ID > 0 {

        return existingUser.ID, nil // Возвращаем существующего пользователя
    }

    hashed, err := s.hasher.Hash(password)
    if err != nil {
        return 0, fmt.Errorf("failed to hash password: %w", err)
    }

    user := &domain.User{
        Phone:    phone,
        Email:    "", // Email опциональный
        Password: hashed,
        Role:     "", // Роль будет назначена через AssignRole
    }
    
    if err := s.repo.Create(user); err != nil {
        return 0, fmt.Errorf("failed to create user: %w", err)
    }
    
    // Record metrics
    if s.metrics != nil {
        s.metrics.IncrementUserCreations()
    }
    
    // Логируем сгенерированный пароль для администраторов
    log.Printf("Generated password for new user with phone ending in %s: %s", 
        sanitizePhone(phone), password)
    
    // Audit log: user created (without password or hash)
    if s.logger != nil {
        s.logger.Info(nil, "user_created", map[string]interface{}{
            "user_id":         user.ID,
            "phone":           sanitizePhone(phone),
            "timestamp":       time.Now().UTC().Format(time.RFC3339),
            "operation":       "create_user",
        })
    }
    
    return user.ID, nil
}

func (s *AuthService) Login(email, password string) (*TokensWithJTIResult, error) {
    user, err := s.repo.GetByEmail(email)
    if err != nil {
        return nil, domain.ErrInvalidCreds
    }
    if !s.hasher.Compare(password, user.Password) {
        return nil, domain.ErrInvalidCreds
    }

    tokens, err := s.jwtManager.GenerateTokens(user.ID, user.Email, user.Role)
    if err != nil {
        return nil, err
    }

    // Save refresh jti
    expiresAt := time.Now().Add(s.jwtManager.RefreshTTL())
    if err := s.refreshRepo.Save(tokens.RefreshJTI, user.ID, expiresAt); err != nil {
        return nil, err
    }

    return &TokensWithJTIResult{
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
        RefreshJTI:   tokens.RefreshJTI,
    }, nil
}

// LoginByIdentifier поддерживает вход как по email, так и по телефону
func (s *AuthService) LoginByIdentifier(identifier, password string) (*TokensWithJTIResult, error) {
    var user *domain.User
    var err error
    
    // Определяем, является ли идентификатор телефоном (начинается с +)
    if len(identifier) > 0 && identifier[0] == '+' {
        user, err = s.repo.GetByPhone(identifier)
        if err != nil {
            if s.logger != nil {
                s.logger.Error(context.Background(), "failed_to_get_user_by_phone", map[string]interface{}{
                    "phone": identifier,
                    "error": err.Error(),
                })
            }
        }
    } else {
        user, err = s.repo.GetByEmail(identifier)
        if err != nil {
            if s.logger != nil {
                s.logger.Error(context.Background(), "failed_to_get_user_by_email", map[string]interface{}{
                    "email": identifier,
                    "error": err.Error(),
                })
            }
        }
    }
    
    if err != nil {
        return nil, domain.ErrInvalidCreds
    }
    
    if s.logger != nil {
        s.logger.Info(context.Background(), "user_found", map[string]interface{}{
            "user_id": user.ID,
            "phone": user.Phone,
            "email": user.Email,
        })
    }
    
    if !s.hasher.Compare(password, user.Password) {
        if s.logger != nil {
            s.logger.Error(context.Background(), "password_comparison_failed", map[string]interface{}{
                "user_id": user.ID,
            })
        }
        return nil, domain.ErrInvalidCreds
    }

    // Используем тот идентификатор, по которому пользователь авторизовался
    var tokenIdentifier string
    if len(identifier) > 0 && identifier[0] == '+' {
        // Авторизация по телефону - используем телефон в токене
        tokenIdentifier = user.Phone
    } else {
        // Авторизация по email - используем email в токене
        tokenIdentifier = user.Email
    }

    tokens, err := s.jwtManager.GenerateTokens(user.ID, tokenIdentifier, user.Role)
    if err != nil {
        return nil, err
    }

    // Save refresh jti
    expiresAt := time.Now().Add(s.jwtManager.RefreshTTL())
    if err := s.refreshRepo.Save(tokens.RefreshJTI, user.ID, expiresAt); err != nil {
        return nil, err
    }

    return &TokensWithJTIResult{
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
        RefreshJTI:   tokens.RefreshJTI,
    }, nil
}

type TokensWithJTIResult struct {
    AccessToken  string
    RefreshToken string
    RefreshJTI   string
}

// Refresh: validate incoming refresh token, check jti in DB, revoke old, issue new pair
func (s *AuthService) Refresh(refreshToken string) (*TokensWithJTIResult, error) {
    claims, err := s.jwtManager.VerifyRefreshToken(refreshToken)
    if err != nil {
        return nil, err
    }

    jtiVal, ok := claims["jti"].(string)
    if !ok || jtiVal == "" {
        return nil, errors.New("invalid refresh token: missing jti")
    }

    // check DB
    valid, err := s.refreshRepo.IsValid(jtiVal)
    if err != nil {
        return nil, err
    }
    if !valid {
        return nil, domain.ErrTokenExpired
    }

    // extract subject (user id)
    sub, ok := claims["sub"].(string)
    if !ok {
        // sometimes numeric
        switch v := claims["sub"].(type) {
        case float64:
            sub = fmt.Sprintf("%.0f", v)
        default:
            return nil, errors.New("invalid subject in token")
        }
    }
    userID, err := strconv.ParseInt(sub, 10, 64)
    if err != nil {
        return nil, err
    }

    // Получаем пользователя из БД для получения актуальной роли
    user, err := s.repo.GetByID(userID)
    if err != nil {
        return nil, domain.ErrInvalidCreds
    }

    // Извлекаем идентификатор из исходного токена (phone или email)
    var identifier string
    if phone, ok := claims["phone"].(string); ok {
        identifier = phone
    } else if email, ok := claims["email"].(string); ok {
        identifier = email
    } else {
        // Fallback к email пользователя из БД
        identifier = user.Email
    }

    // generate new tokens с актуальной ролью из БД
    tokens, err := s.jwtManager.GenerateTokens(userID, identifier, user.Role)
    if err != nil {
        return nil, err
    }

    // save new jti and revoke old
    if err := s.refreshRepo.Save(tokens.RefreshJTI, userID, time.Now().Add(s.jwtManager.RefreshTTL())); err != nil {
        return nil, err
    }
    if err := s.refreshRepo.Revoke(jtiVal); err != nil {
        // non-fatal? but return error to be explicit
        return nil, err
    }

    return &TokensWithJTIResult{
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
        RefreshJTI:   tokens.RefreshJTI,
    }, nil
}

// Logout: revoke provided refresh token jti
func (s *AuthService) Logout(refreshToken string) error {
    claims, err := s.jwtManager.VerifyRefreshToken(refreshToken)
    if err != nil {
        return err
    }
    jtiVal, ok := claims["jti"].(string)
    if !ok {
        return errors.New("invalid refresh token: missing jti")
    }
    return s.refreshRepo.Revoke(jtiVal)
}

// ValidateToken проверяет валидность access токена и возвращает информацию о пользователе
func (s *AuthService) ValidateToken(token string) (int64, string, string, error) {
    return s.jwtManager.VerifyAccessToken(token)
}

// ValidateTokenWithContext проверяет валидность access токена и возвращает информацию о пользователе с контекстом
func (s *AuthService) ValidateTokenWithContext(token string) (int64, string, string, *domain.TokenContext, error) {
    return s.jwtManager.VerifyAccessTokenWithContext(token)
}

// GetUserByID получает пользователя по ID
func (s *AuthService) GetUserByID(userID int64) (*domain.User, error) {
    return s.repo.GetByID(userID)
}

// GetUserPermissions возвращает все разрешения пользователя
func (s *AuthService) GetUserPermissions(userID int64) ([]*domain.UserRoleWithDetails, error) {
    if s.userRoleRepo == nil {
        return nil, errors.New("user role repository not initialized")
    }
    return s.userRoleRepo.GetByUserID(userID)
}

// AssignRoleToUser назначает роль пользователю
func (s *AuthService) AssignRoleToUser(userID int64, roleName string, universityID, branchID, facultyID *int64) error {
	if s.userRoleRepo == nil {
		return errors.New("user role repository not initialized")
	}
	
	// Валидация роли
	if roleName != domain.RoleSuperAdmin && roleName != domain.RoleCurator && roleName != domain.RoleOperator {
		return errors.New("invalid role")
	}
	
	// Получаем роль по имени
	role, err := s.userRoleRepo.GetRoleByName(roleName)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	
	// Создаем user_role запись
	userRole := &domain.UserRole{
		UserID: userID,
		RoleID: role.ID,
	}
	
	if universityID != nil {
		userRole.UniversityID = universityID
	}
	if branchID != nil {
		userRole.BranchID = branchID
	}
	if facultyID != nil {
		userRole.FacultyID = facultyID
	}
	
	return s.userRoleRepo.Create(userRole)
}

// RevokeAllUserRoles отзывает все роли пользователя
func (s *AuthService) RevokeAllUserRoles(userID int64) error {
	if s.userRoleRepo == nil {
		return errors.New("user role repository not initialized")
	}
	
	return s.userRoleRepo.DeleteByUserID(userID)
}

// RequestPasswordReset generates a reset token and sends it to the user
func (s *AuthService) RequestPasswordReset(phone string) error {
	if s.resetTokenRepo == nil {
		return errors.New("password reset repository not initialized")
	}
	if s.notificationService == nil {
		return errors.New("notification service not initialized")
	}

	// Find user by phone
	user, err := s.repo.GetByPhone(phone)
	if err != nil {
		return domain.ErrUserNotFound
	}

	// Generate reset token (64 characters, cryptographically secure)
	token, err := generateSecureToken(64)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Create token with configured expiration
	resetToken := &domain.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(s.resetTokenExpiration),
	}

	// Store token in database
	if err := s.resetTokenRepo.Create(resetToken); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// Record metrics for token generation
	if s.metrics != nil {
		s.metrics.IncrementTokensGenerated()
	}

	// Логируем токен сброса пароля для администраторов
	log.Printf("Generated password reset token for user with phone ending in %s: %s", 
		sanitizePhone(phone), token)

	// Send token via notification service
	if err := s.notificationService.SendResetTokenNotification(nil, phone, token); err != nil {
		return fmt.Errorf("failed to send reset token notification: %w", err)
	}

	// Record metrics for password reset
	if s.metrics != nil {
		s.metrics.IncrementPasswordResets()
	}

	// Audit log: password reset requested (without token)
	if s.logger != nil {
		s.logger.Info(nil, "password_reset_requested", map[string]interface{}{
			"user_id":   user.ID,
			"phone":     sanitizePhone(phone),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "request_password_reset",
		})
	}

	return nil
}

// ResetPassword validates token and updates password
func (s *AuthService) ResetPassword(token, newPassword string) error {
	if s.resetTokenRepo == nil {
		return errors.New("password reset repository not initialized")
	}

	// Validate token exists
	resetToken, err := s.resetTokenRepo.GetByToken(token)
	if err != nil {
		if err == domain.ErrNotFound {
			return domain.ErrResetTokenNotFound
		}
		return fmt.Errorf("failed to retrieve reset token: %w", err)
	}

	// Check if token is used
	if resetToken.IsUsed() {
		// Record metrics for invalidated token
		if s.metrics != nil {
			s.metrics.IncrementTokensInvalidated()
		}
		
		// Audit log: token already used
		if s.logger != nil {
			s.logger.Info(nil, "password_reset_token_already_used", map[string]interface{}{
				"user_id":   resetToken.UserID,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"operation": "reset_password_token_used",
			})
		}
		return domain.ErrResetTokenUsed
	}

	// Check if token is expired
	if resetToken.IsExpired() {
		// Record metrics for expired token
		if s.metrics != nil {
			s.metrics.IncrementTokensExpired()
		}
		
		// Audit log: token expired
		if s.logger != nil {
			s.logger.Info(nil, "password_reset_token_expired", map[string]interface{}{
				"user_id":   resetToken.UserID,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"operation": "reset_password_token_expired",
			})
		}
		return domain.ErrResetTokenExpired
	}

	// Validate new password meets requirements
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashed, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Get user
	user, err := s.repo.GetByID(resetToken.UserID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	// Update user password
	user.Password = hashed
	if err := s.repo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate token
	if err := s.resetTokenRepo.Invalidate(token); err != nil {
		return fmt.Errorf("failed to invalidate token: %w", err)
	}

	// Record metrics for token usage
	if s.metrics != nil {
		s.metrics.IncrementTokensUsed()
	}

	// Revoke all refresh tokens for this user
	if err := s.refreshRepo.RevokeAllForUser(resetToken.UserID); err != nil {
		// Log but don't fail - password was already updated
		fmt.Printf("Warning: failed to revoke refresh tokens for user %d: %v\n", resetToken.UserID, err)
	}

	// Audit log: password reset completed (without password or hash)
	if s.logger != nil {
		s.logger.Info(nil, "password_reset_completed", map[string]interface{}{
			"user_id":   resetToken.UserID,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "reset_password",
		})
	}

	return nil
}

// ChangePassword allows authenticated user to change password
func (s *AuthService) ChangePassword(userID int64, currentPassword, newPassword string) error {
	// Get user by ID
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	// Verify current password
	if !s.hasher.Compare(currentPassword, user.Password) {
		return domain.ErrInvalidCreds
	}

	// Validate new password meets requirements
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashed, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	user.Password = hashed
	if err := s.repo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke all refresh tokens for this user
	if err := s.refreshRepo.RevokeAllForUser(userID); err != nil {
		// Log but don't fail - password was already updated
		fmt.Printf("Warning: failed to revoke refresh tokens for user %d: %v\n", userID, err)
	}

	// Record metrics for password change
	if s.metrics != nil {
		s.metrics.IncrementPasswordChanges()
	}

	// Audit log: password changed (without password or hash)
	if s.logger != nil {
		s.logger.Info(nil, "password_changed", map[string]interface{}{
			"user_id":   userID,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "change_password",
		})
	}

	return nil
}

// validatePassword checks if a password meets security requirements
func (s *AuthService) validatePassword(password string) error {
	// Check minimum length
	if len(password) < s.minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", s.minPasswordLength)
	}

	// Check for uppercase letter
	hasUpper := false
	for _, c := range password {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Check for lowercase letter
	hasLower := false
	for _, c := range password {
		if c >= 'a' && c <= 'z' {
			hasLower = true
			break
		}
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for digit
	hasDigit := false
	for _, c := range password {
		if c >= '0' && c <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	// Check for special character
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	hasSpecial := false
	for _, c := range password {
		for _, special := range specialChars {
			if c == special {
				hasSpecial = true
				break
			}
		}
		if hasSpecial {
			break
		}
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := make([]byte, length)
	
	for i := range token {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		token[i] = charset[num.Int64()]
	}
	
	return string(token), nil
}

// sanitizePhone returns only the last 4 digits of a phone number for logging
func sanitizePhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return "****" + phone[len(phone)-4:]
}
// AuthenticateMAX authenticates a user using MAX Mini App initData
func (s *AuthService) AuthenticateMAX(initData string) (*TokensWithJTIResult, error) {
	if s.maxAuthValidator == nil {
		return nil, errors.New("MAX auth validator not configured")
	}
	
	if s.maxBotToken == "" {
		return nil, errors.New("MAX bot token not configured")
	}

	// Validate initData and extract user information
	maxUserData, err := s.maxAuthValidator.ValidateInitData(initData, s.maxBotToken)
	if err != nil {
		// Audit log: hash verification failure (without sensitive data)
		if s.logger != nil {
			s.logger.Error(nil, "max_auth_validation_failed", map[string]interface{}{
				"error":     "initData validation failed",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"operation": "authenticate_max",
			})
		}
		
		// Map validation errors to appropriate HTTP status codes
		errMsg := err.Error()
		if strings.Contains(errMsg, "hash verification failed") {
			return nil, appErrors.UnauthorizedError("Invalid authentication data")
		} else if strings.Contains(errMsg, "hash parameter is missing") || 
				  strings.Contains(errMsg, "failed to parse initData") ||
				  strings.Contains(errMsg, "initData cannot be empty") {
			return nil, appErrors.ValidationError("Invalid initData format")
		} else {
			return nil, appErrors.UnauthorizedError("Authentication failed")
		}
	}

	// Try to find existing user by max_id
	user, err := s.repo.GetByMaxID(maxUserData.MaxID)
	if err != nil {
		// User doesn't exist - return authentication error instead of creating
		if s.logger != nil {
			s.logger.Error(nil, "max_user_not_found", map[string]interface{}{
				"max_id":    maxUserData.MaxID,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"operation": "authenticate_max",
			})
		}
		return nil, appErrors.UnauthorizedError("User not found. Please contact administrator to register your account.")
	}

	// User exists, update their information with current MAX data
	user.Username = &maxUserData.Username
	displayName := buildDisplayName(maxUserData.FirstName, maxUserData.LastName)
	user.Name = &displayName
	
	if err := s.repo.Update(user); err != nil {
		// Audit log: database error
		if s.logger != nil {
			s.logger.Error(nil, "max_user_update_failed", map[string]interface{}{
				"user_id":   user.ID,
				"max_id":    maxUserData.MaxID,
				"error":     err.Error(),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"operation": "authenticate_max",
			})
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	// Audit log: existing user updated
	if s.logger != nil {
		s.logger.Info(nil, "max_user_updated", map[string]interface{}{
			"user_id":   user.ID,
			"max_id":    maxUserData.MaxID,
			"username":  maxUserData.Username,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "authenticate_max",
		})
	}

	// Generate JWT tokens using max_id as identifier
	identifier := fmt.Sprintf("max_%d", maxUserData.MaxID)
	tokens, err := s.jwtManager.GenerateTokens(user.ID, identifier, user.Role)
	if err != nil {
		// Audit log: JWT generation failure
		if s.logger != nil {
			s.logger.Error(nil, "max_jwt_generation_failed", map[string]interface{}{
				"user_id":   user.ID,
				"max_id":    maxUserData.MaxID,
				"error":     err.Error(),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"operation": "authenticate_max",
			})
		}
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Save refresh token JTI
	expiresAt := time.Now().Add(s.jwtManager.RefreshTTL())
	if err := s.refreshRepo.Save(tokens.RefreshJTI, user.ID, expiresAt); err != nil {
		// Audit log: refresh token save failure
		if s.logger != nil {
			s.logger.Error(nil, "max_refresh_token_save_failed", map[string]interface{}{
				"user_id":   user.ID,
				"max_id":    maxUserData.MaxID,
				"error":     err.Error(),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"operation": "authenticate_max",
			})
		}
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Audit log: successful authentication
	if s.logger != nil {
		s.logger.Info(nil, "max_authentication_successful", map[string]interface{}{
			"user_id":   user.ID,
			"max_id":    maxUserData.MaxID,
			"username":  maxUserData.Username,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"operation": "authenticate_max",
		})
	}

	return &TokensWithJTIResult{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		RefreshJTI:   tokens.RefreshJTI,
	}, nil
}

// buildDisplayName creates a display name from first and last name
func buildDisplayName(firstName, lastName string) string {
	if lastName != "" {
		return firstName + " " + lastName
	}
	return firstName
}

// GetBotInfo retrieves bot information from MaxBot service
func (s *AuthService) GetBotInfo(ctx context.Context) (*domain.BotInfo, error) {
	if s.maxBotClient == nil {
		return nil, fmt.Errorf("MaxBot client not configured")
	}

	botInfo, err := s.maxBotClient.GetBotInfo(ctx)
	if err != nil {
		if s.logger != nil {
			s.logger.Error(ctx, "failed_to_get_bot_info", map[string]interface{}{
				"error":     err.Error(),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		}
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}

	if s.logger != nil {
		s.logger.Info(ctx, "bot_info_retrieved", map[string]interface{}{
			"bot_name":  botInfo.Name,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	return botInfo, nil
}