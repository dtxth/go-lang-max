package notification

import (
	"context"
	"fmt"
	"time"

	"auth-service/internal/infrastructure/logger"
	grpcretry "auth-service/internal/infrastructure/grpc"
	maxbotproto "maxbot-service/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MaxNotificationService is a real implementation of NotificationService using MaxBot Service
type MaxNotificationService struct {
	conn    *grpc.ClientConn
	client  maxbotproto.MaxBotServiceClient
	logger  *logger.Logger
	timeout time.Duration
}

// NewMaxNotificationService creates a new MAX notification service
func NewMaxNotificationService(maxBotAddr string, log *logger.Logger) (*MaxNotificationService, error) {
	conn, err := grpc.NewClient(maxBotAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MaxBot service: %w", err)
	}

	return &MaxNotificationService{
		conn:    conn,
		client:  maxbotproto.NewMaxBotServiceClient(conn),
		logger:  log,
		timeout: 10 * time.Second,
	}, nil
}

// Close closes the gRPC connection
func (s *MaxNotificationService) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// SendPasswordNotification sends a temporary password to a user via MAX Messenger
func (s *MaxNotificationService) SendPasswordNotification(ctx context.Context, phone, password string) error {
	sanitizedPhone := sanitizePhone(phone)
	
	// Format message in Russian with password and instructions
	message := fmt.Sprintf(
		"Ваш временный пароль для входа в систему: %s\n\n"+
			"Рекомендуем сменить пароль после первого входа.",
		password,
	)
	
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	
	// Send notification with retry logic
	var resp *maxbotproto.SendNotificationResponse
	err := grpcretry.WithRetry(timeoutCtx, "MaxBot.SendNotification", func() error {
		var callErr error
		resp, callErr = s.client.SendNotification(timeoutCtx, &maxbotproto.SendNotificationRequest{
			Phone: phone,
			Text:  message,
		})
		return callErr
	})
	
	if err != nil {
		s.logger.Error(ctx, "Failed to send password notification", map[string]interface{}{
			"phone_suffix": sanitizedPhone,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to send notification: %w", err)
	}
	
	// Check response for errors
	if resp.Error != "" {
		s.logger.Error(ctx, "MaxBot returned error for password notification", map[string]interface{}{
			"phone_suffix": sanitizedPhone,
			"error_code":   resp.ErrorCode.String(),
			"error":        resp.Error,
		})
		return fmt.Errorf("MaxBot error: %s", resp.Error)
	}
	
	if !resp.Success {
		s.logger.Error(ctx, "Password notification delivery failed", map[string]interface{}{
			"phone_suffix": sanitizedPhone,
		})
		return fmt.Errorf("notification delivery failed")
	}
	
	s.logger.Info(ctx, "Password notification sent successfully", map[string]interface{}{
		"phone_suffix": sanitizedPhone,
	})
	
	return nil
}

// SendResetTokenNotification sends a password reset token to a user via MAX Messenger
func (s *MaxNotificationService) SendResetTokenNotification(ctx context.Context, phone, token string) error {
	sanitizedPhone := sanitizePhone(phone)
	
	// Format message in Russian with reset token and instructions
	message := fmt.Sprintf(
		"Ваш код для сброса пароля: %s\n\n"+
			"Код действителен в течение 15 минут.\n"+
			"Если вы не запрашивали сброс пароля, проигнорируйте это сообщение.",
		token,
	)
	
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	
	// Send notification with retry logic
	var resp *maxbotproto.SendNotificationResponse
	err := grpcretry.WithRetry(timeoutCtx, "MaxBot.SendNotification", func() error {
		var callErr error
		resp, callErr = s.client.SendNotification(timeoutCtx, &maxbotproto.SendNotificationRequest{
			Phone: phone,
			Text:  message,
		})
		return callErr
	})
	
	if err != nil {
		s.logger.Error(ctx, "Failed to send reset token notification", map[string]interface{}{
			"phone_suffix": sanitizedPhone,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to send notification: %w", err)
	}
	
	// Check response for errors
	if resp.Error != "" {
		s.logger.Error(ctx, "MaxBot returned error for reset token notification", map[string]interface{}{
			"phone_suffix": sanitizedPhone,
			"error_code":   resp.ErrorCode.String(),
			"error":        resp.Error,
		})
		return fmt.Errorf("MaxBot error: %s", resp.Error)
	}
	
	if !resp.Success {
		s.logger.Error(ctx, "Reset token notification delivery failed", map[string]interface{}{
			"phone_suffix": sanitizedPhone,
		})
		return fmt.Errorf("notification delivery failed")
	}
	
	s.logger.Info(ctx, "Reset token notification sent successfully", map[string]interface{}{
		"phone_suffix": sanitizedPhone,
	})
	
	return nil
}
