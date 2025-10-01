import { Injectable, signal, computed, inject } from '@angular/core';
import { Subscription } from 'rxjs';
import { WebSocketService, CardReaderPayload } from '../../websocket.service';
import { KioskApiService, CardData, TicketResponse } from './kiosk-api.service';
import * as QRCode from 'qrcode';

export interface TicketData extends TicketResponse {
  qrCodeDataUrl?: string;
}

export interface CardReaderStatus {
  connected: boolean;
  status: string;
}

export interface DataField {
  label: string;
  value: string | null;
  type?: 'text' | 'date' | 'datetime' | 'image';
  imageAlt?: string;
}

@Injectable({
  providedIn: 'root'
})
export class CardReaderStateService {
  private readonly wsService = inject(WebSocketService);
  private readonly kioskApiService = inject(KioskApiService);
  private wsSubscription?: Subscription;
  
  // State signals
  readonly cardData = signal<CardData | null>(null);
  readonly ticketData = signal<TicketData | null>(null);
  readonly error = signal<string | null>(null);
  readonly isReading = signal<boolean>(false);
  readonly cardReaderStatus = signal<CardReaderStatus>({ connected: false, status: 'disconnected' });
  readonly wsConnectionStatus = signal<string>('disconnected');
  readonly cardReaderState = signal<string>('waiting');
  readonly cardReaderMessage = signal<string>('Please insert your ID card');
  
  // Debouncing for state changes
  private lastStateChange = 0;
  private readonly stateChangeDebounce = 1000; // 1 second debounce
  
  // Track processed card data to prevent duplicate API calls
  private processedCardData = new Set<string>();

  // Computed signals
  readonly cardDataFields = computed(() => {
    const data = this.cardData();
    if (!data) return [];

    const name = `${data.first_name || ''} ${data.last_name || ''}`.trim();
    
    const fields: DataField[] = [
      { label: 'Name', value: name || null },
      { label: 'ID Number', value: data.id_number || null },
      { label: 'Date of Birth', value: data.date_of_birth || null, type: 'date' },
      { label: 'Gender', value: data.gender || null },
      { label: 'Nationality', value: data.nationality || null },
      { label: 'Address', value: data.address || null },
      { label: 'Issued Date', value: data.issued_date || null, type: 'date' },
      { label: 'Expiry Date', value: data.expiry_date || null, type: 'date' },
      { label: 'Photo', value: data.photo || null, type: 'image', imageAlt: 'Card Photo' },
      { label: 'Source', value: data.source || null },
      { label: 'Read Time', value: data.read_time || null, type: 'datetime' }
    ];

    return fields.filter(field => field.value !== null && field.value !== '');
  });

  initialize(): void {
    this.wsService.connect();
    
    // Subscribe to WebSocket connection status
    this.wsService.connectionStatusObservable$.subscribe({
      next: (status: string) => {
        console.log('WebSocket status changed:', status);
        this.wsConnectionStatus.set(status);
      },
      error: (err) => {
        console.error('WebSocket connection status error:', err);
      }
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
  }

  disconnect(): void {
    if (this.wsSubscription) {
      this.wsSubscription.unsubscribe();
    }
    this.wsService.disconnect();
  }

  checkCardReaderStatus(): void {
    this.kioskApiService.getCardReaderStatus().subscribe({
      next: (status) => this.cardReaderStatus.set(status),
      error: (err) => {
        console.error('Failed to check card reader status:', err);
        this.cardReaderStatus.set({ connected: false, status: 'error' });
      }
    });
  }

  private handleStateUpdate(payload: CardReaderPayload): void {
    if (payload.state && payload.message) {
      const now = Date.now();

      // Allow critical state transitions to bypass debounce
      const criticalStates = ['removed', 'error'];
      const shouldDebounce = !criticalStates.includes(payload.state) && 
                            (now - this.lastStateChange < this.stateChangeDebounce);

      if (shouldDebounce) {
        console.log('State change debounced:', payload.state);
        return;
      }

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
        
        // Clear card data and ticket when card is actually removed
        setTimeout(() => {
          this.cardData.set(null);
          this.ticketData.set(null);
          // Clear processed card data to allow new cards to be processed
          this.processedCardData.clear();
        }, 3000); // 3 seconds after card removal
      } else if (payload.state === 'error') {
        this.error.set(payload.message);
      }
    }
  }

  private handleCardData(payload: CardReaderPayload): void {
    console.log('handleCardData called with payload:', payload);
    if (payload.cardData) {
      console.log('Processing card data:', payload.cardData);
      
      // Create a unique key for this card data to prevent duplicate processing
      const cardKey = `${payload.cardData.id_number}_${payload.cardData.first_name}_${payload.cardData.last_name}_${payload.occurredAt}`;
      
      // Check if we've already processed this card data
      if (this.processedCardData.has(cardKey)) {
        console.log('Card data already processed, skipping duplicate API call');
        return;
      }
      
      // Mark this card data as processed
      this.processedCardData.add(cardKey);
      
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
        source: payload.cardData.source,
        read_time: payload.occurredAt
      };
      console.log('Setting card data signal:', cardData);
      this.cardData.set(cardData);
      this.error.set(null);
      
      // Generate ticket (only once per unique card data)
      this.generateTicket(cardData);
    } else {
      console.log('No card data in payload');
    }
  }

  private generateTicket(cardData: CardData): void {
    console.log('Generating ticket for card data:', cardData);
    
    // Create a mock ID card raw data for the API
    const idCardRaw = JSON.stringify({
      id_number: cardData.id_number,
      first_name: cardData.first_name,
      last_name: cardData.last_name
    });

    console.log('Calling API with idCardRaw:', idCardRaw);

    this.kioskApiService.generateTicket('triage-1', idCardRaw).subscribe({
      next: (response) => {
        console.log('Ticket generated successfully:', response);
        this.generateQRCode(response.qrUrl).then(qrDataUrl => {
          console.log('QR code generated:', qrDataUrl);
          this.ticketData.set({
            ...response,
            qrCodeDataUrl: qrDataUrl
          });
        }).catch(qrError => {
          console.error('Failed to generate QR code:', qrError);
          this.error.set('Failed to generate QR code');
        });
      },
      error: (error) => {
        console.error('Failed to generate ticket:', error);
        console.error('Error details:', error.status, error.statusText, error.message);
        this.error.set(`Failed to generate ticket: ${error.status} ${error.statusText}`);
      }
    });
  }

  private async generateQRCode(qrUrl: string): Promise<string> {
    try {
      return await QRCode.toDataURL(qrUrl, {
        width: 200,
        margin: 2,
        color: {
          dark: '#000000',
          light: '#FFFFFF'
        }
      });
    } catch (error) {
      console.error('Failed to generate QR code:', error);
      return '';
    }
  }
}
