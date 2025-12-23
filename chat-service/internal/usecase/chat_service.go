package usecase

import (
	"chat-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ChatService struct {
	chatRepo                              domain.ChatRepository
	administratorRepo                     domain.AdministratorRepository
	maxService                            domain.MaxService
	participantsCache                     domain.ParticipantsCache
	participantsUpdater                   domain.ParticipantsUpdater
	participantsConfig                    *domain.ParticipantsConfig
	listChatsWithRoleFilterUC             *ListChatsWithRoleFilterUseCase
	addAdministratorWithPermissionCheckUC *AddAdministratorWithPermissionCheckUseCase
	removeAdministratorWithValidationUC   *RemoveAdministratorWithValidationUseCase
}

func NewChatService(
	chatRepo domain.ChatRepository,
	administratorRepo domain.AdministratorRepository,
	maxService domain.MaxService,
) *ChatService {
	return &ChatService{
		chatRepo:                              chatRepo,
		administratorRepo:                     administratorRepo,
		maxService:                            maxService,
		participantsCache:                     nil,
		participantsUpdater:                   nil,
		participantsConfig:                    nil,
		listChatsWithRoleFilterUC:             NewListChatsWithRoleFilterUseCase(chatRepo),
		addAdministratorWithPermissionCheckUC: NewAddAdministratorWithPermissionCheckUseCase(administratorRepo, chatRepo, maxService),
		removeAdministratorWithValidationUC:   NewRemoveAdministratorWithValidationUseCase(administratorRepo, chatRepo),
	}
}

