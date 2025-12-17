package domain

import (
	"context"
	"time"
)

// ParticipantsCache определяет интерфейс для кэширования количества участников
type ParticipantsCache interface {
	// Get получает количество участников из кэша
	Get(ctx context.Context, chatID int64) (*ParticipantsInfo, error)
	
	// Set сохраняет количество участников в кэш
	Set(ctx context.Context, chatID int64, count int, ttl time.Duration) error
	
	// GetMultiple получает количество участников для нескольких чатов
	GetMultiple(ctx context.Context, chatIDs []int64) (map[int64]*ParticipantsInfo, error)
	
	// SetMultiple сохраняет количество участников для нескольких чатов
	SetMultiple(ctx context.Context, data map[int64]int, ttl time.Duration) error
	
	// Delete удаляет данные из кэша
	Delete(ctx context.Context, chatID int64) error
	
	// GetStaleChats возвращает чаты с устаревшими данными
	GetStaleChats(ctx context.Context, olderThan time.Duration, limit int) ([]int64, error)
}

// ParticipantsInfo содержит информацию о количестве участников
type ParticipantsInfo struct {
	Count     int       `json:"count"`
	UpdatedAt time.Time `json:"updated_at"`
	Source    string    `json:"source"` // "cache", "api", "database"
}

// ParticipantsUpdater определяет интерфейс для обновления количества участников
type ParticipantsUpdater interface {
	// UpdateSingle обновляет количество участников для одного чата
	UpdateSingle(ctx context.Context, chatID int64, maxChatID string) (*ParticipantsInfo, error)
	
	// UpdateBatch обновляет количество участников для нескольких чатов
	UpdateBatch(ctx context.Context, chats []ChatUpdateRequest) (map[int64]*ParticipantsInfo, error)
	
	// UpdateStale обновляет устаревшие данные
	UpdateStale(ctx context.Context, olderThan time.Duration, batchSize int) (int, error)
	
	// UpdateAll обновляет все чаты (для ночного обновления)
	UpdateAll(ctx context.Context, batchSize int) (int, error)
}

// ChatUpdateRequest содержит данные для обновления чата
type ChatUpdateRequest struct {
	ChatID    int64  `json:"chat_id"`
	MaxChatID string `json:"max_chat_id"`
}

// ParticipantsConfig содержит конфигурацию для работы с участниками
type ParticipantsConfig struct {
	CacheTTL              time.Duration `env:"PARTICIPANTS_CACHE_TTL" default:"1h"`
	UpdateInterval        time.Duration `env:"PARTICIPANTS_UPDATE_INTERVAL" default:"15m"`
	FullUpdateHour        int           `env:"PARTICIPANTS_FULL_UPDATE_HOUR" default:"3"`
	BatchSize             int           `env:"PARTICIPANTS_BATCH_SIZE" default:"50"`
	MaxAPITimeout         time.Duration `env:"PARTICIPANTS_MAX_API_TIMEOUT" default:"30s"`
	StaleThreshold        time.Duration `env:"PARTICIPANTS_STALE_THRESHOLD" default:"1h"`
	EnableBackgroundSync  bool          `env:"PARTICIPANTS_ENABLE_BACKGROUND_SYNC" default:"true"`
	EnableLazyUpdate      bool          `env:"PARTICIPANTS_ENABLE_LAZY_UPDATE" default:"true"`
	MaxRetries            int           `env:"PARTICIPANTS_MAX_RETRIES" default:"3"`
}