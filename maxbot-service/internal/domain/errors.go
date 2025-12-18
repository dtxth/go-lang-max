package domain

import (
	"maxbot-service/internal/infrastructure/errors"
)

var (
	ErrInvalidPhone     = errors.InvalidPhoneError("")
	ErrMaxIDNotFound    = errors.NotFoundError("MAX_id")
	ErrMaxAPIError      = errors.ExternalServiceError("MAX API", nil)
	ErrCacheUnavailable = errors.ExternalServiceError("Profile Cache", nil)
)
