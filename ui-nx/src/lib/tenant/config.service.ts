import { Injectable, inject, InjectionToken } from '@angular/core';

// Injection token for API URL - must be provided by each app
export const TENANT_API_URL = new InjectionToken<string>('TENANT_API_URL');

@Injectable({
  providedIn: 'root'
})
export class TenantConfigService {
  private apiUrl = inject(TENANT_API_URL, { optional: true }) || 'http://localhost:8080/api';

  get adminTenantsUrl(): string {
    return `${this.apiUrl}/admin/tenants`;
  }
}

