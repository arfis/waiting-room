import { ApplicationConfig, provideBrowserGlobalErrorListeners, provideZonelessChangeDetection } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { environment } from '../environments/environment';
import { tenantInterceptor, TENANT_API_URL, TenantService } from '@lib/tenant';
import { TENANT_SERVICE_TOKEN, API_URL_TOKEN } from '@waiting-room/api-client';

import { routes } from './app.routes';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideZonelessChangeDetection(),
    provideRouter(routes),
    provideHttpClient(
      // Use functional interceptors with withInterceptors
      withInterceptors([tenantInterceptor])
    ),
    // Provide API URL for tenant service
    { provide: TENANT_API_URL, useValue: environment.apiUrl },
    // Provide API URL for api-client WebSocket
    { provide: API_URL_TOKEN, useValue: environment.apiUrl },
    // Provide TenantService to the injection token for api-client
    { provide: TENANT_SERVICE_TOKEN, useExisting: TenantService }
  ]
};
