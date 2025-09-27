import { Component, signal, inject, OnInit, OnDestroy, ChangeDetectionStrategy, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { CardComponent, DataGridComponent, DataField } from 'ui';
import { WebSocketService, CardReaderPayload } from '../websocket.service';
import { Subscription } from 'rxjs';
import * as QRCode from 'qrcode';

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
  source?: string;
  read_time: string;
}

interface CardReaderStatus {
  connected: boolean;
  status: string;
}

interface TicketData {
  entryId: string;
  ticketNumber: string;
  qrUrl: string;
  qrCodeDataUrl?: string;
}

@Component({
  selector: 'card-loader',
  standalone: true,
  imports: [CommonModule, CardComponent, DataGridComponent],
  templateUrl: './card-loader.component.html',
  styleUrls: ['./card-loader.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class CardLoaderPageComponent implements OnInit, OnDestroy {
  private http = inject(HttpClient);
  private wsService = inject(WebSocketService);
  private wsSubscription?: Subscription;

  // State signals
  cardData = signal<CardData | null>(null);
  ticketData = signal<TicketData | null>(null);
  error = signal<string | null>(null);
  isReading = signal<boolean>(false);
  cardReaderStatus = signal<CardReaderStatus>({ connected: false, status: 'disconnected' });
  wsConnectionStatus = signal<string>('disconnected');
  cardReaderState = signal<string>('waiting');
  cardReaderMessage = signal<string>('Please insert your ID card');

  // Debouncing for state changes
  private lastStateChange = 0;
  private stateChangeDebounce = 1000; // 1 second debounce

  // Computed signals for better performance
  cardDataFields = computed(() => {
    const data = this.cardData();
    if (!data) return [];

    const name = `${data.first_name || ''} ${data.last_name || ''}`.trim();
    
    return [
      { label: 'Name', value: name || null },
      { label: 'ID Number', value: data.id_number || null },
      { label: 'Date of Birth', value: data.date_of_birth || null, type: 'date' as const },
      { label: 'Gender', value: data.gender || null },
      { label: 'Nationality', value: data.nationality || null },
      { label: 'Address', value: data.address || null },
      { label: 'Issued Date', value: data.issued_date || null, type: 'date' as const },
      { label: 'Expiry Date', value: data.expiry_date || null, type: 'date' as const },
      { label: 'Photo', value: data.photo || null, type: 'image' as const, imageAlt: 'Card Photo' },
      { label: 'Source', value: data.source || null },
      { label: 'Read Time', value: data.read_time || null, type: 'datetime' as const }
    ].filter(field => field.value !== null && field.value !== '');
  });

  ngOnInit() {
    // Connect to WebSocket
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

    // Note: We're using WebSocket for card reading, not HTTP API
  }

  ngOnDestroy() {
    if (this.wsSubscription) {
      this.wsSubscription.unsubscribe();
    }
    this.wsService.disconnect();
  }

  handleStateUpdate(payload: CardReaderPayload) {
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
        }, 3000); // 3 seconds after card removal
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
        source: payload.cardData.source,
        read_time: payload.occurredAt
      };
      console.log('Setting card data signal:', cardData);
      this.cardData.set(cardData);
      this.error.set(null);
      
      // Generate ticket
      this.generateTicket(cardData);
    } else {
      console.log('No card data in payload');
    }
  }

  private generateTicket(cardData: CardData) {
    console.log('Generating ticket for card data:', cardData);
    
    // Create a mock ID card raw data for the API
    const idCardRaw = JSON.stringify({
      id_number: cardData.id_number,
      first_name: cardData.first_name,
      last_name: cardData.last_name
    });

    console.log('Calling API with idCardRaw:', idCardRaw);

    // Call the API to create a ticket
    this.http.post<any>('http://localhost:8080/waiting-rooms/triage-1/swipe', {
      idCardRaw: idCardRaw
    }).subscribe({
      next: (response) => {
        console.log('Ticket generated successfully:', response);
        this.generateQRCode(response.qrUrl).then(qrDataUrl => {
          console.log('QR code generated:', qrDataUrl);
          this.ticketData.set({
            entryId: response.entryId,
            ticketNumber: response.ticketNumber,
            qrUrl: response.qrUrl,
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
    if (this.isReading()) return;
    
    this.isReading.set(true);
    this.error.set(null);
    
    // Simulate card reading (this would normally call the API)
    setTimeout(() => {
      this.isReading.set(false);
      this.error.set('Manual card reading not implemented. Use the card reader app.');
    }, 2000);
  }
}