package domain

// AdministratorRepository определяет интерфейс для работы с администраторами чатов
type AdministratorRepository interface {
	// Create создает нового администратора чата
	Create(admin *Administrator) error

	// GetByID получает администратора по ID
	GetByID(id int64) (*Administrator, error)

	// GetByChatID получает всех администраторов чата
	GetByChatID(chatID int64) ([]*Administrator, error)

	// GetByPhoneAndChatID получает администратора по телефону и ID чата
	GetByPhoneAndChatID(phone string, chatID int64) (*Administrator, error)

	// Delete удаляет администратора
	Delete(id int64) error

	// CountByChatID подсчитывает количество администраторов у чата
	CountByChatID(chatID int64) (int, error)
}

