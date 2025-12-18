package monitoring

import (
	"context"
	"testing"
	"time"

	"maxbot-service/internal/domain"
)

// TestMonitoringIntegration tests the complete monitoring flow
func TestMonitoringIntegration(t *testing.T) {
	// Create mock monitoring service
	monitoring := NewMockMonitoringService()
	
	ctx := context.Background()
	
	// Test 1: Record webhook events
	events := []domain.WebhookEventMetric{
		{
			EventType:      "message_new",
			UserID:         "user1",
			ProcessedAt:    time.Now().Add(-2 * time.Hour),
			Success:        true,
			ProcessingTime: 100,
			ProfileFound:   true,
			ProfileStored:  true,
		},
		{
			EventType:      "callback_query",
			UserID:         "user2",
			ProcessedAt:    time.Now().Add(-1 * time.Hour),
			Success:        true,
			ProcessingTime: 150,
			ProfileFound:   true,
			ProfileStored:  true,
		},
		{
			EventType:      "message_new",
			UserID:         "user3",
			ProcessedAt:    time.Now().Add(-30 * time.Minute),
			Success:        false,
			ErrorMessage:   "validation error",
			ProcessingTime: 50,
			ProfileFound:   false,
			ProfileStored:  false,
		},
	}
	
	// Record all events
	for _, event := range events {
		err := monitoring.RecordWebhookEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to record webhook event: %v", err)
		}
	}
	
	// Test 2: Get webhook statistics for last day
	period := domain.TimePeriod{
		From: time.Now().Add(-24 * time.Hour),
		To:   time.Now(),
	}
	
	stats, err := monitoring.GetWebhookStats(ctx, period)
	if err != nil {
		t.Fatalf("Failed to get webhook stats: %v", err)
	}
	
	// Verify statistics
	if stats.TotalEvents != 3 {
		t.Errorf("Expected 3 total events, got %d", stats.TotalEvents)
	}
	
	if stats.SuccessfulEvents != 2 {
		t.Errorf("Expected 2 successful events, got %d", stats.SuccessfulEvents)
	}
	
	if stats.FailedEvents != 1 {
		t.Errorf("Expected 1 failed event, got %d", stats.FailedEvents)
	}
	
	if stats.ProfilesExtracted != 2 {
		t.Errorf("Expected 2 profiles extracted, got %d", stats.ProfilesExtracted)
	}
	
	if stats.ProfilesStored != 2 {
		t.Errorf("Expected 2 profiles stored, got %d", stats.ProfilesStored)
	}
	
	// Verify event types
	if stats.EventsByType["message_new"] != 2 {
		t.Errorf("Expected 2 message_new events, got %d", stats.EventsByType["message_new"])
	}
	
	if stats.EventsByType["callback_query"] != 1 {
		t.Errorf("Expected 1 callback_query event, got %d", stats.EventsByType["callback_query"])
	}
	
	// Test 3: Get profile coverage
	coverage, err := monitoring.GetProfileCoverage(ctx)
	if err != nil {
		t.Fatalf("Failed to get profile coverage: %v", err)
	}
	
	// Verify coverage metrics
	if coverage.TotalUsers == 0 {
		t.Error("Expected non-zero total users")
	}
	
	if coverage.CoveragePercentage < 0 || coverage.CoveragePercentage > 100 {
		t.Errorf("Coverage percentage should be 0-100, got %f", coverage.CoveragePercentage)
	}
	
	if coverage.FullNamePercentage < 0 || coverage.FullNamePercentage > 100 {
		t.Errorf("Full name percentage should be 0-100, got %f", coverage.FullNamePercentage)
	}
	
	// Test 4: Get profile quality report
	report, err := monitoring.GetProfileQualityReport(ctx)
	if err != nil {
		t.Fatalf("Failed to get profile quality report: %v", err)
	}
	
	// Verify quality report
	if report.TotalProfiles == 0 {
		t.Error("Expected non-zero total profiles")
	}
	
	if report.QualityMetrics.QualityScore < 0 || report.QualityMetrics.QualityScore > 100 {
		t.Errorf("Quality score should be 0-100, got %f", report.QualityMetrics.QualityScore)
	}
	
	if len(report.SourceBreakdown) == 0 {
		t.Error("Expected non-empty source breakdown")
	}
	
	if len(report.RecommendedActions) == 0 {
		t.Error("Expected some recommended actions")
	}
	
	// Verify source breakdown contains expected sources
	expectedSources := []domain.ProfileSource{
		domain.SourceWebhook,
		domain.SourceUserInput,
		domain.SourceDefault,
	}
	
	for _, source := range expectedSources {
		if _, exists := report.SourceBreakdown[source]; !exists {
			t.Errorf("Expected source %s in breakdown", source)
		}
	}
}

