package domain

type PasswordResetRepository interface {
	// Create stores a new password reset token
	Create(token *PasswordResetToken) error

	// GetByToken retrieves a token by its value
	GetByToken(token string) (*PasswordResetToken, error)

	// Invalidate marks a token as used
	Invalidate(token string) error

	// DeleteExpired removes expired tokens
	DeleteExpired() error
}
