import { Injectable, signal, computed, inject, effect } from '@angular/core';
import { QueueWebSocketService, WebSocketQueueEntry } from 'api-client';
import { QueueApiService } from './queue-api.service';

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
  
  readonly estimatedWaitTime = computed(() => {
    const waiting = this.waitingEntries();
    // Simple calculation: 5 minutes per person
    return waiting.length * 5;
  });

  constructor() {
    // Update last updated timestamp when queue entries change
    effect(() => {
      this.queueEntries();
      this.lastUpdated.set(new Date());
    });
  }

  async initialize(roomId: string): Promise<void> {
    await this.queueWebSocket.initialize(roomId);
  }

  disconnect(): void {
    this.queueWebSocket.disconnect();
  }

  callNext(roomId: string): void {
    this.isLoading.set(true);
    
    this.queueApiService.callNext(roomId).subscribe({
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

  refreshQueue(roomId: string): void {
    this.isLoading.set(true);
    
    this.queueApiService.getQueue(roomId).subscribe({
      next: (entries) => {
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
