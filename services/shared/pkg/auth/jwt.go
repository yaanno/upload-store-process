package auth

import (
	"fmt"
	"time"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// TokenGenerator creates a new JWT token
type TokenGenerator struct {
	secretKey []byte
	issuer    string
}

// NewTokenGenerator creates a new token generator
func NewTokenGenerator(secretKey string, issuer string) *TokenGenerator {
	return &TokenGenerator{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

// GenerateToken creates a new JWT token with specified claims
func (tg *TokenGenerator) GenerateToken(userID, email string, roles, permissions []string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tg.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
		UserID:      userID,
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tg.secretKey)
}

// ValidateToken validates and parses a JWT token
func (tg *TokenGenerator) ValidateToken(tokenString string) (*Claims, error) {
	// Basic input validation
	if tokenString == "" {
		return nil, ErrMissingToken
	}

	// Parse token with custom claims
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tg.secretKey, nil
	}, jwt.WithLeeway(5*time.Second))

	if err != nil {
		// Detailed error handling
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, fmt.Errorf("%w: token is malformed", ErrInvalidToken)
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, fmt.Errorf("%w: invalid token signature", ErrInvalidToken)
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, fmt.Errorf("%w: token has expired", ErrTokenExpired)
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, fmt.Errorf("%w: token is not valid yet", ErrInvalidToken)
		default:
			return nil, fmt.Errorf("token validation failed: %w", err)
		}
	}

	// Type assert and validate claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Additional custom validations
	if claims.IsExpired() {
		return nil, ErrTokenExpired
	}

	return claims, nil
}
