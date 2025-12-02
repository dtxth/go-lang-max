package usecase

import (
	"chat-service/internal/domain"
)

// RemoveAdministratorWithValidationUseCase удаляет администратора из чата с валидацией
type RemoveAdministratorWithValidationUseCase struct {
	administratorRepo domain.AdministratorRepository
	chatRepo          domain.ChatRepository
}

// NewRemoveAdministratorWithValidationUseCase создает новый use case для удаления администратора
func NewRemoveAdministratorWithValidationUseCase(
	administratorRepo domain.AdministratorRepository,
	chatRepo domain.ChatRepository,
) *RemoveAdministratorWithValidationUseCase {
	return &RemoveAdministratorWithValidationUseCase{
		administratorRepo: administratorRepo,
		chatRepo:          chatRepo,
	}
}

// Execute удаляет администратора с проверкой, что он не последний
// Validates: Requirements 6.3, 6.4
func (uc *RemoveAdministratorWithValidationUseCase) Execute(adminID int64) error {
	// Получаем администратора
	admin, err := uc.administratorRepo.GetByID(adminID)
	if err != nil {
		return domain.ErrAdministratorNotFound
	}

	// Проверяем существование чата
	_, err = uc.chatRepo.GetByID(admin.ChatID)
	if err != nil {
		return domain.ErrChatNotFound
	}

	// Проверяем количество администраторов у чата
	count, err := uc.administratorRepo.CountByChatID(admin.ChatID)
	if err != nil {
		return err
	}

	// Нельзя удалить последнего администратора
	if count <= 1 {
		return domain.ErrCannotDeleteLastAdmin
	}

	// Удаляем администратора
	return uc.administratorRepo.Delete(adminID)
}
