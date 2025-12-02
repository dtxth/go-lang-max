package usecase

import (
	"chat-service/internal/domain"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock repositories for testing
type mockAdminRepoForAdd struct {
	admins         map[int64]*domain.Administrator
	phoneToAdmin   map[string]map[int64]*domain.Administrator // phone -> chatID -> admin
	nextID         int64
	createError    error
}

func (m *mockAdminRepoForAdd) Create(admin *domain.Administrator) error {
	if m.createError != nil {
		return m.createError
	}
	m.nextID++
	admin.ID = m.nextID
	m.admins[admin.ID] = admin
	
	if m.phoneToAdmin == nil {
		m.phoneToAdmin = make(map[string]map[int64]*domain.Administrator)
	}
	if m.phoneToAdmin[admin.Phone] == nil {
		m.phoneToAdmin[admin.Phone] = make(map[int64]*domain.Administrator)
	}
	m.phoneToAdmin[admin.Phone][admin.ChatID] = admin
	
	return nil
}

func (m *mockAdminRepoForAdd) GetByPhoneAndChatID(phone string, chatID int64) (*domain.Administrator, error) {
	if m.phoneToAdmin == nil {
		return nil, sql.ErrNoRows
	}
	if chatAdmins, ok := m.phoneToAdmin[phone]; ok {
		if admin, ok := chatAdmins[chatID]; ok {
			return admin, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockAdminRepoForAdd) GetByID(id int64) (*domain.Administrator, error) {
	return nil, nil
}

func (m *mockAdminRepoForAdd) GetByChatID(chatID int64) ([]*domain.Administrator, error) {
	return nil, nil
}

func (m *mockAdminRepoForAdd) Delete(id int64) error {
	return nil
}

func (m *mockAdminRepoForAdd) CountByChatID(chatID int64) (int, error) {
	return 0, nil
}

type mockChatRepoForAdd struct {
	chats map[int64]*domain.Chat
}

func (m *mockChatRepoForAdd) GetByID(id int64) (*domain.Chat, error) {
	chat, ok := m.chats[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return chat, nil
}

func (m *mockChatRepoForAdd) Create(chat *domain.Chat) error {
	return nil
}

func (m *mockChatRepoForAdd) GetByMaxChatID(maxChatID string) (*domain.Chat, error) {
	return nil, nil
}

func (m *mockChatRepoForAdd) Search(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return nil, 0, nil
}

func (m *mockChatRepoForAdd) GetAll(limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	return nil, 0, nil
}

func (m *mockChatRepoForAdd) Update(chat *domain.Chat) error {
	return nil
}

func (m *mockChatRepoForAdd) Delete(id int64) error {
	return nil
}

type mockMaxServiceForAdd struct {
	maxIDs       map[string]string // phone -> maxID
	validateFunc func(string) bool
	getMaxIDFunc func(string) (string, error)
}

func (m *mockMaxServiceForAdd) ValidatePhone(phone string) bool {
	if m.validateFunc != nil {
		return m.validateFunc(phone)
	}
	return true
}

func (m *mockMaxServiceForAdd) GetMaxIDByPhone(phone string) (string, error) {
	if m.getMaxIDFunc != nil {
		return m.getMaxIDFunc(phone)
	}
	if maxID, ok := m.maxIDs[phone]; ok {
		return maxID, nil
	}
	return "", domain.ErrMaxIDNotFound
}

func TestAddAdministratorWithPermissionCheck_Superadmin_Success(t *testing.T) {
	// Setup
	chatID := int64(1)
	universityID := int64(10)
	phone := "+79991234567"

	adminRepo := &mockAdminRepoForAdd{
		admins:       make(map[int64]*domain.Administrator),
		phoneToAdmin: make(map[string]map[int64]*domain.Administrator),
		nextID:       0,
	}

	chatRepo := &mockChatRepoForAdd{
		chats: map[int64]*domain.Chat{
			chatID: {
				ID:           chatID,
				Name:         "Test Chat",
				UniversityID: &universityID,
			},
		},
	}

	maxService := &mockMaxServiceForAdd{
		maxIDs: map[string]string{
			phone: "max123",
		},
	}

	uc := NewAddAdministratorWithPermissionCheckUseCase(adminRepo, chatRepo, maxService)

	// Execute - superadmin can add to any chat
	admin, err := uc.Execute(chatID, phone, "superadmin", nil, nil, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, admin)
	assert.Equal(t, chatID, admin.ChatID)
	assert.Equal(t, phone, admin.Phone)
	assert.Equal(t, "max123", admin.MaxID)
}

func TestAddAdministratorWithPermissionCheck_Curator_Success(t *testing.T) {
	// Setup
	chatID := int64(1)
	universityID := int64(10)
	phone := "+79991234567"

	adminRepo := &mockAdminRepoForAdd{
		admins:       make(map[int64]*domain.Administrator),
		phoneToAdmin: make(map[string]map[int64]*domain.Administrator),
		nextID:       0,
	}

	chatRepo := &mockChatRepoForAdd{
		chats: map[int64]*domain.Chat{
			chatID: {
				ID:           chatID,
				Name:         "Test Chat",
				UniversityID: &universityID,
			},
		},
	}

	maxService := &mockMaxServiceForAdd{
		maxIDs: map[string]string{
			phone: "max123",
		},
	}

	uc := NewAddAdministratorWithPermissionCheckUseCase(adminRepo, chatRepo, maxService)

	// Execute - curator can add to their university's chat
	admin, err := uc.Execute(chatID, phone, "curator", &universityID, nil, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, admin)
	assert.Equal(t, chatID, admin.ChatID)
	assert.Equal(t, phone, admin.Phone)
	assert.Equal(t, "max123", admin.MaxID)
}

func TestAddAdministratorWithPermissionCheck_Curator_WrongUniversity(t *testing.T) {
	// Setup
	chatID := int64(1)
	chatUniversityID := int64(10)
	curatorUniversityID := int64(20) // Different university
	phone := "+79991234567"

	adminRepo := &mockAdminRepoForAdd{
		admins:       make(map[int64]*domain.Administrator),
		phoneToAdmin: make(map[string]map[int64]*domain.Administrator),
		nextID:       0,
	}

	chatRepo := &mockChatRepoForAdd{
		chats: map[int64]*domain.Chat{
			chatID: {
				ID:           chatID,
				Name:         "Test Chat",
				UniversityID: &chatUniversityID,
			},
		},
	}

	maxService := &mockMaxServiceForAdd{
		maxIDs: map[string]string{
			phone: "max123",
		},
	}

	uc := NewAddAdministratorWithPermissionCheckUseCase(adminRepo, chatRepo, maxService)

	// Execute - curator cannot add to different university's chat
	admin, err := uc.Execute(chatID, phone, "curator", &curatorUniversityID, nil, nil)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrForbidden, err)
	assert.Nil(t, admin)
}

func TestAddAdministratorWithPermissionCheck_Operator_Success(t *testing.T) {
	// Setup
	chatID := int64(1)
	universityID := int64(10)
	phone := "+79991234567"

	adminRepo := &mockAdminRepoForAdd{
		admins:       make(map[int64]*domain.Administrator),
		phoneToAdmin: make(map[string]map[int64]*domain.Administrator),
		nextID:       0,
	}

	chatRepo := &mockChatRepoForAdd{
		chats: map[int64]*domain.Chat{
			chatID: {
				ID:           chatID,
				Name:         "Test Chat",
				UniversityID: &universityID,
			},
		},
	}

	maxService := &mockMaxServiceForAdd{
		maxIDs: map[string]string{
			phone: "max123",
		},
	}

	uc := NewAddAdministratorWithPermissionCheckUseCase(adminRepo, chatRepo, maxService)

	// Execute - operator can add to their university's chat
	admin, err := uc.Execute(chatID, phone, "operator", &universityID, nil, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, admin)
}

func TestAddAdministratorWithPermissionCheck_InvalidPhone(t *testing.T) {
	// Setup
	chatID := int64(1)
	universityID := int64(10)
	phone := "invalid"

	adminRepo := &mockAdminRepoForAdd{
		admins:       make(map[int64]*domain.Administrator),
		phoneToAdmin: make(map[string]map[int64]*domain.Administrator),
		nextID:       0,
	}

	chatRepo := &mockChatRepoForAdd{
		chats: map[int64]*domain.Chat{
			chatID: {
				ID:           chatID,
				Name:         "Test Chat",
				UniversityID: &universityID,
			},
		},
	}

	maxService := &mockMaxServiceForAdd{
		validateFunc: func(p string) bool {
			return false // Invalid phone
		},
	}

	uc := NewAddAdministratorWithPermissionCheckUseCase(adminRepo, chatRepo, maxService)

	// Execute
	admin, err := uc.Execute(chatID, phone, "superadmin", nil, nil, nil)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidPhone, err)
	assert.Nil(t, admin)
}

func TestAddAdministratorWithPermissionCheck_ChatNotFound(t *testing.T) {
	// Setup
	chatID := int64(999) // Non-existent chat
	phone := "+79991234567"

	adminRepo := &mockAdminRepoForAdd{
		admins:       make(map[int64]*domain.Administrator),
		phoneToAdmin: make(map[string]map[int64]*domain.Administrator),
		nextID:       0,
	}

	chatRepo := &mockChatRepoForAdd{
		chats: map[int64]*domain.Chat{},
	}

	maxService := &mockMaxServiceForAdd{
		maxIDs: map[string]string{
			phone: "max123",
		},
	}

	uc := NewAddAdministratorWithPermissionCheckUseCase(adminRepo, chatRepo, maxService)

	// Execute
	admin, err := uc.Execute(chatID, phone, "superadmin", nil, nil, nil)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrChatNotFound, err)
	assert.Nil(t, admin)
}

func TestAddAdministratorWithPermissionCheck_AdministratorExists(t *testing.T) {
	// Setup
	chatID := int64(1)
	universityID := int64(10)
	phone := "+79991234567"

	adminRepo := &mockAdminRepoForAdd{
		admins: make(map[int64]*domain.Administrator),
		phoneToAdmin: map[string]map[int64]*domain.Administrator{
			phone: {
				chatID: {
					ID:     1,
					ChatID: chatID,
					Phone:  phone,
					MaxID:  "existing_max",
				},
			},
		},
		nextID: 1,
	}

	chatRepo := &mockChatRepoForAdd{
		chats: map[int64]*domain.Chat{
			chatID: {
				ID:           chatID,
				Name:         "Test Chat",
				UniversityID: &universityID,
			},
		},
	}

	maxService := &mockMaxServiceForAdd{
		maxIDs: map[string]string{
			phone: "max123",
		},
	}

	uc := NewAddAdministratorWithPermissionCheckUseCase(adminRepo, chatRepo, maxService)

	// Execute
	admin, err := uc.Execute(chatID, phone, "superadmin", nil, nil, nil)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrAdministratorExists, err)
	assert.Nil(t, admin)
}

func TestAddAdministratorWithPermissionCheck_MaxIDNotFound(t *testing.T) {
	// Setup
	chatID := int64(1)
	universityID := int64(10)
	phone := "+79991234567"

	adminRepo := &mockAdminRepoForAdd{
		admins:       make(map[int64]*domain.Administrator),
		phoneToAdmin: make(map[string]map[int64]*domain.Administrator),
		nextID:       0,
	}

	chatRepo := &mockChatRepoForAdd{
		chats: map[int64]*domain.Chat{
			chatID: {
				ID:           chatID,
				Name:         "Test Chat",
				UniversityID: &universityID,
			},
		},
	}

	maxService := &mockMaxServiceForAdd{
		getMaxIDFunc: func(p string) (string, error) {
			return "", domain.ErrMaxIDNotFound
		},
	}

	uc := NewAddAdministratorWithPermissionCheckUseCase(adminRepo, chatRepo, maxService)

	// Execute
	admin, err := uc.Execute(chatID, phone, "superadmin", nil, nil, nil)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrMaxIDNotFound, err)
	assert.Nil(t, admin)
}
