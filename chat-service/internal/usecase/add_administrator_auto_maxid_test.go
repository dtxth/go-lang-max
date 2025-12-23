package usecase

import (
	"chat-service/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddAdministratorWithFlags_AutoMaxID_Success(t *testing.T) {
	// Setup
	chatRepo := &mockChatRepoForAdd{
		chats: map[int64]*domain.Chat{
			1: {ID: 1, Name: "Test Chat"},
		},
	}
	
	adminRepo := &mockAdminRepoForAdd{
		admins:       make(map[int64]*domain.Administrator),
		phoneToAdmin: make(map[string]map[int64]*domain.Administrator),
	}
	
	maxService := newMockMaxServiceForAdd()

	chatService := &ChatService{
		chatRepo:          chatRepo,
		administratorRepo: adminRepo,
		maxService:        maxService,
	}

	chatID := int64(1)
	phone := "+79001234567"
	expectedUserID := int64(123456789)

	// Настраиваем мок для GetInternalUsers
	maxService.internalUsers[phone] = []*domain.InternalUser{
		{
			UserID:      expectedUserID,
			FirstName:   "Иван",
			LastName:    "Иванов",
			PhoneNumber: phone,
		},
	}

	// Execute
	admin, err := chatService.AddAdministratorWithFlags(chatID, phone, "", true, true, false)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, admin)
	assert.Equal(t, chatID, admin.ChatID)
	assert.Equal(t, phone, admin.Phone)
	assert.Equal(t, "123456789", admin.MaxID)
	assert.True(t, admin.AddUser)
	assert.True(t, admin.AddAdmin)
}

func TestAddAdministratorWithFlags_AutoMaxID_UserNotFound(t *testing.T) {
	// Setup
	chatRepo := &mockChatRepoForAdd{
		chats: map[int64]*domain.Chat{
			1: {ID: 1, Name: "Test Chat"},
		},
	}
	
	adminRepo := &mockAdminRepoForAdd{
		admins:       make(map[int64]*domain.Administrator),
		phoneToAdmin: make(map[string]map[int64]*domain.Administrator),
	}
	
	maxService := newMockMaxServiceForAdd()

	chatService := &ChatService{
		chatRepo:          chatRepo,
		administratorRepo: adminRepo,
		maxService:        maxService,
	}

	chatID := int64(1)
	phone := "+79001234567"

	// Настраиваем мок для GetInternalUsers - пользователь не найден
	maxService.internalUsers[phone] = []*domain.InternalUser{} // пустой список
	maxService.failedPhones[phone] = []string{phone}           // телефон в списке неудачных

	// Execute
	admin, err := chatService.AddAdministratorWithFlags(chatID, phone, "", true, true, false)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrMaxIDNotFound, err)
	assert.Nil(t, admin)
}

func TestAddAdministratorWithFlags_ExplicitMaxID_SkipsAutoRetrieval(t *testing.T) {
	// Setup
	chatRepo := &mockChatRepoForAdd{
		chats: map[int64]*domain.Chat{
			1: {ID: 1, Name: "Test Chat"},
		},
	}
	
	adminRepo := &mockAdminRepoForAdd{
		admins:       make(map[int64]*domain.Administrator),
		phoneToAdmin: make(map[string]map[int64]*domain.Administrator),
	}
	
	maxService := newMockMaxServiceForAdd()

	chatService := &ChatService{
		chatRepo:          chatRepo,
		administratorRepo: adminRepo,
		maxService:        maxService,
	}

	chatID := int64(1)
	phone := "+79001234567"
	explicitMaxID := "987654321"

	// Execute
	admin, err := chatService.AddAdministratorWithFlags(chatID, phone, explicitMaxID, true, true, false)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, admin)
	assert.Equal(t, chatID, admin.ChatID)
	assert.Equal(t, phone, admin.Phone)
	assert.Equal(t, explicitMaxID, admin.MaxID)
	assert.True(t, admin.AddUser)
	assert.True(t, admin.AddAdmin)
}