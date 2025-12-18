package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"maxbot-service/internal/domain"
)

// RedisMonitoringService реализует MonitoringService используя Redis
type RedisMonitoringService struct {
	client       *redis.Client
	profileCache domain.ProfileCacheService
}

// NewRedisMonitoringService создает новый экземпляр RedisMonitoringService
func NewRedisMonitoringService(client *redis.Client, profileCache domain.ProfileCacheService) *RedisMonitoringService {
	return &RedisMonitoringService{
		client:       client,
		profileCache: profileCache,
	}
}

// RecordWebhookEvent записывает событие обработки webhook
func (m *RedisMonitoringService) RecordWebhookEvent(ctx context.Context, event domain.WebhookEventMetric) error {
	// Добавляем таймаут для Redis операции
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Сериализуем событие
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook event: %w", err)
	}

	// Создаем ключи для хранения
	timestamp := event.ProcessedAt.Unix()
	eventKey := fmt.Sprintf("webhook:events:%d:%s", timestamp, event.UserID)
	
	// Используем pipeline для атомарности операций
	pipe := m.client.Pipeline()
	
	// Сохраняем событие с TTL 7 дней
	pipe.Set(ctx, eventKey, data, 7*24*time.Hour)
	
	// Обновляем счетчики по типам событий
	dailyKey := fmt.Sprintf("webhook:stats:daily:%s", event.ProcessedAt.Format("2006-01-02"))
	pipe.HIncrBy(ctx, dailyKey, "total_events", 1)
	pipe.HIncrBy(ctx, dailyKey, fmt.Sprintf("events_%s", event.EventType), 1)
	
	if event.Success {
		pipe.HIncrBy(ctx, dailyKey, "successful_events", 1)
	} else {
		pipe.HIncrBy(ctx, dailyKey, "failed_events", 1)
	}
	
	if event.ProfileFound {
		pipe.HIncrBy(ctx, dailyKey, "profiles_extracted", 1)
	}
	
	if event.ProfileStored {
		pipe.HIncrBy(ctx, dailyKey, "profiles_stored", 1)
	}
	
	// Добавляем время обработки для расчета среднего
	pipe.LPush(ctx, fmt.Sprintf("webhook:processing_times:%s", event.ProcessedAt.Format("2006-01-02")), event.ProcessingTime)
	pipe.LTrim(ctx, fmt.Sprintf("webhook:processing_times:%s", event.ProcessedAt.Format("2006-01-02")), 0, 999) // Храним последние 1000 значений
	
	// Устанавливаем TTL для счетчиков (30 дней)
	pipe.Expire(ctx, dailyKey, 30*24*time.Hour)
	pipe.Expire(ctx, fmt.Sprintf("webhook:processing_times:%s", event.ProcessedAt.Format("2006-01-02")), 30*24*time.Hour)
	
	// Выполняем pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record webhook event: %w", err)
	}
	
	return nil
}

// GetWebhookStats возвращает статистику обработки webhook событий
func (m *RedisMonitoringService) GetWebhookStats(ctx context.Context, period domain.TimePeriod) (*domain.WebhookStats, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	stats := &domain.WebhookStats{
		Period:       period,
		EventsByType: make(map[string]int64),
		ErrorsByType: make(map[string]int64),
	}

	// Получаем все дни в периоде
	days := m.getDaysInPeriod(period)
	
	var totalProcessingTime int64
	var processingTimeCount int64
	
	for _, day := range days {
		dailyKey := fmt.Sprintf("webhook:stats:daily:%s", day)
		
		// Получаем статистику за день
		dailyStats, err := m.client.HGetAll(ctx, dailyKey).Result()
		if err != nil && err != redis.Nil {
			continue // Пропускаем дни с ошибками
		}
		
		// Агрегируем данные
		for field, value := range dailyStats {
			count, _ := strconv.ParseInt(value, 10, 64)
			
			switch field {
			case "total_events":
				stats.TotalEvents += count
			case "successful_events":
				stats.SuccessfulEvents += count
			case "failed_events":
				stats.FailedEvents += count
			case "profiles_extracted":
				stats.ProfilesExtracted += count
			case "profiles_stored":
				stats.ProfilesStored += count
			default:
				if strings.HasPrefix(field, "events_") {
					eventType := strings.TrimPrefix(field, "events_")
					stats.EventsByType[eventType] += count
				}
			}
		}
		
		// Получаем времена обработки для расчета среднего
		processingTimesKey := fmt.Sprintf("webhook:processing_times:%s", day)
		times, err := m.client.LRange(ctx, processingTimesKey, 0, -1).Result()
		if err == nil {
			for _, timeStr := range times {
				if time, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
					totalProcessingTime += time
					processingTimeCount++
				}
			}
		}
	}
	
	// Рассчитываем среднее время обработки
	if processingTimeCount > 0 {
		stats.AverageProcessingTime = float64(totalProcessingTime) / float64(processingTimeCount)
	}
	
	return stats, nil
}

