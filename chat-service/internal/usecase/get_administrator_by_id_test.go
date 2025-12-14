package usecase

import (
	"chat-service/internal/domain"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockAdminRepoForGetByID struct {
	admins map[int64]*domain.Administrator
}

func (m *mockAdminRepoForGetByID) GetByID(id int64) (*domain.Administrator, error) {
	admin, ok := m.admins[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return admin, nil
}

func (m *mockAdminRepoForGetByID) Create(admin *domain.Administrator) error {
	return nil
}

func (m *mockAdminRepoForGetByID) GetByChatID(chatID int64) ([]*domain.Administrator, error) {
	return nil, nil
}

func (m *mockAdminRepoForGetByID) GetByPhoneAndChatID(phone string, chatID int64) (*domain.Administrator, error) {
	return nil, nil
}

func (m *mockAdminRepoForGetByID) Delete(id int64) error {
	return nil
}

func (m *mockAdminRepoForGetByID) CountByChatID(chatID int64) (int, error) {
	return 0, nil
}

func (m *mockAdminRepoForGetByID) GetAll(query string, limit, offset int) ([]*domain.Administrator, int, error) {
	return nil, 0, nil
}

func TestGetAdministratorByID_Success(t *testing.T) {
	// Setup
	adminID := int64(100)
	expectedAdmin := &domain.Administrator{
		ID:       adminID,
		ChatID:   1,
		Phone:    "+79991234567",
		MaxID:    "max123",
		AddUser:  true,
		AddAdmin: true,
	}

	adminRepo := &mockAdminRepoForGetByID{
		admins: map[int64]*domain.Administrator{
			adminID: expectedAdmin,
		},
	}

	chatService := &ChatService{
		administratorRepo: adminRepo,
	}

	// Execute
	admin, err := chatService.GetAdministratorByID(adminID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, admin)
	assert.Equal(t, expectedAdmin.ID, admin.ID)
	assert.Equal(t, expectedAdmin.Phone, admin.Phone)
	assert.Equal(t, expectedAdmin.MaxID, admin.MaxID)
}

func TestGetAdministratorByID_NotFound(t *testing.T) {
	// Setup
	adminRepo := &mockAdminRepoForGetByID{
		admins: map[int64]*domain.Administrator{},
	}

	chatService := &ChatService{
		administratorRepo: adminRepo,
	}

	// Execute
	admin, err := chatService.GetAdministratorByID(999)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, admin)
	assert.Equal(t, domain.ErrAdministratorNotFound, err)
}