func NewChatServiceWithParticipants(
	chatRepo domain.ChatRepository,
	administratorRepo domain.AdministratorRepository,
	maxService domain.MaxService,
	participantsCache domain.ParticipantsCache,
	participantsUpdater domain.ParticipantsUpdater,
	participantsConfig *domain.ParticipantsConfig,
) *ChatService {
	return &ChatService{
		chatRepo:                              chatRepo,
		administratorRepo:                     administratorRepo,
		maxService:                            maxService,
		participantsCache:                     participantsCache,
		participantsUpdater:                   participantsUpdater,
		participantsConfig:                    participantsConfig,
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

// GetAllChatsWithSortingAndSearch получает все чаты с пагинацией, сортировкой и поиском
func (s *ChatService) GetAllChatsWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	chats, totalCount, err := s.chatRepo.GetAllWithSortingAndSearch(limit, offset, sortBy, sortOrder, search, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// Обогащаем чаты актуальными данными об участниках с cache-first логикой
	enrichedChats, err := s.enrichChatsWithParticipantsLazy(context.Background(), chats)
	if err != nil {
		// Логируем ошибку, но не прерываем выполнение - возвращаем данные из БД
		// s.logger.Error("Failed to enrich chats with participants", "error", err)
		return chats, totalCount, nil
	}
	
	return enrichedChats, totalCount, nil
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

	// Если MAX_id не передан, получаем его по телефону через GetInternalUsers
	if maxID == "" {
		users, failed, err := s.maxService.GetInternalUsers([]string{phone})
		if err != nil {
			return nil, err
		}

		// Проверяем, что пользователь найден
		if len(users) == 0 {
			if len(failed) > 0 {
				return nil, domain.ErrMaxIDNotFound
			}
			return nil, domain.ErrInvalidPhone
		}

		// Используем UserID как MaxID
		maxID = strings.TrimSpace(fmt.Sprintf("%d", users[0].UserID))
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

// RefreshParticipantsCount принудительно обновляет количество участников для чата
func (s *ChatService) RefreshParticipantsCount(ctx context.Context, chatID int64) (*domain.ParticipantsInfo, error) {
	// Проверяем, что у нас есть ParticipantsUpdater
	if s.participantsUpdater == nil {
		return nil, errors.New("participants updater not available")
	}
	
	// Получаем чат для проверки существования и получения MAX Chat ID
	chat, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		return nil, domain.ErrChatNotFound
	}
	
	// Если нет MAX Chat ID, возвращаем данные из БД как fallback
	if chat.MaxChatID == "" {
		return &domain.ParticipantsInfo{
			Count:     chat.ParticipantsCount,
			UpdatedAt: chat.UpdatedAt,
			Source:    "database",
		}, nil
	}
	
	// Используем ParticipantsUpdater для обновления
	info, err := s.participantsUpdater.UpdateSingle(ctx, chatID, chat.MaxChatID)
	if err != nil {
		// При ошибке возвращаем fallback данные из БД
		return &domain.ParticipantsInfo{
			Count:     chat.ParticipantsCount,
			UpdatedAt: chat.UpdatedAt,
			Source:    "database",
		}, nil
	}
	
	return info, nil
}

// enrichChatsWithParticipantsLazy обогащает чаты актуальными данными об участниках с cache-first логикой
func (s *ChatService) enrichChatsWithParticipantsLazy(ctx context.Context, chats []*domain.Chat) ([]*domain.Chat, error) {
	// Проверяем, что у нас есть все необходимые компоненты для работы с участниками
	if len(chats) == 0 || s.participantsCache == nil || s.participantsConfig == nil || !s.participantsConfig.EnableLazyUpdate {
		return chats, nil
	}
	
	// Собираем ID чатов для батчевого запроса к кэшу
	chatIDs := make([]int64, len(chats))
	chatMap := make(map[int64]*domain.Chat)
	for i, chat := range chats {
		chatIDs[i] = chat.ID
		chatMap[chat.ID] = chat
	}
	
	// Шаг 1: Проверяем кэш для всех чатов
	cachedData, err := s.participantsCache.GetMultiple(ctx, chatIDs)
	if err != nil {
		// При ошибке кэша возвращаем данные из БД как fallback
		// TODO: добавить метрику participants_cache_errors_total
		return chats, nil
	}
	
	// Шаг 2: Определяем стратегию для каждого чата
	chatsToUpdate := make([]domain.ChatUpdateRequest, 0)
	staleThreshold := time.Now().Add(-s.participantsConfig.StaleThreshold)
	
	for _, chat := range chats {
		cachedInfo, exists := cachedData[chat.ID]
		
		if exists && cachedInfo.UpdatedAt.After(staleThreshold) {
			// Данные свежие - используем из кэша
			chat.ParticipantsCount = cachedInfo.Count
		} else if chat.MaxChatID != "" {
			// Данные отсутствуют или устарели - добавляем в список для обновления
			chatsToUpdate = append(chatsToUpdate, domain.ChatUpdateRequest{
				ChatID:    chat.ID,
				MaxChatID: chat.MaxChatID,
			})
		}
		// Если нет MaxChatID, оставляем данные из БД без изменений
	}
	
	// Шаг 3: Обновляем устаревшие данные
	if len(chatsToUpdate) > 0 {
		// Для небольшого количества чатов делаем синхронное обновление
		if len(chatsToUpdate) <= 5 {
			updatedData, err := s.participantsUpdater.UpdateBatch(ctx, chatsToUpdate)
			if err != nil {
				// При ошибке API возвращаем данные из БД/кэша как fallback
				return chats, nil
			}
			
			// Применяем обновленные данные
			for chatID, info := range updatedData {
				if chat, exists := chatMap[chatID]; exists {
					chat.ParticipantsCount = info.Count
				}
			}
		} else {
			// Для большого количества чатов запускаем асинхронное обновление
			go func() {
				updateCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				
				_, err := s.participantsUpdater.UpdateBatch(updateCtx, chatsToUpdate)
				if err != nil {
					// Логирование ошибки уже внутри UpdateBatch
					return
				}
				// Асинхронное обновление не влияет на текущий ответ
			}()
		}
	}
	
	return chats, nil
}



