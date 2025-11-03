import { Component, inject, signal, output, HostListener, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TenantService, CreateTenantRequest, Tenant } from '@lib/tenant';

@Component({
  selector: 'app-create-tenant-modal',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    @if (isOpen()) {
      <div class="fixed inset-0 z-50 overflow-y-auto" (click)="closeOnBackdrop($event)">
        <div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
          <!-- Background overlay -->
          <div class="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75"></div>
          
          <!-- Modal panel -->
          <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full"
               (click)="$event.stopPropagation()">
            <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
              <div class="flex items-center justify-between mb-4">
                <h3 class="text-lg leading-6 font-medium text-gray-900">
                  Create New Tenant
                </h3>
                <button
                  type="button"
                  (click)="close()"
                  class="text-gray-400 hover:text-gray-500 focus:outline-none">
                  <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                  </svg>
                </button>
              </div>
              
              <form (ngSubmit)="createTenant($event)" class="space-y-4">
                <div>
                  <label for="modal-buildingId" class="block text-sm font-medium text-gray-700 mb-2">
                    Building ID *
                  </label>
                  <input
                    type="text"
                    id="modal-buildingId"
                    name="buildingId"
                    [(ngModel)]="newTenant.buildingId"
                    required
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="e.g., Building-A">
                </div>
                
                <div>
                  <label for="modal-sectionId" class="block text-sm font-medium text-gray-700 mb-2">
                    Section ID *
                  </label>
                  <input
                    type="text"
                    id="modal-sectionId"
                    name="sectionId"
                    [(ngModel)]="newTenant.sectionId"
                    required
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="e.g., Section-1">
                </div>
                
                <div>
                  <label for="modal-name" class="block text-sm font-medium text-gray-700 mb-2">
                    Tenant Name *
                  </label>
                  <input
                    type="text"
                    id="modal-name"
                    name="name"
                    [(ngModel)]="newTenant.name"
                    required
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="e.g., Main Office">
                </div>
                
                <div>
                  <label for="modal-description" class="block text-sm font-medium text-gray-700 mb-2">
                    Description (Optional)
                  </label>
                  <textarea
                    id="modal-description"
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
                
                <div class="flex space-x-3 pt-4">
                  <button
                    type="button"
                    (click)="close()"
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
          </div>
        </div>
      </div>
    }
  `
})
export class CreateTenantModalComponent {
  tenantService = inject(TenantService);
  
  newTenant: CreateTenantRequest = {
    buildingId: '',
    sectionId: '',
    name: '',
    description: ''
  };
  
  tenantCreated = output<void>();
  
  // Get isOpen from service signal
  get isOpen() {
    return this.tenantService.showCreateModal;
  }
  
  constructor() {
    // Watch for service signal changes to reset form
    effect(() => {
      if (this.tenantService.showCreateModal()) {
        // Reset form when opening
        this.newTenant = {
          buildingId: '',
          sectionId: '',
          name: '',
          description: ''
        };
      }
    });
  }
  
  close(): void {
    this.tenantService.closeCreateModal();
    this.tenantService.clearError();
  }
  
  closeOnBackdrop(event: Event): void {
    const target = event.target as HTMLElement;
    if (target.classList.contains('fixed') && target.classList.contains('inset-0')) {
      this.close();
    }
  }
  
  @HostListener('document:keydown.escape', ['$event'])
  onEscapeKey(event: Event): void {
    if (this.isOpen()) {
      this.close();
    }
  }
  
  createTenant(event: Event): void {
    event.preventDefault();
    event.stopPropagation();
    
    if (!this.newTenant.buildingId || !this.newTenant.sectionId || !this.newTenant.name) {
      this.tenantService.setError('All required fields must be filled');
      return;
    }
    
    this.tenantService.createTenant(this.newTenant).subscribe({
      next: (tenant: Tenant) => {
        // Auto-select the newly created tenant
        this.tenantService.setSelectedTenant(tenant.id);
        // Close modal
        this.close();
        // Emit event
        this.tenantCreated.emit();
      },
      error: (error: any) => {
        // Error is already handled in the service
        console.error('Error creating tenant:', error);
      }
    });
  }
}
