package usecase

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"auth-service/internal/domain"
)

type AuthService struct {
    repo         domain.UserRepository
    refreshRepo  domain.RefreshTokenRepository
    hasher       domain.PasswordHasher
    jwtManager   domain.JWTManager
    userRoleRepo domain.UserRoleRepository
}

func NewAuthService(repo domain.UserRepository, refreshRepo domain.RefreshTokenRepository, hasher domain.PasswordHasher, jwtManager domain.JWTManager, userRoleRepo domain.UserRoleRepository) *AuthService {
    return &AuthService{
        repo: repo, refreshRepo: refreshRepo, hasher: hasher, jwtManager: jwtManager, userRoleRepo: userRoleRepo,
    }
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

    // generate new tokens с актуальной ролью из БД
    tokens, err := s.jwtManager.GenerateTokens(userID, user.Email, user.Role)
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

// 
AssignRoleToUser назначает роль пользователю
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
