import { Injectable, signal, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { TenantConfigService } from './config.service';
import { Observable, tap } from 'rxjs';

export interface Tenant {
  id: string;
  buildingId: string;
  sectionId: string;
  name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateTenantRequest {
  buildingId: string;
  sectionId: string;
  name: string;
  description?: string;
}

@Injectable({
  providedIn: 'root'
})
export class TenantService {
  private http = inject(HttpClient);
  private configService = inject(TenantConfigService);
  
  private _selectedTenantId = signal<string>('');
  private _tenants = signal<Tenant[]>([]);
  private _loading = signal<boolean>(false);
  private _error = signal<string>('');
  private _showCreateForm = signal<boolean>(false);
  private _showCreateModal = signal<boolean>(false);

  // Public readonly signals
  readonly selectedTenantId = this._selectedTenantId.asReadonly();
  readonly tenants = this._tenants.asReadonly();
  readonly loading = this._loading.asReadonly();
  readonly error = this._error.asReadonly();
  readonly showCreateForm = this._showCreateForm.asReadonly();
  readonly showCreateModal = this._showCreateModal.asReadonly();

  constructor() {
    // Load selected tenant from localStorage on service initialization
    // Check if we're in browser environment (not SSR)
    if (typeof window !== 'undefined' && window.localStorage) {
      const savedTenantId = localStorage.getItem('selectedTenantId');
      if (savedTenantId) {
        console.log(`[TenantService] Loading saved tenant from localStorage: ${savedTenantId}`);
        
        // If it's in the old format (database ID without colon), we'll need to convert it
        // after tenants are loaded. For now, just store it as-is.
        // If it's already in the new format (buildingId:sectionId), use it directly.
        this._selectedTenantId.set(savedTenantId);
      } else {
        console.log(`[TenantService] No saved tenant found in localStorage`);
      }
    } else {
      console.log(`[TenantService] Not in browser environment (SSR), skipping localStorage load`);
    }
    console.log(`[TenantService] Initialized with tenant ID: ${this._selectedTenantId() || 'none'}`);
  }

  getTenants(): Observable<Tenant[]> {
    this._loading.set(true);
    this._error.set('');
    
    return this.http.get<Tenant[]>(this.configService.adminTenantsUrl);
  }

  getTenant(id: string): Observable<Tenant> {
    return this.http.get<Tenant>(`${this.configService.adminTenantsUrl}/${id}`);
  }

  createTenant(tenant: CreateTenantRequest): Observable<Tenant> {
    this._loading.set(true);
    this._error.set('');
    
    return this.http.post<Tenant>(this.configService.adminTenantsUrl, tenant).pipe(
      tap({
        next: (createdTenant) => {
          this._loading.set(false);
          this._error.set('');
          // Reload tenants list after creation
          this.loadTenants();
        },
        error: (error) => {
          this._loading.set(false);
          this._error.set(error.error?.message || 'Failed to create tenant');
        }
      })
    );
  }

  updateTenant(id: string, tenant: Partial<CreateTenantRequest>): Observable<Tenant> {
    this._loading.set(true);
    this._error.set('');
    
    return this.http.put<Tenant>(`${this.configService.adminTenantsUrl}/${id}`, tenant).pipe(
      tap({
        next: () => {
          this._loading.set(false);
          this._error.set('');
        },
        error: (error) => {
          this._loading.set(false);
          this._error.set(error.error?.message || 'Failed to update tenant');
        }
      })
    );
  }

  deleteTenant(id: string): Observable<void> {
    this._loading.set(true);
    this._error.set('');
    
    return this.http.delete<void>(`${this.configService.adminTenantsUrl}/${id}`).pipe(
      tap({
        next: () => {
          this._loading.set(false);
          this._error.set('');
          // Reload tenants list after deletion
          this.loadTenants();
        },
        error: (error) => {
          this._loading.set(false);
          this._error.set(error.error?.message || 'Failed to delete tenant');
        }
      })
    );
  }

  loadTenants(): void {
    this._loading.set(true);
    this._error.set('');
    
    this.getTenants().subscribe({
      next: (tenants) => {
        this._tenants.set(tenants);
        this._loading.set(false);
        this._error.set('');
        
        // Migrate old format (database ID) to new format (buildingId:sectionId) if needed
        const currentTenantId = this._selectedTenantId();
        if (currentTenantId && !currentTenantId.includes(':')) {
          // It's in the old format (database ID), convert to new format
          const tenant = tenants.find(t => t.id === currentTenantId);
          if (tenant) {
            const fullTenantId = `${tenant.buildingId}:${tenant.sectionId}`;
            console.log(`[TenantService] Migrating tenant ID from old format to new: ${currentTenantId} -> ${fullTenantId}`);
            this._selectedTenantId.set(fullTenantId);
            if (typeof window !== 'undefined' && window.localStorage) {
              localStorage.setItem('selectedTenantId', fullTenantId);
            }
          } else {
            console.warn(`[TenantService] Saved tenant ID ${currentTenantId} not found in tenants list, clearing`);
            this._selectedTenantId.set('');
            if (typeof window !== 'undefined' && window.localStorage) {
              localStorage.removeItem('selectedTenantId');
            }
          }
        }
        
        // Auto-select first tenant if none selected and auto-select is enabled
        if (!this._selectedTenantId() && tenants.length > 0) {
          // Only auto-select if explicitly enabled (for kiosk apps)
          // For admin, this should be false
        }
      },
      error: (error) => {
        this._loading.set(false);
        this._error.set(error.error?.message || 'Failed to load tenants');
      }
    });
  }

  setSelectedTenant(tenantId: string): void {
    console.log(`[TenantService] Setting selected tenant: ${tenantId}`);
    console.log(`[TenantService] Previous tenant ID: ${this._selectedTenantId()}`);
    
    // Find the tenant to get the full identifier format
    const tenant = this._tenants().find(t => t.id === tenantId);
    let fullTenantId = tenantId; // Default to the ID if tenant not found
    
    if (tenant) {
      // Store the full identifier format: "buildingId:sectionId"
      fullTenantId = `${tenant.buildingId}:${tenant.sectionId}`;
      console.log(`[TenantService] Converting tenant ID to full format: ${tenantId} -> ${fullTenantId}`);
    } else {
      // If tenant not found, check if it's already in the full format
      if (tenantId.includes(':')) {
        fullTenantId = tenantId;
        console.log(`[TenantService] Tenant ID already in full format: ${fullTenantId}`);
      } else {
        console.warn(`[TenantService] Tenant with ID ${tenantId} not found in tenants list, using as-is`);
      }
    }
    
    this._selectedTenantId.set(fullTenantId);
    // Check if we're in browser environment (not SSR)
    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.setItem('selectedTenantId', fullTenantId);
      console.log(`[TenantService] Saved full tenant ID to localStorage: ${fullTenantId}`);
      // Verify it was saved
      const verify = localStorage.getItem('selectedTenantId');
      console.log(`[TenantService] Verified localStorage value: ${verify}`);
    }
    // Double-check the signal was updated
    const currentValue = this._selectedTenantId();
    console.log(`[TenantService] Current selected tenant ID signal value after set: "${currentValue}"`);
    console.log(`[TenantService] Signal value type: ${typeof currentValue}, length: ${currentValue?.length || 0}`);
  }
  
  // Helper method to get tenant ID synchronously (for interceptors)
  getSelectedTenantIdSync(): string {
    const value = this._selectedTenantId();
    return value || '';
  }

  getSelectedTenant(): Tenant | null {
    const fullTenantId = this._selectedTenantId();
    if (!fullTenantId) return null;
    
    // Parse the full tenant ID format: "buildingId:sectionId"
    const [buildingId, sectionId] = fullTenantId.split(':');
    
    // Find tenant by buildingId and sectionId
    return this._tenants().find(t => t.buildingId === buildingId && t.sectionId === sectionId) || null;
  }
  
  // Get the tenant database ID from the full identifier format
  getSelectedTenantDatabaseId(): string | null {
    const tenant = this.getSelectedTenant();
    return tenant?.id || null;
  }

  getTenantDisplayName(tenant: Tenant): string {
    // Format: "Building:Section" (e.g., "Nemocnica Spiska nova ves:Kardiologia pavilon B")
    return `${tenant.buildingId}:${tenant.sectionId}`;
  }

  clearError(): void {
    this._error.set('');
  }

  setLoading(loading: boolean): void {
    this._loading.set(loading);
  }

  setError(error: string): void {
    this._error.set(error);
  }
  
  requestCreateForm(): void {
    this._showCreateForm.set(true);
  }
  
  clearCreateFormRequest(): void {
    this._showCreateForm.set(false);
  }
  
  openCreateModal(): void {
    this._showCreateModal.set(true);
  }
  
  closeCreateModal(): void {
    this._showCreateModal.set(false);
  }
}

