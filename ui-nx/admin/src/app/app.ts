import { Component, signal, inject, OnInit, computed } from '@angular/core';
import { RouterOutlet, Router, NavigationEnd, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { filter } from 'rxjs/operators';
import { TenantSelectorComponent, TenantService } from '@lib/tenant';
import { TenantSelectionScreenComponent } from './components/tenant-selection-screen/tenant-selection-screen';
import { CreateTenantModalComponent } from './shared/components/create-tenant-modal/create-tenant-modal';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, RouterOutlet, RouterModule, TenantSelectorComponent, TenantSelectionScreenComponent, CreateTenantModalComponent],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class AppComponent implements OnInit {
  title = 'admin';
  
  sidebarOpen = signal(true); // Start with sidebar open
  currentRoute = signal('');
  
  private tenantService = inject(TenantService);
  
  // Check if tenant is selected
  hasSelectedTenant = computed(() => {
    const tenantId = this.tenantService.selectedTenantId();
    return !!tenantId;
  });

  constructor(private router: Router) {
    // Track current route for page title
    this.router.events
      .pipe(filter(event => event instanceof NavigationEnd))
      .subscribe((event: NavigationEnd) => {
        this.currentRoute.set(event.url);
        // Close sidebar on mobile after navigation
        if (window.innerWidth < 1024) {
          this.sidebarOpen.set(false);
        }
      });
  }

  ngOnInit(): void {
    // Load tenants on app initialization
    this.tenantService.loadTenants();
  }

  toggleSidebar(): void {
    this.sidebarOpen.set(!this.sidebarOpen());
  }

  closeSidebar(): void {
    this.sidebarOpen.set(false);
  }

  showCreateTenantForm(): void {
    // Open the create tenant modal via service
    this.tenantService.openCreateModal();
  }
  
  onTenantCreated(): void {
    // Tenant was created, modal will auto-close and select the tenant
    // No additional action needed
  }

  getPageTitle(): string {
    const route = this.currentRoute();
    switch (route) {
      case '/dashboard':
        return 'Dashboard';
      case '/configuration':
        return 'Configuration';
      case '/card-readers':
        return 'Card Readers';
      default:
        return 'Admin Panel';
    }
  }
}