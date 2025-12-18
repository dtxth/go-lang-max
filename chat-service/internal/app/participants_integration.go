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
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// ParticipantsIntegration содержит компоненты для работы с участниками
type ParticipantsIntegration struct {
	Cache         domain.ParticipantsCache
	Updater       domain.ParticipantsUpdater
	Worker        *worker.ParticipantsWorker
	Config        *domain.ParticipantsConfig
	logger        *logger.Logger
	redisClient   *redis.Client
	
	// Circuit breaker state
	circuitBreaker *CircuitBreaker
	
	// Health monitoring
	healthMutex    sync.RWMutex
	redisHealthy   bool
	maxAPIHealthy  bool
	lastHealthCheck time.Time
}

// CircuitBreaker implements circuit breaker pattern for MAX API failures
type CircuitBreaker struct {
	mutex           sync.RWMutex
	failureCount    int
	lastFailureTime time.Time
	state          usecase.CircuitState
	threshold      int
	timeout        time.Duration
}

// NewParticipantsIntegration создает интеграцию для работы с участниками
func NewParticipantsIntegration(
	chatRepo domain.ChatRepository,
	maxService domain.MaxService,
	logger *logger.Logger,
) (*ParticipantsIntegration, error) {
	initStart := time.Now()
	
	// Загружаем конфигурацию
	config := config.LoadParticipantsConfig()
	
	logger.Info(context.Background(), "Initializing participants integration", map[string]interface{}{
		"component":               "participants_integration",
		"initialization_stage":    "configuration_loaded",
		"cache_ttl":               config.CacheTTL.String(),
		"update_interval":         config.UpdateInterval.String(),
		"full_update_hour":        config.FullUpdateHour,
		"batch_size":              config.BatchSize,
		"max_api_timeout":         config.MaxAPITimeout.String(),
		"stale_threshold":         config.StaleThreshold.String(),
		"enable_background_sync":  config.EnableBackgroundSync,
		"enable_lazy_update":      config.EnableLazyUpdate,
		"max_retries":             config.MaxRetries,
	})
	
	// Создаем circuit breaker для MAX API
	circuitBreaker := &CircuitBreaker{
		threshold: 5,  // 5 consecutive failures
		timeout:   5 * time.Minute, // 5 minutes timeout
		state:     usecase.CircuitClosed,
	}
	
	logger.Info(context.Background(), "Circuit breaker initialized", map[string]interface{}{
		"component":               "participants_integration",
		"initialization_stage":    "circuit_breaker_created",
		"failure_threshold":       5,
		"timeout_duration":        "5m",
		"initial_state":           "closed",
	})
	
	// Создаем Redis клиент с retry logic
	redisClient, err := createRedisClientWithRetry(logger)
	if err != nil {
		logger.Error(context.Background(), "Failed to create Redis client, participants integration will be disabled", map[string]interface{}{
			"component":               "participants_integration",
			"initialization_stage":    "redis_connection_failed",
			"error":                   err.Error(),
			"fallback_mode":           "disabled",
			"initialization_duration": time.Since(initStart).String(),
		})
		return &ParticipantsIntegration{
			logger:         logger,
			circuitBreaker: circuitBreaker,
			Config:         config,
			redisHealthy:   false,
			maxAPIHealthy:  true,
		}, nil
	}
	
	// Создаем кэш с логгером
	participantsCache := cache.NewParticipantsRedisCacheWithLogger(redisClient, logger)
	logger.Info(context.Background(), "Redis cache component initialized", map[string]interface{}{
		"component":               "participants_integration",
		"initialization_stage":    "cache_created",
		"cache_type":              "redis",
	})
	
	// Создаем updater с circuit breaker
	participantsUpdater := usecase.NewParticipantsUpdaterServiceWithCircuitBreaker(
		chatRepo,
		participantsCache,
		maxService,
		config,
		logger,
		circuitBreaker,
	)
	logger.Info(context.Background(), "Participants updater service initialized", map[string]interface{}{
		"component":               "participants_integration",
		"initialization_stage":    "updater_created",
		"circuit_breaker_enabled": true,
	})
	
	// Создаем воркер
	participantsWorker := worker.NewParticipantsWorker(
		participantsUpdater,
		config,
		logger,
	)
	logger.Info(context.Background(), "Participants worker initialized", map[string]interface{}{
		"component":               "participants_integration",
		"initialization_stage":    "worker_created",
		"background_sync_enabled": config.EnableBackgroundSync,
	})
	
	integration := &ParticipantsIntegration{
		Cache:          participantsCache,
		Updater:        participantsUpdater,
		Worker:         participantsWorker,
		Config:         config,
		logger:         logger,
		redisClient:    redisClient,
		circuitBreaker: circuitBreaker,
		redisHealthy:   true,
		maxAPIHealthy:  true,
		lastHealthCheck: time.Now(),
	}
	
	initDuration := time.Since(initStart)
	logger.Info(context.Background(), "Participants integration initialized successfully", map[string]interface{}{
		"component":               "participants_integration",
		"initialization_stage":    "completed",
		"initialization_duration": initDuration.String(),
		"redis_healthy":           true,
		"max_api_healthy":         true,
		"components_initialized":  []string{"cache", "updater", "worker", "circuit_breaker"},
	})
	return integration, nil
}

