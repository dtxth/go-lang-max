package domain

import (
	"employee-service/internal/infrastructure/errors"
)

var (
	ErrEmployeeNotFound   = errors.NotFoundError("employee")
	ErrEmployeeExists     = errors.AlreadyExistsError("employee", "")
	ErrUniversityNotFound = errors.NotFoundError("university")
	ErrUniversityExists   = errors.AlreadyExistsError("university", "INN")
	ErrInvalidPhone       = errors.InvalidPhoneError("")
	ErrMaxIDNotFound      = errors.NotFoundError("MAX_id")
	ErrInvalidRole        = errors.ValidationError("invalid role")
	ErrForbidden          = errors.ForbiddenError("insufficient permissions")
)

