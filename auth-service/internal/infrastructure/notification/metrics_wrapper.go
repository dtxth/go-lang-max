package notification

import (
	"context"
	
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/metrics"
)

// MetricsWrapper wraps a NotificationService and records metrics
type MetricsWrapper struct {
	service domain.NotificationService
	metrics *metrics.Metrics
}

// NewMetricsWrapper creates a new metrics wrapper for a notification service
func NewMetricsWrapper(service domain.NotificationService, m *metrics.Metrics) *MetricsWrapper {
	return &MetricsWrapper{
		service: service,
		metrics: m,
	}
}

// SendPasswordNotification sends a password notification and records metrics
func (w *MetricsWrapper) SendPasswordNotification(ctx context.Context, phone, password string) error {
	err := w.service.SendPasswordNotification(ctx, phone, password)
	
	if err != nil {
		w.metrics.IncrementNotificationsFailed()
	} else {
		w.metrics.IncrementNotificationsSent()
	}
	
	return err
}

// SendResetTokenNotification sends a reset token notification and records metrics
func (w *MetricsWrapper) SendResetTokenNotification(ctx context.Context, phone, token string) error {
	err := w.service.SendResetTokenNotification(ctx, phone, token)
	
	if err != nil {
		w.metrics.IncrementNotificationsFailed()
	} else {
		w.metrics.IncrementNotificationsSent()
	}
	
	return err
}
