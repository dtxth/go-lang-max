package notification

import (
	"context"
	"fmt"

	"auth-service/internal/infrastructure/logger"
)

// MockNotificationService is a mock implementation of NotificationService for testing and development
type MockNotificationService struct {
	logger *logger.Logger
}

// NewMockNotificationService creates a new mock notification service
func NewMockNotificationService(log *logger.Logger) *MockNotificationService {
	return &MockNotificationService{
		logger: log,
	}
}

// SendPasswordNotification logs that a password notification would be sent (without the actual password)
func (s *MockNotificationService) SendPasswordNotification(ctx context.Context, phone, password string) error {
	sanitizedPhone := sanitizePhone(phone)
	
	s.logger.Info(ctx, "MOCK: Would send password notification", map[string]interface{}{
		"phone_suffix": sanitizedPhone,
		"action":       "send_password_notification",
	})
	
	return nil
}

// SendResetTokenNotification logs that a reset token notification would be sent (without the actual token)
func (s *MockNotificationService) SendResetTokenNotification(ctx context.Context, phone, token string) error {
	sanitizedPhone := sanitizePhone(phone)
	
	s.logger.Info(ctx, "MOCK: Would send reset token notification", map[string]interface{}{
		"phone_suffix": sanitizedPhone,
		"action":       "send_reset_token_notification",
	})
	
	return nil
}

// sanitizePhone returns only the last 4 digits of a phone number for logging
func sanitizePhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return fmt.Sprintf("****%s", phone[len(phone)-4:])
}
