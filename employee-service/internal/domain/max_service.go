package domain

// MaxService определяет интерфейс для работы с MAX API
// Используется для замены номера телефона на MAX_id
type MaxService interface {
	// GetMaxIDByPhone получает MAX_id по номеру телефона
	GetMaxIDByPhone(phone string) (string, error)
	
	// BatchGetMaxIDByPhone получает MAX_id для нескольких телефонов
	// Возвращает map[phone]maxID для успешных запросов
	BatchGetMaxIDByPhone(phones []string) (map[string]string, error)
	
	// GetUserProfileByPhone получает профиль пользователя по номеру телефона
	GetUserProfileByPhone(phone string) (*UserProfile, error)
	
	// GetInternalUsers получает детальную информацию о пользователях по номерам телефонов
	GetInternalUsers(phones []string) ([]*InternalUser, []string, error)
}

// UserProfile содержит профиль пользователя MAX Messenger
type UserProfile struct {
	MaxID     string
	FirstName string
	LastName  string
	Phone     string
}

// InternalUser содержит детальную информацию о пользователе из MAX API /internal/users
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

