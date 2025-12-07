import { inject } from '@angular/core';
import { HttpInterceptorFn } from '@angular/common/http';
import { TenantService } from './tenant.service';

/**
 * Tenant HTTP Interceptor
 *
 * Adds the X-Tenant-ID header to all outgoing HTTP requests.
 * Format: "1001" (tenant only) or "1001:2001" (tenant with section)
 *
 * The header is skipped for:
 * - /admin/tenants endpoints (tenant management operations)
 * - Requests without a selected tenant
 */
export const tenantInterceptor: HttpInterceptorFn = (req, next) => {
  const tenantService = inject(TenantService);

  console.log(`[TenantInterceptor] Processing request: ${req.url}`);

  // Get composite tenant ID (format: "1001" or "1001:2001")
  const compositeId = tenantService.getSelectedTenantIdSync();

  // Don't add tenant header for tenant management endpoints
  const url = req.url;
  const isTenantManagementRequest = url.includes('/admin/tenants');

  console.log(`[TenantInterceptor] Composite ID: "${compositeId || 'none'}"`);
  console.log(`[TenantInterceptor] Is tenant management: ${isTenantManagementRequest}`);

  // Add header if we have a tenant ID and this is not a tenant management request
  if (compositeId && compositeId.trim() !== '' && !isTenantManagementRequest) {
    const tenantReq = req.clone({
      setHeaders: {
        'X-Tenant-ID': compositeId  // Format: "1001" or "1001:2001"
      }
    });

    console.log(`[TenantInterceptor] ✓ Added X-Tenant-ID header: ${compositeId}`);

    return next(tenantReq);
  }

  console.log(`[TenantInterceptor] ✗ Skipped tenant header`);
  return next(req);
};
