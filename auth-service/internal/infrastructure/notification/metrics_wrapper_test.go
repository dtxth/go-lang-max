package notification

import (
	"context"
	"errors"
	"testing"

	"auth-service/internal/infrastructure/metrics"
	"github.com/stretchr/testify/assert"
)

// MockNotificationService is a mock implementation for testing
type MockNotificationServiceForMetrics struct {
	shouldFail bool
}

func (m *MockNotificationServiceForMetrics) SendPasswordNotification(ctx context.Context, phone, password string) error {
	if m.shouldFail {
		return errors.New("notification failed")
	}
	return nil
}

func (m *MockNotificationServiceForMetrics) SendResetTokenNotification(ctx context.Context, phone, token string) error {
	if m.shouldFail {
		return errors.New("notification failed")
	}
	return nil
}

// TestMetricsWrapperPasswordNotificationSuccess tests successful password notification
func TestMetricsWrapperPasswordNotificationSuccess(t *testing.T) {
	mockService := &MockNotificationServiceForMetrics{shouldFail: false}
	m := metrics.NewMetrics()
	wrapper := NewMetricsWrapper(mockService, m)

	err := wrapper.SendPasswordNotification(context.Background(), "+79991234567", "TestPass123!")
	assert.NoError(t, err, "Notification should succeed")

	snapshot := m.GetMetrics()
	assert.Equal(t, int64(1), snapshot.NotificationsSent, "Should have 1 successful notification")
	assert.Equal(t, int64(0), snapshot.NotificationsFailed, "Should have 0 failed notifications")
}

// TestMetricsWrapperPasswordNotificationFailure tests failed password notification
func TestMetricsWrapperPasswordNotificationFailure(t *testing.T) {
	mockService := &MockNotificationServiceForMetrics{shouldFail: true}
	m := metrics.NewMetrics()
	wrapper := NewMetricsWrapper(mockService, m)

	err := wrapper.SendPasswordNotification(context.Background(), "+79991234567", "TestPass123!")
	assert.Error(t, err, "Notification should fail")

	snapshot := m.GetMetrics()
	assert.Equal(t, int64(0), snapshot.NotificationsSent, "Should have 0 successful notifications")
	assert.Equal(t, int64(1), snapshot.NotificationsFailed, "Should have 1 failed notification")
}

// TestMetricsWrapperResetTokenNotificationSuccess tests successful reset token notification
func TestMetricsWrapperResetTokenNotificationSuccess(t *testing.T) {
	mockService := &MockNotificationServiceForMetrics{shouldFail: false}
	m := metrics.NewMetrics()
	wrapper := NewMetricsWrapper(mockService, m)

	err := wrapper.SendResetTokenNotification(context.Background(), "+79991234567", "reset-token-123")
	assert.NoError(t, err, "Notification should succeed")

	snapshot := m.GetMetrics()
	assert.Equal(t, int64(1), snapshot.NotificationsSent, "Should have 1 successful notification")
	assert.Equal(t, int64(0), snapshot.NotificationsFailed, "Should have 0 failed notifications")
}

// TestMetricsWrapperResetTokenNotificationFailure tests failed reset token notification
func TestMetricsWrapperResetTokenNotificationFailure(t *testing.T) {
	mockService := &MockNotificationServiceForMetrics{shouldFail: true}
	m := metrics.NewMetrics()
	wrapper := NewMetricsWrapper(mockService, m)

	err := wrapper.SendResetTokenNotification(context.Background(), "+79991234567", "reset-token-123")
	assert.Error(t, err, "Notification should fail")

	snapshot := m.GetMetrics()
	assert.Equal(t, int64(0), snapshot.NotificationsSent, "Should have 0 successful notifications")
	assert.Equal(t, int64(1), snapshot.NotificationsFailed, "Should have 1 failed notification")
}

// TestMetricsWrapperMultipleNotifications tests multiple notifications
func TestMetricsWrapperMultipleNotifications(t *testing.T) {
	m := metrics.NewMetrics()

	// Create wrapper with successful service
	successService := &MockNotificationServiceForMetrics{shouldFail: false}
	successWrapper := NewMetricsWrapper(successService, m)

	// Send 3 successful notifications
	successWrapper.SendPasswordNotification(context.Background(), "+79991234567", "Pass1")
	successWrapper.SendPasswordNotification(context.Background(), "+79991234568", "Pass2")
	successWrapper.SendResetTokenNotification(context.Background(), "+79991234569", "token1")

	// Create wrapper with failing service
	failService := &MockNotificationServiceForMetrics{shouldFail: true}
	failWrapper := NewMetricsWrapper(failService, m)

	// Send 2 failed notifications
	failWrapper.SendPasswordNotification(context.Background(), "+79991234570", "Pass3")
	failWrapper.SendResetTokenNotification(context.Background(), "+79991234571", "token2")

	// Check metrics
	snapshot := m.GetMetrics()
	assert.Equal(t, int64(3), snapshot.NotificationsSent, "Should have 3 successful notifications")
	assert.Equal(t, int64(2), snapshot.NotificationsFailed, "Should have 2 failed notifications")

	// Check rates
	successRate := m.GetNotificationSuccessRate()
	assert.Equal(t, 0.6, successRate, "Success rate should be 60%")

	failureRate := m.GetNotificationFailureRate()
	assert.Equal(t, 0.4, failureRate, "Failure rate should be 40%")
}
