package cache

import (
	"context"
	"sync"
	"time"

	"maxbot-service/internal/domain"
)

// CircuitBreakerState представляет состояние circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// ProfileCacheCircuitBreaker реализует circuit breaker pattern для ProfileCacheService
type ProfileCacheCircuitBreaker struct {
	cache           domain.ProfileCacheService
	failureCount    int
	lastFailureTime time.Time
	state           CircuitBreakerState
	mutex           sync.RWMutex
	
	// Конфигурация
	maxFailures     int
	timeout         time.Duration
	resetTimeout    time.Duration
}

// NewProfileCacheCircuitBreaker создает новый circuit breaker для profile cache
func NewProfileCacheCircuitBreaker(cache domain.ProfileCacheService) *ProfileCacheCircuitBreaker {
	return &ProfileCacheCircuitBreaker{
		cache:        cache,
		state:        StateClosed,
		maxFailures:  5,           // Максимум 5 ошибок подряд
		timeout:      30 * time.Second, // Таймаут для операций
		resetTimeout: 60 * time.Second, // Время до попытки восстановления
	}
}

// StoreProfile сохраняет профиль с circuit breaker логикой
func (cb *ProfileCacheCircuitBreaker) StoreProfile(ctx context.Context, userID string, profile domain.UserProfileCache) error {
	if !cb.canExecute() {
		// Circuit breaker открыт - возвращаем ошибку для graceful degradation
		return domain.ErrCacheUnavailable
	}
	
	ctx, cancel := context.WithTimeout(ctx, cb.timeout)
	defer cancel()
	
	err := cb.cache.StoreProfile(ctx, userID, profile)
	cb.recordResult(err)
	
	return err
}

// GetProfile получает профиль с circuit breaker логикой
func (cb *ProfileCacheCircuitBreaker) GetProfile(ctx context.Context, userID string) (*domain.UserProfileCache, error) {
	if !cb.canExecute() {
		// Circuit breaker открыт - возвращаем nil для graceful degradation (Requirements 3.4, 7.5)
		return nil, nil
	}
	
	ctx, cancel := context.WithTimeout(ctx, cb.timeout)
	defer cancel()
	
	profile, err := cb.cache.GetProfile(ctx, userID)
	cb.recordResult(err)
	
	if err != nil {
		// Возвращаем nil вместо ошибки для graceful degradation
		return nil, nil
	}
	
	return profile, nil
}

// UpdateProfile обновляет профиль с circuit breaker логикой
func (cb *ProfileCacheCircuitBreaker) UpdateProfile(ctx context.Context, userID string, updates domain.ProfileUpdates) error {
	if !cb.canExecute() {
		return domain.ErrCacheUnavailable
	}
	
	ctx, cancel := context.WithTimeout(ctx, cb.timeout)
	defer cancel()
	
	err := cb.cache.UpdateProfile(ctx, userID, updates)
	cb.recordResult(err)
	
	return err
}

// GetProfileStats получает статистику с circuit breaker логикой
func (cb *ProfileCacheCircuitBreaker) GetProfileStats(ctx context.Context) (*domain.ProfileStats, error) {
	if !cb.canExecute() {
		// Возвращаем пустую статистику при недоступности кэша
		return &domain.ProfileStats{
			TotalProfiles:       0,
			ProfilesWithFullName: 0,
			ProfilesBySource:    make(map[domain.ProfileSource]int64),
		}, nil
	}
	
	ctx, cancel := context.WithTimeout(ctx, cb.timeout)
	defer cancel()
	
	stats, err := cb.cache.GetProfileStats(ctx)
	cb.recordResult(err)
	
	if err != nil {
		// Возвращаем пустую статистику при ошибке
		return &domain.ProfileStats{
			TotalProfiles:       0,
			ProfilesWithFullName: 0,
			ProfilesBySource:    make(map[domain.ProfileSource]int64),
		}, nil
	}
	
	return stats, nil
}

// canExecute проверяет, можно ли выполнить операцию
func (cb *ProfileCacheCircuitBreaker) canExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Проверяем, не пора ли попробовать снова
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = StateHalfOpen
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult записывает результат операции
func (cb *ProfileCacheCircuitBreaker) recordResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()
		
		if cb.failureCount >= cb.maxFailures {
			cb.state = StateOpen
		}
	} else {
		// Успешная операция - сбрасываем счетчик
		cb.failureCount = 0
		cb.state = StateClosed
	}
}

// GetState возвращает текущее состояние circuit breaker
func (cb *ProfileCacheCircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetFailureCount возвращает количество ошибок
func (cb *ProfileCacheCircuitBreaker) GetFailureCount() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failureCount
}