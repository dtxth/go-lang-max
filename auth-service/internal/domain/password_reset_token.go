package domain

import "time"

type PasswordResetToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// IsValid checks if token is valid (not expired, not used)
func (t *PasswordResetToken) IsValid() bool {
	return t.UsedAt == nil && time.Now().Before(t.ExpiresAt)
}

// IsExpired checks if token has expired
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if token has been used
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}
