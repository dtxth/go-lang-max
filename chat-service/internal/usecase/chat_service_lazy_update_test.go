package usecase

import (
	"chat-service/internal/domain"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing lazy update functionality
type MockChatRepoForLazyUpdate struct {
	mock.Mock
}

func (m *MockChatRepoForLazyUpdate) GetAllWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	args := m.Called(limit, offset, sortBy, sortOrder, search, filter)
	return args.Get(0).([]*domain.Chat), args.Int(1), args.Error(2)
}

func (m *MockChatRepoForLazyUpdate) GetByID(id int64) (*domain.Chat, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *MockChatRepoForLazyUpdate) Create(chat *domain.Chat) error {
	args := m.Called(chat)
	return args.Error(0)
}

func (m *MockChatRepoForLazyUpdate) Update(chat *domain.Chat) error {
	args := m.Called(chat)
	return args.Error(0)
}

func (m *MockChatRepoForLazyUpdate) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockChatRepoForLazyUpdate) GetAll(limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	args := m.Called(limit, offset, filter)
	return args.Get(0).([]*domain.Chat), args.Int(1), args.Error(2)
}

func (m *MockChatRepoForLazyUpdate) GetByMaxChatID(maxChatID string) (*domain.Chat, error) {
	args := m.Called(maxChatID)
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *MockChatRepoForLazyUpdate) Search(query string, limit, offset int, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
	args := m.Called(query, limit, offset, filter)
	return args.Get(0).([]*domain.Chat), args.Int(1), args.Error(2)
}

type MockParticipantsCacheForLazyUpdate struct {
	mock.Mock
}

func (m *MockParticipantsCacheForLazyUpdate) Get(ctx context.Context, chatID int64) (*domain.ParticipantsInfo, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).(*domain.ParticipantsInfo), args.Error(1)
}

func (m *MockParticipantsCacheForLazyUpdate) Set(ctx context.Context, chatID int64, count int, ttl time.Duration) error {
	args := m.Called(ctx, chatID, count, ttl)
	return args.Error(0)
}

func (m *MockParticipantsCacheForLazyUpdate) GetMultiple(ctx context.Context, chatIDs []int64) (map[int64]*domain.ParticipantsInfo, error) {
	args := m.Called(ctx, chatIDs)
	return args.Get(0).(map[int64]*domain.ParticipantsInfo), args.Error(1)
}

func (m *MockParticipantsCacheForLazyUpdate) SetMultiple(ctx context.Context, data map[int64]int, ttl time.Duration) error {
	args := m.Called(ctx, data, ttl)
	return args.Error(0)
}

func (m *MockParticipantsCacheForLazyUpdate) Delete(ctx context.Context, chatID int64) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func (m *MockParticipantsCacheForLazyUpdate) GetStaleChats(ctx context.Context, olderThan time.Duration, limit int) ([]int64, error) {
	args := m.Called(ctx, olderThan, limit)
	return args.Get(0).([]int64), args.Error(1)
}

type MockParticipantsUpdaterForLazyUpdate struct {
	mock.Mock
}

func (m *MockParticipantsUpdaterForLazyUpdate) UpdateSingle(ctx context.Context, chatID int64, maxChatID string) (*domain.ParticipantsInfo, error) {
	args := m.Called(ctx, chatID, maxChatID)
	return args.Get(0).(*domain.ParticipantsInfo), args.Error(1)
}

func (m *MockParticipantsUpdaterForLazyUpdate) UpdateBatch(ctx context.Context, chats []domain.ChatUpdateRequest) (map[int64]*domain.ParticipantsInfo, error) {
	args := m.Called(ctx, chats)
	return args.Get(0).(map[int64]*domain.ParticipantsInfo), args.Error(1)
}

func (m *MockParticipantsUpdaterForLazyUpdate) UpdateStale(ctx context.Context, olderThan time.Duration, batchSize int) (int, error) {
	args := m.Called(ctx, olderThan, batchSize)
	return args.Int(0), args.Error(1)
}

func (m *MockParticipantsUpdaterForLazyUpdate) UpdateAll(ctx context.Context, batchSize int) (int, error) {
	args := m.Called(ctx, batchSize)
	return args.Int(0), args.Error(1)
}

