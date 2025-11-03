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
        class="flex items-center px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 min-w-[200px] justify-between"
        [class.opacity-50]="loading()"
        [disabled]="loading()">
        <div class="flex items-center space-x-2">
          <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path>
          </svg>
          <span class="truncate">{{ getSelectedTenantName() }}</span>
        </div>
        <svg 
          class="w-4 h-4 text-gray-400 transition-transform duration-200"
          [class.rotate-180]="isOpen()"
          fill="none" 
          stroke="currentColor" 
          viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
        </svg>
      </button>

      @if (isOpen()) {
        <div 
          class="absolute right-0 mt-1 w-full bg-white border border-gray-300 rounded-md shadow-lg z-50 max-h-60 overflow-auto">
          @if (loading()) {
            <div class="px-4 py-3 text-sm text-gray-500 text-center">
              <div class="inline-block animate-spin rounded-full h-4 w-4 border-b-2 border-gray-900 mr-2"></div>
              Loading tenants...
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
            <div class="py-1">
              @for (tenant of tenants(); track tenant.id) {
                <button
                  type="button"
                  (click)="onTenantChange(tenant.id)"
                  class="w-full text-left px-4 py-2 text-sm hover:bg-gray-100 transition-colors"
                  [class.bg-blue-50]="selectedTenantId() === tenant.id"
                  [class.text-blue-700]="selectedTenantId() === tenant.id"
                  [class.font-medium]="selectedTenantId() === tenant.id">
                  <div class="font-medium">{{tenant.buildingId}}:{{ tenant.name }}</div>
                  <div class="text-xs text-gray-500">{{ tenant.sectionId }}</div>
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
  
  selectedTenantId = signal<string>('');
  isOpen = signal<boolean>(false);
  showCreateForm = signal<boolean>(false);
  
  // Input to control whether to show create tenant button (for admin apps)
  showCreateButton = input<boolean>(false);
  
  // Output event to trigger create form in parent
  createTenantRequested = output<void>();

  // Expose readonly signals from service
  get tenants() {
    return this.tenantService.tenants;
  }

  get loading() {
    return this.tenantService.loading;
  }

  get error() {
    return this.tenantService.error;
  }

  constructor() {
    // Sync with service's selected tenant (convert full ID to database ID for comparison)
    effect(() => {
      const serviceFullTenantId = this.tenantService.selectedTenantId();
      const selectedTenant = this.tenantService.getSelectedTenant();
      const serviceTenantDatabaseId = selectedTenant?.id || '';
      
      if (serviceTenantDatabaseId && serviceTenantDatabaseId !== this.selectedTenantId()) {
        this.selectedTenantId.set(serviceTenantDatabaseId);
      }
    });
  }

  ngOnInit(): void {
    // Initialize selected tenant from service (convert full ID to database ID for display)
    const selectedTenant = this.tenantService.getSelectedTenant();
    if (selectedTenant) {
      this.selectedTenantId.set(selectedTenant.id);
    }
    
    // Load tenants if not already loaded
    if (this.tenantService.tenants().length === 0) {
      this.tenantService.loadTenants();
    }
  }

  toggleDropdown(): void {
    this.isOpen.update(val => !val);
  }

  onTenantChange(tenantId: string): void {
    this.tenantService.setSelectedTenant(tenantId);
    this.isOpen.set(false);
  }

  onCreateTenantClick(): void {
    this.createTenantRequested.emit();
    this.isOpen.set(false);
  }

  getSelectedTenantName(): string {
    const tenant = this.tenantService.getSelectedTenant();
    if (tenant) {
      return this.tenantService.getTenantDisplayName(tenant);
    }
    return 'Select Tenant';
  }

  @HostListener('document:click', ['$event'])
  onDocumentClick(event: MouseEvent): void {
    const target = event.target as HTMLElement;
    if (!target.closest('.tenant-selector')) {
      this.isOpen.set(false);
    }
  }
}

