package usecase

import (
	"chat-service/internal/domain"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock repositories for testing
type mockAdminRepoForRemove struct {
	admins map[int64]*domain.Administrator
	counts map[int64]int
}

func (m *mockAdminRepoForRemove) GetByID(id int64) (*domain.Administrator, error) {
	admin, ok := m.admins[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return admin, nil
}

func (m *mockAdminRepoForRemove) CountByChatID(chatID int64) (int, error) {
	count, ok := m.counts[chatID]
	if !ok {
		return 0, nil
	}
	return count, nil
}

func (m *mockAdminRepoForRemove) Delete(id int64) error {
	delete(m.admins, id)
	return nil
}

func (m *mockAdminRepoForRemove) Create(admin *domain.Administrator) error {
	return nil
}

func (m *mockAdminRepoForRemove) GetByChatID(chatID int64) ([]*domain.Administrator, error) {
	return nil, nil
}

func (m *mockAdminRepoForRemove) GetByPhoneAndChatID(phone string, chatID int64) (*domain.Administrator, error) {
	return nil, nil
}

type mockChatRepoForRemove struct {
	chats map[int64]*domain.Chat
}

func (m *mockChatRepoForRemove) GetByID(id int64) (*domain.Chat, error) {
	chat, ok := m.chats[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return chat, nil
}

func (m *mockChatRepoForRemove) Create(chat *domain.Chat) error {
	return nil
}

func (m *mockChatRepoForRemove) GetByMaxChatID(maxChatID string) (*domain.Chat, error) {
	return nil, nil
}

func (m *mockChatRepoForRemove) Search(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return nil, 0, nil
}

func (m *mockChatRepoForRemove) GetAll(limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return nil, 0, nil
}

func (m *mockChatRepoForRemove) Update(chat *domain.Chat) error {
	return nil
}

func (m *mockChatRepoForRemove) Delete(id int64) error {
	return nil
}

func TestRemoveAdministratorWithValidation_Success(t *testing.T) {
	// Setup
	chatID := int64(1)
	adminID := int64(100)

	adminRepo := &mockAdminRepoForRemove{
		admins: map[int64]*domain.Administrator{
			adminID: {
				ID:     adminID,
				ChatID: chatID,
				Phone:  "+79991234567",
				MaxID:  "max123",
			},
		},
		counts: map[int64]int{
			chatID: 2, // 2 administrators, so we can delete one
		},
	}

	chatRepo := &mockChatRepoForRemove{
		chats: map[int64]*domain.Chat{
			chatID: {
				ID:   chatID,
				Name: "Test Chat",
			},
		},
	}

	uc := NewRemoveAdministratorWithValidationUseCase(adminRepo, chatRepo)

	// Execute
	err := uc.Execute(adminID)

	// Assert
	assert.NoError(t, err)
	_, exists := adminRepo.admins[adminID]
	assert.False(t, exists, "Administrator should be deleted")
}

func TestRemoveAdministratorWithValidation_AdminNotFound(t *testing.T) {
	// Setup
	adminRepo := &mockAdminRepoForRemove{
		admins: map[int64]*domain.Administrator{},
		counts: map[int64]int{},
	}

	chatRepo := &mockChatRepoForRemove{
		chats: map[int64]*domain.Chat{},
	}

	uc := NewRemoveAdministratorWithValidationUseCase(adminRepo, chatRepo)

	// Execute
	err := uc.Execute(999)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrAdministratorNotFound, err)
}

func TestRemoveAdministratorWithValidation_ChatNotFound(t *testing.T) {
	// Setup
	adminID := int64(100)
	chatID := int64(1)

	adminRepo := &mockAdminRepoForRemove{
		admins: map[int64]*domain.Administrator{
			adminID: {
				ID:     adminID,
				ChatID: chatID,
				Phone:  "+79991234567",
				MaxID:  "max123",
			},
		},
		counts: map[int64]int{},
	}

	chatRepo := &mockChatRepoForRemove{
		chats: map[int64]*domain.Chat{}, // Chat doesn't exist
	}

	uc := NewRemoveAdministratorWithValidationUseCase(adminRepo, chatRepo)

	// Execute
	err := uc.Execute(adminID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrChatNotFound, err)
}

func TestRemoveAdministratorWithValidation_CannotDeleteLastAdmin(t *testing.T) {
	// Setup
	chatID := int64(1)
	adminID := int64(100)

	adminRepo := &mockAdminRepoForRemove{
		admins: map[int64]*domain.Administrator{
			adminID: {
				ID:     adminID,
				ChatID: chatID,
				Phone:  "+79991234567",
				MaxID:  "max123",
			},
		},
		counts: map[int64]int{
			chatID: 1, // Only 1 administrator
		},
	}

	chatRepo := &mockChatRepoForRemove{
		chats: map[int64]*domain.Chat{
			chatID: {
				ID:   chatID,
				Name: "Test Chat",
			},
		},
	}

	uc := NewRemoveAdministratorWithValidationUseCase(adminRepo, chatRepo)

	// Execute
	err := uc.Execute(adminID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCannotDeleteLastAdmin, err)
	_, exists := adminRepo.admins[adminID]
	assert.True(t, exists, "Administrator should not be deleted")
}
