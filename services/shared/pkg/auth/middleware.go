package auth

import (
	"context"
	"reflect"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ExtractTokenFromContext retrieves JWT token from gRPC metadata
func ExtractTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrMissingToken
	}

	// Try multiple possible token headers
	tokenHeaders := []string{
		"authorization",
		"jwt_token",
		"x-jwt-token",
	}

	for _, header := range tokenHeaders {
		tokens := md.Get(header)
		if len(tokens) > 0 {
			// Remove "Bearer " prefix if present
			token := tokens[0]
			if strings.HasPrefix(token, "Bearer ") {
				token = strings.TrimPrefix(token, "Bearer ")
			}
			return token, nil
		}
	}

	return "", ErrMissingToken
}

// ExtractTokenFromRequest attempts to extract JWT token from request
func ExtractTokenFromRequest(req interface{}, tokenGenerator *TokenGenerator) (string, error) {
	// Use reflection to find jwt_token field
	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	tokenField := v.FieldByName("JwtToken")
	if !tokenField.IsValid() {
		return "", ErrMissingToken
	}

	token := tokenField.String()
	if token == "" {
		return "", ErrMissingToken
	}

	return token, nil
}

// UnaryServerInterceptor creates a JWT authentication interceptor
func UnaryServerInterceptor(tokenValidator TokenValidator, skipMethods ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for specified methods
		for _, method := range skipMethods {
			if strings.Contains(info.FullMethod, method) {
				return handler(ctx, req)
			}
		}

		// Extract token from request
		token, err := ExtractTokenFromRequest(req, nil)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "token extraction failed: %v", err)
		}

		// Validate token
		claims, err := tokenValidator.ValidateToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "token validation failed: %v", err)
		}

		// Attach claims to context
		ctx = context.WithValue(ctx, "claims", claims)

		return handler(ctx, req)
	}
}

// GetClaimsFromContext retrieves claims from context
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value("claims").(*Claims)
	return claims, ok
}
