import { Component, inject, OnInit, OnDestroy, ChangeDetectionStrategy, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { QueueStateService } from '../core/services/queue-state.service';
import { QueueActionsComponent } from './components/queue-actions/queue-actions.component';
import { CurrentEntryCardComponent } from './components/current-entry-card/current-entry-card.component';
import { WaitingQueueListComponent } from './components/waiting-queue-list/waiting-queue-list.component';
import { ActivityLogComponent } from './components/activity-log/activity-log.component';
import { WebSocketQueueEntry, Api, ConfigurationResponse, ServicePointConfiguration, QueueEntryStatus } from 'api-client';

@Component({
  selector: 'app-queue-management',
  standalone: true,
  imports: [
    CommonModule,
    QueueActionsComponent,
    CurrentEntryCardComponent,
    WaitingQueueListComponent,
    ActivityLogComponent,

  ],
  templateUrl: './queue-management.component.html',
  styleUrl: './queue-management.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class QueueManagementComponent implements OnInit, OnDestroy {
  private readonly queueState = inject(QueueStateService);
  private readonly api = new Api();
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

  async ngOnInit(): Promise<void> {
    // Load configuration first, then initialize queue state
    await this.loadConfiguration();
  
    // Only initialize queue state if we have a selected service point
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
   
    // For now, just call the next person
    // In a real implementation, you might want a specific endpoint for calling a particular entry
    this.queueState.callNext(this.roomId, this.selectedServicePoint()?.id || '');
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
      const config = await this.api.configuration.getConfiguration();
      this.configuration.set(config);
      
      // If there's only one service point, auto-select it
      const servicePoints = this.availableServicePoints();
      if (servicePoints.length === 1) {
        this.selectedServicePoint.set(servicePoints[0]);
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
    // Initialize queue state after service point selection - load waiting and current entries
    this.queueState.initialize(this.roomId, ['WAITING', 'CALLED']);
  }

  protected clearServicePoint(): void {
    this.selectedServicePoint.set(null);
    this.queueState.disconnect();
  }

  protected retryConfiguration(): void {
    this.loadConfiguration();
  }
}
