import { Injectable, signal, computed, inject, effect } from '@angular/core';
import { QueueWebSocketService, WebSocketQueueEntry, QueueEntryStatus } from '@waiting-room/api-client';
import { QueueApiService } from './queue-api.service';
import { TenantService } from '@lib/tenant';

export interface ActivityLog {
  id: string;
  ticketNumber: string;
  action: string;
  timestamp: Date;
}

@Injectable({
  providedIn: 'root'
})
export class QueueStateService {
  private readonly queueWebSocket = inject(QueueWebSocketService);
  private readonly queueApiService = inject(QueueApiService);
  private readonly tenantService = inject(TenantService);
  
  // WebSocket state
  readonly queueEntries = this.queueWebSocket.queueEntries;
  readonly isConnected = this.queueWebSocket.isConnected;
  readonly wsError = this.queueWebSocket.error;
  
  // Local state
  readonly isLoading = signal<boolean>(false);
  readonly lastUpdated = signal<Date>(new Date());
  readonly recentActivity = signal<ActivityLog[]>([]);
  // Computed state
  readonly currentEntry = computed(() => {
    const entries = this.queueEntries();
    return entries.find(entry => 
      entry.status === 'CALLED' || entry.status === 'IN_SERVICE'
    ) || null;
  });
  
  readonly waitingEntries = computed(() => {
    const entries = this.queueEntries();
    return entries
      .filter(entry => entry.status === 'WAITING')
      .sort((a, b) => a.position - b.position);
  });

  readonly calledEntries = computed(() => {
    const entries = this.queueEntries();
    return entries
      .filter(entry => entry.status === 'CALLED')
      .sort((a, b) => a.position - b.position);
  });
  
  readonly estimatedWaitTime = computed(() => {
    const waiting = this.waitingEntries();
    // Simple calculation: 5 minutes per person
    return waiting.length * 5;
  });

  private currentRoomId: string | null = null;
  private currentStates: QueueEntryStatus[] | undefined = undefined;
  private lastTenantId: string | null = null;

  constructor() {
    // Update last updated timestamp when queue entries change
    effect(() => {
      const entries = this.queueEntries();
      console.log('[QueueStateService] Queue entries updated:', entries.length, 'entries');
      console.log('[QueueStateService] Entry statuses:', entries.map(e => `${e.ticketNumber}:${e.status}`).join(', '));
      this.lastUpdated.set(new Date());
    });
    
    // Watch for tenant changes and reload queue data
    effect(() => {
      const tenantId = this.tenantService.selectedTenantId();
      const roomId = this.currentRoomId;
      const states = this.currentStates;
      
      // Only reload if tenant actually changed (not just initial load)
      if (tenantId && roomId && tenantId !== this.lastTenantId && this.lastTenantId !== null) {
        console.log('Tenant changed from', this.lastTenantId, 'to', tenantId, '- reloading queue data with states:', states);
        // Clear old queue entries before loading new tenant's data
        this.queueEntries.set([]);
        this.lastTenantId = tenantId;
        // Reload queue data when tenant changes, using the same states that were originally requested
        this.queueWebSocket.initialize(roomId, states);
      } else if (tenantId && this.lastTenantId === null) {
        // Store initial tenant ID to prevent unnecessary reloads on first load
        this.lastTenantId = tenantId;
      }
    });
  }

  async initialize(roomId: string, states?: QueueEntryStatus[]): Promise<void> {
    this.currentRoomId = roomId;
    this.currentStates = states;
    // Store current tenant ID to track changes
    const currentTenantId = this.tenantService.getSelectedTenantIdSync();
    this.lastTenantId = currentTenantId;
    await this.queueWebSocket.initialize(roomId, states);
  }

  disconnect(): void {
    this.queueWebSocket.disconnect();
  }

  callNext(roomId: string, servicePointId: string): void {
    this.isLoading.set(true);
    
    this.queueApiService.callNext(roomId, servicePointId).subscribe({
      next: (response) => {
        this.addActivity(response.ticketNumber, 'Called to service');
        this.isLoading.set(false);
      },
      error: (error) => {
        console.error('Failed to call next person:', error);
        this.isLoading.set(false);
      }
    });
  }

  finishCurrent(roomId: string): void {
    this.isLoading.set(true);

    this.queueApiService.finishCurrent(roomId).subscribe({
      next: (response) => {
        this.addActivity(response.ticketNumber, 'Finished service');
        this.isLoading.set(false);
      },
      error: (error) => {
        console.error('Failed to finish current person:', error);
        this.isLoading.set(false);
      }
    });
  }

