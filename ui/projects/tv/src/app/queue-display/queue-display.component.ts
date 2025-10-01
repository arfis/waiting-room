import { Component, signal, inject, OnInit, OnDestroy, ChangeDetectionStrategy, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { CardComponent } from 'ui';
import { QueueWebSocketService, WebSocketQueueEntry } from 'api-client';

// Using WebSocketQueueEntry from api-client

@Component({
  selector: 'app-queue-display',
  standalone: true,
  imports: [CommonModule, CardComponent],
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
    await this.queueWebSocket.initialize('triage-1');
  }

  ngOnDestroy() {
    // Disconnect from WebSocket
    this.queueWebSocket.disconnect();
  }

  private updateComputedSignals(entries: WebSocketQueueEntry[]) {
    // Find currently being served (CALLED or IN_SERVICE)
    const current = entries.find(entry => 
      entry.status === 'CALLED' || entry.status === 'IN_SERVICE'
    );
    this.currentEntry.set(current || null);

    // Get waiting entries, sorted by position
    const waiting = entries
      .filter(entry => entry.status === 'WAITING')
      .sort((a, b) => a.position - b.position);
    this.waitingEntries.set(waiting);
  }

  estimatedWaitTime(): number {
    const waiting = this.waitingEntries();
    // Simple calculation: 5 minutes per person
    return waiting.length * 5;
  }
}
