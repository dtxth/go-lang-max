package usecase

import (
	"chat-service/internal/domain"
	"chat-service/internal/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Моки для тестирования
type MockChatRepositoryForParticipants struct {
	mock.Mock
}

func (m *MockChatRepositoryForParticipants) Create(chat *domain.Chat) error {
	args := m.Called(chat)
	return args.Error(0)
}

func (m *MockChatRepositoryForParticipants) GetByID(id int64) (*domain.Chat, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *MockChatRepositoryForParticipants) GetByMaxChatID(maxChatID string) (*domain.Chat, error) {
	args := m.Called(maxChatID)
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *MockChatRepositoryForParticipants) Search(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	args := m.Called(query, limit, offset, filter)
	return args.Get(0).([]*domain.Chat), args.Int(1), args.Error(2)
}

func (m *MockChatRepositoryForParticipants) GetAll(limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	args := m.Called(limit, offset, filter)
	return args.Get(0).([]*domain.Chat), args.Int(1), args.Error(2)
}

func (m *MockChatRepositoryForParticipants) GetAllWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	args := m.Called(limit, offset, sortBy, sortOrder, search, filter)
	return args.Get(0).([]*domain.Chat), args.Int(1), args.Error(2)
}

func (m *MockChatRepositoryForParticipants) Update(chat *domain.Chat) error {
	args := m.Called(chat)
	return args.Error(0)
}

func (m *MockChatRepositoryForParticipants) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockParticipantsCache struct {
	mock.Mock
}

func (m *MockParticipantsCache) Get(ctx context.Context, chatID int64) (*domain.ParticipantsInfo, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).(*domain.ParticipantsInfo), args.Error(1)
}

func (m *MockParticipantsCache) Set(ctx context.Context, chatID int64, count int, ttl time.Duration) error {
	args := m.Called(ctx, chatID, count, ttl)
	return args.Error(0)
}

func (m *MockParticipantsCache) GetMultiple(ctx context.Context, chatIDs []int64) (map[int64]*domain.ParticipantsInfo, error) {
	args := m.Called(ctx, chatIDs)
	return args.Get(0).(map[int64]*domain.ParticipantsInfo), args.Error(1)
}

func (m *MockParticipantsCache) SetMultiple(ctx context.Context, data map[int64]int, ttl time.Duration) error {
	args := m.Called(ctx, data, ttl)
	return args.Error(0)
}

func (m *MockParticipantsCache) Delete(ctx context.Context, chatID int64) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func (m *MockParticipantsCache) GetStaleChats(ctx context.Context, olderThan time.Duration, limit int) ([]int64, error) {
	args := m.Called(ctx, olderThan, limit)
	return args.Get(0).([]int64), args.Error(1)
}

type MockMaxServiceForParticipants struct {
	mock.Mock
}

func (m *MockMaxServiceForParticipants) GetMaxIDByPhone(phone string) (string, error) {
	args := m.Called(phone)
	return args.String(0), args.Error(1)
}

func (m *MockMaxServiceForParticipants) ValidatePhone(phone string) bool {
	args := m.Called(phone)
	return args.Bool(0)
}

func (m *MockMaxServiceForParticipants) GetChatInfo(ctx context.Context, chatID int64) (*domain.ChatInfo, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).(*domain.ChatInfo), args.Error(1)
}