  callSpecificEntry(roomId: string, servicePointId: string, entryId: string): void {
    console.log('[QueueStateService] callSpecificEntry called:', { roomId, servicePointId, entryId });
    this.isLoading.set(true);
    
    this.queueApiService.callSpecificEntry(roomId, servicePointId, entryId).subscribe({
      next: (response) => {
        console.log('[QueueStateService] callSpecificEntry success:', response);
        this.addActivity(response.ticketNumber, 'Called to service');
        this.isLoading.set(false);
        // Note: WebSocket should update the queue entries automatically
        // But we can also manually refresh if needed
        console.log('[QueueStateService] Current queue entries after call:', this.queueEntries().length);
      },
      error: (error) => {
        console.error('[QueueStateService] Failed to call specific entry:', error);
        this.isLoading.set(false);
      }
    });
  }

  refreshQueue(roomId: string): void {
    this.isLoading.set(true);
    
    const tenantId = this.tenantService.getSelectedTenantIdSync();
    console.log(`Refreshing queue for room ${roomId}, tenant: ${tenantId || 'none'}`);
    
    this.queueApiService.getQueue(roomId).subscribe({
      next: (entries) => {
        console.log(`Queue refreshed: ${entries.length} entries for tenant ${tenantId || 'none'}`);
        // The WebSocket service will handle updating the entries
        this.queueEntries.set(entries as WebSocketQueueEntry[]);
        this.isLoading.set(false);
      },
      error: (error) => {
        console.error('Failed to refresh queue:', error);
        this.isLoading.set(false);
      }
    });
  }

  refreshWaitingEntries(roomId: string): void {
    this.isLoading.set(true);
    
    const tenantId = this.tenantService.getSelectedTenantIdSync();
    console.log(`Refreshing waiting entries for room ${roomId}, tenant: ${tenantId || 'none'}`);
    
    this.queueApiService.getQueue(roomId, ['WAITING']).subscribe({
      next: (entries) => {
        console.log(`Waiting entries refreshed: ${entries.length} entries for tenant ${tenantId || 'none'}`);
        // Update only waiting entries
        const currentEntries = this.queueEntries();
        const nonWaitingEntries = currentEntries.filter(entry => entry.status !== 'WAITING');
        const allEntries = [...nonWaitingEntries, ...entries as WebSocketQueueEntry[]];
        this.queueEntries.set(allEntries);
        this.isLoading.set(false);
      },
      error: (error) => {
        console.error('Failed to refresh waiting entries:', error);
        this.isLoading.set(false);
      }
    });
  }

  refreshCurrentEntry(roomId: string): void {
    this.isLoading.set(true);
    
    const tenantId = this.tenantService.getSelectedTenantIdSync();
    console.log(`Refreshing current entry for room ${roomId}, tenant: ${tenantId || 'none'}`);
    
    this.queueApiService.getQueue(roomId, ['IN_SERVICE']).subscribe({
      next: (entries) => {
        console.log(`Current entry refreshed: ${entries.length} entries for tenant ${tenantId || 'none'}`);
        // Update current entry (IN_SERVICE)
        const currentEntries = this.queueEntries();
        const nonCurrentEntries = currentEntries.filter(entry => entry.status !== 'IN_SERVICE');
        const allEntries = [...nonCurrentEntries, ...entries as WebSocketQueueEntry[]];
        this.queueEntries.set(allEntries);
        this.isLoading.set(false);
      },
      error: (error) => {
        console.error('Failed to refresh current entry:', error);
        this.isLoading.set(false);
      }
    });
  }

  getStatusText(status: string): string {
    const statusMap: Record<string, string> = {
      'WAITING': 'Waiting',
      'CALLED': 'Called',
      'IN_SERVICE': 'In Service',
      'COMPLETED': 'Completed',
      'CANCELLED': 'Cancelled'
    };
    return statusMap[status] || status;
  }

  getCalledEntriesForServicePoint(servicePointId: string): WebSocketQueueEntry[] {
    return this.calledEntries().filter(entry => entry.servicePoint === servicePointId);
  }

  getCurrentEntryForServicePoint(servicePointId: string): WebSocketQueueEntry | null {
    const entries = this.queueEntries();
    // Find the current entry (CALLED or IN_SERVICE) for the specific service point
    return entries.find(entry => 
      entry.servicePoint === servicePointId && 
      (entry.status === 'CALLED' || entry.status === 'IN_SERVICE')
    ) || null;
  }

  private addActivity(ticketNumber: string, action: string): void {
    const activity: ActivityLog = {
      id: Date.now().toString(),
      ticketNumber,
      action,
      timestamp: new Date()
    };
    
    this.recentActivity.update(activities => {
      const newActivities = [activity, ...activities].slice(0, 10); // Keep only last 10
      return newActivities;
    });
  }
}
