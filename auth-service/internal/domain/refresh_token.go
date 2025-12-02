package domain

import "time"

type RefreshToken struct {
	ID        int64
	JTI       string
	UserID    int64
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}
