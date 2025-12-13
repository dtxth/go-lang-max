package cache

import (
	"chat-service/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type ParticipantsRedisCache struct {
	client *redis.Client
	prefix string
}

func NewParticipantsRedisCache(client *redis.Client) *ParticipantsRedisCache {
	return &ParticipantsRedisCache{
		client: client,
		prefix: "chat_participants:",
	}
}

func (c *ParticipantsRedisCache) Get(ctx context.Context, chatID int64) (*domain.ParticipantsInfo, error) {
	key := c.key(chatID)
	
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrParticipantsNotCached
		}
		return nil, fmt.Errorf("failed to get participants from cache: %w", err)
	}
	
	var info domain.ParticipantsInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal participants info: %w", err)
	}
	
	info.Source = "cache"
	return &info, nil
}

func (c *ParticipantsRedisCache) Set(ctx context.Context, chatID int64, count int, ttl time.Duration) error {
	key := c.key(chatID)
	
	info := domain.ParticipantsInfo{
		Count:     count,
		UpdatedAt: time.Now(),
		Source:    "cache",
	}
	
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal participants info: %w", err)
	}
	
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set participants in cache: %w", err)
	}
	
	return nil
}

func (c *ParticipantsRedisCache) GetMultiple(ctx context.Context, chatIDs []int64) (map[int64]*domain.ParticipantsInfo, error) {
	if len(chatIDs) == 0 {
		return make(map[int64]*domain.ParticipantsInfo), nil
	}
	
	keys := make([]string, len(chatIDs))
	for i, chatID := range chatIDs {
		keys[i] = c.key(chatID)
	}
	
	values, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple participants from cache: %w", err)
	}
	
	result := make(map[int64]*domain.ParticipantsInfo)
	for i, val := range values {
		if val == nil {
			continue // ключ не найден
		}
		
		var info domain.ParticipantsInfo
		if err := json.Unmarshal([]byte(val.(string)), &info); err != nil {
			continue // пропускаем поврежденные данные
		}
		
		info.Source = "cache"
		result[chatIDs[i]] = &info
	}
	
	return result, nil
}

func (c *ParticipantsRedisCache) SetMultiple(ctx context.Context, data map[int64]int, ttl time.Duration) error {
	if len(data) == 0 {
		return nil
	}
	
	pipe := c.client.Pipeline()
	now := time.Now()
	
	for chatID, count := range data {
		key := c.key(chatID)
		
		info := domain.ParticipantsInfo{
			Count:     count,
			UpdatedAt: now,
			Source:    "cache",
		}
		
		jsonData, err := json.Marshal(info)
		if err != nil {
			continue // пропускаем ошибочные данные
		}
		
		pipe.Set(ctx, key, jsonData, ttl)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set multiple participants in cache: %w", err)
	}
	
	return nil
}

func (c *ParticipantsRedisCache) Delete(ctx context.Context, chatID int64) error {
	key := c.key(chatID)
	
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete participants from cache: %w", err)
	}
	
	return nil
}

func (c *ParticipantsRedisCache) GetStaleChats(ctx context.Context, olderThan time.Duration, limit int) ([]int64, error) {
	// Используем SCAN для поиска всех ключей с префиксом
	var cursor uint64
	var staleChats []int64
	cutoffTime := time.Now().Add(-olderThan)
	
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, c.prefix+"*", int64(limit*2)).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan cache keys: %w", err)
		}
		
		if len(keys) > 0 {
			// Получаем данные для найденных ключей
			values, err := c.client.MGet(ctx, keys...).Result()
			if err != nil {
				return nil, fmt.Errorf("failed to get values for stale check: %w", err)
			}
			
			for i, val := range values {
				if val == nil {
					continue
				}
				
				var info domain.ParticipantsInfo
				if err := json.Unmarshal([]byte(val.(string)), &info); err != nil {
					continue
				}
				
				// Проверяем, устарели ли данные
				if info.UpdatedAt.Before(cutoffTime) {
					// Извлекаем chat_id из ключа
					chatIDStr := strings.TrimPrefix(keys[i], c.prefix)
					if chatID, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
						staleChats = append(staleChats, chatID)
						
						// Ограничиваем количество результатов
						if len(staleChats) >= limit {
							return staleChats, nil
						}
					}
				}
			}
		}
		
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	
	return staleChats, nil
}

func (c *ParticipantsRedisCache) key(chatID int64) string {
	return c.prefix + strconv.FormatInt(chatID, 10)
}