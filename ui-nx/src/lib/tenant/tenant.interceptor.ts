import { inject } from '@angular/core';
import { HttpInterceptorFn } from '@angular/common/http';
import { TenantService } from './tenant.service';

// Functional interceptor (recommended for standalone apps)
// Adds X-Tenant-ID header to all requests when a tenant is selected
export const tenantInterceptor: HttpInterceptorFn = (req, next) => {
  const tenantService = inject(TenantService);
  
  // ALWAYS log that interceptor is running - this is critical for debugging
  console.log(`[TenantInterceptor] ===== INTERCEPTOR CALLED =====`);
  console.log(`[TenantInterceptor] Request URL: ${req.url}`);
  
  // Try multiple ways to get the tenant ID
  const signalValue = tenantService.selectedTenantId();
  const syncValue = tenantService.getSelectedTenantIdSync();
  
  // Also try to read from localStorage directly as a fallback
  let localStorageValue = '';
  if (typeof window !== 'undefined' && window.localStorage) {
    localStorageValue = localStorage.getItem('selectedTenantId') || '';
  }
  
  const selectedTenantId = signalValue || syncValue || localStorageValue || '';
  
  // Don't add tenant header for tenant management endpoints (GET/POST /admin/tenants)
  // These endpoints are used to list/create tenants and should not require a tenant context
  const url = req.url;
  const isTenantManagementRequest = url.includes('/admin/tenants');
  
  // Debug: Log ALL tenant service state
  console.log(`[TenantInterceptor] Signal value: "${signalValue || 'empty'}" (type: ${typeof signalValue})`);
  console.log(`[TenantInterceptor] Sync value: "${syncValue || 'empty'}"`);
  console.log(`[TenantInterceptor] LocalStorage value: "${localStorageValue || 'empty'}"`);
  console.log(`[TenantInterceptor] Final selectedTenantId: "${selectedTenantId}"`);
  console.log(`[TenantInterceptor] Is tenant management request: ${isTenantManagementRequest}`);
  console.log(`[TenantInterceptor] TenantService instance exists: ${!!tenantService}`);
  
  // For tenant management requests, don't add tenant header
  // For all other requests, add tenant header if tenant is selected
  if (selectedTenantId && selectedTenantId.trim() !== '' && !isTenantManagementRequest) {
    // Clone the request and add the X-Tenant-ID header
    const headerValue = selectedTenantId.trim();
    const tenantReq = req.clone({
      setHeaders: {
        'X-Tenant-ID': headerValue
      }
    });
    console.log(`[TenantInterceptor] ✓✓✓ ADDING HEADER ✓✓✓`);
    console.log(`[TenantInterceptor] Header value: "${headerValue}"`);
    console.log(`[TenantInterceptor] For URL: ${url}`);
    
    // Verify the header was added
    const addedHeader = tenantReq.headers.get('X-Tenant-ID');
    console.log(`[TenantInterceptor] Verified header in cloned request: "${addedHeader}"`);
    
    // Also log all headers
    const allHeaders: string[] = [];
    tenantReq.headers.keys().forEach(key => {
      allHeaders.push(`${key}: ${tenantReq.headers.get(key)}`);
    });
    console.log(`[TenantInterceptor] All request headers:`, allHeaders);
    
    return next(tenantReq);
  } else {
    console.log(`[TenantInterceptor] ✗✗✗ NOT ADDING HEADER ✗✗✗`);
    console.log(`[TenantInterceptor] Reason: Tenant="${selectedTenantId || 'none'}", IsTenantManagement=${isTenantManagementRequest}, URL=${url}`);
    console.log(`[TenantInterceptor] Conditions check: hasValue=${!!selectedTenantId}, trimmed="${selectedTenantId ? selectedTenantId.trim() : 'N/A'}", isEmpty=${selectedTenantId.trim() === ''}, isTenantManagement=${isTenantManagementRequest}`);
  }
  
  console.log(`[TenantInterceptor] ===== END INTERCEPTOR =====`);
  return next(req);
};

