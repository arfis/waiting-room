import { Injectable, signal, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { TenantConfigService } from './config.service';
import { Observable, tap } from 'rxjs';

export interface Tenant {
  id: string;              // MongoDB _id
  tenantId: number;        // Numeric tenant ID
  name: string;
  description?: string;
  databaseName: string;
  status: 'active' | 'inactive';
  createdAt: string;
  updatedAt: string;
}

export interface Section {
  id: string;              // MongoDB _id
  sectionId: number;       // Numeric section ID
  tenantId: number;        // Parent tenant ID
  name: string;
  description?: string;
  status: 'active' | 'inactive';
  createdAt: string;
  updatedAt: string;
}

export interface CreateTenantRequest {
  tenantId?: number;       // Optional: will be auto-generated if not provided
  name: string;
  description?: string;
}

export interface CreateSectionRequest {
  sectionId?: number;      // Optional: will be auto-generated if not provided
  name: string;
  description?: string;
}

@Injectable({
  providedIn: 'root'
})
export class TenantService {
  private http = inject(HttpClient);
  private configService = inject(TenantConfigService);

  // Tenant signals
  private _selectedTenantId = signal<number>(0);
  private _selectedSectionId = signal<number>(0);
  private _tenants = signal<Tenant[]>([]);
  private _sections = signal<Section[]>([]);
  private _loading = signal<boolean>(false);
  private _error = signal<string>('');
  private _showCreateForm = signal<boolean>(false);
  private _showCreateModal = signal<boolean>(false);

  // Public readonly signals
  readonly selectedTenantId = this._selectedTenantId.asReadonly();
  readonly selectedSectionId = this._selectedSectionId.asReadonly();
  readonly tenants = this._tenants.asReadonly();
  readonly sections = this._sections.asReadonly();
  readonly loading = this._loading.asReadonly();
  readonly error = this._error.asReadonly();
  readonly showCreateForm = this._showCreateForm.asReadonly();
  readonly showCreateModal = this._showCreateModal.asReadonly();

  constructor() {
    // Load selected tenant from localStorage on service initialization
    if (typeof window !== 'undefined' && window.localStorage) {
      const savedCompositeId = localStorage.getItem('selectedTenantId');
      if (savedCompositeId) {
        console.log(`[TenantService] Loading saved tenant from localStorage: ${savedCompositeId}`);

        // Parse composite ID format: "1001" or "1001:2001"
        const { tenantId, sectionId } = this.parseCompositeId(savedCompositeId);
        this._selectedTenantId.set(tenantId);
        if (sectionId) {
          this._selectedSectionId.set(sectionId);
        }
      } else {
        console.log(`[TenantService] No saved tenant found in localStorage`);
      }
    } else {
      console.log(`[TenantService] Not in browser environment (SSR), skipping localStorage load`);
    }
    console.log(`[TenantService] Initialized with tenant ID: ${this._selectedTenantId() || 'none'}`);
  }

  // Parse composite ID format "1001" or "1001:2001"
  private parseCompositeId(compositeId: string): { tenantId: number; sectionId: number | null } {
    const parts = compositeId.split(':');
    const tenantId = parseInt(parts[0], 10);
    const sectionId = parts.length > 1 ? parseInt(parts[1], 10) : null;
    return { tenantId, sectionId };
  }

  // Format composite ID as "1001" or "1001:2001"
  private formatCompositeId(tenantId: number, sectionId?: number): string {
    if (sectionId) {
      return `${tenantId}:${sectionId}`;
    }
    return `${tenantId}`;
  }

  // Tenant Management

  getTenants(): Observable<Tenant[]> {
    this._loading.set(true);
    this._error.set('');

    return this.http.get<Tenant[]>(this.configService.adminTenantsUrl);
  }

  getTenant(tenantId: number): Observable<Tenant> {
    return this.http.get<Tenant>(`${this.configService.adminTenantsUrl}/${tenantId}`);
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

  updateTenant(tenantId: number, tenant: Partial<CreateTenantRequest>): Observable<Tenant> {
    this._loading.set(true);
    this._error.set('');

    return this.http.put<Tenant>(`${this.configService.adminTenantsUrl}/${tenantId}`, {
      ...tenant,
      tenantId
    }).pipe(
      tap({
        next: () => {
          this._loading.set(false);
          this._error.set('');
          // Reload tenants after update
          this.loadTenants();
        },
        error: (error) => {
          this._loading.set(false);
          this._error.set(error.error?.message || 'Failed to update tenant');
        }
      })
    );
  }

  deleteTenant(tenantId: number): Observable<void> {
    this._loading.set(true);
    this._error.set('');

    return this.http.delete<void>(`${this.configService.adminTenantsUrl}/${tenantId}`).pipe(
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

        // Verify selected tenant still exists
        const currentTenantId = this._selectedTenantId();
        if (currentTenantId && !tenants.find(t => t.tenantId === currentTenantId)) {
          console.warn(`[TenantService] Selected tenant ${currentTenantId} not found in tenants list, clearing`);
          this.clearSelectedTenant();
        }
      },
      error: (error) => {
        this._loading.set(false);
        this._error.set(error.error?.message || 'Failed to load tenants');
      }
    });
  }

  // Section Management

  getSections(): Observable<Section[]> {
    const tenantId = this._selectedTenantId();
    if (!tenantId) {
      throw new Error('No tenant selected');
    }

    this._loading.set(true);
    this._error.set('');

    return this.http.get<Section[]>(`${this.configService.adminTenantsUrl}/${tenantId}/sections`);
  }

  getSection(sectionId: number): Observable<Section> {
    const tenantId = this._selectedTenantId();
    if (!tenantId) {
      throw new Error('No tenant selected');
    }

    return this.http.get<Section>(`${this.configService.adminTenantsUrl}/${tenantId}/sections/${sectionId}`);
  }

  createSection(section: CreateSectionRequest): Observable<Section> {
    const tenantId = this._selectedTenantId();
    if (!tenantId) {
      throw new Error('No tenant selected');
    }

    this._loading.set(true);
    this._error.set('');

    return this.http.post<Section>(`${this.configService.adminTenantsUrl}/${tenantId}/sections`, {
      ...section,
      tenantId
    }).pipe(
      tap({
        next: (createdSection) => {
          this._loading.set(false);
          this._error.set('');
          // Reload sections list after creation
          this.loadSections();
        },
        error: (error) => {
          this._loading.set(false);
          this._error.set(error.error?.message || 'Failed to create section');
        }
      })
    );
  }

  updateSection(sectionId: number, section: Partial<CreateSectionRequest>): Observable<Section> {
    const tenantId = this._selectedTenantId();
    if (!tenantId) {
      throw new Error('No tenant selected');
    }

    this._loading.set(true);
    this._error.set('');

    return this.http.put<Section>(`${this.configService.adminTenantsUrl}/${tenantId}/sections/${sectionId}`, {
      ...section,
      sectionId,
      tenantId
    }).pipe(
      tap({
        next: () => {
          this._loading.set(false);
          this._error.set('');
          // Reload sections after update
          this.loadSections();
        },
        error: (error) => {
          this._loading.set(false);
          this._error.set(error.error?.message || 'Failed to update section');
        }
      })
    );
  }

  deleteSection(sectionId: number): Observable<void> {
    const tenantId = this._selectedTenantId();
    if (!tenantId) {
      throw new Error('No tenant selected');
    }

    this._loading.set(true);
    this._error.set('');

    return this.http.delete<void>(`${this.configService.adminTenantsUrl}/${tenantId}/sections/${sectionId}`).pipe(
      tap({
        next: () => {
          this._loading.set(false);
          this._error.set('');
          // Reload sections list after deletion
          this.loadSections();
        },
        error: (error) => {
          this._loading.set(false);
          this._error.set(error.error?.message || 'Failed to delete section');
        }
      })
    );
  }

  loadSections(): void {
    const tenantId = this._selectedTenantId();
    if (!tenantId) {
      this._sections.set([]);
      return;
    }

    this._loading.set(true);
    this._error.set('');

    this.getSections().subscribe({
      next: (sections) => {
        this._sections.set(sections);
        this._loading.set(false);
        this._error.set('');

        // Verify selected section still exists
        const currentSectionId = this._selectedSectionId();
        if (currentSectionId && !sections.find(s => s.sectionId === currentSectionId)) {
          console.warn(`[TenantService] Selected section ${currentSectionId} not found in sections list, clearing`);
          this._selectedSectionId.set(0);
        }
      },
      error: (error) => {
        this._loading.set(false);
        this._error.set(error.error?.message || 'Failed to load sections');
      }
    });
  }

  // Selection Management

  setSelectedTenant(tenantId: number, sectionId?: number): void {
    console.log(`[TenantService] Setting selected tenant: ${tenantId}, section: ${sectionId || 'none'}`);

    this._selectedTenantId.set(tenantId);
    this._selectedSectionId.set(sectionId || 0);

    // Save composite ID to localStorage
    const compositeId = this.formatCompositeId(tenantId, sectionId);

    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.setItem('selectedTenantId', compositeId);
      console.log(`[TenantService] Saved composite ID to localStorage: ${compositeId}`);
    }

    // Load sections when tenant is selected
    if (tenantId) {
      this.loadSections();
    }
  }

  setSelectedSection(sectionId: number): void {
    console.log(`[TenantService] Setting selected section: ${sectionId}`);

    const tenantId = this._selectedTenantId();
    if (!tenantId) {
      console.error('[TenantService] Cannot set section without tenant');
      return;
    }

    this._selectedSectionId.set(sectionId);

    // Update localStorage with composite ID
    const compositeId = this.formatCompositeId(tenantId, sectionId);
    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.setItem('selectedTenantId', compositeId);
      console.log(`[TenantService] Updated composite ID in localStorage: ${compositeId}`);
    }
  }

  clearSelectedTenant(): void {
    console.log('[TenantService] Clearing selected tenant');
    this._selectedTenantId.set(0);
    this._selectedSectionId.set(0);
    this._sections.set([]);

    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.removeItem('selectedTenantId');
    }
  }

  // Helper method to get composite ID synchronously (for interceptors)
  getSelectedTenantIdSync(): string {
    const tenantId = this._selectedTenantId();
    const sectionId = this._selectedSectionId();

    if (!tenantId) {
      return '';
    }

    return this.formatCompositeId(tenantId, sectionId || undefined);
  }

  getSelectedTenant(): Tenant | null {
    const tenantId = this._selectedTenantId();
    if (!tenantId) return null;

    return this._tenants().find(t => t.tenantId === tenantId) || null;
  }

  getSelectedSection(): Section | null {
    const sectionId = this._selectedSectionId();
    if (!sectionId) return null;

    return this._sections().find(s => s.sectionId === sectionId) || null;
  }

  getTenantDisplayName(tenant: Tenant): string {
    // Format: "Tenant Name (ID: 1001)"
    return `${tenant.name} (ID: ${tenant.tenantId})`;
  }

  getSectionDisplayName(section: Section): string {
    // Format: "Section Name (ID: 2001)"
    return `${section.name} (ID: ${section.sectionId})`;
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
