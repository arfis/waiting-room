import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { TenantService, Section, CreateSectionRequest } from '@lib/tenant';

@Component({
  selector: 'app-sections',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './sections.html',
  styleUrl: './sections.scss'
})
export class SectionsComponent implements OnInit {
  tenantService = inject(TenantService);
  router = inject(Router);

  showCreateForm = signal<boolean>(false);
  editingSection = signal<Section | null>(null);
  testingWebSocket = signal<number | null>(null);
  wsTestStatus = signal<Map<number, 'connected' | 'disconnected' | 'testing'>>(new Map());

  newSection: CreateSectionRequest = {
    name: '',
    description: ''
  };

  editSection: CreateSectionRequest = {
    name: '',
    description: ''
  };

  // Computed values
  sections = computed(() => this.tenantService.sections());
  selectedTenant = computed(() => this.tenantService.getSelectedTenant());
  selectedTenantId = computed(() => this.tenantService.selectedTenantId());
  loading = computed(() => this.tenantService.loading());
  error = computed(() => this.tenantService.error());

  ngOnInit(): void {
    // Load sections if tenant is selected
    const tenantId = this.selectedTenantId();
    if (tenantId) {
      this.tenantService.loadSections();
    }
  }

  openCreateForm(): void {
    this.showCreateForm.set(true);
    this.editingSection.set(null);
    this.newSection = {
      name: '',
      description: ''
    };
  }

  closeCreateForm(): void {
    this.showCreateForm.set(false);
    this.newSection = {
      name: '',
      description: ''
    };
  }

  createSection(): void {
    if (!this.newSection.name) {
      this.tenantService.setError('Section name is required');
      return;
    }

    this.tenantService.createSection(this.newSection).subscribe({
      next: () => {
        this.closeCreateForm();
        this.tenantService.clearError();
      },
      error: (error) => {
        console.error('Error creating section:', error);
      }
    });
  }

  openEditForm(section: Section): void {
    this.editingSection.set(section);
    this.showCreateForm.set(false);
    this.editSection = {
      name: section.name,
      description: section.description
    };
  }

  closeEditForm(): void {
    this.editingSection.set(null);
    this.editSection = {
      name: '',
      description: ''
    };
  }

  saveSection(): void {
    const section = this.editingSection();
    if (!section) return;

    if (!this.editSection.name) {
      this.tenantService.setError('Section name is required');
      return;
    }

    this.tenantService.updateSection(section.sectionId, this.editSection).subscribe({
      next: () => {
        this.closeEditForm();
        this.tenantService.clearError();
      },
      error: (error) => {
        console.error('Error updating section:', error);
      }
    });
  }

  deleteSection(section: Section): void {
    if (!confirm(`Are you sure you want to delete "${section.name}"?`)) {
      return;
    }

    this.tenantService.deleteSection(section.sectionId).subscribe({
      next: () => {
        this.tenantService.clearError();
      },
      error: (error) => {
        console.error('Error deleting section:', error);
      }
    });
  }

  configureSection(section: Section): void {
    // Select this section and navigate to configuration
    this.tenantService.setSelectedTenant(this.selectedTenantId(), section.sectionId);
    this.router.navigate(['/configuration']);
  }

  testWebSocket(section: Section): void {
    this.testingWebSocket.set(section.sectionId);
    this.wsTestStatus.update(map => {
      map.set(section.sectionId, 'testing');
      return new Map(map);
    });

    const tenant = this.selectedTenant();
    if (!tenant) return;

    // Construct WebSocket URL
    const wsUrl = `ws://localhost:8080/ws/test-room?tenantId=${tenant.tenantId}:${section.sectionId}`;

    try {
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        console.log('WebSocket connected:', wsUrl);
        this.wsTestStatus.update(map => {
          map.set(section.sectionId, 'connected');
          return new Map(map);
        });

        // Close after 2 seconds
        setTimeout(() => {
          ws.close();
          this.testingWebSocket.set(null);
        }, 2000);
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.wsTestStatus.update(map => {
          map.set(section.sectionId, 'disconnected');
          return new Map(map);
        });
        this.testingWebSocket.set(null);
      };

      ws.onclose = () => {
        console.log('WebSocket closed');
        setTimeout(() => {
          this.wsTestStatus.update(map => {
            map.delete(section.sectionId);
            return new Map(map);
          });
        }, 3000);
      };
    } catch (error) {
      console.error('Failed to create WebSocket:', error);
      this.wsTestStatus.update(map => {
        map.set(section.sectionId, 'disconnected');
        return new Map(map);
      });
      this.testingWebSocket.set(null);
    }
  }

  getWebSocketStatus(sectionId: number): 'connected' | 'disconnected' | 'testing' | null {
    return this.wsTestStatus().get(sectionId) || null;
  }

  getStatusBadgeClass(status: string | undefined): string {
    if (status === 'active') {
      return 'bg-green-100 text-green-800';
    } else if (status === 'inactive') {
      return 'bg-red-100 text-red-800';
    }
    return 'bg-gray-100 text-gray-800';
  }
}
