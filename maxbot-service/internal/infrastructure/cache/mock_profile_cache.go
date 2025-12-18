package cache

import (
	"context"
	"sync"
	"time"

	"maxbot-service/internal/domain"
)

// MockProfileCache реализует ProfileCacheService для тестирования
type MockProfileCache struct {
	profiles map[string]domain.UserProfileCache
	mutex    sync.RWMutex
}

// NewMockProfileCache создает новый mock кэш профилей
func NewMockProfileCache() *MockProfileCache {
	return &MockProfileCache{
		profiles: make(map[string]domain.UserProfileCache),
	}
}

// StoreProfile сохраняет профиль в памяти
func (m *MockProfileCache) StoreProfile(ctx context.Context, userID string, profile domain.UserProfileCache) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	profile.LastUpdated = time.Now()
	m.profiles[userID] = profile
	return nil
}

// GetProfile получает профиль из памяти
func (m *MockProfileCache) GetProfile(ctx context.Context, userID string) (*domain.UserProfileCache, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if profile, exists := m.profiles[userID]; exists {
		return &profile, nil
	}
	return nil, nil
}

// UpdateProfile обновляет профиль в памяти
func (m *MockProfileCache) UpdateProfile(ctx context.Context, userID string, updates domain.ProfileUpdates) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	profile, exists := m.profiles[userID]
	if !exists {
		profile = domain.UserProfileCache{
			UserID: userID,
		}
	}
	
	// Применяем обновления
	if updates.MaxFirstName != nil {
		profile.MaxFirstName = *updates.MaxFirstName
	}
	if updates.MaxLastName != nil {
		profile.MaxLastName = *updates.MaxLastName
	}
	if updates.UserProvidedName != nil {
		profile.UserProvidedName = *updates.UserProvidedName
	}
	if updates.Source != nil {
		profile.Source = *updates.Source
	}
	
	profile.LastUpdated = time.Now()
	m.profiles[userID] = profile
	return nil
}

// GetProfileStats возвращает статистику профилей
func (m *MockProfileCache) GetProfileStats(ctx context.Context) (*domain.ProfileStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	stats := &domain.ProfileStats{
		TotalProfiles:    int64(len(m.profiles)),
		ProfilesBySource: make(map[domain.ProfileSource]int64),
	}
	
	for _, profile := range m.profiles {
		if profile.HasFullName() {
			stats.ProfilesWithFullName++
		}
		stats.ProfilesBySource[profile.Source]++
	}
	
	return stats, nil
}