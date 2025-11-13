package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
)

type APP_CONTEXT string

const (
	TENANT_HEADER             = "X-Tenant-ID"
	TENANT        APP_CONTEXT = "TENANT"
	USER_INFO     string      = "USER_INFO"
)

type TenantMiddleware struct{}

func NewTenantMiddleware() *TenantMiddleware {
	return &TenantMiddleware{}
}

// Middleware extracts tenant ID from header or query parameter and adds it to context
func (m *TenantMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Try to get tenant ID from header first
			tenantID := r.Header.Get(TENANT_HEADER)
			log.Printf("[TenantMiddleware] Request to %s - Header %s: %s", r.URL.Path, TENANT_HEADER, tenantID)

			// If not in header, try query parameter
			if tenantID == "" {
				tenantID = r.URL.Query().Get("tenantId")
				if tenantID != "" {
					log.Printf("[TenantMiddleware] Request to %s - Query param tenantId: %s", r.URL.Path, tenantID)
				}
			}

			// Normalize tenant ID: trim whitespace for consistency
			normalizedTenantID := strings.TrimSpace(tenantID)

			// If tenant ID is provided, add normalized version to context
			if normalizedTenantID != "" {
				ctx = context.WithValue(ctx, TENANT, normalizedTenantID)
				log.Printf("[TenantMiddleware] Added normalized tenant ID to context: '%s' (original: '%s')", normalizedTenantID, tenantID)
			} else {
				log.Printf("[TenantMiddleware] No tenant ID found in request to %s", r.URL.Path)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
