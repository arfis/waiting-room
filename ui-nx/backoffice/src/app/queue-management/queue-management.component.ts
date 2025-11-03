import { Component, inject, OnInit, OnDestroy, ChangeDetectionStrategy, signal, computed, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { QueueStateService } from '../core/services/queue-state.service';
import { QueueActionsComponent } from './components/queue-actions/queue-actions.component';
import { CurrentEntryCardComponent } from './components/current-entry-card/current-entry-card.component';
import { WaitingQueueListComponent } from './components/waiting-queue-list/waiting-queue-list.component';
import { ActivityLogComponent } from './components/activity-log/activity-log.component';
import { WebSocketQueueEntry, ConfigurationResponse, ServicePointConfiguration, QueueEntryStatus } from '@waiting-room/api-client';
import { TenantSelectorComponent, TenantService } from '@lib/tenant';
import { environment } from '../../environments/environment';

@Component({
  selector: 'app-queue-management',
  standalone: true,
  imports: [
    CommonModule,
    QueueActionsComponent,
    CurrentEntryCardComponent,
    WaitingQueueListComponent,
    ActivityLogComponent,
    TenantSelectorComponent,
  ],
  templateUrl: './queue-management.component.html',
  styleUrl: './queue-management.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class QueueManagementComponent implements OnInit, OnDestroy {
  private readonly queueState = inject(QueueStateService);
  private readonly tenantService = inject(TenantService);
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl || 'http://localhost:8080/api';
  private readonly roomId = 'triage-1';

  // Configuration state
  private readonly configuration = signal<ConfigurationResponse | null>(null);
  protected readonly isConfigLoading = signal(true);
  protected readonly configError = signal<string | null>(null);
  protected readonly selectedServicePoint = signal<ServicePointConfiguration | null>(null);

  // Expose state to template
  protected readonly currentEntry = this.queueState.currentEntry;
  protected readonly waitingEntries = this.queueState.waitingEntries;
  protected readonly calledEntries = this.queueState.calledEntries;
  protected readonly isLoading = this.queueState.isLoading;
  protected readonly lastUpdated = this.queueState.lastUpdated;
  protected readonly recentActivity = this.queueState.recentActivity;
  protected readonly estimatedWaitTime = this.queueState.estimatedWaitTime;
  protected readonly isConnected = this.queueState.isConnected;
  protected readonly error = this.queueState.wsError;

  // Configuration-related computed properties
  protected readonly availableServicePoints = computed(() => {
    const config = this.configuration();
    if (!config || !config.rooms.length) return [];
    
    // Get service points from the default room or first room
    const defaultRoom = config.rooms.find((room: any) => room.id === config.defaultRoom) || config.rooms[0];
    return defaultRoom?.servicePoints || [];
  });

  protected readonly showServicePointSelection = computed(() => {
    return !this.isConfigLoading() && !this.configError() && !this.selectedServicePoint();
  });

  protected readonly showQueueManagement = computed(() => {
    return !this.isConfigLoading() && !this.configError() && !!this.selectedServicePoint();
  });

  private lastTenantId: string | null = null;

  constructor() {
    // Watch for tenant and service point availability to initialize queue state
    effect(() => {
      const tenantId = this.tenantService.selectedTenantId();
      const servicePoint = this.selectedServicePoint();
      const configLoaded = !this.isConfigLoading() && !this.configError();
      
      // Initialize queue state if we have tenant, service point, and config loaded
      if (tenantId && servicePoint && configLoaded && this.lastTenantId === null) {
        console.log('Initializing queue state with tenant:', tenantId, 'service point:', servicePoint.name);
        this.lastTenantId = tenantId;
        // Initialize queue state
        this.queueState.initialize(this.roomId, ['WAITING', 'IN_SERVICE']);
      }
      // Reload if tenant actually changed (not just initial load)
      else if (tenantId && servicePoint && configLoaded && tenantId !== this.lastTenantId && this.lastTenantId !== null) {
        console.log('Tenant changed in component from', this.lastTenantId, 'to', tenantId, '- reloading queue data');
        this.lastTenantId = tenantId;
        // Use the same states that were used during initialization
        this.queueState.initialize(this.roomId, ['WAITING', 'IN_SERVICE']);
      }
    });
  }

  async ngOnInit(): Promise<void> {
    // Check if tenant is selected before proceeding
    const tenantId = this.tenantService.getSelectedTenantIdSync();
    if (!tenantId) {
      console.warn('Cannot initialize queue management: No tenant selected');
      return;
    }
    
    // Load configuration first, then initialize queue state
    await this.loadConfiguration();
  
    // Initialize queue state if we have a selected service point
    // (service point might be auto-selected during loadConfiguration if there's only one)
    if (this.selectedServicePoint()) {
      await this.queueState.initialize(this.roomId, ['WAITING', 'IN_SERVICE']);
    }
  }

  ngOnDestroy(): void {
    this.queueState.disconnect();
  }

  protected onCallNext(): void {
    this.queueState.callNext(this.roomId, this.selectedServicePoint()?.name || '');
  }

  protected onFinishCurrent(): void {
    this.queueState.finishCurrent(this.roomId);
  }

  protected onRefresh(): void {
    // Refresh both waiting entries and current entry
    this.queueState.refreshWaitingEntries(this.roomId);
    this.queueState.refreshCurrentEntry(this.roomId);
  }

  protected onCallSpecificEntry(entry: WebSocketQueueEntry): void {
    const servicePoint = this.selectedServicePoint();
    if (!servicePoint) {
      console.warn('Cannot call entry: No service point selected');
      return;
    }
    
    // Call the specific entry by ID
    this.queueState.callSpecificEntry(this.roomId, servicePoint.id, entry.id);
  }

  protected getStatusText(status: string): string {
    return this.queueState.getStatusText(status);
  }

  protected getCalledEntryForSelectedServicePoint(): WebSocketQueueEntry | null {
    const selectedServicePoint = this.selectedServicePoint();
    if (!selectedServicePoint) {
      return null;
    }
    return this.queueState.getCalledEntriesForServicePoint(selectedServicePoint.name)[0] || null;
  }

  // Configuration methods
  private async loadConfiguration(): Promise<void> {
    this.isConfigLoading.set(true);
    this.configError.set(null);
    
    try {
      // Use HttpClient directly so it goes through the tenant interceptor
      const config = await this.http.get<ConfigurationResponse>(`${this.apiUrl}/config`).toPromise();
      if (config) {
        this.configuration.set(config);
        
        // If there's only one service point, auto-select it
        const servicePoints = this.availableServicePoints();
        if (servicePoints.length === 1) {
          this.selectedServicePoint.set(servicePoints[0]);
        }
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load configuration';
      this.configError.set(errorMessage);
      console.error('Configuration loading failed:', err);
    } finally {
      this.isConfigLoading.set(false);
    }
  }

  protected selectServicePoint(servicePoint: ServicePointConfiguration): void {
    this.selectedServicePoint.set(servicePoint);
    
    // Check if tenant is selected before initializing
    const tenantId = this.tenantService.getSelectedTenantIdSync();
    if (!tenantId) {
      console.warn('Cannot initialize queue state: No tenant selected');
      return;
    }
    
    // Initialize queue state after service point selection - load waiting and current entries
    // Use same states as ngOnInit to ensure consistency
    this.queueState.initialize(this.roomId, ['WAITING', 'IN_SERVICE']);
  }

  protected clearServicePoint(): void {
    this.selectedServicePoint.set(null);
    this.queueState.disconnect();
  }

  protected retryConfiguration(): void {
    this.loadConfiguration();
  }
}
