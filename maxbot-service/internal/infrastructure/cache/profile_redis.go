package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"maxbot-service/internal/domain"
)

// ProfileRedisCache реализует ProfileCacheService используя Redis
type ProfileRedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewProfileRedisCache создает новый экземпляр ProfileRedisCache
func NewProfileRedisCache(client *redis.Client, ttl time.Duration) *ProfileRedisCache {
	return &ProfileRedisCache{
		client: client,
		ttl:    ttl,
	}
}

// StoreProfile сохраняет профиль пользователя в Redis
func (c *ProfileRedisCache) StoreProfile(ctx context.Context, userID string, profile domain.UserProfileCache) error {
	key := c.getProfileKey(userID)
	
	// Устанавливаем время последнего обновления
	profile.LastUpdated = time.Now()
	
	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}
	
	// Добавляем таймаут для Redis операции (Requirements 3.4)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	err = c.client.Set(ctx, key, data, c.ttl).Err()
	if err != nil {
		// Проверяем тип ошибки для лучшей диагностики
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout storing profile in Redis: %w", err)
		}
		return fmt.Errorf("failed to store profile in Redis: %w", err)
	}
	
	return nil
}

// GetProfile получает профиль пользователя из Redis
func (c *ProfileRedisCache) GetProfile(ctx context.Context, userID string) (*domain.UserProfileCache, error) {
	key := c.getProfileKey(userID)
	
	// Добавляем таймаут для Redis операции (Requirements 3.4)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Профиль не найден - возвращаем nil без ошибки (Requirements 3.5)
			return nil, nil
		}
		// Проверяем тип ошибки для лучшей диагностики
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout getting profile from Redis: %w", err)
		}
		return nil, fmt.Errorf("failed to get profile from Redis: %w", err)
	}
	
	var profile domain.UserProfileCache
	err = json.Unmarshal([]byte(data), &profile)
	if err != nil {
		// Если данные повреждены, логируем и возвращаем nil для graceful degradation
		return nil, fmt.Errorf("failed to unmarshal profile (corrupted data): %w", err)
	}
	
	return &profile, nil
}

// UpdateProfile обновляет профиль пользователя в Redis
func (c *ProfileRedisCache) UpdateProfile(ctx context.Context, userID string, updates domain.ProfileUpdates) error {
	// Получаем существующий профиль
	profile, err := c.GetProfile(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get existing profile: %w", err)
	}
	
	// Если профиль не существует, создаем новый
	if profile == nil {
		profile = &domain.UserProfileCache{
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
	
	// Сохраняем обновленный профиль
	return c.StoreProfile(ctx, userID, *profile)
}

// GetProfileStats возвращает статистику профилей
func (c *ProfileRedisCache) GetProfileStats(ctx context.Context) (*domain.ProfileStats, error) {
	// Получаем все ключи профилей
	pattern := c.getProfileKey("*")
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get profile keys: %w", err)
	}
	
	stats := &domain.ProfileStats{
		TotalProfiles:    int64(len(keys)),
		ProfilesBySource: make(map[domain.ProfileSource]int64),
	}
	
	// Если нет профилей, возвращаем пустую статистику
	if len(keys) == 0 {
		return stats, nil
	}
	
	// Получаем все профили для анализа
	pipe := c.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles for stats: %w", err)
	}
	
	// Анализируем профили
	for _, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			continue // Пропускаем ошибочные записи
		}
		
		var profile domain.UserProfileCache
		if err := json.Unmarshal([]byte(data), &profile); err != nil {
			continue // Пропускаем некорректные записи
		}
		
		// Подсчитываем профили с полным именем
		if profile.HasFullName() {
			stats.ProfilesWithFullName++
		}
		
		// Подсчитываем по источникам
		stats.ProfilesBySource[profile.Source]++
	}
	
	return stats, nil
}

// getProfileKey генерирует ключ для профиля в Redis
func (c *ProfileRedisCache) getProfileKey(userID string) string {
	return fmt.Sprintf("profile:user:%s", userID)
}

// IsHealthy проверяет доступность Redis для graceful degradation
func (c *ProfileRedisCache) IsHealthy(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	
	// Простая проверка через PING команду
	err := c.client.Ping(ctx).Err()
	return err == nil
}

// GetConnectionStatus возвращает статус подключения к Redis
func (c *ProfileRedisCache) GetConnectionStatus(ctx context.Context) map[string]interface{} {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	
	status := make(map[string]interface{})
	
	// Проверяем PING
	pingErr := c.client.Ping(ctx).Err()
	status["ping_ok"] = pingErr == nil
	if pingErr != nil {
		status["ping_error"] = pingErr.Error()
	}
	
	// Получаем информацию о Redis
	info, err := c.client.Info(ctx, "server").Result()
	if err == nil {
		status["redis_info"] = info
	} else {
		status["info_error"] = err.Error()
	}
	
	return status
}