package usecase

import (
	"chat-service/internal/domain"
	"errors"
	"strings"
)

type ChatService struct {
	chatRepo          domain.ChatRepository
	administratorRepo domain.AdministratorRepository
	universityRepo    domain.UniversityRepository
	maxService        domain.MaxService
}

func NewChatService(
	chatRepo domain.ChatRepository,
	administratorRepo domain.AdministratorRepository,
	universityRepo domain.UniversityRepository,
	maxService domain.MaxService,
) *ChatService {
	return &ChatService{
		chatRepo:          chatRepo,
		administratorRepo: administratorRepo,
		universityRepo:    universityRepo,
		maxService:        maxService,
	}
}

// SearchChats выполняет поиск чатов по названию с учетом роли пользователя
func (s *ChatService) SearchChats(query string, limit, offset int, userRole string, universityID *int64) ([]*domain.Chat, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.chatRepo.Search(query, limit, offset, userRole, universityID)
}

// GetAllChats получает все чаты с пагинацией (с учетом роли пользователя)
func (s *ChatService) GetAllChats(limit, offset int, userRole string, universityID *int64) ([]*domain.Chat, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.chatRepo.GetAll(limit, offset, userRole, universityID)
}

// GetChatByID получает чат по ID
func (s *ChatService) GetChatByID(id int64) (*domain.Chat, error) {
	chat, err := s.chatRepo.GetByID(id)
	if err != nil {
		return nil, domain.ErrChatNotFound
	}
	return chat, nil
}

// AddAdministrator добавляет администратора к чату
func (s *ChatService) AddAdministrator(chatID int64, phone string) (*domain.Administrator, error) {
	// Валидация телефона
	if !s.maxService.ValidatePhone(phone) {
		return nil, domain.ErrInvalidPhone
	}

	// Проверяем существование чата
	_, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		return nil, domain.ErrChatNotFound
	}

	// Проверяем, не существует ли уже администратор с таким телефоном в этом чате
	existing, _ := s.administratorRepo.GetByPhoneAndChatID(phone, chatID)
	if existing != nil {
		return nil, domain.ErrAdministratorExists
	}

	// Получаем MAX_id по телефону
	maxID, err := s.maxService.GetMaxIDByPhone(phone)
	if err != nil {
		return nil, err
	}

	// Создаем администратора
	admin := &domain.Administrator{
		ChatID: chatID,
		Phone:  phone,
		MaxID:  maxID,
	}

	if err := s.administratorRepo.Create(admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// RemoveAdministrator удаляет администратора из чата
// Нельзя удалить последнего администратора (должно быть минимум 2)
func (s *ChatService) RemoveAdministrator(adminID int64) error {
	// Получаем администратора
	admin, err := s.administratorRepo.GetByID(adminID)
	if err != nil {
		return domain.ErrAdministratorNotFound
	}

	// Проверяем количество администраторов у чата
	count, err := s.administratorRepo.CountByChatID(admin.ChatID)
	if err != nil {
		return err
	}

	if count < 2 {
		return domain.ErrCannotDeleteLastAdmin
	}

	// Удаляем администратора
	return s.administratorRepo.Delete(adminID)
}

// CreateChat создает новый чат
func (s *ChatService) CreateChat(
	name, url, maxChatID, source string,
	participantsCount int,
	universityID *int64,
	department string,
) (*domain.Chat, error) {
	// Валидация
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("chat name is required")
	}

	url = strings.TrimSpace(url)
	if url == "" {
		return nil, errors.New("chat URL is required")
	}

	// Проверяем валидность источника
	validSources := map[string]bool{
		"admin_panel":    true,
		"bot_registrar":  true,
		"academic_group": true,
	}
	if !validSources[source] {
		return nil, errors.New("invalid chat source")
	}

	// Проверяем существование вуза, если указан
	if universityID != nil {
		_, err := s.universityRepo.GetByID(*universityID)
		if err != nil {
			return nil, domain.ErrUniversityNotFound
		}
	}

	// Создаем чат
	chat := &domain.Chat{
		Name:              name,
		URL:               url,
		MaxChatID:         maxChatID,
		ParticipantsCount: participantsCount,
		UniversityID:      universityID,
		Department:        strings.TrimSpace(department),
		Source:            source,
	}

	if err := s.chatRepo.Create(chat); err != nil {
		return nil, err
	}

	// Загружаем полную информацию о чате
	return s.chatRepo.GetByID(chat.ID)
}

// UpdateChat обновляет данные чата
func (s *ChatService) UpdateChat(chat *domain.Chat) error {
	// Проверяем существование чата
	_, err := s.chatRepo.GetByID(chat.ID)
	if err != nil {
		return domain.ErrChatNotFound
	}

	// Проверяем существование вуза, если указан
	if chat.UniversityID != nil {
		_, err := s.universityRepo.GetByID(*chat.UniversityID)
		if err != nil {
			return domain.ErrUniversityNotFound
		}
	}

	return s.chatRepo.Update(chat)
}

// DeleteChat удаляет чат
func (s *ChatService) DeleteChat(id int64) error {
	_, err := s.chatRepo.GetByID(id)
	if err != nil {
		return domain.ErrChatNotFound
	}

	return s.chatRepo.Delete(id)
}

