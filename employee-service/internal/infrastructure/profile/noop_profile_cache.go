package profile

import (
	"context"
	"employee-service/internal/domain"
	"time"
)

// NoOpProfileCacheClient реализует ProfileCacheService как заглушку
// Используется когда profile cache недоступен
type NoOpProfileCacheClient struct{}

// GetProfile всегда возвращает пустой профиль
func (c *NoOpProfileCacheClient) GetProfile(ctx context.Context, userID string) (*domain.CachedUserProfile, error) {
	return &domain.CachedUserProfile{
		UserID:           userID,
		MaxFirstName:     "",
		MaxLastName:      "",
		UserProvidedName: "",
		LastUpdated:      time.Now(),
		Source:           domain.SourceDefault,
	}, nil
}