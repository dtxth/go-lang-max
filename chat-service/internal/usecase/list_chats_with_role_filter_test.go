package usecase

import (
	"chat-service/internal/domain"
	"errors"
	"testing"
)

// MockChatRepository для тестирования
type MockChatRepository struct {
	searchFunc func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error)
}

func (m *MockChatRepository) Create(chat *domain.Chat) error {
	return nil
}

func (m *MockChatRepository) GetByID(id int64) (*domain.Chat, error) {
	return nil, nil
}

func (m *MockChatRepository) GetByMaxChatID(maxChatID string) (*domain.Chat, error) {
	return nil, nil
}

func (m *MockChatRepository) Search(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	if m.searchFunc != nil {
		return m.searchFunc(query, limit, offset, filter)
	}
	return nil, 0, nil
}

func (m *MockChatRepository) GetAll(limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return m.Search("", limit, offset, filter)
}

func (m *MockChatRepository) Update(chat *domain.Chat) error {
	return nil
}

func (m *MockChatRepository) Delete(id int64) error {
	return nil
}

func (m *MockChatRepository) GetAllWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return m.Search(search, limit, offset, filter)
}

func TestListChatsWithRoleFilterUseCase_Execute_Superadmin(t *testing.T) {
	// Arrange
	universityID := int64(1)
	mockRepo := &MockChatRepository{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			// Проверяем, что фильтр для суперадмина не ограничивает по университету
			if filter != nil && !filter.IsSuperadmin() {
				t.Error("Expected superadmin filter")
			}
			return []*domain.Chat{
				{ID: 1, Name: "Chat 1", UniversityID: &universityID},
				{ID: 2, Name: "Chat 2", UniversityID: nil},
			}, 2, nil
		},
	}

	uc := NewListChatsWithRoleFilterUseCase(mockRepo)

	filter := &domain.ChatFilter{
		Role: "superadmin",
	}

	// Act
	chats, total, err := uc.Execute("", 50, 0, filter)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}
	if len(chats) != 2 {
		t.Errorf("Expected 2 chats, got %d", len(chats))
	}
}

func TestListChatsWithRoleFilterUseCase_Execute_Curator(t *testing.T) {
	// Arrange
	universityID := int64(1)
	mockRepo := &MockChatRepository{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			// Проверяем, что фильтр для куратора ограничивает по университету
			if filter == nil || !filter.IsCurator() {
				t.Error("Expected curator filter")
			}
			if filter.UniversityID == nil || *filter.UniversityID != universityID {
				t.Errorf("Expected university_id %d, got %v", universityID, filter.UniversityID)
			}
			return []*domain.Chat{
				{ID: 1, Name: "Chat 1", UniversityID: &universityID},
			}, 1, nil
		},
	}

	uc := NewListChatsWithRoleFilterUseCase(mockRepo)

	filter := &domain.ChatFilter{
		Role:         "curator",
		UniversityID: &universityID,
	}

	// Act
	chats, total, err := uc.Execute("", 50, 0, filter)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if total != 1 {
		t.Errorf("Expected total 1, got %d", total)
	}
	if len(chats) != 1 {
		t.Errorf("Expected 1 chat, got %d", len(chats))
	}
}

func TestListChatsWithRoleFilterUseCase_Execute_CuratorWithoutUniversityID(t *testing.T) {
	// Arrange
	mockRepo := &MockChatRepository{}
	uc := NewListChatsWithRoleFilterUseCase(mockRepo)

	filter := &domain.ChatFilter{
		Role:         "curator",
		UniversityID: nil, // Куратор без university_id
	}

	// Act
	_, _, err := uc.Execute("", 50, 0, filter)

	// Assert
	if err != domain.ErrForbidden {
		t.Errorf("Expected ErrForbidden, got %v", err)
	}
}

func TestListChatsWithRoleFilterUseCase_Execute_NilFilter(t *testing.T) {
	// Arrange
	mockRepo := &MockChatRepository{}
	uc := NewListChatsWithRoleFilterUseCase(mockRepo)

	// Act
	_, _, err := uc.Execute("", 50, 0, nil)

	// Assert
	if err != domain.ErrInvalidRole {
		t.Errorf("Expected ErrInvalidRole, got %v", err)
	}
}

func TestListChatsWithRoleFilterUseCase_Execute_PaginationDefaults(t *testing.T) {
	// Arrange
	mockRepo := &MockChatRepository{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			// Проверяем, что применяются значения по умолчанию
			if limit != 50 {
				t.Errorf("Expected default limit 50, got %d", limit)
			}
			if offset != 0 {
				t.Errorf("Expected default offset 0, got %d", offset)
			}
			return []*domain.Chat{}, 0, nil
		},
	}

	uc := NewListChatsWithRoleFilterUseCase(mockRepo)

	filter := &domain.ChatFilter{
		Role: "superadmin",
	}

	// Act
	_, _, err := uc.Execute("", 0, -1, filter)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestListChatsWithRoleFilterUseCase_Execute_PaginationLimitCap(t *testing.T) {
	// Arrange
	mockRepo := &MockChatRepository{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			// Проверяем, что лимит ограничен 100
			if limit != 100 {
				t.Errorf("Expected limit capped at 100, got %d", limit)
			}
			return []*domain.Chat{}, 0, nil
		},
	}

	uc := NewListChatsWithRoleFilterUseCase(mockRepo)

	filter := &domain.ChatFilter{
		Role: "superadmin",
	}

	// Act
	_, _, err := uc.Execute("", 200, 0, filter)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestListChatsWithRoleFilterUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	expectedError := errors.New("database error")
	mockRepo := &MockChatRepository{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			return nil, 0, expectedError
		},
	}

	uc := NewListChatsWithRoleFilterUseCase(mockRepo)

	filter := &domain.ChatFilter{
		Role: "superadmin",
	}

	// Act
	_, _, err := uc.Execute("", 50, 0, filter)

	// Assert
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}
