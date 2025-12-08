package usecase

import (
	"chat-service/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockAdminRepoForGetAll struct {
	administrators []*domain.Administrator
	totalCount     int
}

func (m *mockAdminRepoForGetAll) GetByID(id int64) (*domain.Administrator, error) {
	return nil, nil
}

func (m *mockAdminRepoForGetAll) Create(admin *domain.Administrator) error {
	return nil
}

func (m *mockAdminRepoForGetAll) GetByChatID(chatID int64) ([]*domain.Administrator, error) {
	return nil, nil
}

func (m *mockAdminRepoForGetAll) GetByPhoneAndChatID(phone string, chatID int64) (*domain.Administrator, error) {
	return nil, nil
}

func (m *mockAdminRepoForGetAll) Delete(id int64) error {
	return nil
}

func (m *mockAdminRepoForGetAll) CountByChatID(chatID int64) (int, error) {
	return 0, nil
}

func (m *mockAdminRepoForGetAll) GetAll(query string, limit, offset int) ([]*domain.Administrator, int, error) {
	// Фильтруем по query если указан
	if query != "" {
		filtered := []*domain.Administrator{}
		for _, admin := range m.administrators {
			if contains(admin.Phone, query) || contains(admin.MaxID, query) {
				filtered = append(filtered, admin)
			}
		}
		return filtered, len(filtered), nil
	}

	// Применяем пагинацию
	start := offset
	end := offset + limit
	if start > len(m.administrators) {
		return []*domain.Administrator{}, m.totalCount, nil
	}
	if end > len(m.administrators) {
		end = len(m.administrators)
	}

	return m.administrators[start:end], m.totalCount, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0)
}

func TestGetAllAdministrators_Success(t *testing.T) {
	// Setup
	now := time.Now()
	administrators := []*domain.Administrator{
		{
			ID:        1,
			ChatID:    1,
			Phone:     "+79991234567",
			MaxID:     "max123",
			AddUser:   true,
			AddAdmin:  true,
			CreatedAt: now,
		},
		{
			ID:        2,
			ChatID:    2,
			Phone:     "+79997654321",
			MaxID:     "max456",
			AddUser:   true,
			AddAdmin:  false,
			CreatedAt: now,
		},
	}

	adminRepo := &mockAdminRepoForGetAll{
		administrators: administrators,
		totalCount:     2,
	}

	chatService := &ChatService{
		administratorRepo: adminRepo,
	}

	// Execute
	admins, totalCount, err := chatService.GetAllAdministrators("", 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(admins))
	assert.Equal(t, 2, totalCount)
	assert.Equal(t, "+79991234567", admins[0].Phone)
	assert.Equal(t, "+79997654321", admins[1].Phone)
}

func TestGetAllAdministrators_WithPagination(t *testing.T) {
	// Setup
	now := time.Now()
	administrators := []*domain.Administrator{
		{ID: 1, ChatID: 1, Phone: "+79991234567", MaxID: "max1", CreatedAt: now},
		{ID: 2, ChatID: 2, Phone: "+79991234568", MaxID: "max2", CreatedAt: now},
		{ID: 3, ChatID: 3, Phone: "+79991234569", MaxID: "max3", CreatedAt: now},
	}

	adminRepo := &mockAdminRepoForGetAll{
		administrators: administrators,
		totalCount:     3,
	}

	chatService := &ChatService{
		administratorRepo: adminRepo,
	}

	// Execute - получаем первую страницу
	admins, totalCount, err := chatService.GetAllAdministrators("", 2, 0)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(admins))
	assert.Equal(t, 3, totalCount)
	assert.Equal(t, int64(1), admins[0].ID)
	assert.Equal(t, int64(2), admins[1].ID)

	// Execute - получаем вторую страницу
	admins, totalCount, err = chatService.GetAllAdministrators("", 2, 2)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(admins))
	assert.Equal(t, 3, totalCount)
	assert.Equal(t, int64(3), admins[0].ID)
}

func TestGetAllAdministrators_WithSearch(t *testing.T) {
	// Setup
	now := time.Now()
	administrators := []*domain.Administrator{
		{ID: 1, ChatID: 1, Phone: "+79991234567", MaxID: "max123", CreatedAt: now},
		{ID: 2, ChatID: 2, Phone: "+79997654321", MaxID: "max456", CreatedAt: now},
	}

	adminRepo := &mockAdminRepoForGetAll{
		administrators: administrators,
		totalCount:     2,
	}

	chatService := &ChatService{
		administratorRepo: adminRepo,
	}

	// Execute - поиск по телефону
	admins, totalCount, err := chatService.GetAllAdministrators("+79991234567", 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(admins))
	assert.Equal(t, 1, totalCount)
	assert.Equal(t, "+79991234567", admins[0].Phone)
}

func TestGetAllAdministrators_EmptyResult(t *testing.T) {
	// Setup
	adminRepo := &mockAdminRepoForGetAll{
		administrators: []*domain.Administrator{},
		totalCount:     0,
	}

	chatService := &ChatService{
		administratorRepo: adminRepo,
	}

	// Execute
	admins, totalCount, err := chatService.GetAllAdministrators("", 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, len(admins))
	assert.Equal(t, 0, totalCount)
}
