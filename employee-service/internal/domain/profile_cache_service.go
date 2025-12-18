package domain

import (
	"context"
	"time"
)

// ProfileCacheService определяет интерфейс для работы с кэшем профилей пользователей
type ProfileCacheService interface {
	// GetProfile получает профиль пользователя из кэша по MAX_id
	GetProfile(ctx context.Context, userID string) (*CachedUserProfile, error)
}

// CachedUserProfile представляет кэшированный профиль пользователя
type CachedUserProfile struct {
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

// GetDisplayName возвращает наиболее приоритетное имя для отображения
// Приоритет: user_provided_name > max_first_name + max_last_name > max_first_name
func (p *CachedUserProfile) GetDisplayName() (firstName, lastName string) {
	if p.UserProvidedName != "" {
		// Пытаемся разделить user_provided_name на имя и фамилию
		// Простое разделение по пробелу
		parts := splitName(p.UserProvidedName)
		if len(parts) >= 2 {
			return parts[0], parts[1]
		}
		return parts[0], ""
	}
	
	return p.MaxFirstName, p.MaxLastName
}

// GetPrioritySource возвращает источник с наивысшим приоритетом
func (p *CachedUserProfile) GetPrioritySource() ProfileSource {
	if p.UserProvidedName != "" {
		return SourceUserInput
	}
	if p.MaxFirstName != "" || p.MaxLastName != "" {
		return SourceWebhook
	}
	return SourceDefault
}

// splitName разделяет полное имя на части
func splitName(fullName string) []string {
	// Простое разделение по пробелам
	// В реальной реализации можно использовать более сложную логику
	parts := make([]string, 0)
	current := ""
	
	for _, char := range fullName {
		if char == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}