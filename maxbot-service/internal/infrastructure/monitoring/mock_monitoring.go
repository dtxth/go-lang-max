package monitoring

import (
	"context"
	"time"

	"maxbot-service/internal/domain"
)

// MockMonitoringService реализует MonitoringService для тестирования
type MockMonitoringService struct {
	events []domain.WebhookEventMetric
}

// NewMockMonitoringService создает новый экземпляр MockMonitoringService
func NewMockMonitoringService() *MockMonitoringService {
	return &MockMonitoringService{
		events: make([]domain.WebhookEventMetric, 0),
	}
}

// RecordWebhookEvent записывает событие обработки webhook (mock)
func (m *MockMonitoringService) RecordWebhookEvent(ctx context.Context, event domain.WebhookEventMetric) error {
	m.events = append(m.events, event)
	return nil
}

// GetWebhookStats возвращает статистику обработки webhook событий (mock)
func (m *MockMonitoringService) GetWebhookStats(ctx context.Context, period domain.TimePeriod) (*domain.WebhookStats, error) {
	// Фильтруем события по периоду
	var filteredEvents []domain.WebhookEventMetric
	for _, event := range m.events {
		if event.ProcessedAt.After(period.From) && event.ProcessedAt.Before(period.To) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	stats := &domain.WebhookStats{
		Period:       period,
		EventsByType: make(map[string]int64),
		ErrorsByType: make(map[string]int64),
	}

	var totalProcessingTime int64
	
	for _, event := range filteredEvents {
		stats.TotalEvents++
		stats.EventsByType[event.EventType]++
		
		if event.Success {
			stats.SuccessfulEvents++
		} else {
			stats.FailedEvents++
			if event.ErrorMessage != "" {
				stats.ErrorsByType[event.ErrorMessage]++
			}
		}
		
		if event.ProfileFound {
			stats.ProfilesExtracted++
		}
		
		if event.ProfileStored {
			stats.ProfilesStored++
		}
		
		totalProcessingTime += event.ProcessingTime
	}

	if len(filteredEvents) > 0 {
		stats.AverageProcessingTime = float64(totalProcessingTime) / float64(len(filteredEvents))
	}

	return stats, nil
}

// GetProfileCoverage возвращает метрики покрытия профилей (mock)
func (m *MockMonitoringService) GetProfileCoverage(ctx context.Context) (*domain.ProfileCoverage, error) {
	return &domain.ProfileCoverage{
		TotalUsers:         1000,
		UsersWithProfiles:  800,
		UsersWithFullNames: 600,
		CoveragePercentage: 80.0,
		FullNamePercentage: 60.0,
		ProfilesBySource: map[domain.ProfileSource]int64{
			domain.SourceWebhook:   500,
			domain.SourceUserInput: 200,
			domain.SourceDefault:   100,
		},
		LastUpdated: time.Now(),
	}, nil
}

// GetProfileQualityReport возвращает отчет о качестве профильных данных (mock)
func (m *MockMonitoringService) GetProfileQualityReport(ctx context.Context) (*domain.ProfileQualityReport, error) {
	return &domain.ProfileQualityReport{
		GeneratedAt:   time.Now(),
		TotalProfiles: 800,
		QualityMetrics: domain.ProfileQualityMetrics{
			CompleteProfiles:  600,
			PartialProfiles:   150,
			EmptyProfiles:     50,
			StaleProfiles:     100,
			QualityScore:      75.0,
			CompletenessScore: 75.0,
			FreshnessScore:    80.0,
		},
		SourceBreakdown: map[domain.ProfileSource]domain.SourceQuality{
			domain.SourceWebhook: {
				Count:            500,
				CompleteProfiles: 400,
				AverageAge:       7.0,
				QualityScore:     85.0,
			},
			domain.SourceUserInput: {
				Count:            200,
				CompleteProfiles: 180,
				AverageAge:       3.0,
				QualityScore:     95.0,
			},
			domain.SourceDefault: {
				Count:            100,
				CompleteProfiles: 20,
				AverageAge:       30.0,
				QualityScore:     30.0,
			},
		},
		RecommendedActions: []string{
			"Увеличить покрытие полных имен пользователей через webhook события",
			"Добавить больше возможностей для пользователей указывать свои имена",
		},
		DataIssues: []domain.ProfileDataIssue{
			{
				Type:        "incomplete_profiles",
				Description: "Профили без полных имен",
				Count:       200,
				Severity:    "medium",
			},
			{
				Type:        "stale_profiles",
				Description: "Устаревшие профили (>30 дней)",
				Count:       100,
				Severity:    "low",
			},
		},
	}, nil
}