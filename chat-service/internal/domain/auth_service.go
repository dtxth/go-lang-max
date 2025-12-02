package domain

// AuthService определяет интерфейс для работы с Auth Service
type AuthService interface {
	// ValidateToken проверяет валидность токена и возвращает информацию о пользователе
	ValidateToken(token string) (*TokenInfo, error)
}

// TokenInfo содержит информацию о пользователе из токена
type TokenInfo struct {
	Valid        bool
	UserID       int64
	Email        string
	Role         string
	UniversityID *int64
	BranchID     *int64
	FacultyID    *int64
}
