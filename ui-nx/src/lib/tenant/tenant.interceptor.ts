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
 * - /admin/tenants CRUD endpoints (tenant management operations on master DB)
 * - Requests without a selected tenant
 *
 * The header IS added for:
 * - /admin/tenants/{id}/sections endpoints (section management on tenant DB)
 * - All other configuration, queue, and card reader endpoints
 */
export const tenantInterceptor: HttpInterceptorFn = (req, next) => {
  const tenantService = inject(TenantService);

  console.log(`[TenantInterceptor] Processing request: ${req.url}`);

  // Get composite tenant ID (format: "1001" or "1001:2001")
  const compositeId = tenantService.getSelectedTenantIdSync();

  // Don't add tenant header for tenant CRUD endpoints, but DO add it for section endpoints
  // Examples:
  // - /admin/tenants (list) -> skip header (master DB)
  // - /admin/tenants/123 (get/update/delete) -> skip header (master DB)
  // - /admin/tenants/123/sections -> ADD header (tenant DB)
  const url = req.url;
  const isTenantCrudRequest = url.includes('/admin/tenants') && !url.includes('/sections');

  console.log(`[TenantInterceptor] Composite ID: "${compositeId || 'none'}"`);
  console.log(`[TenantInterceptor] Is tenant CRUD (skip header): ${isTenantCrudRequest}`);

  // Add header if we have a tenant ID and this is not a tenant CRUD request
  if (compositeId && compositeId.trim() !== '' && !isTenantCrudRequest) {
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
