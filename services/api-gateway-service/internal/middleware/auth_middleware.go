package middleware

import (
	"net/http"
	"strings"

	"github.com/yaanno/upload-store-process/services/shared/pkg/auth"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

type JWTAuthMiddleware struct {
	logger         logger.Logger
	maxFileSize    int64
	tokenGenerator auth.TokenGenerator
}

func NewJWTAuthMiddleware(
	logger logger.Logger,
	maxFileSize int64,
	tokenGenerator auth.TokenGenerator,
) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		logger:         logger,
		maxFileSize:    maxFileSize,
		tokenGenerator: tokenGenerator,
	}
}

func (m *JWTAuthMiddleware) JWTAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Error().Str("field", "Authorization").Msg("Authorization header cannot be empty")
			http.Error(w, "Authorization header cannot be empty", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		_, err := m.tokenGenerator.ValidateToken(tokenString)
		if err != nil {
			m.logger.Error().Err(err).Msg("Failed to validate token")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func JWTAuthMiddlewareWrapper(middleware *JWTAuthMiddleware, next http.HandlerFunc) http.HandlerFunc {
	return middleware.JWTAuthMiddleware(next)
}
