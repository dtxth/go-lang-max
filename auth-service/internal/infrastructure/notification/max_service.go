package notification

import (
	"context"
	"time"

	"auth-service/internal/infrastructure/logger"
	// grpcretry "auth-service/internal/infrastructure/grpc"
	// maxbotproto "maxbot-service/api/proto/maxbotproto"

	// "google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"
)

// MaxNotificationService is a real implementation of NotificationService using MaxBot Service
type MaxNotificationService struct {
	// conn    *grpc.ClientConn
	// client  maxbotproto.MaxBotServiceClient
	logger  *logger.Logger
	timeout time.Duration
}

// NewMaxNotificationService creates a new MAX notification service
func NewMaxNotificationService(maxBotAddr string, log *logger.Logger) (*MaxNotificationService, error) {
	// Temporary mock implementation
	return &MaxNotificationService{
		logger:  log,
		timeout: 10 * time.Second,
	}, nil
}

// Close closes the gRPC connection
func (s *MaxNotificationService) Close() error {
	// No connection to close in mock implementation
	return nil
}

// SendPasswordNotification sends a temporary password to a user via MAX Messenger
func (s *MaxNotificationService) SendPasswordNotification(ctx context.Context, phone, password string) error {
	sanitizedPhone := sanitizePhone(phone)
	
	// Mock implementation - just log the notification
	s.logger.Info(ctx, "Mock: Password notification sent", map[string]interface{}{
		"phone_suffix": sanitizedPhone,
		// "password":     password,  // Commented out to avoid logging passwords
	})
	
	return nil
}

// SendResetTokenNotification sends a password reset token to a user via MAX Messenger
func (s *MaxNotificationService) SendResetTokenNotification(ctx context.Context, phone, token string) error {
	sanitizedPhone := sanitizePhone(phone)
	
	// Mock implementation - just log the notification
	s.logger.Info(ctx, "Mock: Reset token notification sent", map[string]interface{}{
		"phone_suffix": sanitizedPhone,
		// "token":        token,  // Commented out to avoid logging tokens
	})
	
	return nil
}