// GetProfileCoverage возвращает метрики покрытия профилей
func (m *RedisMonitoringService) GetProfileCoverage(ctx context.Context) (*domain.ProfileCoverage, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Получаем статистику профилей из кэша
	profileStats, err := m.profileCache.GetProfileStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile stats: %w", err)
	}

	coverage := &domain.ProfileCoverage{
		TotalUsers:       profileStats.TotalProfiles,
		UsersWithProfiles: profileStats.TotalProfiles,
		UsersWithFullNames: profileStats.ProfilesWithFullName,
		ProfilesBySource: profileStats.ProfilesBySource,
		LastUpdated:      time.Now(),
	}

	// Рассчитываем проценты
	if coverage.TotalUsers > 0 {
		coverage.CoveragePercentage = float64(coverage.UsersWithProfiles) / float64(coverage.TotalUsers) * 100
		coverage.FullNamePercentage = float64(coverage.UsersWithFullNames) / float64(coverage.TotalUsers) * 100
	}

	return coverage, nil
}

// GetProfileQualityReport возвращает отчет о качестве профильных данных
func (m *RedisMonitoringService) GetProfileQualityReport(ctx context.Context) (*domain.ProfileQualityReport, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// Получаем базовую статистику профилей
	profileStats, err := m.profileCache.GetProfileStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile stats: %w", err)
	}

	report := &domain.ProfileQualityReport{
		GeneratedAt:      time.Now(),
		TotalProfiles:    profileStats.TotalProfiles,
		SourceBreakdown:  make(map[domain.ProfileSource]domain.SourceQuality),
		RecommendedActions: []string{},
		DataIssues:       []domain.ProfileDataIssue{},
	}

	// Анализируем качество по источникам
	for source, count := range profileStats.ProfilesBySource {
		quality := domain.SourceQuality{
			Count: count,
		}
		
		// Рассчитываем метрики качества для каждого источника
		switch source {
		case domain.SourceWebhook:
			quality.QualityScore = 85.0 // Webhook данные обычно хорошего качества
			quality.AverageAge = 7.0    // Предполагаем свежие данные
		case domain.SourceUserInput:
			quality.QualityScore = 95.0 // Пользовательский ввод самого высокого качества
			quality.AverageAge = 3.0    // Очень свежие данные
		case domain.SourceDefault:
			quality.QualityScore = 30.0 // Данные по умолчанию низкого качества
			quality.AverageAge = 30.0   // Могут быть старыми
		}
		
		report.SourceBreakdown[source] = quality
	}

	// Рассчитываем общие метрики качества
	report.QualityMetrics = m.calculateQualityMetrics(profileStats)

	// Генерируем рекомендации
	report.RecommendedActions = m.generateRecommendations(report.QualityMetrics, profileStats)

	// Выявляем проблемы с данными
	report.DataIssues = m.identifyDataIssues(profileStats)

	return report, nil
}

