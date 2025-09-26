import { Injectable, signal } from '@angular/core';
import { Observable, Subject, BehaviorSubject } from 'rxjs';

export interface CardReaderPayload {
  deviceId: string;
  roomId: string;
  token: string;
  reader: string;
  atr: string;
  protocol: string;
  occurredAt: string;
  state?: string; // "waiting", "reading", "success", "error"
  message?: string; // Human readable message
  cardData?: {
    id_number?: string;
    first_name?: string;
    last_name?: string;
    date_of_birth?: string;
    gender?: string;
    nationality?: string;
    address?: string;
    issued_date?: string;
    expiry_date?: string;
    photo?: string;
    source?: string;
  };
}

@Injectable({
  providedIn: 'root'
})
export class WebSocketService {
  private ws: WebSocket | null = null;
  private connectionStatus = signal<'disconnected' | 'connecting' | 'connected'>('disconnected');
  private connectionStatusSubject = new BehaviorSubject<'disconnected' | 'connecting' | 'connected'>('disconnected');
  private cardDataSubject = new Subject<CardReaderPayload>();
  private stateUpdateSubject = new Subject<CardReaderPayload>();
  private healthCheckInterval: any = null;
  
  public readonly connectionStatus$ = this.connectionStatus.asReadonly();
  public readonly connectionStatusObservable$ = this.connectionStatusSubject.asObservable();
  public readonly cardData$ = this.cardDataSubject.asObservable();
  public readonly stateUpdate$ = this.stateUpdateSubject.asObservable();

  connect(url: string = 'ws://localhost:4201/ws/card-reader'): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('WebSocket already connected');
      return;
    }

    this.connectionStatus.set('connecting');
    this.connectionStatusSubject.next('connecting');
    
    try {
      this.ws = new WebSocket(url);
      
      this.ws.onopen = () => {
        console.log('WebSocket connected to:', url);
        this.connectionStatus.set('connected');
        this.connectionStatusSubject.next('connected');
        
        // Start health check
        this.startHealthCheck();
      };
      
      this.ws.onmessage = (event) => {
        try {
          let data: string;
          
          // Handle different data types
          if (typeof event.data === 'string') {
            data = event.data;
          } else if (event.data instanceof Blob) {
            // Convert Blob to text asynchronously
            const reader = new FileReader();
            reader.onload = () => {
              try {
                const textData = reader.result as string;
                console.log('Received Blob data as text:', textData);
                
                // Handle ping/pong messages
                if (textData === 'pong') {
                  console.log('Received pong from Blob, connection is healthy');
                  return;
                }
                
                const payload: CardReaderPayload = JSON.parse(textData);
                console.log('Parsed payload:', payload);
                this.routePayload(payload);
              } catch (error) {
                console.error('Failed to parse WebSocket message from Blob:', error, 'Data:', reader.result);
              }
            };
            reader.onerror = () => {
              console.error('Failed to read Blob data');
            };
            reader.readAsText(event.data);
            return; // Exit early, processing will continue in reader.onload
          } else if (event.data instanceof ArrayBuffer) {
            // Handle ArrayBuffer
            const decoder = new TextDecoder();
            data = decoder.decode(event.data);
          } else {
            console.error('Unexpected WebSocket message type:', typeof event.data, 'Constructor:', event.data?.constructor?.name);
            return;
          }
          
          console.log('Received string data:', data);
          
          // Handle ping/pong messages
          if (data === 'pong') {
            console.log('Received pong, connection is healthy');
            return;
          }
          
          const payload: CardReaderPayload = JSON.parse(data);
          console.log('Parsed payload:', payload);
          this.routePayload(payload);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error, 'Event data:', event.data);
        }
      };
      
      this.ws.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        this.connectionStatus.set('disconnected');
        this.connectionStatusSubject.next('disconnected');
        
        // Stop health check
        this.stopHealthCheck();
        
        // Clear the WebSocket reference
        this.ws = null;
        
        // Attempt to reconnect after 3 seconds
        setTimeout(() => {
          if (this.connectionStatus() === 'disconnected') {
            console.log('Attempting to reconnect WebSocket...');
            this.connect(url);
          }
        }, 3000);
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.connectionStatus.set('disconnected');
        this.connectionStatusSubject.next('disconnected');
        this.stopHealthCheck();
      };
      
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.connectionStatus.set('disconnected');
      this.connectionStatusSubject.next('disconnected');
    }
  }

  disconnect(): void {
    this.stopHealthCheck();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
      this.connectionStatus.set('disconnected');
      this.connectionStatusSubject.next('disconnected');
    }
  }

  send(message: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected');
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  private routePayload(payload: CardReaderPayload): void {
    // Route based on payload content
    if (payload.cardData) {
      // This is card data - send to card data subject
      console.log('Card data received:', payload.cardData);
      this.cardDataSubject.next(payload);
    }
    
    if (payload.state && payload.message) {
      // This is also a state update - send to state update subject
      console.log('State update:', payload.state, '-', payload.message);
      this.stateUpdateSubject.next(payload);
    }
    
    // If neither cardData nor state/message, send to both as fallback
    if (!payload.cardData && !(payload.state && payload.message)) {
      console.log('Unknown payload type, sending to both subjects');
      this.stateUpdateSubject.next(payload);
      this.cardDataSubject.next(payload);
    }
  }

  private startHealthCheck(): void {
    this.stopHealthCheck(); // Clear any existing interval
    this.healthCheckInterval = setInterval(() => {
      if (this.ws?.readyState !== WebSocket.OPEN) {
        console.log('WebSocket health check failed - connection lost, state:', this.ws?.readyState);
        this.connectionStatus.set('disconnected');
        this.connectionStatusSubject.next('disconnected');
        this.stopHealthCheck();
        this.ws = null;
      } else {
        // Send a ping to keep connection alive and detect if it's really working
        try {
          this.ws.send('ping');
        } catch (error) {
          console.log('WebSocket ping failed - connection lost');
          this.connectionStatus.set('disconnected');
          this.connectionStatusSubject.next('disconnected');
          this.stopHealthCheck();
          this.ws = null;
        }
      }
    }, 3000); // Check every 3 seconds
  }

  private stopHealthCheck(): void {
    if (this.healthCheckInterval) {
      clearInterval(this.healthCheckInterval);
      this.healthCheckInterval = null;
    }
  }
}
