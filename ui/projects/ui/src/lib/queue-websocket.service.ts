import { Injectable, signal } from '@angular/core';

export interface QueueEntry {
  id: string;
  waitingRoomId: string;
  ticketNumber: string;
  status: string;
  position: number;
  createdAt: string;
  cardData?: {
    idNumber: string;
    firstName: string;
    lastName: string;
    source: string;
  };
}

export interface QueueUpdate {
  type: 'queue_update';
  roomId: string;
  entries: QueueEntry[];
}

@Injectable({
  providedIn: 'root'
})
export class QueueWebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  // Signals for reactive state
  queueEntries = signal<QueueEntry[]>([]);
  isConnected = signal<boolean>(false);
  error = signal<string | null>(null);

  connect(roomId: string): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('WebSocket already connected');
      return;
    }

    const wsUrl = `ws://localhost:8080/ws/queue/${roomId}`;
    console.log('Connecting to WebSocket:', wsUrl);

    try {
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.isConnected.set(true);
        this.error.set(null);
        this.reconnectAttempts = 0;
      };

      this.ws.onmessage = (event) => {
        try {
          const data: QueueUpdate = JSON.parse(event.data);
          if (data.type === 'queue_update') {
            console.log('Queue update received:', data.entries);
            this.queueEntries.set(data.entries);
          }
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        this.isConnected.set(false);
        
        if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
          this.reconnectAttempts++;
          console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
          setTimeout(() => this.connect(roomId), this.reconnectDelay * this.reconnectAttempts);
        } else if (this.reconnectAttempts >= this.maxReconnectAttempts) {
          this.error.set('Failed to reconnect to queue updates');
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.error.set('WebSocket connection error');
      };

    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.error.set('Failed to connect to queue updates');
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
      this.isConnected.set(false);
    }
  }

  // Computed signals for common queue operations
  getWaitingEntries() {
    return this.queueEntries().filter(entry => entry.status === 'WAITING');
  }

  getCurrentEntry() {
    return this.queueEntries().find(entry => 
      entry.status === 'CALLED' || entry.status === 'IN_SERVICE'
    );
  }

  getCompletedEntries() {
    return this.queueEntries().filter(entry => entry.status === 'COMPLETED');
  }

  getQueueStats() {
    const entries = this.queueEntries();
    return {
      total: entries.length,
      waiting: entries.filter(e => e.status === 'WAITING').length,
      called: entries.filter(e => e.status === 'CALLED').length,
      completed: entries.filter(e => e.status === 'COMPLETED').length
    };
  }
}
