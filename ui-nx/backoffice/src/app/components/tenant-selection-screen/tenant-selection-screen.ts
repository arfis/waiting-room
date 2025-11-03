import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TenantSelectorComponent, TenantService } from '@lib/tenant';

@Component({
  selector: 'app-tenant-selection-screen',
  standalone: true,
  imports: [CommonModule, TenantSelectorComponent],
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
              No tenants available. Please contact an administrator.
            } @else {
              Please select a tenant to continue.
            }
          </p>
        </div>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Tenant
            </label>
            <app-tenant-selector></app-tenant-selector>
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
        </div>
        
        @if (tenantService.selectedTenantId()) {
          <div class="mt-6 bg-blue-50 border border-blue-200 rounded-md p-4">
            <p class="text-sm text-blue-800">
              <span class="font-medium">Selected:</span> 
              {{ getSelectedTenantName() }}
            </p>
            <p class="text-xs text-blue-600 mt-1">
              You can now use the backoffice to manage the queue.
            </p>
          </div>
        }
      </div>
    </div>
  `
})
export class TenantSelectionScreenComponent implements OnInit {
  tenantService = inject(TenantService);

  ngOnInit(): void {
    // Load tenants on component initialization
    if (this.tenantService.tenants().length === 0) {
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
}