func TestGetAllChatsWithSortingAndSearch_LazyUpdate_FreshCache(t *testing.T) {
	// Setup
	mockRepo := new(MockChatRepoForLazyUpdate)
	mockCache := new(MockParticipantsCacheForLazyUpdate)
	mockUpdater := new(MockParticipantsUpdaterForLazyUpdate)
	
	config := &domain.ParticipantsConfig{
		EnableLazyUpdate: true,
		StaleThreshold:   time.Hour,
	}
	
	service := NewChatServiceWithParticipants(
		mockRepo,
		nil, // administratorRepo not needed for this test
		nil, // maxService not needed for this test
		mockCache,
		mockUpdater,
		config,
	)
	
	// Test data
	now := time.Now()
	chats := []*domain.Chat{
		{ID: 1, Name: "Chat 1", MaxChatID: "123", ParticipantsCount: 10},
		{ID: 2, Name: "Chat 2", MaxChatID: "456", ParticipantsCount: 20},
	}
	
	cachedData := map[int64]*domain.ParticipantsInfo{
		1: {Count: 15, UpdatedAt: now.Add(-30 * time.Minute), Source: "cache"}, // Fresh data
		2: {Count: 25, UpdatedAt: now.Add(-30 * time.Minute), Source: "cache"}, // Fresh data
	}
	
	// Setup expectations
	mockRepo.On("GetAllWithSortingAndSearch", 50, 0, "name", "asc", "", (*domain.ChatFilter)(nil)).Return(chats, 2, nil)
	mockCache.On("GetMultiple", mock.Anything, []int64{1, 2}).Return(cachedData, nil)
	
	// Execute
	result, totalCount, err := service.GetAllChatsWithSortingAndSearch(50, 0, "name", "asc", "", nil)
	
	// Verify
	assert.NoError(t, err)
	assert.Equal(t, 2, totalCount)
	assert.Len(t, result, 2)
	
	// Verify that cached data was used (participants count updated from cache)
	assert.Equal(t, 15, result[0].ParticipantsCount) // Updated from cache
	assert.Equal(t, 25, result[1].ParticipantsCount) // Updated from cache
	
	// Verify that UpdateBatch was NOT called (data was fresh)
	mockUpdater.AssertNotCalled(t, "UpdateBatch")
	
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestGetAllChatsWithSortingAndSearch_LazyUpdate_StaleCache(t *testing.T) {
	// Setup
	mockRepo := new(MockChatRepoForLazyUpdate)
	mockCache := new(MockParticipantsCacheForLazyUpdate)
	mockUpdater := new(MockParticipantsUpdaterForLazyUpdate)
	
	config := &domain.ParticipantsConfig{
		EnableLazyUpdate: true,
		StaleThreshold:   time.Hour,
	}
	
	service := NewChatServiceWithParticipants(
		mockRepo,
		nil, // administratorRepo not needed for this test
		nil, // maxService not needed for this test
		mockCache,
		mockUpdater,
		config,
	)
	
	// Test data
	now := time.Now()
	chats := []*domain.Chat{
		{ID: 1, Name: "Chat 1", MaxChatID: "123", ParticipantsCount: 10},
		{ID: 2, Name: "Chat 2", MaxChatID: "456", ParticipantsCount: 20},
	}
	
	cachedData := map[int64]*domain.ParticipantsInfo{
		1: {Count: 15, UpdatedAt: now.Add(-2 * time.Hour), Source: "cache"}, // Stale data
		2: {Count: 25, UpdatedAt: now.Add(-2 * time.Hour), Source: "cache"}, // Stale data
	}
	
	updatedData := map[int64]*domain.ParticipantsInfo{
		1: {Count: 18, UpdatedAt: now, Source: "api"},
		2: {Count: 28, UpdatedAt: now, Source: "api"},
	}
	
	expectedUpdateRequests := []domain.ChatUpdateRequest{
		{ChatID: 1, MaxChatID: "123"},
		{ChatID: 2, MaxChatID: "456"},
	}
	
	// Setup expectations
	mockRepo.On("GetAllWithSortingAndSearch", 50, 0, "name", "asc", "", (*domain.ChatFilter)(nil)).Return(chats, 2, nil)
	mockCache.On("GetMultiple", mock.Anything, []int64{1, 2}).Return(cachedData, nil)
	mockUpdater.On("UpdateBatch", mock.Anything, expectedUpdateRequests).Return(updatedData, nil)
	
	// Execute
	result, totalCount, err := service.GetAllChatsWithSortingAndSearch(50, 0, "name", "asc", "", nil)
	
	// Verify
	assert.NoError(t, err)
	assert.Equal(t, 2, totalCount)
	assert.Len(t, result, 2)
	
	// Verify that updated data was used (participants count updated from API)
	assert.Equal(t, 18, result[0].ParticipantsCount) // Updated from API
	assert.Equal(t, 28, result[1].ParticipantsCount) // Updated from API
	
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockUpdater.AssertExpectations(t)
}

func TestGetAllChatsWithSortingAndSearch_LazyUpdate_CacheError_Fallback(t *testing.T) {
	// Setup
	mockRepo := new(MockChatRepoForLazyUpdate)
	mockCache := new(MockParticipantsCacheForLazyUpdate)
	mockUpdater := new(MockParticipantsUpdaterForLazyUpdate)
	
	config := &domain.ParticipantsConfig{
		EnableLazyUpdate: true,
		StaleThreshold:   time.Hour,
	}
	
	service := NewChatServiceWithParticipants(
		mockRepo,
		nil, // administratorRepo not needed for this test
		nil, // maxService not needed for this test
		mockCache,
		mockUpdater,
		config,
	)
	
	// Test data
	chats := []*domain.Chat{
		{ID: 1, Name: "Chat 1", MaxChatID: "123", ParticipantsCount: 10},
		{ID: 2, Name: "Chat 2", MaxChatID: "456", ParticipantsCount: 20},
	}
	
	// Setup expectations - cache returns error
	mockRepo.On("GetAllWithSortingAndSearch", 50, 0, "name", "asc", "", (*domain.ChatFilter)(nil)).Return(chats, 2, nil)
	mockCache.On("GetMultiple", mock.Anything, []int64{1, 2}).Return(map[int64]*domain.ParticipantsInfo{}, assert.AnError)
	
	// Execute
	result, totalCount, err := service.GetAllChatsWithSortingAndSearch(50, 0, "name", "asc", "", nil)
	
	// Verify - should fallback to database data
	assert.NoError(t, err)
	assert.Equal(t, 2, totalCount)
	assert.Len(t, result, 2)
	
	// Verify that database data was used (no updates from cache/API)
	assert.Equal(t, 10, result[0].ParticipantsCount) // Original database value
	assert.Equal(t, 20, result[1].ParticipantsCount) // Original database value
	
	// Verify that UpdateBatch was NOT called (cache error triggered fallback)
	mockUpdater.AssertNotCalled(t, "UpdateBatch")
	
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}