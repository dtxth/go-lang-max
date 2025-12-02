package usecase

import (
	"chat-service/internal/domain"
	"errors"
	"testing"
)

// MockChatRepositoryForSearch is a mock implementation for testing search
type MockChatRepositoryForSearch struct {
	searchFunc func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error)
}

func (m *MockChatRepositoryForSearch) Create(chat *domain.Chat) error {
	return nil
}

func (m *MockChatRepositoryForSearch) GetByID(id int64) (*domain.Chat, error) {
	return nil, nil
}

func (m *MockChatRepositoryForSearch) GetByMaxChatID(maxChatID string) (*domain.Chat, error) {
	return nil, nil
}

func (m *MockChatRepositoryForSearch) Search(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	if m.searchFunc != nil {
		return m.searchFunc(query, limit, offset, filter)
	}
	return nil, 0, nil
}

func (m *MockChatRepositoryForSearch) GetAll(limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return m.Search("", limit, offset, filter)
}

func (m *MockChatRepositoryForSearch) Update(chat *domain.Chat) error {
	return nil
}

func (m *MockChatRepositoryForSearch) Delete(id int64) error {
	return nil
}

func TestSearchChats_EmptyQuery(t *testing.T) {
	// Arrange
	mockRepo := &MockChatRepositoryForSearch{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			if query != "" {
				t.Errorf("Expected empty query, got %s", query)
			}
			return []*domain.Chat{
				{ID: 1, Name: "Chat 1"},
				{ID: 2, Name: "Chat 2"},
			}, 2, nil
		},
	}

	uc := NewListChatsWithRoleFilterUseCase(mockRepo)
	filter := &domain.ChatFilter{Role: "superadmin"}

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

func TestSearchChats_WithQuery(t *testing.T) {
	// Arrange
	mockRepo := &MockChatRepositoryForSearch{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			if query != "математика" {
				t.Errorf("Expected query 'математика', got %s", query)
			}
			return []*domain.Chat{
				{ID: 1, Name: "Группа математика 101"},
			}, 1, nil
		},
	}

	uc := NewListChatsWithRoleFilterUseCase(mockRepo)
	filter := &domain.ChatFilter{Role: "superadmin"}

	// Act
	chats, total, err := uc.Execute("математика", 50, 0, filter)

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
	if chats[0].Name != "Группа математика 101" {
		t.Errorf("Expected chat name 'Группа математика 101', got %s", chats[0].Name)
	}
}

func TestSearchChats_MultiWordQuery(t *testing.T) {
	// Arrange
	mockRepo := &MockChatRepositoryForSearch{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			if query != "группа математика" {
				t.Errorf("Expected query 'группа математика', got %s", query)
			}
			// Simulate multi-word search - only chats with both words
			return []*domain.Chat{
				{ID: 1, Name: "Группа математика 101"},
			}, 1, nil
		},
	}

	uc := NewListChatsWithRoleFilterUseCase(mockRepo)
	filter := &domain.ChatFilter{Role: "superadmin"}

	// Act
	chats, total, err := uc.Execute("группа математика", 50, 0, filter)

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

func TestSearchChats_NoMatches(t *testing.T) {
	// Arrange
	mockRepo := &MockChatRepositoryForSearch{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			// No matches found
			return []*domain.Chat{}, 0, nil
		},
	}

	uc := NewListChatsWithRoleFilterUseCase(mockRepo)
	filter := &domain.ChatFilter{Role: "superadmin"}

	// Act
	chats, total, err := uc.Execute("несуществующий чат", 50, 0, filter)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}
	if len(chats) != 0 {
		t.Errorf("Expected 0 chats, got %d", len(chats))
	}
}

func TestSearchChats_WithRoleFiltering(t *testing.T) {
	// Arrange
	universityID := int64(1)
	mockRepo := &MockChatRepositoryForSearch{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			// Verify filter is applied
			if filter == nil {
				t.Error("Expected filter to be set")
			}
			if filter.Role != "curator" {
				t.Errorf("Expected role 'curator', got %s", filter.Role)
			}
			if filter.UniversityID == nil || *filter.UniversityID != 1 {
				t.Error("Expected university_id to be 1")
			}
			// Return only chats from university 1
			return []*domain.Chat{
				{ID: 1, Name: "Chat from University 1", UniversityID: &universityID},
			}, 1, nil
		},
	}

	uc := NewListChatsWithRoleFilterUseCase(mockRepo)
	filter := &domain.ChatFilter{
		Role:         "curator",
		UniversityID: &universityID,
	}

	// Act
	chats, total, err := uc.Execute("chat", 50, 0, filter)

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

func TestSearchChats_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockChatRepositoryForSearch{
		searchFunc: func(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
			return nil, 0, errors.New("database error")
		},
	}

	uc := NewListChatsWithRoleFilterUseCase(mockRepo)
	filter := &domain.ChatFilter{Role: "superadmin"}

	// Act
	_, _, err := uc.Execute("test", 50, 0, filter)

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
