import { Component, signal, inject, OnInit, OnDestroy, ChangeDetectionStrategy, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { CardComponent } from '@waiting-room/primeng-components';
import { QueueWebSocketService, WebSocketQueueEntry } from '@waiting-room/api-client';
import { CalledTicketComponent } from './components/called-ticket/called-ticket.component';
import { WaitingTicketComponent } from './components/waiting-ticket/waiting-ticket.component';
import { CurrentEntryComponent } from './components/current-entry/current-entry.component';
import { TenantSelectorComponent, TenantService } from '@lib/tenant';

// Using WebSocketQueueEntry from api-client

@Component({
  selector: 'app-queue-display',
  standalone: true,
  imports: [CommonModule, CardComponent, CalledTicketComponent, WaitingTicketComponent, CurrentEntryComponent, TenantSelectorComponent],
  templateUrl: './queue-display.component.html',
  styleUrls: ['./queue-display.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class QueueDisplayComponent implements OnInit, OnDestroy {
  private queueWebSocket = inject(QueueWebSocketService);
  private tenantService = inject(TenantService);
  private readonly roomId = 'triage-1';
  private lastTenantId: string | null = null;
  lastUpdated = signal<Date>(new Date());

  // Use WebSocket service signals directly
  queueEntries = this.queueWebSocket.queueEntries;
  isConnected = this.queueWebSocket.isConnected;
  error = this.queueWebSocket.error;

  // Computed signals using WebSocket service methods
  currentEntry = signal<WebSocketQueueEntry | null>(null);
  waitingEntries = signal<WebSocketQueueEntry[]>([]);
  calledEntries = signal<WebSocketQueueEntry[]>([]);

  constructor() {
    // Update computed signals when queue entries change
    // Use effect to watch for changes in the signal - must be in constructor
    effect(() => {
      const entries = this.queueEntries();
      this.lastUpdated.set(new Date());
      this.updateComputedSignals(entries);
    });

    // Watch for tenant changes and reload queue data
    effect(() => {
      const tenantId = this.tenantService.selectedTenantId();
      
      // Initialize on first tenant selection
      if (tenantId && this.lastTenantId === null) {
        console.log('[TV] Initializing with tenant:', tenantId);
        this.lastTenantId = tenantId;
        this.queueWebSocket.initialize(this.roomId, ['CALLED', 'IN_SERVICE', 'WAITING']);
      }
      // Reload if tenant changed
      else if (tenantId && tenantId !== this.lastTenantId && this.lastTenantId !== null) {
        console.log('[TV] Tenant changed from', this.lastTenantId, 'to', tenantId, '- reloading queue data');
        this.lastTenantId = tenantId;
        // Clear old entries and reload
        this.queueWebSocket.disconnect();
        setTimeout(() => {
          this.queueWebSocket.initialize(this.roomId, ['CALLED', 'IN_SERVICE', 'WAITING']);
        }, 100);
      }
    });
  }

  async ngOnInit() {
    // Check if tenant is selected before initializing
    const tenantId = this.tenantService.getSelectedTenantIdSync();
    if (tenantId) {
      this.lastTenantId = tenantId;
      // Initialize with HTTP API first, then connect WebSocket
      // Fetch CALLED entries to show which tickets are called and where
      await this.queueWebSocket.initialize(this.roomId, ['CALLED', 'IN_SERVICE', 'WAITING']);
    }
  }

  ngOnDestroy() {
    // Disconnect from WebSocket
    this.queueWebSocket.disconnect();
  }

  private updateComputedSignals(entries: WebSocketQueueEntry[]) {
    // Find currently being served (IN_SERVICE only for current)
    const current = entries.find(entry => 
      entry.status === 'IN_SERVICE'
    );
    this.currentEntry.set(current || null);

    // Get waiting entries, sorted by position
    const waiting = entries
      .filter(entry => entry.status === 'WAITING')
      .sort((a, b) => a.position - b.position);
    this.waitingEntries.set(waiting);

    // Get called entries, sorted by position
    const called = entries
      .filter(entry => entry.status === 'CALLED')
      .sort((a, b) => a.position - b.position);
    this.calledEntries.set(called);
  }

  estimatedWaitTime(): number {
    const waiting = this.waitingEntries();
    // Calculate based on actual service durations
    // Sum up durations of all people ahead in queue
    let totalSeconds = 0;
    for (const entry of waiting) {
      // serviceDuration is in minutes from API, convert to seconds
      const durationSeconds = (entry.serviceDuration || 5) * 60; // Default 5 minutes if not specified
      totalSeconds += durationSeconds;
    }
    // Convert back to minutes for display
    return Math.round(totalSeconds / 60);
  }

  averageServiceTime(): number {
    const waiting = this.waitingEntries();
    if (waiting.length === 0) return 0;
    
    // Calculate average service duration
    let totalMinutes = 0;
    let count = 0;
    for (const entry of waiting) {
      if (entry.serviceDuration && entry.serviceDuration > 0) {
        totalMinutes += entry.serviceDuration;
        count++;
      }
    }
    
    // If no durations specified, use default
    if (count === 0) {
      return 5; // Default 5 minutes
    }
    
    return Math.round(totalMinutes / count);
  }

}
