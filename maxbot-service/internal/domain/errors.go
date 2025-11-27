package domain

import "errors"

var (
	ErrInvalidPhone  = errors.New("invalid phone number")
	ErrMaxIDNotFound = errors.New("max id not found")
)