// Start запускает фоновые процессы
func (pi *ParticipantsIntegration) Start() {
	startTime := time.Now()
	
	if pi.Worker != nil && pi.redisHealthy {
		pi.logger.Info(context.Background(), "Starting participants integration worker", map[string]interface{}{
			"component":               "participants_integration",
			"operation":               "start",
			"redis_healthy":           pi.redisHealthy,
			"background_sync_enabled": pi.Config.EnableBackgroundSync,
			"lazy_update_enabled":     pi.Config.EnableLazyUpdate,
		})
		
		pi.Worker.Start()
		
		// Запускаем мониторинг здоровья
		go pi.startHealthMonitoring()
		
		startDuration := time.Since(startTime)
		pi.logger.Info(context.Background(), "Participants integration started successfully", map[string]interface{}{
			"component":      "participants_integration",
			"operation":      "start_completed",
			"start_duration": startDuration.String(),
			"health_monitoring_enabled": true,
		})
	} else {
		reasons := []string{}
		if pi.Worker == nil {
			reasons = append(reasons, "worker_not_initialized")
		}
		if !pi.redisHealthy {
			reasons = append(reasons, "redis_unhealthy")
		}
		
		pi.logger.Warn(context.Background(), "Participants integration worker not started", map[string]interface{}{
			"component":      "participants_integration",
			"operation":      "start_skipped",
			"reasons":        reasons,
			"redis_healthy":  pi.redisHealthy,
			"worker_exists":  pi.Worker != nil,
		})
	}
}

// Stop останавливает фоновые процессы
func (pi *ParticipantsIntegration) Stop() {
	stopStart := time.Now()
	
	pi.logger.Info(context.Background(), "Stopping participants integration", map[string]interface{}{
		"component": "participants_integration",
		"operation": "stop",
	})
	
	// Останавливаем воркер
	if pi.Worker != nil {
		workerStopStart := time.Now()
		pi.Worker.Stop()
		workerStopDuration := time.Since(workerStopStart)
		
		pi.logger.Info(context.Background(), "Participants worker stopped", map[string]interface{}{
			"component":           "participants_integration",
			"operation":           "worker_stopped",
			"worker_stop_duration": workerStopDuration.String(),
		})
	}
	
	// Закрываем Redis клиент
	if pi.redisClient != nil {
		redisCloseStart := time.Now()
		if err := pi.redisClient.Close(); err != nil {
			pi.logger.Error(context.Background(), "Failed to close Redis client", map[string]interface{}{
				"component": "participants_integration",
				"operation": "redis_close_failed",
				"error":     err.Error(),
			})
		} else {
			redisCloseDuration := time.Since(redisCloseStart)
			pi.logger.Info(context.Background(), "Redis client closed successfully", map[string]interface{}{
				"component":             "participants_integration",
				"operation":             "redis_closed",
				"redis_close_duration":  redisCloseDuration.String(),
			})
		}
	}
	
	stopDuration := time.Since(stopStart)
	pi.logger.Info(context.Background(), "Participants integration stopped", map[string]interface{}{
		"component":      "participants_integration",
		"operation":      "stop_completed",
		"stop_duration":  stopDuration.String(),
	})
}

