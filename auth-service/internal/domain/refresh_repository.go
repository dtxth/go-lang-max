package domain

import "time"

type RefreshTokenRepository interface {
    Save(jti string, userID int64, expiresAt time.Time) error
    IsValid(jti string) (bool, error)
    Revoke(jti string) error
    RevokeAllForUser(userID int64) error // опционально: logout all devices
    // (можно добавить FindByJTI, если нужно вернуть запись)
}