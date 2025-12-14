package config

import (
	"chat-service/internal/domain"
	"os"
	"strconv"
	"time"
)

// LoadParticipantsConfig загружает конфигурацию для работы с участниками
func LoadParticipantsConfig() *domain.ParticipantsConfig {
	config := &domain.ParticipantsConfig{
		CacheTTL:              1 * time.Hour,
		UpdateInterval:        15 * time.Minute,
		FullUpdateHour:        3,
		BatchSize:             50,
		MaxAPITimeout:         30 * time.Second,
		StaleThreshold:        1 * time.Hour,
		EnableBackgroundSync:  true,
		EnableLazyUpdate:      true,
	}
	
	// Загружаем из переменных окружения
	if val := os.Getenv("PARTICIPANTS_CACHE_TTL"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.CacheTTL = duration
		}
	}
	
	if val := os.Getenv("PARTICIPANTS_UPDATE_INTERVAL"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.UpdateInterval = duration
		}
	}
	
	if val := os.Getenv("PARTICIPANTS_FULL_UPDATE_HOUR"); val != "" {
		if hour, err := strconv.Atoi(val); err == nil && hour >= 0 && hour <= 23 {
			config.FullUpdateHour = hour
		}
	}
	
	if val := os.Getenv("PARTICIPANTS_BATCH_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil && size > 0 {
			config.BatchSize = size
		}
	}
	
	if val := os.Getenv("PARTICIPANTS_MAX_API_TIMEOUT"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.MaxAPITimeout = duration
		}
	}
	
	if val := os.Getenv("PARTICIPANTS_STALE_THRESHOLD"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.StaleThreshold = duration
		}
	}
	
	if val := os.Getenv("PARTICIPANTS_ENABLE_BACKGROUND_SYNC"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.EnableBackgroundSync = enabled
		}
	}
	
	if val := os.Getenv("PARTICIPANTS_ENABLE_LAZY_UPDATE"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.EnableLazyUpdate = enabled
		}
	}
	
	return config
}