// IsHealthy проверяет состояние интеграции
func (pi *ParticipantsIntegration) IsHealthy() bool {
	pi.healthMutex.RLock()
	defer pi.healthMutex.RUnlock()
	
	// Считаем интеграцию здоровой, если Redis работает
	// MAX API может быть временно недоступен
	return pi.redisHealthy
}

// GetHealthStatus возвращает детальную информацию о состоянии
func (pi *ParticipantsIntegration) GetHealthStatus() map[string]interface{} {
	pi.healthMutex.RLock()
	defer pi.healthMutex.RUnlock()
	
	return map[string]interface{}{
		"redis_healthy":     pi.redisHealthy,
		"max_api_healthy":   pi.maxAPIHealthy,
		"last_health_check": pi.lastHealthCheck,
		"circuit_breaker_state": pi.getCircuitBreakerState(),
	}
}

// startHealthMonitoring запускает мониторинг здоровья компонентов с настраиваемым интервалом
func (pi *ParticipantsIntegration) startHealthMonitoring() {
	healthCheckInterval := getRedisHealthCheckInterval()
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()
	
	pi.logger.Info(context.Background(), "Starting Redis health monitoring", map[string]interface{}{
		"component":              "participants_integration",
		"operation":              "health_monitoring_start",
		"health_check_interval":  healthCheckInterval.String(),
		"automatic_reconnection": true,
	})
	
	for {
		select {
		case <-ticker.C:
			pi.checkHealth()
		}
	}
}

// getRedisHealthCheckInterval получает интервал проверки здоровья Redis
func getRedisHealthCheckInterval() time.Duration {
	if val := os.Getenv("REDIS_HEALTH_CHECK_INTERVAL"); val != "" {
		if interval, err := time.ParseDuration(val); err == nil && interval >= 10*time.Second && interval <= 5*time.Minute {
			return interval
		}
	}
	return 30 * time.Second // default
}

// checkHealth проверяет здоровье Redis и MAX API
func (pi *ParticipantsIntegration) checkHealth() {
	healthCheckStart := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Проверяем Redis
	redisHealthy := true
	var redisError error
	redisCheckStart := time.Now()
	
	if pi.redisClient != nil {
		if err := pi.redisClient.Ping(ctx).Err(); err != nil {
			redisHealthy = false
			redisError = err
		}
	} else {
		redisHealthy = false
	}
	
	redisCheckDuration := time.Since(redisCheckStart)
	
	// Получаем предыдущее состояние для сравнения
	pi.healthMutex.RLock()
	previousRedisHealthy := pi.redisHealthy
	pi.healthMutex.RUnlock()
	
	// Обновляем состояние
	pi.healthMutex.Lock()
	pi.redisHealthy = redisHealthy
	pi.lastHealthCheck = time.Now()
	pi.healthMutex.Unlock()
	
	healthCheckDuration := time.Since(healthCheckStart)
	
	// Логируем результаты проверки здоровья
	healthLogData := map[string]interface{}{
		"component":                "participants_integration",
		"operation":                "health_check",
		"redis_healthy":            redisHealthy,
		"redis_check_duration":     redisCheckDuration.String(),
		"total_health_check_duration": healthCheckDuration.String(),
		"circuit_breaker_state":    pi.getCircuitBreakerState(),
	}
	
	if redisError != nil {
		healthLogData["redis_error"] = redisError.Error()
	}
	
	// Логируем изменения состояния
	if !redisHealthy && previousRedisHealthy {
		pi.logger.Error(ctx, "Redis became unhealthy", healthLogData)
	} else if redisHealthy && !previousRedisHealthy {
		pi.logger.Info(ctx, "Redis became healthy", healthLogData)
	} else {
		// Периодическое логирование состояния (каждые 10 минут)
		if time.Since(pi.lastHealthCheck) > 10*time.Minute {
			pi.logger.Debug(ctx, "Health check completed", healthLogData)
		}
	}
	
	// Предупреждение о медленных проверках здоровья
	if healthCheckDuration > 5*time.Second {
		pi.logger.Warn(ctx, "Health check was slow", map[string]interface{}{
			"component":                "participants_integration",
			"operation":                "health_check_slow",
			"duration":                 healthCheckDuration.String(),
			"expected_max":             "5s",
		})
	}
}

