import { Injectable, signal, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { QueueEntry as BaseQueueEntry, QueueEntryStatus } from './api';

export interface WebSocketQueueEntry extends BaseQueueEntry {
  createdAt: string;
  cardData?: {
    idNumber: string;
    firstName: string;
    lastName: string;
    source: string;
  };
  serviceName?: string;
  serviceDuration?: number;
}

export interface QueueUpdate {
  type: 'queue_update';
  roomId: string;
  entries: WebSocketQueueEntry[];
}

@Injectable({
  providedIn: 'root'
})
export class QueueWebSocketService {
  private http = inject(HttpClient);
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  // Signals for reactive state
  queueEntries = signal<WebSocketQueueEntry[]>([]);
  isConnected = signal<boolean>(false);
  error = signal<string | null>(null);

  async initialize(roomId: string, states?: QueueEntryStatus[]): Promise<void> {
    // First, load initial data via HTTP API
    await this.loadInitialData(roomId, states);
    
    // Then connect WebSocket for real-time updates
    this.connect(roomId);
  }

  private async loadInitialData(roomId: string, states?: QueueEntryStatus[]): Promise<void> {
    try {
      console.log('Loading initial queue data via HTTP API...');
      
      let url = `http://localhost:8080/api/waiting-rooms/${roomId}/queue`;
      
      if (states && states.length > 0) {
        // Build query string with multiple state parameters
        const params = new URLSearchParams();
        states.forEach(state => params.append('state', state));
        url += `?${params.toString()}`;
        
        console.log(`Loading entries for states [${states.join(', ')}]`);
      } else {
        console.log('Loading all queue entries');
      }
      
      const entries = await this.http.get<WebSocketQueueEntry[]>(url).toPromise();
      if (entries) {
        console.log('Initial queue data loaded:', entries);
        this.queueEntries.set(entries);
      }
    } catch (error) {
      console.error('Failed to load initial queue data:', error);
      this.error.set('Failed to load initial queue data');
    }
  }

  connect(roomId: string): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('WebSocket already connected');
      return;
    }

    // Don't connect if we're in a browser environment that doesn't support WebSocket
    if (typeof WebSocket === 'undefined') {
      console.warn('WebSocket not supported in this environment');
      this.error.set('WebSocket not supported');
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
