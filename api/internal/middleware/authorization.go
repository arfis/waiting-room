package middleware

import (
	"net/http"
)

type AuthorizationMiddleware struct{}

func NewAuthorizationMiddleware() *AuthorizationMiddleware {
	return &AuthorizationMiddleware{}
}

func (m *AuthorizationMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For now, just pass through - no actual authorization
			// In a real implementation, you would validate JWT tokens here
			next.ServeHTTP(w, r)
		})
	}
}
