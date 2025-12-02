package domain

// ChatRepository определяет интерфейс для работы с чатами
type ChatRepository interface {
	// Create создает новый чат
	Create(chat *Chat) error

	// GetByID получает чат по ID с администраторами
	GetByID(id int64) (*Chat, error)

	// GetByMaxChatID получает чат по MAX chat ID
	GetByMaxChatID(maxChatID string) (*Chat, error)

	// Search выполняет поиск чатов по названию
	Search(query string, limit, offset int, userRole string, universityID *int64) ([]*Chat, int, error)

	// GetAll получает все чаты с пагинацией (с учетом роли пользователя)
	GetAll(limit, offset int, userRole string, universityID *int64) ([]*Chat, int, error)

	// Update обновляет данные чата
	Update(chat *Chat) error

	// Delete удаляет чат
	Delete(id int64) error
}
