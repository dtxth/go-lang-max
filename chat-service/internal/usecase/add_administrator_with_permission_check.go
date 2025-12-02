package usecase

import (
	"chat-service/internal/domain"
)

// AddAdministratorWithPermissionCheckUseCase добавляет администратора к чату с проверкой прав
type AddAdministratorWithPermissionCheckUseCase struct {
	administratorRepo domain.AdministratorRepository
	chatRepo          domain.ChatRepository
	maxService        domain.MaxService
}

// NewAddAdministratorWithPermissionCheckUseCase создает новый use case для добавления администратора
func NewAddAdministratorWithPermissionCheckUseCase(
	administratorRepo domain.AdministratorRepository,
	chatRepo domain.ChatRepository,
	maxService domain.MaxService,
) *AddAdministratorWithPermissionCheckUseCase {
	return &AddAdministratorWithPermissionCheckUseCase{
		administratorRepo: administratorRepo,
		chatRepo:          chatRepo,
		maxService:        maxService,
	}
}

// Execute добавляет администратора с проверкой прав доступа
// Validates: Requirements 6.1, 6.2
func (uc *AddAdministratorWithPermissionCheckUseCase) Execute(
	chatID int64,
	phone string,
	userRole string,
	userUniversityID *int64,
	userBranchID *int64,
	userFacultyID *int64,
) (*domain.Administrator, error) {
	// Валидация телефона
	if !uc.maxService.ValidatePhone(phone) {
		return nil, domain.ErrInvalidPhone
	}

	// Проверяем существование чата
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return nil, domain.ErrChatNotFound
	}

	// Проверяем права доступа пользователя к чату
	if err := uc.checkPermission(chat, userRole, userUniversityID, userBranchID, userFacultyID); err != nil {
		return nil, err
	}

	// Проверяем, не существует ли уже администратор с таким телефоном в этом чате
	existing, _ := uc.administratorRepo.GetByPhoneAndChatID(phone, chatID)
	if existing != nil {
		return nil, domain.ErrAdministratorExists
	}

	// Получаем MAX_id по телефону через MaxBot Service
	maxID, err := uc.maxService.GetMaxIDByPhone(phone)
	if err != nil {
		return nil, err
	}

	// Создаем администратора
	admin := &domain.Administrator{
		ChatID: chatID,
		Phone:  phone,
		MaxID:  maxID,
	}

	if err := uc.administratorRepo.Create(admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// checkPermission проверяет, имеет ли пользователь право добавлять администратора к чату
func (uc *AddAdministratorWithPermissionCheckUseCase) checkPermission(
	chat *domain.Chat,
	userRole string,
	userUniversityID *int64,
	userBranchID *int64,
	userFacultyID *int64,
) error {
	// Суперадмин имеет доступ ко всем чатам
	if userRole == "superadmin" {
		return nil
	}

	// Куратор имеет доступ только к чатам своего вуза
	if userRole == "curator" {
		if userUniversityID == nil {
			return domain.ErrForbidden
		}
		if chat.UniversityID == nil || *chat.UniversityID != *userUniversityID {
			return domain.ErrForbidden
		}
		return nil
	}

	// Оператор имеет доступ только к чатам своего подразделения
	if userRole == "operator" {
		if userUniversityID == nil {
			return domain.ErrForbidden
		}
		if chat.UniversityID == nil || *chat.UniversityID != *userUniversityID {
			return domain.ErrForbidden
		}
		// TODO: В будущем добавить проверку branch_id и faculty_id
		return nil
	}

	// Неизвестная роль или отсутствие роли
	return domain.ErrForbidden
}
