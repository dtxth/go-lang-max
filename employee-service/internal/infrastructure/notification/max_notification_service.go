package notification

import (
	"context"
	"fmt"
	"log"

	maxbotproto "maxbot-service/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MaxNotificationService implements domain.NotificationService using MaxBot gRPC client
type MaxNotificationService struct {
	conn   *grpc.ClientConn
	client maxbotproto.MaxBotServiceClient
	logger *log.Logger
}

// NewMaxNotificationService creates a new MaxNotificationService
func NewMaxNotificationService(maxBotAddress string, logger *log.Logger) (*MaxNotificationService, error) {
	conn, err := grpc.NewClient(maxBotAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MaxBot service: %w", err)
	}

	return &MaxNotificationService{
		conn:   conn,
		client: maxbotproto.NewMaxBotServiceClient(conn),
		logger: logger,
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
	message := fmt.Sprintf(
		"Ваш временный пароль для входа в систему: %s\n\n"+
			"Рекомендуем сменить пароль после первого входа.",
		password,
	)

	resp, err := s.client.SendNotification(ctx, &maxbotproto.SendNotificationRequest{
		Phone: phone,
		Text:  message,
	})

	if err != nil {
		s.logger.Printf("Failed to send password notification to phone ending in %s: %v",
			sanitizePhone(phone), err)
		return fmt.Errorf("failed to send notification: %w", err)
	}

	if resp.Error != "" {
		s.logger.Printf("MaxBot returned error for phone ending in %s: %s",
			sanitizePhone(phone), resp.Error)
		return fmt.Errorf("notification service error: %s", resp.Error)
	}

	if !resp.Success {
		s.logger.Printf("Failed to send notification to phone ending in %s",
			sanitizePhone(phone))
		return fmt.Errorf("notification delivery failed")
	}

	s.logger.Printf("Successfully sent password notification to phone ending in %s",
		sanitizePhone(phone))
	return nil
}

// sanitizePhone returns only the last 4 digits of a phone number for logging
func sanitizePhone(phone string) string {
	if len(phone) < 4 {
		return "****"
	}
	return "****" + phone[len(phone)-4:]
}
