import { Component, signal, inject, OnInit, OnDestroy, ChangeDetectionStrategy, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { CardComponent } from '@waiting-room/primeng-components';
import { QueueWebSocketService, WebSocketQueueEntry } from '@waiting-room/api-client';
import { CalledTicketComponent } from './components/called-ticket/called-ticket.component';
import { WaitingTicketComponent } from './components/waiting-ticket/waiting-ticket.component';
import { CurrentEntryComponent } from './components/current-entry/current-entry.component';
import { TenantSelectorComponent } from '@lib/tenant';

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
  }

  async ngOnInit() {
    // Initialize with HTTP API first, then connect WebSocket
    // Fetch CALLED entries to show which tickets are called and where
    await this.queueWebSocket.initialize('triage-1', ['CALLED', 'IN_SERVICE', 'WAITING']);
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
    // Simple calculation: 5 minutes per person
    return waiting.length * 5;
  }

}
