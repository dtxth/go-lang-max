package domain

// ChatServiceInterface определяет интерфейс для сервиса чатов
type ChatServiceInterface interface {
	// SearchChats выполняет поиск чатов по названию с фильтрацией по роли
	SearchChats(query string, limit, offset int, filter *ChatFilter) ([]*Chat, int, error)
	
	// GetAllChats получает все чаты с пагинацией и фильтрацией по роли
	GetAllChats(limit, offset int, filter *ChatFilter) ([]*Chat, int, error)
	
	// GetAllChatsWithSortingAndSearch получает все чаты с пагинацией, сортировкой и поиском
	GetAllChatsWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *ChatFilter) ([]*Chat, int, error)
	
	// GetChatByID получает чат по ID
	GetChatByID(id int64) (*Chat, error)
	
	// AddAdministratorWithFlags добавляет администратора к чату с указанием флагов
	AddAdministratorWithFlags(chatID int64, phone string, maxID string, addUser bool, addAdmin bool, skipPhoneValidation bool) (*Administrator, error)
	
	// GetAdministratorByID получает администратора по ID
	GetAdministratorByID(id int64) (*Administrator, error)
	
	// GetAllAdministrators получает всех администраторов с пагинацией и поиском
	GetAllAdministrators(query string, limit, offset int) ([]*Administrator, int, error)
	
	// RemoveAdministrator удаляет администратора из чата
	RemoveAdministrator(adminID int64) error
	
	// CreateChat creates a new chat
	CreateChat(name, url, maxChatID, source string, participantsCount int, universityID *int64, department string) (*Chat, error)
}