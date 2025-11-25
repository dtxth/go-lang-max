package domain

import "errors"

var (
	ErrEmployeeNotFound   = errors.New("employee not found")
	ErrEmployeeExists     = errors.New("employee already exists")
	ErrUniversityNotFound = errors.New("university not found")
	ErrUniversityExists   = errors.New("university already exists")
	ErrInvalidPhone       = errors.New("invalid phone number")
	ErrMaxIDNotFound      = errors.New("max_id not found for phone")
)

