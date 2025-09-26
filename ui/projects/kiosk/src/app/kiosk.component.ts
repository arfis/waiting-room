import { Component, signal, inject, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { CardComponent } from 'ui';
import { WebSocketService, CardReaderPayload } from './websocket.service';
import { Subscription } from 'rxjs';

interface CardData {
  id_number: string;
  first_name: string;
  last_name: string;
  date_of_birth: string;
  gender: string;
  nationality: string;
  address: string;
  issued_date: string;
  expiry_date: string;
  photo?: string;
  read_time: string;
}

interface CardReaderStatus {
  connected: boolean;
  status: string;
}

@Component({
  selector: 'app-kiosk',
  standalone: true,
  imports: [CommonModule, CardComponent],
  template: `
    <div class="min-h-screen bg-slate-50 p-6">
      <ui-card>
        <h1 class="text-3xl font-bold mb-8 text-center">Welcome to Waiting Room</h1>
        
        <!-- Card Reader Status -->
        <div class="mb-6 p-4 rounded-lg" [class]="wsConnectionStatus() === 'connected' ? 'bg-green-100 border border-green-300' : 'bg-yellow-100 border border-yellow-300'">
          <div class="flex items-center gap-2">
            <div class="w-3 h-3 rounded-full" [class]="wsConnectionStatus() === 'connected' ? 'bg-green-500' : 'bg-yellow-500'"></div>
            <span class="font-medium">
              Card Reader: {{ wsConnectionStatus() === 'connected' ? 'Ready' : 'Waiting for Connection' }}
            </span>
          </div>
          <div class="text-sm text-gray-600 mt-1">
            {{ wsConnectionStatus() === 'connected' ? 'Standalone card reader app is connected' : 'Start the card reader app to begin' }}
          </div>
        </div>

        <!-- WebSocket Status -->
        <div class="mb-6 p-4 rounded-lg" [class]="wsConnectionStatus() === 'connected' ? 'bg-blue-100 border border-blue-300' : wsConnectionStatus() === 'connecting' ? 'bg-yellow-100 border border-yellow-300' : 'bg-red-100 border border-red-300'">
          <div class="flex items-center gap-2">
            <div class="w-3 h-3 rounded-full" [class]="wsConnectionStatus() === 'connected' ? 'bg-blue-500' : wsConnectionStatus() === 'connecting' ? 'bg-yellow-500' : 'bg-red-500'"></div>
            <span class="font-medium">
              WebSocket: {{ wsConnectionStatus() === 'connected' ? 'Connected' : wsConnectionStatus() === 'connecting' ? 'Connecting...' : 'Disconnected' }}
            </span>
          </div>
        </div>

        <!-- Card Reading Section -->
        <div class="text-center mb-8">
          <div class="mb-4">
            <!-- Different icons based on state -->
            <svg *ngIf="cardReaderState() === 'waiting'" class="mx-auto w-16 h-16 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path>
            </svg>
            <svg *ngIf="cardReaderState() === 'reading'" class="mx-auto w-16 h-16 text-yellow-500 animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path>
            </svg>
            <svg *ngIf="cardReaderState() === 'success'" class="mx-auto w-16 h-16 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <svg *ngIf="cardReaderState() === 'error'" class="mx-auto w-16 h-16 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <svg *ngIf="cardReaderState() === 'removed'" class="mx-auto w-16 h-16 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path>
            </svg>
            <svg *ngIf="wsConnectionStatus() === 'disconnected'" class="mx-auto w-16 h-16 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path>
            </svg>
          </div>
          <h2 class="text-xl font-semibold mb-2">{{ cardReaderMessage() }}</h2>
          <p class="text-slate-600 mb-6">
            <span *ngIf="wsConnectionStatus() === 'connected'">
              <span *ngIf="cardReaderState() === 'waiting'">Card reader is ready. Insert your card to automatically read it.</span>
              <span *ngIf="cardReaderState() === 'reading'">Reading your card data, please wait...</span>
              <span *ngIf="cardReaderState() === 'success'">Card read successfully! Please remove your card.</span>
              <span *ngIf="cardReaderState() === 'removed'">Card removed successfully. Ready for next card.</span>
              <span *ngIf="cardReaderState() === 'error'">There was an error reading your card. Please try again.</span>
            </span>
            <span *ngIf="wsConnectionStatus() === 'connecting'">Connecting to card reader...</span>
            <span *ngIf="wsConnectionStatus() === 'disconnected'">
              <strong>To start card reading:</strong><br>
              1. Open terminal and run: <code class="bg-gray-200 px-2 py-1 rounded">cd card-reader && go run main.go</code><br>
              2. Insert your ID card when prompted
            </span>
          </p>
          
          <!-- Fallback manual read button -->
          <button 
            *ngIf="wsConnectionStatus() !== 'connected'"
            class="px-8 py-4 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            (click)="readCard()"
            [disabled]="!cardReaderStatus().connected || isReading()"
          >
            <span *ngIf="!isReading()">Read Card Manually</span>
            <span *ngIf="isReading()" class="flex items-center gap-2">
              <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Reading Card...
            </span>
          </button>
        </div>

        <!-- Card Data Display -->
        <div *ngIf="cardData()" class="mt-8 p-6 bg-slate-100 rounded-lg">
          <!-- Debug info -->
          <div class="text-xs text-gray-500 mb-2">Debug: cardData() = {{ cardData() | json }}</div>
          <h3 class="text-lg font-semibold mb-4">Card Information</h3>
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span class="font-medium">Name:</span>
              <span class="ml-2">{{ cardData()?.first_name }} {{ cardData()?.last_name }}</span>
            </div>
            <div>
              <span class="font-medium">ID Number:</span>
              <span class="ml-2">{{ cardData()?.id_number }}</span>
            </div>
            <div>
              <span class="font-medium">Date of Birth:</span>
              <span class="ml-2">{{ formatDate(cardData()?.date_of_birth) }}</span>
            </div>
            <div>
              <span class="font-medium">Gender:</span>
              <span class="ml-2">{{ cardData()?.gender }}</span>
            </div>
            <div>
              <span class="font-medium">Nationality:</span>
              <span class="ml-2">{{ cardData()?.nationality }}</span>
            </div>
            <div>
              <span class="font-medium">Address:</span>
              <span class="ml-2">{{ cardData()?.address }}</span>
            </div>
          </div>
          
          <div class="mt-6 flex gap-4">
            <button 
              class="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
              (click)="joinQueue()"
            >
              Join Queue
            </button>
            <button 
              class="px-6 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors"
              (click)="clearCard()"
            >
              Clear
            </button>
          </div>
        </div>

        <!-- Error Display -->
        <div *ngIf="error()" class="mt-6 p-4 bg-red-100 border border-red-300 rounded-lg">
          <div class="flex items-center gap-2 text-red-700">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <span class="font-medium">Error:</span>
            <span>{{ error() }}</span>
          </div>
        </div>
      </ui-card>
    </div>
  `
})
export class KioskComponent implements OnInit, OnDestroy {
  private http = inject(HttpClient);
  private wsService = inject(WebSocketService);
  private wsSubscription?: Subscription;
  
  cardReaderStatus = signal<CardReaderStatus>({ connected: false, status: 'unknown' });
  cardData = signal<CardData | null>(null);
  isReading = signal(false);
  error = signal<string | null>(null);
  wsConnectionStatus = signal<'disconnected' | 'connecting' | 'connected'>('disconnected');
  cardReaderState = signal<string>('waiting');
  cardReaderMessage = signal<string>('Please insert your ID card');

  ngOnInit() {
    // Connect to WebSocket
    this.wsService.connect();
    
    // Subscribe to WebSocket connection status changes
    this.wsService.connectionStatusObservable$.subscribe((status: 'disconnected' | 'connecting' | 'connected') => {
      console.log('WebSocket status changed to:', status);
      this.wsConnectionStatus.set(status);
    });
    
    // Subscribe to state updates from WebSocket
    this.wsService.stateUpdate$.subscribe({
      next: (payload: CardReaderPayload) => {
        this.handleStateUpdate(payload);
      },
      error: (err) => {
        console.error('WebSocket state update error:', err);
      }
    });

    // Subscribe to card data from WebSocket
    this.wsSubscription = this.wsService.cardData$.subscribe({
      next: (payload: CardReaderPayload) => {
        this.handleCardData(payload);
      },
      error: (err) => {
        console.error('WebSocket card data error:', err);
        this.error.set('Failed to receive card data');
      }
    });

    // Note: We're using WebSocket for card reading, not HTTP API
  }

  ngOnDestroy() {
    if (this.wsSubscription) {
      this.wsSubscription.unsubscribe();
    }
    this.wsService.disconnect();
  }

  private lastStateChange = 0;
  private stateChangeDebounce = 1000; // 1 second debounce

  handleStateUpdate(payload: CardReaderPayload) {
    if (payload.state && payload.message) {
      const now = Date.now();
      
      // Debounce rapid state changes to prevent flickering
      if (now - this.lastStateChange < this.stateChangeDebounce) {
        console.log('State change debounced:', payload.state);
        return;
      }
      
      console.log('State update received:', payload.state, '-', payload.message);
      this.lastStateChange = now;
      this.cardReaderState.set(payload.state);
      this.cardReaderMessage.set(payload.message);
      
      // Update reading state based on card reader state
      this.isReading.set(payload.state === 'reading');
      
      // Clear error when we get a successful state
      if (payload.state === 'success') {
        this.error.set(null);
      } else if (payload.state === 'removed') {
        this.error.set(null);
        // After a short delay, go back to waiting state
        setTimeout(() => {
          this.cardReaderState.set('waiting');
          this.cardReaderMessage.set('Please insert your ID card');
        }, 2000);
        
        // Clear card data after a longer delay so user can see it
        setTimeout(() => {
          this.cardData.set(null);
        }, 10000); // 10 seconds
      } else if (payload.state === 'error') {
        this.error.set(payload.message);
      }
    }
  }

  handleCardData(payload: CardReaderPayload) {
    console.log('handleCardData called with payload:', payload);
    if (payload.cardData) {
      console.log('Processing card data:', payload.cardData);
      const cardData: CardData = {
        id_number: payload.cardData.id_number || '',
        first_name: payload.cardData.first_name || '',
        last_name: payload.cardData.last_name || '',
        date_of_birth: payload.cardData.date_of_birth || '',
        gender: payload.cardData.gender || '',
        nationality: payload.cardData.nationality || '',
        address: payload.cardData.address || '',
        issued_date: payload.cardData.issued_date || '',
        expiry_date: payload.cardData.expiry_date || '',
        photo: payload.cardData.photo,
        read_time: payload.occurredAt
      };
      console.log('Setting card data signal:', cardData);
      this.cardData.set(cardData);
      this.error.set(null);
    } else {
      console.log('No card data in payload');
    }
  }

  checkCardReaderStatus() {
    this.http.get<CardReaderStatus>('http://localhost:8080/api/card-reader/status')
      .subscribe({
        next: (status) => this.cardReaderStatus.set(status),
        error: (err) => {
          console.error('Failed to check card reader status:', err);
          this.cardReaderStatus.set({ connected: false, status: 'error' });
        }
      });
  }

  readCard() {
    if (!this.cardReaderStatus().connected) {
      this.error.set('Card reader is not connected');
      return;
    }

    this.isReading.set(true);
    this.error.set(null);
    this.cardData.set(null);

    this.http.post<CardData>('http://localhost:8080/api/card-reader/read', {})
      .subscribe({
        next: (data) => {
          this.cardData.set(data);
          this.isReading.set(false);
        },
        error: (err) => {
          this.error.set(err.error?.message || 'Failed to read card');
          this.isReading.set(false);
        }
      });
  }

  joinQueue() {
    // TODO: Implement queue joining logic
    console.log('Joining queue with card data:', this.cardData());
    // This would typically call your existing API to create a queue entry
  }

  clearCard() {
    this.cardData.set(null);
    this.error.set(null);
  }

  formatDate(dateString: string | undefined): string {
    if (!dateString) return '';
    try {
      return new Date(dateString).toLocaleDateString();
    } catch {
      return dateString;
    }
  }
}
