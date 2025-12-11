package domain

import "time"

type PasswordHasher interface {
	Hash(s string) (string, error)
	Compare(s, hashed string) bool
}

type TokensWithJTI struct {
	TokenPair
	RefreshJTI string
}

// TokenContext содержит контекстную информацию для токена
type TokenContext struct {
	UniversityID *int64
	BranchID     *int64
	FacultyID    *int64
}

type JWTManager interface {
	// Генерация токенов возвращает access + refresh и JTI refresh токена
	GenerateTokens(userID int64, identifier, role string) (*TokensWithJTI, error)
	
	// Генерация токенов с контекстом роли
	GenerateTokensWithContext(userID int64, identifier, role string, ctx *TokenContext) (*TokensWithJTI, error)

	// Проверка access токена, возвращает userID, identifier (email или phone) и role
	VerifyAccessToken(token string) (int64, string, string, error)
	
	// Проверка access токена с контекстом, возвращает userID, identifier (email или phone), role и контекст
	VerifyAccessTokenWithContext(token string) (int64, string, string, *TokenContext, error)

	// Проверка refresh токена, возвращает claims (включая jti, userID, identifier, role)
	VerifyRefreshToken(token string) (map[string]interface{}, error)

	// TTL refresh токена
	RefreshTTL() time.Duration
}
