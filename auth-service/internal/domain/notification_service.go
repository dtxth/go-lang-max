package domain

import "context"

// NotificationService defines the interface for sending notifications to users
type NotificationService interface {
	// SendPasswordNotification sends a temporary password to a user
	SendPasswordNotification(ctx context.Context, phone, password string) error
	
	// SendResetTokenNotification sends a password reset token to a user
	SendResetTokenNotification(ctx context.Context, phone, token string) error
}
