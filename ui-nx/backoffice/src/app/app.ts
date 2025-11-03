import { Component, inject, computed, OnInit } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { TenantSelectionScreenComponent } from './components/tenant-selection-screen/tenant-selection-screen';
import { TenantService } from '@lib/tenant';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, TenantSelectionScreenComponent],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class AppComponent implements OnInit {
  protected readonly title = 'backoffice';
  
  private tenantService = inject(TenantService);
  
  // Check if tenant is selected
  hasSelectedTenant = computed(() => {
    const tenantId = this.tenantService.selectedTenantId();
    return !!tenantId;
  });

  ngOnInit(): void {
    // Load tenants on app initialization
    this.tenantService.loadTenants();
  }
}
