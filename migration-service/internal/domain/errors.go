package domain

import (
	"migration-service/internal/infrastructure/errors"
)

var (
	// ErrMigrationJobNotFound is returned when a migration job is not found
	ErrMigrationJobNotFound = errors.NotFoundError("migration job")

	// ErrInvalidSourceType is returned when an invalid source type is provided
	ErrInvalidSourceType = errors.ValidationError("invalid source type")

	// ErrInvalidStatus is returned when an invalid status is provided
	ErrInvalidStatus = errors.ValidationError("invalid status")

	// ErrMigrationAlreadyRunning is returned when trying to start a migration that's already running
	ErrMigrationAlreadyRunning = errors.AlreadyExistsError("migration job", "running")

	// ErrInvalidFileFormat is returned when the uploaded file has an invalid format
	ErrInvalidFileFormat = errors.ValidationError("invalid file format")

	// ErrMissingRequiredColumns is returned when required columns are missing from the file
	ErrMissingRequiredColumns = errors.ValidationError("missing required columns")

	// ErrUniversityNotFound is returned when a university is not found
	ErrUniversityNotFound = errors.NotFoundError("university")

	// ErrChatServiceError is returned when chat service call fails
	ErrChatServiceError = errors.ExternalServiceError("Chat Service", nil)

	// ErrStructureServiceError is returned when structure service call fails
	ErrStructureServiceError = errors.ExternalServiceError("Structure Service", nil)
)