// calculateQualityMetrics рассчитывает метрики качества профилей
func (m *RedisMonitoringService) calculateQualityMetrics(stats *domain.ProfileStats) domain.ProfileQualityMetrics {
	metrics := domain.ProfileQualityMetrics{}

	if stats.TotalProfiles == 0 {
		return metrics
	}

	// Рассчитываем полноту профилей
	metrics.CompleteProfiles = stats.ProfilesWithFullName
	metrics.PartialProfiles = stats.TotalProfiles - stats.ProfilesWithFullName
	
	// Рассчитываем балл полноты (0-100)
	metrics.CompletenessScore = float64(metrics.CompleteProfiles) / float64(stats.TotalProfiles) * 100

	// Рассчитываем балл свежести на основе источников
	webhookProfiles := stats.ProfilesBySource[domain.SourceWebhook]
	userInputProfiles := stats.ProfilesBySource[domain.SourceUserInput]
	
	if stats.TotalProfiles > 0 {
		freshProfilesRatio := float64(webhookProfiles+userInputProfiles) / float64(stats.TotalProfiles)
		metrics.FreshnessScore = freshProfilesRatio * 100
	}

	// Общий балл качества (среднее между полнотой и свежестью)
	metrics.QualityScore = (metrics.CompletenessScore + metrics.FreshnessScore) / 2

	return metrics
}

// generateRecommendations генерирует рекомендации по улучшению качества данных
func (m *RedisMonitoringService) generateRecommendations(metrics domain.ProfileQualityMetrics, stats *domain.ProfileStats) []string {
	var recommendations []string

	// Рекомендации по полноте данных
	if metrics.CompletenessScore < 70 {
		recommendations = append(recommendations, "Увеличить покрытие полных имен пользователей через webhook события")
	}

	// Рекомендации по источникам данных
	defaultProfiles := stats.ProfilesBySource[domain.SourceDefault]
	if defaultProfiles > stats.TotalProfiles/2 {
		recommendations = append(recommendations, "Слишком много профилей с данными по умолчанию - активизировать сбор через webhook")
	}

	// Рекомендации по пользовательскому вводу
	userInputProfiles := stats.ProfilesBySource[domain.SourceUserInput]
	if userInputProfiles < stats.TotalProfiles/10 {
		recommendations = append(recommendations, "Добавить больше возможностей для пользователей указывать свои имена")
	}

	// Общие рекомендации по качеству
	if metrics.QualityScore < 60 {
		recommendations = append(recommendations, "Общее качество данных требует улучшения - проверить настройки webhook")
	}

	return recommendations
}

// identifyDataIssues выявляет проблемы с данными профилей
func (m *RedisMonitoringService) identifyDataIssues(stats *domain.ProfileStats) []domain.ProfileDataIssue {
	var issues []domain.ProfileDataIssue

	// Проблема: слишком много неполных профилей
	incompleteProfiles := stats.TotalProfiles - stats.ProfilesWithFullName
	if incompleteProfiles > stats.TotalProfiles/2 {
		issues = append(issues, domain.ProfileDataIssue{
			Type:        "incomplete_profiles",
			Description: "Более 50% профилей не содержат полных имен",
			Count:       incompleteProfiles,
			Severity:    "high",
		})
	}

	// Проблема: отсутствие webhook данных
	webhookProfiles := stats.ProfilesBySource[domain.SourceWebhook]
	if webhookProfiles == 0 && stats.TotalProfiles > 0 {
		issues = append(issues, domain.ProfileDataIssue{
			Type:        "no_webhook_data",
			Description: "Отсутствуют данные из webhook событий",
			Count:       stats.TotalProfiles,
			Severity:    "high",
		})
	}

	// Проблема: слишком много данных по умолчанию
	defaultProfiles := stats.ProfilesBySource[domain.SourceDefault]
	if defaultProfiles > stats.TotalProfiles*3/4 {
		issues = append(issues, domain.ProfileDataIssue{
			Type:        "too_many_defaults",
			Description: "Более 75% профилей используют данные по умолчанию",
			Count:       defaultProfiles,
			Severity:    "medium",
		})
	}

	return issues
}

// getDaysInPeriod возвращает список дней в заданном периоде
func (m *RedisMonitoringService) getDaysInPeriod(period domain.TimePeriod) []string {
	var days []string
	
	current := period.From.Truncate(24 * time.Hour)
	end := period.To.Truncate(24 * time.Hour)
	
	for current.Before(end) || current.Equal(end) {
		days = append(days, current.Format("2006-01-02"))
		current = current.Add(24 * time.Hour)
	}
	
	return days
}

// IsHealthy проверяет доступность Redis для мониторинга
func (m *RedisMonitoringService) IsHealthy(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	
	err := m.client.Ping(ctx).Err()
	return err == nil
}