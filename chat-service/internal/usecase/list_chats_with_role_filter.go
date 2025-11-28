package usecase

import (
	"chat-service/internal/domain"
)

// ListChatsWithRoleFilterUseCase реализует получение списка чатов с фильтрацией по роли
type ListChatsWithRoleFilterUseCase struct {
	chatRepo domain.ChatRepository
}

func NewListChatsWithRoleFilterUseCase(chatRepo domain.ChatRepository) *ListChatsWithRoleFilterUseCase {
	return &ListChatsWithRoleFilterUseCase{
		chatRepo: chatRepo,
	}
}

// Execute выполняет получение списка чатов с применением фильтрации по роли
// Фильтрация:
// - Superadmin: видит все чаты из всех университетов
// - Curator: видит только чаты своего университета
// - Operator: видит только чаты своего подразделения (branch/faculty)
func (uc *ListChatsWithRoleFilterUseCase) Execute(
	query string,
	limit, offset int,
	filter *domain.ChatFilter,
) ([]*domain.Chat, int, error) {
	// Валидация фильтра
	if filter == nil {
		return nil, 0, domain.ErrInvalidRole
	}

	// Проверяем, что у куратора и оператора указан university_id
	if (filter.IsCurator() || filter.IsOperator()) && filter.UniversityID == nil {
		return nil, 0, domain.ErrForbidden
	}

	// Применяем пагинацию по умолчанию
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Получаем чаты с применением фильтра
	return uc.chatRepo.Search(query, limit, offset, filter)
}

// GetAll получает все чаты с фильтрацией по роли
func (uc *ListChatsWithRoleFilterUseCase) GetAll(
	limit, offset int,
	filter *domain.ChatFilter,
) ([]*domain.Chat, int, error) {
	return uc.Execute("", limit, offset, filter)
}
