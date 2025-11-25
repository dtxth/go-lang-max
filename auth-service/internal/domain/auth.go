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

type JWTManager interface {
	// Генерация токенов возвращает access + refresh и JTI refresh токена
	GenerateTokens(userID int64, email, role string) (*TokensWithJTI, error)

	// Проверка access токена, возвращает userID, email и role
	VerifyAccessToken(token string) (int64, string, string, error)

	// Проверка refresh токена, возвращает claims (включая jti, userID, email, role)
	VerifyRefreshToken(token string) (map[string]interface{}, error)

	// TTL refresh токена
	RefreshTTL() time.Duration
}
