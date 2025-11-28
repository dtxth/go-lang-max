package domain

import (
	"auth-service/internal/infrastructure/errors"
)

var (
	ErrUserExists   = errors.AlreadyExistsError("user", "email")
	ErrInvalidCreds = errors.UnauthorizedError("invalid email or password")
	ErrTokenExpired = errors.ExpiredTokenError()
	ErrUserNotFound = errors.NotFoundError("user")
	ErrInvalidToken = errors.InvalidTokenError()
	ErrInvalidRole  = errors.ValidationError("invalid role")
)