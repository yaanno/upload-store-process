package auth

// TokenValidator defines the interface for token validation
type TokenValidator interface {
	ValidateToken(tokenString string) (*Claims, error)
}
