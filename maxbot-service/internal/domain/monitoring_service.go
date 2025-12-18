package domain

import (
	"context"
	"time"
)

// MonitoringService определяет интерфейс для мониторинга и аналитики
type MonitoringService interface {
	// RecordWebhookEvent записывает событие обработки webhook
	RecordWebhookEvent(ctx context.Context, event WebhookEventMetric) error
	// GetWebhookStats возвращает статистику обработки webhook событий
	GetWebhookStats(ctx context.Context, period TimePeriod) (*WebhookStats, error)
	// GetProfileCoverage возвращает метрики покрытия профилей
	GetProfileCoverage(ctx context.Context) (*ProfileCoverage, error)
	// GetProfileQualityReport возвращает отчет о качестве профильных данных
	GetProfileQualityReport(ctx context.Context) (*ProfileQualityReport, error)
}

// WebhookEventMetric представляет метрику события webhook
type WebhookEventMetric struct {
	EventType     string    `json:"event_type"`     // Тип события (message_new, callback_query)
	UserID        string    `json:"user_id"`        // ID пользователя
	ProcessedAt   time.Time `json:"processed_at"`   // Время обработки
	Success       bool      `json:"success"`        // Успешность обработки
	ErrorMessage  string    `json:"error_message"`  // Сообщение об ошибке (если есть)
	ProcessingTime int64    `json:"processing_time"` // Время обработки в миллисекундах
	ProfileFound  bool      `json:"profile_found"`  // Найден ли профиль в событии
	ProfileStored bool      `json:"profile_stored"` // Сохранен ли профиль в кэше
}

// TimePeriod определяет временной период для статистики
type TimePeriod struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// WebhookStats содержит статистику обработки webhook событий
type WebhookStats struct {
	Period              TimePeriod                `json:"period"`
	TotalEvents         int64                     `json:"total_events"`
	SuccessfulEvents    int64                     `json:"successful_events"`
	FailedEvents        int64                     `json:"failed_events"`
	EventsByType        map[string]int64          `json:"events_by_type"`
	ProfilesExtracted   int64                     `json:"profiles_extracted"`
	ProfilesStored      int64                     `json:"profiles_stored"`
	AverageProcessingTime float64                 `json:"average_processing_time_ms"`
	ErrorsByType        map[string]int64          `json:"errors_by_type"`
}

// ProfileCoverage содержит метрики покрытия профилей
type ProfileCoverage struct {
	TotalUsers           int64   `json:"total_users"`            // Общее количество пользователей
	UsersWithProfiles    int64   `json:"users_with_profiles"`    // Пользователи с профилями
	UsersWithFullNames   int64   `json:"users_with_full_names"`  // Пользователи с полными именами
	CoveragePercentage   float64 `json:"coverage_percentage"`    // Процент покрытия профилями
	FullNamePercentage   float64 `json:"full_name_percentage"`   // Процент полных имен
	ProfilesBySource     map[ProfileSource]int64 `json:"profiles_by_source"`
	LastUpdated          time.Time `json:"last_updated"`
}

// ProfileQualityReport содержит отчет о качестве профильных данных
type ProfileQualityReport struct {
	GeneratedAt          time.Time                    `json:"generated_at"`
	TotalProfiles        int64                        `json:"total_profiles"`
	QualityMetrics       ProfileQualityMetrics        `json:"quality_metrics"`
	SourceBreakdown      map[ProfileSource]SourceQuality `json:"source_breakdown"`
	RecommendedActions   []string                     `json:"recommended_actions"`
	DataIssues           []ProfileDataIssue           `json:"data_issues"`
}

// ProfileQualityMetrics содержит метрики качества профилей
type ProfileQualityMetrics struct {
	CompleteProfiles     int64   `json:"complete_profiles"`      // Профили с полной информацией
	PartialProfiles      int64   `json:"partial_profiles"`       // Профили с частичной информацией
	EmptyProfiles        int64   `json:"empty_profiles"`         // Пустые профили
	StaleProfiles        int64   `json:"stale_profiles"`         // Устаревшие профили (>30 дней)
	QualityScore         float64 `json:"quality_score"`          // Общий балл качества (0-100)
	CompletenessScore    float64 `json:"completeness_score"`     // Балл полноты данных
	FreshnessScore       float64 `json:"freshness_score"`        // Балл свежести данных
}

// SourceQuality содержит метрики качества по источнику
type SourceQuality struct {
	Count             int64   `json:"count"`
	CompleteProfiles  int64   `json:"complete_profiles"`
	AverageAge        float64 `json:"average_age_days"`
	QualityScore      float64 `json:"quality_score"`
}

// ProfileDataIssue представляет проблему с данными профиля
type ProfileDataIssue struct {
	Type        string `json:"type"`        // Тип проблемы
	Description string `json:"description"` // Описание проблемы
	Count       int64  `json:"count"`       // Количество затронутых профилей
	Severity    string `json:"severity"`    // Серьезность (low, medium, high)
}

// Предопределенные временные периоды
var (
	LastHour  = func() TimePeriod {
		now := time.Now()
		return TimePeriod{From: now.Add(-time.Hour), To: now}
	}
	LastDay   = func() TimePeriod {
		now := time.Now()
		return TimePeriod{From: now.Add(-24*time.Hour), To: now}
	}
	LastWeek  = func() TimePeriod {
		now := time.Now()
		return TimePeriod{From: now.Add(-7*24*time.Hour), To: now}
	}
	LastMonth = func() TimePeriod {
		now := time.Now()
		return TimePeriod{From: now.Add(-30*24*time.Hour), To: now}
	}
)