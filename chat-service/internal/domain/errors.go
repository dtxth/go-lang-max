package domain

import (
	"chat-service/internal/infrastructure/errors"
)

var (
	ErrChatNotFound           = errors.NotFoundError("chat")
	ErrChatExists             = errors.AlreadyExistsError("chat", "")
	ErrAdministratorNotFound  = errors.NotFoundError("administrator")
	ErrAdministratorExists    = errors.AlreadyExistsError("administrator", "")
	ErrInvalidPhone           = errors.InvalidPhoneError("")
	ErrMaxIDNotFound          = errors.NotFoundError("MAX_id")
	ErrCannotDeleteLastAdmin  = errors.CannotDeleteError("administrator", "last administrator cannot be removed")
	ErrUniversityNotFound     = errors.NotFoundError("university")
	ErrInvalidToken           = errors.InvalidTokenError()
	ErrUnauthorized           = errors.UnauthorizedError("unauthorized")
	ErrForbidden              = errors.ForbiddenError("insufficient permissions")
	ErrInvalidRole            = errors.ValidationError("invalid role")
	ErrParticipantsNotCached  = errors.NotFoundError("participants count not cached")
)