// getCircuitBreakerState возвращает состояние circuit breaker
func (pi *ParticipantsIntegration) getCircuitBreakerState() string {
	if pi.circuitBreaker == nil {
		return "unknown"
	}
	
	pi.circuitBreaker.mutex.RLock()
	defer pi.circuitBreaker.mutex.RUnlock()
	
	switch pi.circuitBreaker.state {
	case usecase.CircuitClosed:
		return "closed"
	case usecase.CircuitOpen:
		return "open"
	case usecase.CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// createRedisClientWithRetry создает клиент Redis с retry logic и автоматическим переподключением
func createRedisClientWithRetry(logger *logger.Logger) (*redis.Client, error) {
	connectionStart := time.Now()
	
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
		logger.Info(context.Background(), "Using default Redis URL", map[string]interface{}{
			"component":   "participants_integration",
			"operation":   "redis_connection",
			"redis_url":   redisURL,
			"source":      "default",
		})
	} else {
		logger.Info(context.Background(), "Using configured Redis URL", map[string]interface{}{
			"component":   "participants_integration",
			"operation":   "redis_connection",
			"redis_url":   redisURL,
			"source":      "environment",
		})
	}
	
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error(context.Background(), "Failed to parse Redis URL", map[string]interface{}{
			"component":   "participants_integration",
			"operation":   "redis_url_parse_failed",
			"redis_url":   redisURL,
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}
	
	// Настройка автоматического переподключения
	maxRetries := getRedisMaxRetries()
	retryDelay := getRedisRetryDelay()
	
	// Настройка Redis клиента с автоматическим переподключением
	opt.MaxRetries = maxRetries
	opt.MinRetryBackoff = retryDelay
	opt.MaxRetryBackoff = retryDelay * 8 // Максимальная задержка
	opt.DialTimeout = 10 * time.Second
	opt.ReadTimeout = 5 * time.Second
	opt.WriteTimeout = 5 * time.Second
	opt.PoolSize = 10
	opt.MinIdleConns = 2
	opt.MaxConnAge = 30 * time.Minute
	opt.PoolTimeout = 4 * time.Second
	opt.IdleTimeout = 5 * time.Minute
	opt.IdleCheckFrequency = 1 * time.Minute
	
	client := redis.NewClient(opt)
	
	logger.Info(context.Background(), "Starting Redis connection attempts with automatic reconnection", map[string]interface{}{
		"component":              "participants_integration",
		"operation":              "redis_connection_start",
		"max_retries":            maxRetries,
		"initial_delay":          retryDelay.String(),
		"max_retry_backoff":      (retryDelay * 8).String(),
		"dial_timeout":           "10s",
		"pool_size":              10,
		"min_idle_conns":         2,
		"automatic_reconnection": true,
	})
	
	// Первоначальная проверка подключения с retry logic
	initialMaxRetries := 3
	initialRetryDelay := 1 * time.Second
	
	for attempt := 1; attempt <= initialMaxRetries; attempt++ {
		attemptStart := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		if err := client.Ping(ctx).Err(); err != nil {
			cancel()
			attemptDuration := time.Since(attemptStart)
			
			logger.Warn(context.Background(), "Redis initial connection attempt failed", map[string]interface{}{
				"component":        "participants_integration",
				"operation":        "redis_initial_connection_attempt_failed",
				"attempt":          attempt,
				"max_retries":      initialMaxRetries,
				"error":            err.Error(),
				"retry_delay":      initialRetryDelay.String(),
				"attempt_duration": attemptDuration.String(),
			})
			
			if attempt == initialMaxRetries {
				client.Close()
				totalDuration := time.Since(connectionStart)
				
				logger.Error(context.Background(), "All Redis initial connection attempts failed", map[string]interface{}{
					"component":      "participants_integration",
					"operation":      "redis_initial_connection_failed",
					"total_attempts": initialMaxRetries,
					"total_duration": totalDuration.String(),
					"final_error":    err.Error(),
					"note":           "Redis client configured for automatic reconnection, will retry on operations",
				})
				return nil, fmt.Errorf("failed to connect to Redis after %d attempts: %w", initialMaxRetries, err)
			}
			
			logger.Debug(context.Background(), "Waiting before retry", map[string]interface{}{
				"component":   "participants_integration",
				"operation":   "redis_retry_wait",
				"attempt":     attempt,
				"wait_time":   initialRetryDelay.String(),
			})
			
			time.Sleep(initialRetryDelay)
			initialRetryDelay *= 2 // Exponential backoff
			continue
		}
		
		cancel()
		attemptDuration := time.Since(attemptStart)
		totalDuration := time.Since(connectionStart)
		
		logger.Info(context.Background(), "Redis connection established with automatic reconnection", map[string]interface{}{
			"component":              "participants_integration",
			"operation":              "redis_connection_success",
			"attempt":                attempt,
			"redis_url":              redisURL,
			"attempt_duration":       attemptDuration.String(),
			"total_duration":         totalDuration.String(),
			"automatic_reconnection": true,
			"max_retries":            maxRetries,
			"retry_backoff":          retryDelay.String(),
		})
		return client, nil
	}
	
	return nil, fmt.Errorf("unexpected error in Redis connection retry loop")
}

// getRedisMaxRetries получает максимальное количество попыток переподключения Redis
func getRedisMaxRetries() int {
	if val := os.Getenv("REDIS_MAX_RETRIES"); val != "" {
		if retries, err := strconv.Atoi(val); err == nil && retries >= 1 && retries <= 20 {
			return retries
		}
	}
	return 5 // default
}

// getRedisRetryDelay получает задержку между попытками переподключения Redis
func getRedisRetryDelay() time.Duration {
	if val := os.Getenv("REDIS_RETRY_DELAY"); val != "" {
		if delay, err := time.ParseDuration(val); err == nil && delay >= 100*time.Millisecond && delay <= 30*time.Second {
			return delay
		}
	}
	return 1 * time.Second // default
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

// Circuit Breaker methods

// CanExecute проверяет, можно ли выполнить операцию
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	switch cb.state {
	case usecase.CircuitClosed:
		return true
	case usecase.CircuitOpen:
		// Проверяем, не истек ли timeout
		if time.Since(cb.lastFailureTime) > cb.timeout {
			return true // Переходим в half-open
		}
		return false
	case usecase.CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess записывает успешное выполнение
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.failureCount = 0
	cb.state = usecase.CircuitClosed
}

// RecordFailure записывает неудачное выполнение
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	if cb.failureCount >= cb.threshold {
		cb.state = usecase.CircuitOpen
	}
}

// GetState возвращает текущее состояние
func (cb *CircuitBreaker) GetState() usecase.CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	// Проверяем переход из Open в HalfOpen
	if cb.state == usecase.CircuitOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		cb.mutex.RUnlock()
		cb.mutex.Lock()
		if cb.state == usecase.CircuitOpen && time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = usecase.CircuitHalfOpen
		}
		cb.mutex.Unlock()
		cb.mutex.RLock()
	}
	
	return cb.state
}