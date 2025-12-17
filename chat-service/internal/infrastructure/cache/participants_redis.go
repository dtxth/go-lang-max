package cache

import (
	"chat-service/internal/domain"
	"chat-service/internal/infrastructure/logger"
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
	logger *logger.Logger
}

func NewParticipantsRedisCache(client *redis.Client) *ParticipantsRedisCache {
	return &ParticipantsRedisCache{
		client: client,
		prefix: "chat_participants:",
	}
}

func NewParticipantsRedisCacheWithLogger(client *redis.Client, logger *logger.Logger) *ParticipantsRedisCache {
	return &ParticipantsRedisCache{
		client: client,
		prefix: "chat_participants:",
		logger: logger,
	}
}

func (c *ParticipantsRedisCache) Get(ctx context.Context, chatID int64) (*domain.ParticipantsInfo, error) {
	key := c.key(chatID)
	start := time.Now()
	
	val, err := c.client.Get(ctx, key).Result()
	duration := time.Since(start)
	
	if err != nil {
		if err == redis.Nil {
			if c.logger != nil {
				c.logger.Debug(ctx, "Cache miss for participants", map[string]interface{}{
					"chat_id": chatID,
					"duration": duration.String(),
				})
			}
			return nil, domain.ErrParticipantsNotCached
		}
		
		if c.logger != nil {
			c.logger.Error(ctx, "Failed to get participants from cache", map[string]interface{}{
				"chat_id": chatID,
				"error": err.Error(),
				"duration": duration.String(),
			})
		}
		return nil, fmt.Errorf("failed to get participants from cache: %w", err)
	}
	
	var info domain.ParticipantsInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		if c.logger != nil {
			c.logger.Error(ctx, "Failed to unmarshal participants info from cache", map[string]interface{}{
				"chat_id": chatID,
				"error": err.Error(),
				"raw_value": val,
			})
		}
		return nil, fmt.Errorf("failed to unmarshal participants info: %w", err)
	}
	
	info.Source = "cache"
	
	if c.logger != nil {
		c.logger.Debug(ctx, "Cache hit for participants", map[string]interface{}{
			"chat_id": chatID,
			"count": info.Count,
			"age": time.Since(info.UpdatedAt).String(),
			"duration": duration.String(),
		})
	}
	
	return &info, nil
}

func (c *ParticipantsRedisCache) Set(ctx context.Context, chatID int64, count int, ttl time.Duration) error {
	key := c.key(chatID)
	start := time.Now()
	
	info := domain.ParticipantsInfo{
		Count:     count,
		UpdatedAt: time.Now(),
		Source:    "cache",
	}
	
	data, err := json.Marshal(info)
	if err != nil {
		if c.logger != nil {
			c.logger.Error(ctx, "Failed to marshal participants info for cache", map[string]interface{}{
				"chat_id": chatID,
				"count": count,
				"error": err.Error(),
			})
		}
		return fmt.Errorf("failed to marshal participants info: %w", err)
	}
	
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		duration := time.Since(start)
		if c.logger != nil {
			c.logger.Error(ctx, "Failed to set participants in cache", map[string]interface{}{
				"chat_id": chatID,
				"count": count,
				"ttl": ttl.String(),
				"error": err.Error(),
				"duration": duration.String(),
			})
		}
		return fmt.Errorf("failed to set participants in cache: %w", err)
	}
	
	duration := time.Since(start)
	if c.logger != nil {
		c.logger.Debug(ctx, "Successfully cached participants", map[string]interface{}{
			"chat_id": chatID,
			"count": count,
			"ttl": ttl.String(),
			"duration": duration.String(),
		})
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
		if c.logger != nil {
			c.logger.Debug(ctx, "SetMultiple called with empty data", nil)
		}
		return nil
	}
	
	start := time.Now()
	pipe := c.client.Pipeline()
	now := time.Now()
	successCount := 0
	errorCount := 0
	
	for chatID, count := range data {
		key := c.key(chatID)
		
		info := domain.ParticipantsInfo{
			Count:     count,
			UpdatedAt: now,
			Source:    "cache",
		}
		
		jsonData, err := json.Marshal(info)
		if err != nil {
			errorCount++
			if c.logger != nil {
				c.logger.Error(ctx, "Failed to marshal participants info in batch", map[string]interface{}{
					"chat_id": chatID,
					"count": count,
					"error": err.Error(),
				})
			}
			continue // пропускаем ошибочные данные
		}
		
		pipe.Set(ctx, key, jsonData, ttl)
		successCount++
	}
	
	_, err := pipe.Exec(ctx)
	duration := time.Since(start)
	
	logData := map[string]interface{}{
		"total_items": len(data),
		"successful": successCount,
		"errors": errorCount,
		"ttl": ttl.String(),
		"duration": duration.String(),
	}
	
	if err != nil {
		logData["error"] = err.Error()
		if c.logger != nil {
			c.logger.Error(ctx, "Failed to execute batch cache operation", logData)
		}
		return fmt.Errorf("failed to set multiple participants in cache: %w", err)
	}
	
	if c.logger != nil {
		if errorCount > 0 {
			c.logger.Warn(ctx, "Batch cache operation completed with some errors", logData)
		} else {
			c.logger.Debug(ctx, "Successfully cached multiple participants", logData)
		}
		
		// Performance warning for large batches
		if len(data) > 100 && duration > 1*time.Second {
			c.logger.Warn(ctx, "Batch cache operation was slow", map[string]interface{}{
				"items": len(data),
				"duration": duration.String(),
				"items_per_second": fmt.Sprintf("%.2f", float64(len(data))/duration.Seconds()),
			})
		}
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