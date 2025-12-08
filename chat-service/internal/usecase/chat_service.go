package usecase

import (
	"chat-service/internal/domain"
	"errors"
	"strings"
)

type ChatService struct {
	chatRepo                              domain.ChatRepository
	administratorRepo                     domain.AdministratorRepository
	universityRepo                        domain.UniversityRepository
	maxService                            domain.MaxService
	listChatsWithRoleFilterUC             *ListChatsWithRoleFilterUseCase
	addAdministratorWithPermissionCheckUC *AddAdministratorWithPermissionCheckUseCase
	removeAdministratorWithValidationUC   *RemoveAdministratorWithValidationUseCase
}

func NewChatService(
	chatRepo domain.ChatRepository,
	administratorRepo domain.AdministratorRepository,
	universityRepo domain.UniversityRepository,
	maxService domain.MaxService,
) *ChatService {
	return &ChatService{
		chatRepo:                              chatRepo,
		administratorRepo:                     administratorRepo,
		universityRepo:                        universityRepo,
		maxService:                            maxService,
		listChatsWithRoleFilterUC:             NewListChatsWithRoleFilterUseCase(chatRepo),
		addAdministratorWithPermissionCheckUC: NewAddAdministratorWithPermissionCheckUseCase(administratorRepo, chatRepo, maxService),
		removeAdministratorWithValidationUC:   NewRemoveAdministratorWithValidationUseCase(administratorRepo, chatRepo),
	}
}

// SearchChats выполняет поиск чатов по названию с фильтрацией по роли
func (s *ChatService) SearchChats(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return s.listChatsWithRoleFilterUC.Execute(query, limit, offset, filter)
}

// GetAllChats получает все чаты с пагинацией и фильтрацией по роли
func (s *ChatService) GetAllChats(limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return s.listChatsWithRoleFilterUC.GetAll(limit, offset, filter)
}

// GetChatByID получает чат по ID
func (s *ChatService) GetChatByID(id int64) (*domain.Chat, error) {
	chat, err := s.chatRepo.GetByID(id)
	if err != nil {
		return nil, domain.ErrChatNotFound
	}
	return chat, nil
}

// AddAdministrator добавляет администратора к чату (без проверки прав - для обратной совместимости)
func (s *ChatService) AddAdministrator(chatID int64, phone string) (*domain.Administrator, error) {
	return s.AddAdministratorWithFlags(chatID, phone, "", true, true, false)
}

// AddAdministratorWithFlags добавляет администратора к чату с указанием флагов
func (s *ChatService) AddAdministratorWithFlags(chatID int64, phone string, maxID string, addUser bool, addAdmin bool, skipPhoneValidation bool) (*domain.Administrator, error) {
	// Валидация телефона (пропускаем для миграции)
	if !skipPhoneValidation && !s.maxService.ValidatePhone(phone) {
		return nil, domain.ErrInvalidPhone
	}

	// Проверяем существование чата
	_, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		return nil, domain.ErrChatNotFound
	}

	// Проверяем, не существует ли уже администратор с таким телефоном в этом чате
	existing, err := s.administratorRepo.GetByPhoneAndChatID(phone, chatID)
	if err == nil && existing != nil {
		return nil, domain.ErrAdministratorExists
	}

	// Если MAX_id не передан, получаем его по телефону (только если не пропускаем валидацию)
	if maxID == "" && !skipPhoneValidation {
		maxID, err = s.maxService.GetMaxIDByPhone(phone)
		if err != nil {
			return nil, err
		}
	}

	// Создаем администратора
	admin := &domain.Administrator{
		ChatID:   chatID,
		Phone:    phone,
		MaxID:    maxID,
		AddUser:  addUser,
		AddAdmin: addAdmin,
	}

	if err := s.administratorRepo.Create(admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// AddAdministratorWithPermissionCheck добавляет администратора к чату с проверкой прав доступа
func (s *ChatService) AddAdministratorWithPermissionCheck(
	chatID int64,
	phone string,
	userRole string,
	userUniversityID *int64,
	userBranchID *int64,
	userFacultyID *int64,
) (*domain.Administrator, error) {
	return s.addAdministratorWithPermissionCheckUC.Execute(
		chatID,
		phone,
		userRole,
		userUniversityID,
		userBranchID,
		userFacultyID,
	)
}

// RemoveAdministrator удаляет администратора из чата
// Нельзя удалить последнего администратора (должно быть минимум 2)
func (s *ChatService) RemoveAdministrator(adminID int64) error {
	return s.removeAdministratorWithValidationUC.Execute(adminID)
}

// GetAdministratorByID получает администратора по ID
func (s *ChatService) GetAdministratorByID(id int64) (*domain.Administrator, error) {
	admin, err := s.administratorRepo.GetByID(id)
	if err != nil {
		return nil, domain.ErrAdministratorNotFound
	}
	return admin, nil
}

// GetAllAdministrators получает всех администраторов с пагинацией и поиском
func (s *ChatService) GetAllAdministrators(query string, limit, offset int) ([]*domain.Administrator, int, error) {
	return s.administratorRepo.GetAll(query, limit, offset)
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

// CreateOrGetUniversity создает или получает университет по INN/KPP
func (s *ChatService) CreateOrGetUniversity(inn, kpp, name string) (*domain.University, error) {
	// Пытаемся найти существующий университет
	university, err := s.universityRepo.GetByINNAndKPP(inn, kpp)
	if err == nil {
		return university, nil
	}

	// Если не найден, создаем новый
	university = &domain.University{
		INN:  inn,
		KPP:  kpp,
		Name: name,
	}

	if err := s.universityRepo.Create(university); err != nil {
		return nil, err
	}

	return university, nil
}