// TestMonitoringErrorHandling tests error handling in monitoring service
func TestMonitoringErrorHandling(t *testing.T) {
	monitoring := NewMockMonitoringService()
	ctx := context.Background()
	
	// Test with invalid event (should not fail)
	invalidEvent := domain.WebhookEventMetric{
		EventType:   "", // Empty event type
		UserID:      "",
		ProcessedAt: time.Time{}, // Zero time
		Success:     false,
	}
	
	err := monitoring.RecordWebhookEvent(ctx, invalidEvent)
	if err != nil {
		t.Fatalf("Recording invalid event should not fail: %v", err)
	}
	
	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()
	
	// These should still work with mock service
	_, err = monitoring.GetWebhookStats(cancelledCtx, domain.LastDay())
	if err != nil {
		t.Fatalf("GetWebhookStats with cancelled context should not fail with mock: %v", err)
	}
	
	_, err = monitoring.GetProfileCoverage(cancelledCtx)
	if err != nil {
		t.Fatalf("GetProfileCoverage with cancelled context should not fail with mock: %v", err)
	}
	
	_, err = monitoring.GetProfileQualityReport(cancelledCtx)
	if err != nil {
		t.Fatalf("GetProfileQualityReport with cancelled context should not fail with mock: %v", err)
	}
}

// TestMonitoringTimePeriods tests different time periods for statistics
func TestMonitoringTimePeriods(t *testing.T) {
	monitoring := NewMockMonitoringService()
	ctx := context.Background()
	
	// Record events at different times
	now := time.Now()
	events := []domain.WebhookEventMetric{
		{
			EventType:   "message_new",
			UserID:      "user1",
			ProcessedAt: now.Add(-2 * time.Hour), // 2 hours ago
			Success:     true,
		},
		{
			EventType:   "message_new",
			UserID:      "user2",
			ProcessedAt: now.Add(-25 * time.Hour), // 25 hours ago (outside day range)
			Success:     true,
		},
		{
			EventType:   "message_new",
			UserID:      "user3",
			ProcessedAt: now.Add(-30 * time.Minute), // 30 minutes ago
			Success:     true,
		},
	}
	
	for _, event := range events {
		err := monitoring.RecordWebhookEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to record event: %v", err)
		}
	}
	
	// Test different time periods
	testCases := []struct {
		name           string
		period         domain.TimePeriod
		expectedEvents int64
	}{
		{
			name:           "Last hour",
			period:         domain.LastHour(),
			expectedEvents: 1, // Only the 30-minute-ago event
		},
		{
			name:           "Last day",
			period:         domain.LastDay(),
			expectedEvents: 2, // 2-hour and 30-minute events
		},
		{
			name:           "Last week",
			period:         domain.LastWeek(),
			expectedEvents: 3, // All events
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stats, err := monitoring.GetWebhookStats(ctx, tc.period)
			if err != nil {
				t.Fatalf("Failed to get stats for %s: %v", tc.name, err)
			}
			
			if stats.TotalEvents != tc.expectedEvents {
				t.Errorf("Expected %d events for %s, got %d", 
					tc.expectedEvents, tc.name, stats.TotalEvents)
			}
		})
	}
}