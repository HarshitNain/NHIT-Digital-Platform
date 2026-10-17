package utils

import "errors"

var (
	// ErrWeakPassword is returned when password doesn't meet strength requirements
	ErrWeakPassword = errors.New("password must be at least 8 characters with uppercase, lowercase, number, and special character")
	
	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")
	
	// ErrTokenExpired is returned when token has expired
	ErrTokenExpired = errors.New("token has expired")
	
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
)
