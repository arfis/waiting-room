import { Component, inject, OnInit, signal, effect, HostListener, output, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TenantService } from './tenant.service';

@Component({
  selector: 'app-tenant-selector',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="relative tenant-selector">
      <button
        type="button"
        (click)="toggleDropdown()"
        class="flex items-center px-3 py-2 text-sm font-medium bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 min-w-[250px] justify-between shadow-sm"
        [class.opacity-50]="loading()"
        [disabled]="loading()">
        <div class="flex items-center space-x-2 flex-1 min-w-0">
          <svg class="w-4 h-4 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path>
          </svg>
          <span class="truncate">{{ getDisplayName() }}</span>
        </div>
        <svg
          class="w-4 h-4 text-gray-500 dark:text-gray-300 transition-transform duration-200 flex-shrink-0"
          [class.rotate-180]="isOpen()"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
        </svg>
      </button>

      @if (isOpen()) {
        <div
          class="absolute right-0 mt-1 w-full bg-white border border-gray-300 rounded-md shadow-lg z-50 max-h-96 overflow-y-auto">
          @if (loading()) {
            <div class="px-4 py-3 text-sm text-gray-500 text-center">
              <div class="inline-block animate-spin rounded-full h-4 w-4 border-b-2 border-gray-900 mr-2"></div>
              Loading...
            </div>
          } @else if (error()) {
            <div class="px-4 py-3 text-sm text-red-600">
              {{ error() }}
            </div>
          } @else if (tenants().length === 0) {
            <div class="px-4 py-3">
              <div class="text-sm text-gray-500 text-center mb-2">
                No tenants available
              </div>
              @if (showCreateButton()) {
                <button
                  type="button"
                  (click)="onCreateTenantClick()"
                  class="w-full px-3 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors flex items-center justify-center">
                  <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
                  </svg>
                  Create First Tenant
                </button>
              }
            </div>
          } @else {
            <!-- Tenant Selection -->
            @if (!showingSections()) {
              <div class="py-1">
                <div class="px-4 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                  Select Tenant
                </div>
                @for (tenant of tenants(); track tenant.id) {
                  <button
                    type="button"
                    (click)="onTenantSelect(tenant.tenantId)"
                    class="w-full text-left px-4 py-2 text-sm hover:bg-gray-100 transition-colors flex items-center justify-between group"
                    [class.bg-blue-50]="selectedTenantId() === tenant.tenantId && !selectedSectionId()"
                    [class.text-blue-700]="selectedTenantId() === tenant.tenantId && !selectedSectionId()">
                    <div class="flex-1">
                      <div class="font-medium text-gray-900">{{ tenant.name }}</div>
                      <div class="text-xs text-gray-600">ID: {{ tenant.tenantId }}</div>
                    </div>
                    <svg class="w-4 h-4 text-gray-400 group-hover:text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
                    </svg>
                  </button>
                }
                @if (showCreateButton()) {
                  <div class="border-t border-gray-200 mt-1 pt-1">
                    <button
                      type="button"
                      (click)="onCreateTenantClick()"
                      class="w-full text-left px-4 py-2 text-sm text-blue-600 hover:bg-blue-50 transition-colors flex items-center">
                      <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
                      </svg>
                      Create New Tenant
                    </button>
                  </div>
                }
              </div>
            } @else {
              <!-- Section Selection -->
              <div class="py-1">
                <div class="px-4 py-2 border-b border-gray-200 flex items-center justify-between">
                  <button
                    type="button"
                    (click)="backToTenants(); $event.stopPropagation()"
                    class="flex items-center text-sm text-gray-600 hover:text-gray-900 font-medium">
                    <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
                    </svg>
                    Back to Tenants
                  </button>
                  <span class="text-xs font-semibold text-gray-500 uppercase">Sections</span>
                </div>

                <!-- All Sections Option -->
                <button
                  type="button"
                  (click)="onSectionSelect(0)"
                  class="w-full text-left px-4 py-2 text-sm hover:bg-gray-100 transition-colors border-b border-gray-100"
                  [class.bg-blue-50]="!selectedSectionId()"
                  [class.text-blue-700]="!selectedSectionId()"
                  [class.font-medium]="!selectedSectionId()">
                  <div class="flex items-center">
                    <svg class="w-4 h-4 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>
                    </svg>
                    <span class="font-medium">All Sections</span>
                  </div>
                </button>

                @if (loadingSections()) {
                  <div class="px-4 py-3 text-sm text-gray-500 text-center">
                    <div class="inline-block animate-spin rounded-full h-4 w-4 border-b-2 border-gray-900 mr-2"></div>
                    Loading sections...
                  </div>
                } @else if (sections().length === 0) {
                  <div class="px-4 py-3 text-sm text-gray-500 text-center">
                    No sections available for this tenant
                  </div>
                } @else {
                  @for (section of sections(); track section.id) {
                    <button
                      type="button"
                      (click)="onSectionSelect(section.sectionId)"
                      class="w-full text-left px-4 py-2 text-sm hover:bg-gray-100 transition-colors"
                      [class.bg-blue-50]="selectedSectionId() === section.sectionId"
                      [class.text-blue-700]="selectedSectionId() === section.sectionId"
                      [class.font-medium]="selectedSectionId() === section.sectionId">
                      <div class="font-medium text-gray-900">{{ section.name }}</div>
                      <div class="text-xs text-gray-600">ID: {{ section.sectionId }}</div>
                    </button>
                  }
                }
              </div>
            }
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .tenant-selector {
      // Component styles
    }
  `]
})
export class TenantSelectorComponent implements OnInit {
  private tenantService = inject(TenantService);

  selectedTenantId = signal<number>(0);
  selectedSectionId = signal<number>(0);
  isOpen = signal<boolean>(false);
  showCreateForm = signal<boolean>(false);
  showingSections = signal<boolean>(false);
  loadingSections = signal<boolean>(false);
  sectionsCache = signal<Map<number, any[]>>(new Map());
  private pendingTenantId = signal<number | null>(null);

  // Input to control whether to show create tenant button (for admin apps)
  showCreateButton = input<boolean>(false);

  // Output event to trigger create form in parent
  createTenantRequested = output<void>();

  // Expose readonly signals from service
  get tenants() {
    return this.tenantService.tenants;
  }

  get sections() {
    return this.tenantService.sections;
  }

  get loading() {
    return this.tenantService.loading;
  }

  get error() {
    return this.tenantService.error;
  }

  constructor() {
    // Sync with service's selected tenant and section
    effect(() => {
      const serviceTenantId = this.tenantService.selectedTenantId();
      const serviceSectionId = this.tenantService.selectedSectionId();

      if (serviceTenantId && serviceTenantId !== this.selectedTenantId()) {
        this.selectedTenantId.set(serviceTenantId);
      }

      if (serviceSectionId !== this.selectedSectionId()) {
        this.selectedSectionId.set(serviceSectionId);
      }
    });

    // Watch for sections loading completion
    effect(() => {
      const sections = this.tenantService.sections();
      const loading = this.tenantService.loading();
      const pendingId = this.pendingTenantId();

      // If we were waiting for sections to load for a tenant
      if (pendingId && !loading) {
        console.log('[TenantSelector] Sections loaded for tenant', pendingId, sections.length);

        // Cache the sections
        const cache = this.sectionsCache();
        cache.set(pendingId, [...sections]);
        this.sectionsCache.set(new Map(cache));
        this.loadingSections.set(false);
        this.pendingTenantId.set(null);

        // If no sections, automatically select the tenant
        if (sections.length === 0) {
          console.log('[TenantSelector] No sections for tenant, selecting tenant only');
          this.selectTenantOnly(pendingId);
        }
      }
    });
  }

  ngOnInit(): void {
    // Initialize selected tenant and section from service
    const selectedTenant = this.tenantService.getSelectedTenant();
    if (selectedTenant) {
      this.selectedTenantId.set(selectedTenant.tenantId);
    }

    const selectedSection = this.tenantService.getSelectedSection();
    if (selectedSection) {
      this.selectedSectionId.set(selectedSection.sectionId);
    }

    // Load tenants if not already loaded
    if (this.tenantService.tenants().length === 0) {
      this.tenantService.loadTenants();
    }
  }

  toggleDropdown(): void {
    this.isOpen.update(val => !val);
    // Always reset to tenant list when opening/closing dropdown
    if (!this.isOpen()) {
      this.showingSections.set(false);
    }
  }

  onTenantSelect(tenantId: number): void {
    // Set tenant and check if it has sections
    this.selectedTenantId.set(tenantId);

    // Check if sections are already cached
    const cache = this.sectionsCache();
    if (cache.has(tenantId)) {
      // Use cached sections
      const cachedSections = cache.get(tenantId) || [];
      console.log('[TenantSelector] Using cached sections for tenant', tenantId, cachedSections.length);

      if (cachedSections.length === 0) {
        // No sections available - select tenant directly
        this.selectTenantOnly(tenantId);
      } else {
        // Has sections - show section selection
        this.showingSections.set(true);
      }
    } else {
      // Load sections for this tenant
      this.loadSectionsForTenant(tenantId);
    }
  }

  private loadSectionsForTenant(tenantId: number): void {
    this.loadingSections.set(true);
    this.showingSections.set(true);
    this.pendingTenantId.set(tenantId);

    // Load sections via service (async operation)
    // The effect watching sections() and loading() will handle completion
    this.tenantService.setSelectedTenant(tenantId);
  }

  private selectTenantOnly(tenantId: number): void {
    // Select tenant without section (all sections)
    this.selectedTenantId.set(tenantId);
    this.selectedSectionId.set(0);
    this.tenantService.setSelectedTenant(tenantId);
    this.isOpen.set(false);
    this.showingSections.set(false);
  }

  onSectionSelect(sectionId: number): void {
    this.selectedSectionId.set(sectionId);

    if (sectionId === 0) {
      // "All Sections" selected - use tenant only
      this.tenantService.setSelectedTenant(this.selectedTenantId());
    } else {
      // Specific section selected
      this.tenantService.setSelectedTenant(this.selectedTenantId(), sectionId);
    }

    this.isOpen.set(false);
    this.showingSections.set(false);
  }

  backToTenants(): void {
    console.log('[TenantSelector] Back button clicked - showing tenants');
    this.showingSections.set(false);
  }

  onCreateTenantClick(): void {
    this.createTenantRequested.emit();
    this.isOpen.set(false);
  }

  getDisplayName(): string {
    const tenant = this.tenantService.getSelectedTenant();
    const section = this.tenantService.getSelectedSection();

    if (!tenant) {
      return 'Select Tenant & Section';
    }

    if (section) {
      return `${tenant.name} → ${section.name}`;
    }

    return `${tenant.name} (All Sections)`;
  }

  @HostListener('document:click', ['$event'])
  onDocumentClick(event: MouseEvent): void {
    const target = event.target as HTMLElement;
    if (!target.closest('.tenant-selector')) {
      this.isOpen.set(false);
    }
  }
}

