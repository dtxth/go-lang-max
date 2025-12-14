package domain

import (
	"auth-service/internal/infrastructure/errors"
)

var (
	ErrUserExists          = errors.AlreadyExistsError("user", "email")
	ErrInvalidCreds        = errors.UnauthorizedError("invalid email or password")
	ErrTokenExpired        = errors.ExpiredTokenError()
	ErrUserNotFound        = errors.NotFoundError("user")
	ErrInvalidToken        = errors.InvalidTokenError()
	ErrInvalidRole         = errors.ValidationError("invalid role")
	ErrNotFound            = errors.NotFoundError("resource")
	ErrResetTokenNotFound  = errors.NotFoundError("password reset token")
	ErrResetTokenExpired   = errors.UnauthorizedError("password reset token has expired")
	ErrResetTokenUsed      = errors.UnauthorizedError("password reset token has already been used")
)