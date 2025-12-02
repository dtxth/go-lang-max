package domain

// MaxService определяет интерфейс для работы с MAX API
// Используется для получения MAX_id по номеру телефона
type MaxService interface {
	// GetMaxIDByPhone получает MAX_id по номеру телефона
	GetMaxIDByPhone(phone string) (string, error)

	// ValidatePhone проверяет валидность номера телефона
	ValidatePhone(phone string) bool
}

