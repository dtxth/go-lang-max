package app

import (
	"chat-service/internal/config"
	"chat-service/internal/domain"
	"chat-service/internal/infrastructure/cache"
	"chat-service/internal/infrastructure/logger"
	"chat-service/internal/infrastructure/worker"
	"chat-service/internal/usecase"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

// ParticipantsIntegration содержит компоненты для работы с участниками
type ParticipantsIntegration struct {
	Cache   domain.ParticipantsCache
	Updater domain.ParticipantsUpdater
	Worker  *worker.ParticipantsWorker
	Config  *domain.ParticipantsConfig
}

// NewParticipantsIntegration создает интеграцию для работы с участниками
func NewParticipantsIntegration(
	chatRepo domain.ChatRepository,
	maxService domain.MaxService,
	logger *logger.Logger,
) (*ParticipantsIntegration, error) {
	// Загружаем конфигурацию
	config := config.LoadParticipantsConfig()
	
	// Создаем Redis клиент
	redisClient, err := createRedisClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}
	
	// Создаем кэш
	participantsCache := cache.NewParticipantsRedisCache(redisClient)
	
	// Создаем updater
	participantsUpdater := usecase.NewParticipantsUpdaterService(
		chatRepo,
		participantsCache,
		maxService,
		config,
		logger,
	)
	
	// Создаем воркер
	participantsWorker := worker.NewParticipantsWorker(
		participantsUpdater,
		config,
		logger,
	)
	
	return &ParticipantsIntegration{
		Cache:   participantsCache,
		Updater: participantsUpdater,
		Worker:  participantsWorker,
		Config:  config,
	}, nil
}

// Start запускает фоновые процессы
func (pi *ParticipantsIntegration) Start() {
	if pi.Worker != nil {
		pi.Worker.Start()
	}
}

// Stop останавливает фоновые процессы
func (pi *ParticipantsIntegration) Stop() {
	if pi.Worker != nil {
		pi.Worker.Stop()
	}
}

// createRedisClient создает клиент Redis
func createRedisClient() (*redis.Client, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}
	
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}
	
	client := redis.NewClient(opt)
	
	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	return client, nil
}

// IsEnabled проверяет, включена ли интеграция с участниками
func IsParticipantsIntegrationEnabled() bool {
	// Проверяем наличие Redis URL
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return false
	}
	
	// Проверяем, не отключена ли интеграция явно
	if disabled := os.Getenv("PARTICIPANTS_INTEGRATION_DISABLED"); disabled == "true" {
		return false
	}
	
	return true
}