package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"maxbot-service/internal/config"
)

// NewRedisClient создает новый Redis клиент
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	
	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	return client, nil
}

// RedisClient wraps redis.Client for easier access
type RedisClient struct {
	Client *redis.Client
}

// NewProfileCacheService создает новый сервис кэширования профилей
func NewProfileCacheService(cfg *config.Config) (*ProfileRedisCache, error) {
	client, err := NewRedisClient(cfg)
	if err != nil {
		return nil, err
	}
	
	return NewProfileRedisCache(client, cfg.ProfileTTL), nil
}

// NewProfileCacheServiceWithClient создает новый сервис кэширования профилей и возвращает клиент
func NewProfileCacheServiceWithClient(cfg *config.Config) (*ProfileRedisCache, *RedisClient, error) {
	client, err := NewRedisClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	
	cache := NewProfileRedisCache(client, cfg.ProfileTTL)
	redisClient := &RedisClient{Client: client}
	
	return cache, redisClient, nil
}