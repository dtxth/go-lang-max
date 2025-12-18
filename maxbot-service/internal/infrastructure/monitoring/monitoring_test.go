package monitoring

import (
	"context"
	"testing"
	"time"

	"maxbot-service/internal/domain"
)

func TestMockMonitoringService(t *testing.T) {
	service := NewMockMonitoringService()
	ctx := context.Background()

	// Test recording webhook event
	event := domain.WebhookEventMetric{
		EventType:      "message_new",
		UserID:         "test_user_123",
		ProcessedAt:    time.Now(),
		Success:        true,
		ProcessingTime: 150,
		ProfileFound:   true,
		ProfileStored:  true,
	}

	err := service.RecordWebhookEvent(ctx, event)
	if err != nil {
		t.Fatalf("Failed to record webhook event: %v", err)
	}

	// Test getting webhook stats
	period := domain.TimePeriod{
		From: time.Now().Add(-time.Hour),
		To:   time.Now(),
	}

	stats, err := service.GetWebhookStats(ctx, period)
	if err != nil {
		t.Fatalf("Failed to get webhook stats: %v", err)
	}

	if stats.TotalEvents != 1 {
		t.Errorf("Expected 1 total event, got %d", stats.TotalEvents)
	}

	if stats.SuccessfulEvents != 1 {
		t.Errorf("Expected 1 successful event, got %d", stats.SuccessfulEvents)
	}

	if stats.ProfilesExtracted != 1 {
		t.Errorf("Expected 1 profile extracted, got %d", stats.ProfilesExtracted)
	}

	// Test getting profile coverage
	coverage, err := service.GetProfileCoverage(ctx)
	if err != nil {
		t.Fatalf("Failed to get profile coverage: %v", err)
	}

	if coverage.TotalUsers == 0 {
		t.Error("Expected non-zero total users")
	}

	// Test getting profile quality report
	report, err := service.GetProfileQualityReport(ctx)
	if err != nil {
		t.Fatalf("Failed to get profile quality report: %v", err)
	}

	if report.TotalProfiles == 0 {
		t.Error("Expected non-zero total profiles")
	}

	if len(report.SourceBreakdown) == 0 {
		t.Error("Expected non-empty source breakdown")
	}
}

func TestWebhookEventMetricValidation(t *testing.T) {
	service := NewMockMonitoringService()
	ctx := context.Background()

	// Test with failed event
	failedEvent := domain.WebhookEventMetric{
		EventType:      "callback_query",
		UserID:         "test_user_456",
		ProcessedAt:    time.Now(),
		Success:        false,
		ErrorMessage:   "validation error",
		ProcessingTime: 50,
		ProfileFound:   false,
		ProfileStored:  false,
	}

	err := service.RecordWebhookEvent(ctx, failedEvent)
	if err != nil {
		t.Fatalf("Failed to record failed webhook event: %v", err)
	}

	// Get stats and verify failed event is counted
	period := domain.TimePeriod{
		From: time.Now().Add(-time.Hour),
		To:   time.Now(),
	}

	stats, err := service.GetWebhookStats(ctx, period)
	if err != nil {
		t.Fatalf("Failed to get webhook stats: %v", err)
	}

	if stats.FailedEvents == 0 {
		t.Error("Expected at least one failed event")
	}

	if stats.EventsByType["callback_query"] == 0 {
		t.Error("Expected callback_query events to be counted")
	}
}