func TestParticipantsUpdaterService_UpdateSingle(t *testing.T) {
	tests := []struct {
		name           string
		chatID         int64
		maxChatID      string
		setupMocks     func(*MockChatRepositoryForParticipants, *MockParticipantsCache, *MockMaxServiceForParticipants)
		expectedCount  int
		expectedSource string
		expectError    bool
	}{
		{
			name:      "successful update",
			chatID:    1,
			maxChatID: "123456",
			setupMocks: func(chatRepo *MockChatRepositoryForParticipants, cache *MockParticipantsCache, maxService *MockMaxServiceForParticipants) {
				maxService.On("GetChatInfo", mock.Anything, int64(123456)).Return(&domain.ChatInfo{
					ChatID:            123456,
					Title:             "Test Chat",
					ParticipantsCount: 42,
				}, nil)
				cache.On("Set", mock.Anything, int64(1), 42, mock.Anything).Return(nil)
				chatRepo.On("GetByID", int64(1)).Return(&domain.Chat{
					ID:                1,
					ParticipantsCount: 30,
				}, nil)
				chatRepo.On("Update", mock.Anything).Return(nil)
			},
			expectedCount:  42,
			expectedSource: "api",
			expectError:    false,
		},
		{
			name:      "empty max chat id",
			chatID:    1,
			maxChatID: "",
			setupMocks: func(chatRepo *MockChatRepositoryForParticipants, cache *MockParticipantsCache, maxService *MockMaxServiceForParticipants) {
				chatRepo.On("GetByID", int64(1)).Return(&domain.Chat{
					ID:                1,
					ParticipantsCount: 30,
					UpdatedAt:         time.Now(),
				}, nil)
			},
			expectedCount:  30,
			expectedSource: "database",
			expectError:    false,
		},
		{
			name:      "max api error",
			chatID:    1,
			maxChatID: "123456",
			setupMocks: func(chatRepo *MockChatRepositoryForParticipants, cache *MockParticipantsCache, maxService *MockMaxServiceForParticipants) {
				maxService.On("GetChatInfo", mock.Anything, int64(123456)).Return((*domain.ChatInfo)(nil), errors.New("api error"))
				chatRepo.On("GetByID", int64(1)).Return(&domain.Chat{
					ID:                1,
					ParticipantsCount: 30,
					UpdatedAt:         time.Now(),
				}, nil)
			},
			expectedCount:  30,
			expectedSource: "database",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем моки
			chatRepo := new(MockChatRepositoryForParticipants)
			cache := new(MockParticipantsCache)
			maxService := new(MockMaxServiceForParticipants)
			logger := logger.NewDefault()

			// Настраиваем моки
			tt.setupMocks(chatRepo, cache, maxService)

			// Создаем конфигурацию
			config := &domain.ParticipantsConfig{
				CacheTTL: time.Hour,
			}

			// Создаем сервис
			service := NewParticipantsUpdaterService(chatRepo, cache, maxService, config, logger)

			// Выполняем тест
			result, err := service.UpdateSingle(context.Background(), tt.chatID, tt.maxChatID)

			// Проверяем результат
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedCount, result.Count)
				assert.Equal(t, tt.expectedSource, result.Source)
			}

			// Проверяем, что все ожидания моков выполнены
			chatRepo.AssertExpectations(t)
			cache.AssertExpectations(t)
			maxService.AssertExpectations(t)
		})
	}
}

func TestParticipantsUpdaterService_UpdateBatch(t *testing.T) {
	// Создаем моки
	chatRepo := new(MockChatRepositoryForParticipants)
	cache := new(MockParticipantsCache)
	maxService := new(MockMaxServiceForParticipants)
	logger := logger.NewDefault()

	// Настраиваем моки
	maxService.On("GetChatInfo", mock.Anything, int64(123456)).Return(&domain.ChatInfo{
		ChatID:            123456,
		ParticipantsCount: 42,
	}, nil)
	maxService.On("GetChatInfo", mock.Anything, int64(789012)).Return(&domain.ChatInfo{
		ChatID:            789012,
		ParticipantsCount: 15,
	}, nil)

	cache.On("Set", mock.Anything, int64(1), 42, mock.Anything).Return(nil)
	cache.On("Set", mock.Anything, int64(2), 15, mock.Anything).Return(nil)
	cache.On("SetMultiple", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	chatRepo.On("GetByID", int64(1)).Return(&domain.Chat{ID: 1}, nil)
	chatRepo.On("GetByID", int64(2)).Return(&domain.Chat{ID: 2}, nil)
	chatRepo.On("Update", mock.Anything).Return(nil)

	// Создаем конфигурацию
	config := &domain.ParticipantsConfig{
		CacheTTL: time.Hour,
	}

	// Создаем сервис
	service := NewParticipantsUpdaterService(chatRepo, cache, maxService, config, logger)

	// Подготавливаем данные для теста
	chats := []domain.ChatUpdateRequest{
		{ChatID: 1, MaxChatID: "123456"},
		{ChatID: 2, MaxChatID: "789012"},
	}

	// Выполняем тест
	result, err := service.UpdateBatch(context.Background(), chats)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 42, result[1].Count)
	assert.Equal(t, 15, result[2].Count)

	// Проверяем, что все ожидания моков выполнены
	chatRepo.AssertExpectations(t)
	cache.AssertExpectations(t)
	maxService.AssertExpectations(t)
}