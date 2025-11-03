import { Component, inject, OnInit, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TenantSelectorComponent, TenantService, CreateTenantRequest } from '@lib/tenant';

@Component({
  selector: 'app-tenant-selection-screen',
  standalone: true,
  imports: [CommonModule, FormsModule, TenantSelectorComponent],
  template: `
    <div class="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div class="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        <div class="text-center mb-6">
          <div class="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-blue-100 mb-4">
            <svg class="h-8 w-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path>
            </svg>
          </div>
          <h1 class="text-2xl font-bold text-gray-900 mb-2">Select Tenant</h1>
          <p class="text-gray-600">
            @if (tenantService.tenants().length === 0 && !tenantService.loading()) {
              Create your first tenant to get started.
            } @else {
              Please select a tenant to continue accessing the admin panel.
            }
          </p>
        </div>
        
        @if (showCreateForm()) {
          <!-- Create Tenant Form -->
          <div class="space-y-4">
            <h2 class="text-lg font-semibold text-gray-900">Create New Tenant</h2>
            
            <form (ngSubmit)="createTenant($event)" class="space-y-4" #tenantForm="ngForm">
              <div>
                <label for="buildingId" class="block text-sm font-medium text-gray-700 mb-2">
                  Building ID *
                </label>
                <input
                  type="text"
                  id="buildingId"
                  name="buildingId"
                  [(ngModel)]="newTenant.buildingId"
                  required
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="e.g., Building-A">
              </div>
              
              <div>
                <label for="sectionId" class="block text-sm font-medium text-gray-700 mb-2">
                  Section ID *
                </label>
                <input
                  type="text"
                  id="sectionId"
                  name="sectionId"
                  [(ngModel)]="newTenant.sectionId"
                  required
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="e.g., Section-1">
              </div>
              
              <div>
                <label for="name" class="block text-sm font-medium text-gray-700 mb-2">
                  Tenant Name *
                </label>
                <input
                  type="text"
                  id="name"
                  name="name"
                  [(ngModel)]="newTenant.name"
                  required
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="e.g., Main Office">
              </div>
              
              <div>
                <label for="description" class="block text-sm font-medium text-gray-700 mb-2">
                  Description (Optional)
                </label>
                <textarea
                  id="description"
                  name="description"
                  [(ngModel)]="newTenant.description"
                  rows="2"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Optional description"></textarea>
              </div>
              
              @if (tenantService.error()) {
                <div class="bg-red-50 border border-red-200 rounded-md p-3">
                  <p class="text-sm text-red-800">{{ tenantService.error() }}</p>
                </div>
              }
              
              <div class="flex space-x-3">
                <button
                  type="button"
                  (click)="showCreateForm.set(false)"
                  class="flex-1 px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 transition-colors">
                  Cancel
                </button>
                <button
                  type="submit"
                  [disabled]="tenantService.loading()"
                  class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed">
                  @if (tenantService.loading()) {
                    <span class="flex items-center justify-center">
                      <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Creating...
                    </span>
                  } @else {
                    Create Tenant
                  }
                </button>
              </div>
            </form>
          </div>
        } @else {
          <!-- Select Existing Tenant -->
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Tenant
              </label>
              <app-tenant-selector (createTenantRequested)="showCreateForm.set(true)"></app-tenant-selector>
            </div>
            
            @if (tenantService.error()) {
              <div class="bg-red-50 border border-red-200 rounded-md p-3">
                <p class="text-sm text-red-800">{{ tenantService.error() }}</p>
              </div>
            }
            
            @if (tenantService.loading()) {
              <div class="text-center py-4">
                <div class="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
                <p class="mt-2 text-sm text-gray-600">Loading tenants...</p>
              </div>
            }
            
            @if (!tenantService.loading() && tenantService.tenants().length > 0) {
              <button
                type="button"
                (click)="showCreateForm.set(true)"
                class="w-full mt-4 px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 transition-colors text-sm font-medium">
                <span class="flex items-center justify-center">
                  <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
                  </svg>
                  Create New Tenant
                </span>
              </button>
            }
          </div>
          
          @if (tenantService.selectedTenantId()) {
            <div class="mt-6 bg-blue-50 border border-blue-200 rounded-md p-4">
              <p class="text-sm text-blue-800">
                <span class="font-medium">Selected:</span> 
                {{ getSelectedTenantName() }}
              </p>
              <button
                (click)="continue()"
                class="mt-3 w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors font-medium">
                Continue
              </button>
            </div>
          }
        }
      </div>
    </div>
  `
})
export class TenantSelectionScreenComponent implements OnInit {
  tenantService = inject(TenantService);
  
  showCreateForm = signal<boolean>(false);
  newTenant: CreateTenantRequest = {
    buildingId: '',
    sectionId: '',
    name: '',
    description: ''
  };

  constructor() {
    // Show create form if no tenants exist after loading completes
    effect(() => {
      const loading = this.tenantService.loading();
      const tenants = this.tenantService.tenants();
      if (!loading && tenants.length === 0) {
        this.showCreateForm.set(true);
      }
    });
    
    // Watch for create form request from service
    effect(() => {
      if (this.tenantService.showCreateForm()) {
        this.showCreateForm.set(true);
        this.tenantService.clearCreateFormRequest();
      }
    });
  }

  ngOnInit(): void {
    // Load tenants if not already loaded
    if (this.tenantService.tenants().length === 0 && !this.tenantService.loading()) {
      this.tenantService.loadTenants();
    }
  }

  getSelectedTenantName(): string {
    const tenant = this.tenantService.getSelectedTenant();
    if (tenant) {
      return this.tenantService.getTenantDisplayName(tenant);
    }
    return '';
  }

  createTenant(event?: Event): void {
    if (event) {
      event.preventDefault();
      event.stopPropagation();
    }
    
    if (!this.newTenant.buildingId || !this.newTenant.sectionId || !this.newTenant.name) {
      this.tenantService.setError('All required fields must be filled');
      return;
    }
    
    this.tenantService.createTenant(this.newTenant).subscribe({
      next: (tenant) => {
        // Auto-select the newly created tenant
        this.tenantService.setSelectedTenant(tenant.id);
        // Hide create form
        this.showCreateForm.set(false);
        // Reset form
        this.newTenant = {
          buildingId: '',
          sectionId: '',
          name: '',
          description: ''
        };
      },
      error: (error) => {
        // Error is already handled in the service
        console.error('Error creating tenant:', error);
      }
    });
  }

  continue(): void {
    // Tenant is selected, the parent component will handle navigation
    // This is just a UI component
  }
}

