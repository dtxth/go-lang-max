package domain

import "context"

// MaxService определяет интерфейс для работы с MAX API
// Используется для получения MAX_id по номеру телефона и информации о чатах
type MaxService interface {
	// GetMaxIDByPhone получает MAX_id по номеру телефона
	GetMaxIDByPhone(phone string) (string, error)

	// ValidatePhone проверяет валидность номера телефона
	ValidatePhone(phone string) bool

	// GetChatInfo получает информацию о чате из MAX API
	GetChatInfo(ctx context.Context, chatID int64) (*ChatInfo, error)

	// GetInternalUsers получает детальную информацию о пользователях по номерам телефонов
	GetInternalUsers(phones []string) ([]*InternalUser, []string, error)
}

// ChatInfo содержит информацию о чате из MAX API
type ChatInfo struct {
	ChatID            int64
	Title             string
	Type              string
	ParticipantsCount int
	Description       string
}

// InternalUser представляет детальную информацию о пользователе из MAX
type InternalUser struct {
	UserID        int64
	FirstName     string
	LastName      string
	IsBot         bool
	Username      string
	AvatarURL     string
	FullAvatarURL string
	Link          string
	PhoneNumber   string
}

