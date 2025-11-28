package domain

import "errors"

var (
	// ErrMigrationJobNotFound is returned when a migration job is not found
	ErrMigrationJobNotFound = errors.New("migration job not found")

	// ErrInvalidSourceType is returned when an invalid source type is provided
	ErrInvalidSourceType = errors.New("invalid source type")

	// ErrInvalidStatus is returned when an invalid status is provided
	ErrInvalidStatus = errors.New("invalid status")

	// ErrMigrationAlreadyRunning is returned when trying to start a migration that's already running
	ErrMigrationAlreadyRunning = errors.New("migration already running")

	// ErrInvalidFileFormat is returned when the uploaded file has an invalid format
	ErrInvalidFileFormat = errors.New("invalid file format")

	// ErrMissingRequiredColumns is returned when required columns are missing from the file
	ErrMissingRequiredColumns = errors.New("missing required columns")
)
