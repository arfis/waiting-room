package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/arfis/waiting-room/internal/types"
)

type APP_CONTEXT string

const (
	TENANT_HEADER             = "X-Tenant-ID"
	TENANT_ID     APP_CONTEXT = "TENANT_ID"     // Numeric tenant ID
	SECTION_ID    APP_CONTEXT = "SECTION_ID"    // Numeric section ID (optional)
	TENANT        APP_CONTEXT = "TENANT"        // Legacy: full tenant string (for compatibility)
	LOGIN         APP_CONTEXT = "LOGIN"
	USER_INFO     string      = "USER_INFO"
)

type TenantMiddleware struct{}

func NewTenantMiddleware() *TenantMiddleware {
	return &TenantMiddleware{}
}

// Middleware extracts tenant ID from header or query parameter and adds it to context
// Expected format: "tenantId" or "tenantId:sectionId" (numeric values)
func (m *TenantMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Try to get composite tenant ID from header first
			compositeTenantID := r.Header.Get(TENANT_HEADER)
			log.Printf("[TenantMiddleware] Request to %s - Header %s: %s", r.URL.Path, TENANT_HEADER, compositeTenantID)

			// If not in header, try query parameter
			if compositeTenantID == "" {
				compositeTenantID = r.URL.Query().Get("tenantId")
				if compositeTenantID != "" {
					log.Printf("[TenantMiddleware] Request to %s - Query param tenantId: %s", r.URL.Path, compositeTenantID)
				}
			}

			// Normalize: trim whitespace
			compositeTenantID = strings.TrimSpace(compositeTenantID)

			// If tenant ID is provided, parse and add to context
			if compositeTenantID != "" {
				tenantID, sectionID, err := types.ParseTenantAndSectionID(compositeTenantID)
				if err != nil {
					log.Printf("[TenantMiddleware] Failed to parse tenant ID '%s': %v", compositeTenantID, err)
					http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
					return
				}

				// Add parsed IDs to context
				ctx = context.WithValue(ctx, TENANT_ID, tenantID)
				if sectionID != 0 {
					ctx = context.WithValue(ctx, SECTION_ID, sectionID)
				}
				// Keep legacy full string for compatibility
				ctx = context.WithValue(ctx, TENANT, compositeTenantID)

				log.Printf("[TenantMiddleware] Added tenant ID to context: tenantID=%d, sectionID=%d", tenantID, sectionID)
			} else {
				log.Printf("[TenantMiddleware] No tenant ID found in request to %s", r.URL.Path)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetTenantID extracts the numeric tenant ID from context
func GetTenantID(ctx context.Context) (int64, bool) {
	tenantID, ok := ctx.Value(TENANT_ID).(int64)
	return tenantID, ok
}

// GetSectionID extracts the numeric section ID from context
func GetSectionID(ctx context.Context) (int64, bool) {
	sectionID, ok := ctx.Value(SECTION_ID).(int64)
	return sectionID, ok
}

// GetTenantIDOrZero returns the tenant ID or 0 if not present
func GetTenantIDOrZero(ctx context.Context) int64 {
	tenantID, _ := GetTenantID(ctx)
	return tenantID
}

// GetSectionIDOrZero returns the section ID or 0 if not present
func GetSectionIDOrZero(ctx context.Context) int64 {
	sectionID, _ := GetSectionID(ctx)
	return sectionID
}
