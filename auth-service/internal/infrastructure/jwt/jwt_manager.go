package jwt

import (
	"fmt"
	"strconv"
	"time"

	"auth-service/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
    accessSecret  []byte
    refreshSecret []byte

    accessTTL  time.Duration
    refreshTTL time.Duration
}

func NewManager(access, refresh string, accessTTL, refreshTTL time.Duration) *Manager {
    return &Manager{
        accessSecret:  []byte(access),
        refreshSecret: []byte(refresh),
        accessTTL:     accessTTL,
        refreshTTL:    refreshTTL,
    }
}

// GenerateTokens создаёт access и refresh токены с JTI (без контекста)
func (m *Manager) GenerateTokens(userID int64, email, role string) (*domain.TokensWithJTI, error) {
    return m.GenerateTokensWithContext(userID, email, role, nil)
}

// GenerateTokensWithContext создаёт access и refresh токены с JTI и контекстом роли
func (m *Manager) GenerateTokensWithContext(userID int64, email, role string, ctx *domain.TokenContext) (*domain.TokensWithJTI, error) {
    now := time.Now()

    // Access token
    accessClaims := jwt.MapClaims{
        "sub":   fmt.Sprintf("%d", userID),
        "email": email,
        "role":  role,
        "exp":   now.Add(m.accessTTL).Unix(),
        "iat":   now.Unix(),
    }
    
    // Добавляем контекстную информацию, если она есть
    if ctx != nil {
        if ctx.UniversityID != nil {
            accessClaims["university_id"] = *ctx.UniversityID
        }
        if ctx.BranchID != nil {
            accessClaims["branch_id"] = *ctx.BranchID
        }
        if ctx.FacultyID != nil {
            accessClaims["faculty_id"] = *ctx.FacultyID
        }
    }
    
    access := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessStr, err := access.SignedString(m.accessSecret)
    if err != nil {
        return nil, err
    }

    // Refresh token
    jti := uuid.NewString()
    refreshClaims := jwt.MapClaims{
        "jti":   jti,
        "sub":   fmt.Sprintf("%d", userID),
        "email": email,
        "role":  role,
        "exp":   now.Add(m.refreshTTL).Unix(),
        "iat":   now.Unix(),
    }
    
    // Добавляем контекстную информацию в refresh token
    if ctx != nil {
        if ctx.UniversityID != nil {
            refreshClaims["university_id"] = *ctx.UniversityID
        }
        if ctx.BranchID != nil {
            refreshClaims["branch_id"] = *ctx.BranchID
        }
        if ctx.FacultyID != nil {
            refreshClaims["faculty_id"] = *ctx.FacultyID
        }
    }
    
    refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshStr, err := refresh.SignedString(m.refreshSecret)
    if err != nil {
        return nil, err
    }

    return &domain.TokensWithJTI{
        TokenPair: domain.TokenPair{
            AccessToken:  accessStr,
            RefreshToken: refreshStr,
        },
        RefreshJTI: jti,
    }, nil
}

// JWTVerify проверяет access токен и возвращает claims
func (m *Manager) JWTVerify(tokenStr string) (map[string]interface{}, error) {
    token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return m.accessSecret, nil
    })
    if err != nil {
        return nil, err
    }
    if !token.Valid {
        return nil, jwt.ErrTokenSignatureInvalid
    }
    if claims, ok := token.Claims.(jwt.MapClaims); ok {
        return claims, nil
    }
    return nil, jwt.ErrTokenSignatureInvalid
}

// ParseAccessToken возвращает userID, email и role из access токена
func (m *Manager) ParseAccessToken(tokenStr string) (int64, string, string, error) {
    claims, err := m.JWTVerify(tokenStr)
    if err != nil {
        return 0, "", "", err
    }

    sub, ok := claims["sub"].(string)
    if !ok {
        return 0, "", "", fmt.Errorf("sub not found in token")
    }
    email, ok := claims["email"].(string)
    if !ok {
        return 0, "", "", fmt.Errorf("email not found in token")
    }
    role, ok := claims["role"].(string)
    if !ok {
        return 0, "", "", fmt.Errorf("role not found in token")
    }

    userID, err := strconv.ParseInt(sub, 10, 64)
    if err != nil {
        return 0, "", "", err
    }

    return userID, email, role, nil
}

// ParseRefreshToken проверяет refresh токен и возвращает claims
func (m *Manager) ParseRefreshToken(tokenStr string) (map[string]interface{}, error) {
    token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return m.refreshSecret, nil
    })
    if err != nil {
        return nil, err
    }
    if !token.Valid {
        return nil, jwt.ErrTokenSignatureInvalid
    }
    if claims, ok := token.Claims.(jwt.MapClaims); ok {
        return claims, nil
    }
    return nil, jwt.ErrTokenSignatureInvalid
}

// VerifyAccessToken проверяет access токен и возвращает userID, email и role (реализация интерфейса domain.JWTManager)
func (m *Manager) VerifyAccessToken(tokenStr string) (int64, string, string, error) {
    return m.ParseAccessToken(tokenStr)
}

// VerifyAccessTokenWithContext проверяет access токен и возвращает userID, email, role и контекст
func (m *Manager) VerifyAccessTokenWithContext(tokenStr string) (int64, string, string, *domain.TokenContext, error) {
    claims, err := m.JWTVerify(tokenStr)
    if err != nil {
        return 0, "", "", nil, err
    }

    sub, ok := claims["sub"].(string)
    if !ok {
        return 0, "", "", nil, fmt.Errorf("sub not found in token")
    }
    email, ok := claims["email"].(string)
    if !ok {
        return 0, "", "", nil, fmt.Errorf("email not found in token")
    }
    role, ok := claims["role"].(string)
    if !ok {
        return 0, "", "", nil, fmt.Errorf("role not found in token")
    }

    userID, err := strconv.ParseInt(sub, 10, 64)
    if err != nil {
        return 0, "", "", nil, err
    }

    // Извлекаем контекстную информацию
    ctx := &domain.TokenContext{}
    
    if universityID, ok := claims["university_id"].(float64); ok {
        id := int64(universityID)
        ctx.UniversityID = &id
    }
    
    if branchID, ok := claims["branch_id"].(float64); ok {
        id := int64(branchID)
        ctx.BranchID = &id
    }
    
    if facultyID, ok := claims["faculty_id"].(float64); ok {
        id := int64(facultyID)
        ctx.FacultyID = &id
    }

    return userID, email, role, ctx, nil
}

// VerifyRefreshToken проверяет refresh токен и возвращает claims (реализация интерфейса domain.JWTManager)
func (m *Manager) VerifyRefreshToken(tokenStr string) (map[string]interface{}, error) {
    return m.ParseRefreshToken(tokenStr)
}

// RefreshTTL возвращает TTL refresh токена
func (m *Manager) RefreshTTL() time.Duration {
    return m.refreshTTL
}