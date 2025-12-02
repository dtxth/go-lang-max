package domain

import "errors"

var (
    ErrUserExists       = errors.New("user already exists")
    ErrInvalidCreds     = errors.New("invalid email or password")
    ErrTokenExpired     = errors.New("token expired")
)