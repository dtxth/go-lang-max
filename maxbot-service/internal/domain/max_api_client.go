package domain

import "context"

// MaxAPIClient определяет интерфейс для работы с Max Messenger Bot API
type MaxAPIClient interface {
	// GetMaxIDByPhone получает MAX ID по номеру телефона
	GetMaxIDByPhone(ctx context.Context, phone string) (string, error)
	// ValidatePhone проверяет валидность номера телефона
	ValidatePhone(phone string) (bool, string, error)
	
	// SendMessage отправляет сообщение пользователю или в чат
	SendMessage(ctx context.Context, chatID, userID int64, text string) (string, error)
	// SendNotification отправляет VIP-уведомление пользователю по номеру телефона
	SendNotification(ctx context.Context, phone, text string) error
	
	// GetChatInfo получает информацию о чате
	GetChatInfo(ctx context.Context, chatID int64) (*ChatInfo, error)
	// GetChatMembers получает список участников чата
	GetChatMembers(ctx context.Context, chatID int64, limit int, marker int64) (*ChatMembersList, error)
	// GetChatAdmins получает список администраторов чата
	GetChatAdmins(ctx context.Context, chatID int64) ([]*ChatMember, error)
	
	// CheckPhoneNumbers проверяет существование номеров телефонов в Max Messenger
	CheckPhoneNumbers(ctx context.Context, phones []string) ([]string, error)
	
	// BatchGetUsersByPhone получает MAX ID для списка номеров телефонов (до 100)
	BatchGetUsersByPhone(ctx context.Context, phones []string) ([]*UserPhoneMapping, error)
}

// ChatInfo содержит информацию о чате
type ChatInfo struct {
	ChatID            int64
	Title             string
	Type              string
	ParticipantsCount int
	Description       string
}

// ChatMember представляет участника чата
type ChatMember struct {
	UserID  int64
	Name    string
	IsAdmin bool
	IsOwner bool
}

// ChatMembersList содержит список участников чата с маркером для пагинации
type ChatMembersList struct {
	Members []*ChatMember
	Marker  int64
}

// UserPhoneMapping представляет маппинг телефона на MAX_id
type UserPhoneMapping struct {
	Phone string
	MaxID string
	Found bool
}
