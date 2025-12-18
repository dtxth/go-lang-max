package domain

import (
	"context"
	"time"
)

// ProfileCacheService определяет интерфейс для кэширования профилей пользователей
type ProfileCacheService interface {
	// StoreProfile сохраняет профиль пользователя в кэше
	StoreProfile(ctx context.Context, userID string, profile UserProfileCache) error
	// GetProfile получает профиль пользователя из кэша
	GetProfile(ctx context.Context, userID string) (*UserProfileCache, error)
	// UpdateProfile обновляет профиль пользователя в кэше
	UpdateProfile(ctx context.Context, userID string, updates ProfileUpdates) error
	// GetProfileStats возвращает статистику профилей
	GetProfileStats(ctx context.Context) (*ProfileStats, error)
}

// UserProfileCache представляет кэшированный профиль пользователя
type UserProfileCache struct {
	UserID           string        `json:"user_id"`
	MaxFirstName     string        `json:"max_first_name"`
	MaxLastName      string        `json:"max_last_name"`
	UserProvidedName string        `json:"user_provided_name"`
	LastUpdated      time.Time     `json:"last_updated"`
	Source           ProfileSource `json:"source"`
}

// ProfileSource определяет источник профильной информации
type ProfileSource string

const (
	SourceWebhook   ProfileSource = "webhook"
	SourceUserInput ProfileSource = "user_input"
	SourceDefault   ProfileSource = "default"
)

// ProfileUpdates содержит обновления для профиля
type ProfileUpdates struct {
	MaxFirstName     *string        `json:"max_first_name,omitempty"`
	MaxLastName      *string        `json:"max_last_name,omitempty"`
	UserProvidedName *string        `json:"user_provided_name,omitempty"`
	Source           *ProfileSource `json:"source,omitempty"`
}

// ProfileStats содержит статистику профилей
type ProfileStats struct {
	TotalProfiles        int64 `json:"total_profiles"`
	ProfilesWithFullName int64 `json:"profiles_with_full_name"`
	ProfilesBySource     map[ProfileSource]int64 `json:"profiles_by_source"`
}

// GetDisplayName возвращает наиболее приоритетное имя для отображения
func (p *UserProfileCache) GetDisplayName() string {
	// Приоритет: user_provided_name > max_first_name + max_last_name > max_first_name
	if p.UserProvidedName != "" {
		return p.UserProvidedName
	}
	
	if p.MaxFirstName != "" && p.MaxLastName != "" {
		return p.MaxFirstName + " " + p.MaxLastName
	}
	
	if p.MaxFirstName != "" {
		return p.MaxFirstName
	}
	
	return ""
}

// HasFullName проверяет, есть ли полное имя (имя и фамилия)
func (p *UserProfileCache) HasFullName() bool {
	if p.UserProvidedName != "" {
		return true
	}
	return p.MaxFirstName != "" && p.MaxLastName != ""
}