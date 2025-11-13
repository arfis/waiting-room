package service

import (
	"context"
	"log/slog"

	"github.com/arfis/waiting-room/internal/middleware"
)

// GetTenantID returns tenant ID as string from context (used by middleware.TENANT)
func GetTenantID(ctx context.Context) string {
	tenant := ctx.Value(middleware.TENANT)
	if tenant == nil {
		return ""
	}
	if tenantID, ok := tenant.(string); ok {
		return tenantID
	}
	slog.Error("tenant in context is not string", "tenant", tenant)
	return ""
}
