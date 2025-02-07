package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT token claims
type Claims struct {
	jwt.RegisteredClaims
	UserID     string   `json:"user_id"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// HasRole checks if the user has a specific role
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if the user has a specific permission
func (c *Claims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// IsExpired checks if the token is expired
func (c *Claims) IsExpired() bool {
	return time.Now().After(c.ExpiresAt.Time)
}
