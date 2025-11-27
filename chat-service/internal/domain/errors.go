package domain

import "errors"

var (
	ErrChatNotFound          = errors.New("chat not found")
	ErrChatExists            = errors.New("chat already exists")
	ErrAdministratorNotFound = errors.New("administrator not found")
	ErrAdministratorExists   = errors.New("administrator already exists")
	ErrInvalidPhone          = errors.New("invalid phone number")
	ErrMaxIDNotFound         = errors.New("max id not found")
	ErrCannotDeleteLastAdmin = errors.New("cannot delete last administrator")
	ErrUniversityNotFound    = errors.New("university not found")
)
