package auth

import (
	"errors"
)

var (
	// ErrInvalidToken represents an error for an invalid JWT token
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired represents an error for an expired token
	ErrTokenExpired = errors.New("token has expired")

	// ErrMissingToken represents an error for a missing token
	ErrMissingToken = errors.New("missing authentication token")

	// ErrInsufficientPermissions represents an error for insufficient permissions
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)
