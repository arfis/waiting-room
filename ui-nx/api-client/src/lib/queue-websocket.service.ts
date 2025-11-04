import { Injectable, signal, inject, effect, Injector, Optional, InjectionToken } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { QueueEntry as BaseQueueEntry, QueueEntryStatus } from './api-client/api';

// Injection token for TenantService to avoid build-time dependency
export const TENANT_SERVICE_TOKEN = new InjectionToken<any>('TenantService');

// Injection token for API URL - must be provided by each app
export const API_URL_TOKEN = new InjectionToken<string>('API_URL_TOKEN');

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
  private injector = inject(Injector);
  private apiUrl = inject(API_URL_TOKEN, { optional: true }) || 'http://localhost:8080/api';
  private _tenantService: any = null;
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private currentRoomId: string | null = null;
  private currentStates: QueueEntryStatus[] | undefined = undefined;
  private lastTenantId: string | null = null;

  // Signals for reactive state
  queueEntries = signal<WebSocketQueueEntry[]>([]);
  isConnected = signal<boolean>(false);
  error = signal<string | null>(null);

  // Helper methods to get tenant ID without direct dependency on TenantService
  // This avoids build-time TypeScript errors when processing the tenant library
  // We use injection token to get TenantService at runtime
  private getTenantId(): string {
    try {
      // Try to get from cached service
      if (this._tenantService) {
        return (this._tenantService as any).selectedTenantId?.() || '';
      }
      // Try to get from injector using injection token
      const service = this.injector.get(TENANT_SERVICE_TOKEN, null, { optional: true });
      if (service) {
        this._tenantService = service;
        return (service as any).selectedTenantId?.() || '';
      }
      return '';
    } catch (e) {
      return '';
    }
  }

  private getTenantIdSync(): string {
    try {
      // Try to get from cached service
      if (this._tenantService) {
        return (this._tenantService as any).getSelectedTenantIdSync?.() || '';
      }
      // Try to get from injector using injection token
      const service = this.injector.get(TENANT_SERVICE_TOKEN, null, { optional: true });
      if (service) {
        this._tenantService = service;
        return (service as any).getSelectedTenantIdSync?.() || '';
      }
      return '';
    } catch (e) {
      return '';
    }
  }

  constructor() {
    // Watch for tenant changes and reconnect if WebSocket is already connected
    effect(() => {
      const tenantId = this.getTenantId();
      
      // Only reconnect if tenant actually changed (not just initial load)
      if (this.currentRoomId && this.ws && this.isConnected() && tenantId && tenantId !== this.lastTenantId && this.lastTenantId !== null) {
        console.log('Tenant changed in WebSocket service from', this.lastTenantId, 'to', tenantId, '- reconnecting');
        // Clear old queue entries before loading new tenant's data
        this.queueEntries.set([]);
        this.lastTenantId = tenantId;
        // Disconnect and reconnect with new tenant
        this.disconnect();
        // Reconnect with new tenant ID using the same states
        setTimeout(() => {
          if (this.currentRoomId) {
            // Reload initial data first, then reconnect WebSocket
            this.loadInitialData(this.currentRoomId, this.currentStates).then(() => {
              // Pass the new tenant ID to connect()
              this.connect(this.currentRoomId!, tenantId);
            });
          }
        }, 100); // Small delay to ensure disconnect is complete
      } else if (tenantId && this.lastTenantId === null) {
        // Store initial tenant ID to prevent unnecessary reconnects on first load
        this.lastTenantId = tenantId;
      }
    });
  }

  async initialize(roomId: string, states?: QueueEntryStatus[]): Promise<void> {
    // Store roomId and states for reconnection
    this.currentRoomId = roomId;
    this.currentStates = states;
    
    // Check if tenant is selected before connecting
    const tenantId = this.getTenantIdSync();
    console.log('[QueueWebSocket] initialize() called');
    console.log('[QueueWebSocket]   tenantId from service:', tenantId);
    console.log('[QueueWebSocket]   tenantId truthy?', !!tenantId);
    console.log('[QueueWebSocket]   tenantId length:', tenantId ? tenantId.length : 0);
    
    if (!tenantId || tenantId.trim() === '') {
      console.warn('[QueueWebSocket] ❌ Cannot initialize WebSocket: No tenant selected');
      console.warn('[QueueWebSocket]   tenantId value:', JSON.stringify(tenantId));
      this.error.set('No tenant selected');
      return;
    }
    
    console.log('[QueueWebSocket] ✅ Tenant ID available, proceeding with initialization');
    
    // Store current tenant ID to track changes
    if (this.lastTenantId === null) {
      this.lastTenantId = tenantId;
      console.log('[QueueWebSocket] Stored initial tenant ID:', tenantId);
    }
    
    // Disconnect existing connection if any (to reconnect with new tenant)
    if (this.ws) {
      console.log('[QueueWebSocket] Disconnecting existing WebSocket connection');
      this.disconnect();
    }
    
    // Clear old queue entries before loading new data
    this.queueEntries.set([]);
    
    // First, load initial data via HTTP API (tenant-aware via interceptor)
    console.log('[QueueWebSocket] Loading initial data via HTTP...');
    await this.loadInitialData(roomId, states);
    
    // Then connect WebSocket for real-time updates (with tenantID in query)
    // Pass tenantId explicitly to connect() to ensure it's available
    console.log('[QueueWebSocket] Calling connect() with tenant ID:', tenantId);
    this.connect(roomId, tenantId);
  }

  private async loadInitialData(roomId: string, states?: QueueEntryStatus[]): Promise<void> {
    try {
      console.log('Loading initial queue data via HTTP API...');
      
      // Use relative URL to ensure it goes through the interceptor
      let url = `/api/waiting-rooms/${roomId}/queue`;
      
      // Build query parameters properly for HttpClient
      const params: { [key: string]: string | string[] } = {};
      if (states && states.length > 0) {
        params['state'] = states;
        console.log(`Loading entries for states [${states.join(', ')}]`);
      } else {
        console.log('Loading all queue entries');
      }
      
      // Get tenant ID for logging
      const tenantId = this.getTenantIdSync();
      console.log(`Loading queue data for tenant: ${tenantId || 'none'}`);
      
      const entries = await this.http.get<WebSocketQueueEntry[]>(url, { params }).toPromise();
      if (entries) {
        console.log(`Initial queue data loaded: ${entries.length} entries for tenant ${tenantId || 'none'}`);
        this.queueEntries.set(entries);
      }
    } catch (error) {
      console.error('Failed to load initial queue data:', error);
      this.error.set('Failed to load initial queue data');
    }
  }

  private connect(roomId: string, tenantIdOverride?: string): void {
    // Store roomId for reconnection
    this.currentRoomId = roomId;
    
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

    // Get tenant ID - use override if provided, otherwise get from service
    const tenantId = tenantIdOverride || this.getTenantIdSync();
    
    // Debug logging
    console.log('[QueueWebSocket] connect() called');
    console.log('[QueueWebSocket]   tenantIdOverride:', tenantIdOverride || 'none');
      console.log('[QueueWebSocket]   getSelectedTenantIdSync() returned:', this.getTenantIdSync());
    console.log('[QueueWebSocket]   Final tenantId to use:', tenantId);
    console.log('[QueueWebSocket]   tenantId type:', typeof tenantId);
    console.log('[QueueWebSocket]   tenantId length:', tenantId ? tenantId.length : 0);
    console.log('[QueueWebSocket]   tenantId truthy?', !!tenantId);
    console.log('[QueueWebSocket]   getTenantId() returned:', this.getTenantId());
    
    // Don't connect if no tenant is selected
    if (!tenantId || tenantId.trim() === '') {
      console.warn('[QueueWebSocket] ❌ Cannot connect WebSocket: No tenant selected');
      console.warn('[QueueWebSocket]   tenantId value:', JSON.stringify(tenantId));
      console.warn('[QueueWebSocket]   tenantIdOverride:', tenantIdOverride);
      this.error.set('No tenant selected');
      return;
    }
    
    // Build WebSocket URL with tenant ID as query parameter
    // Convert HTTP API URL to WebSocket URL
    // Example: http://localhost:8080/api -> ws://localhost:8080/ws
    let wsBaseUrl: string;
    try {
      const apiUrlObj = new URL(this.apiUrl);
      const wsProtocol = apiUrlObj.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsHost = apiUrlObj.host; // e.g., "localhost:8080"
      // Remove /api suffix if present and replace with /ws
      wsBaseUrl = `${wsProtocol}//${wsHost}/ws`;
    } catch (e) {
      // Fallback if URL parsing fails
      console.warn('[QueueWebSocket] Failed to parse API URL, using default:', e);
      wsBaseUrl = 'ws://localhost:8080/ws';
    }
    
    // Construct WebSocket URL with tenant ID as query parameter
    // The tenant ID format is "buildingId:sectionId"
    const trimmedTenantId = tenantId ? tenantId.trim() : '';
    
    // Log BEFORE constructing URL
    console.log('[QueueWebSocket] BEFORE URL construction:');
    console.log('[QueueWebSocket]   API URL:', this.apiUrl);
    console.log('[QueueWebSocket]   WebSocket base URL:', wsBaseUrl);
    console.log('[QueueWebSocket]   tenantId parameter:', tenantId);
    console.log('[QueueWebSocket]   trimmedTenantId:', trimmedTenantId);
    console.log('[QueueWebSocket]   trimmedTenantId !== "":', trimmedTenantId !== '');
    
    let wsUrl = `${wsBaseUrl}/queue/${roomId}`;
    if (trimmedTenantId !== '') {
      // Add tenant ID as query parameter
      const encodedTenantId = encodeURIComponent(trimmedTenantId);
      wsUrl += `?tenantId=${encodedTenantId}`;
      console.log('[QueueWebSocket] ✅ Adding tenant ID to URL query parameter');
      console.log('[QueueWebSocket]   Encoded tenant ID:', encodedTenantId);
    } else {
      console.warn('[QueueWebSocket] ❌ NOT adding tenant ID - trimmedTenantId is empty!');
      console.warn('[QueueWebSocket]   Original tenantId:', tenantId);
    }
    
    // IMPORTANT: Log the final URL - this is what's being used!
    console.log('===========================================');
    console.log('[QueueWebSocket] Final WebSocket URL:', wsUrl);
    console.log('[QueueWebSocket] Connecting to WebSocket:', wsUrl);
    console.log('[QueueWebSocket] Tenant ID:', trimmedTenantId || 'NONE');
    console.log('[QueueWebSocket] Tenant ID length:', trimmedTenantId ? trimmedTenantId.length : 0);
    console.log('[QueueWebSocket] Tenant ID encoded:', trimmedTenantId ? encodeURIComponent(trimmedTenantId) : 'none');
    console.log('[QueueWebSocket] URL includes tenant ID?', wsUrl.includes('tenantId'));
    console.log('===========================================');

    try {
      // Note: WebSocket API doesn't support custom headers directly
      // We must use query parameters or protocol subprotocols
      // The backend should extract tenantId from query parameters
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.isConnected.set(true);
        this.error.set(null);
        this.reconnectAttempts = 0;
      };

      this.ws.onmessage = (event) => {
        try {
          console.log('[QueueWebSocket] Raw WebSocket message received:', event.data);
          const data: QueueUpdate = JSON.parse(event.data);
          console.log('[QueueWebSocket] Parsed WebSocket message:', data);
          if (data.type === 'queue_update') {
            console.log('[QueueWebSocket] Queue update received:', data.entries.length, 'entries');
            console.log('[QueueWebSocket] Entries:', JSON.stringify(data.entries, null, 2));
            this.queueEntries.set(data.entries);
            console.log('[QueueWebSocket] Queue entries updated. New count:', this.queueEntries().length);
          } else {
            console.warn('[QueueWebSocket] Unknown message type:', data.type);
          }
        } catch (error) {
          console.error('[QueueWebSocket] Failed to parse WebSocket message:', error);
          console.error('[QueueWebSocket] Raw message:', event.data);
        }
      };

      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        this.isConnected.set(false);
        
        if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
          this.reconnectAttempts++;
          console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
          // Get tenant ID for reconnection
          const tenantId = this.getTenantIdSync();
          setTimeout(() => this.connect(roomId, tenantId), this.reconnectDelay * this.reconnectAttempts);
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
    // Clear current roomId to prevent reconnection attempts
    this.currentRoomId = null;
    this.currentStates = undefined;
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
