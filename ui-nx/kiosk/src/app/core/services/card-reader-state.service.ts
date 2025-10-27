import { Injectable, signal, computed, inject } from '@angular/core';
import { Subscription } from 'rxjs';
import { WebSocketService, CardReaderPayload } from '../../websocket.service';
import { KioskApiService, CardData, TicketResponse } from './kiosk-api.service';
import { UserServicesService, UserService, ServiceSection } from './user-services.service';
import * as QRCode from 'qrcode';

export interface TicketData extends TicketResponse {
  qrCodeDataUrl?: string;
}

export interface DataField {
  label: string;
  value: string | number | null | undefined;
  type?: 'text' | 'date' | 'datetime' | 'image';
  imageAlt?: string;
}

export interface CardReaderStatus {
  connected: boolean;
  status: string;
}

@Injectable({
  providedIn: 'root'
})
export class CardReaderStateService {
  private readonly wsService = inject(WebSocketService);
  private readonly kioskApiService = inject(KioskApiService);
  private readonly userServicesService = inject(UserServicesService);
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
  
  // Service selection state
  readonly userServices = signal<UserService[]>([]);
  readonly serviceSections = signal<ServiceSection[]>([]);
  readonly isLoadingServices = signal<boolean>(false);
  readonly selectedService = signal<UserService | null>(null);
  readonly showServiceSelection = signal<boolean>(false);
  
  // Manual ID entry state
  readonly isManualIdSubmitting = signal<boolean>(false);
  
  // Language state
  readonly currentLanguage = signal<string>('en');
  
  // Ticket display state
  readonly ticketCountdown = signal<number>(30);
  readonly isTicketCountdownActive = signal<boolean>(false);
  private countdownInterval?: number;
  
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
      
      // Load user services instead of directly generating ticket
      this.loadUserServices(cardData);
    } else {
      console.log('No card data in payload');
    }
  }

  private loadUserServices(cardData: CardData): void {
    console.log('Loading user services for card data:', cardData);
    
    this.isLoadingServices.set(true);
    this.showServiceSelection.set(true);
    this.error.set(null);
    
    // Use ID number as identifier for external API
    const identifier = cardData.id_number;
    
    // Initialize service sections
    const sections: ServiceSection[] = [
      {
        title: 'Your Appointments',
        services: [],
        type: 'appointment',
        loading: true,
        error: null
      },
      {
        title: 'General Services',
        services: [],
        type: 'generic',
        loading: true,
        error: null
      }
    ];
    
    this.serviceSections.set(sections);
    console.log('ServiceSelectionComponent - Service sections initialized:', sections);
    
    // Load appointment services
    const appointmentLang = this.currentLanguage();
    console.log('CardReaderState: Loading appointment services with language:', appointmentLang);
    this.userServicesService.getAppointmentServices(identifier, appointmentLang).subscribe({
      next: (services) => {
        console.log('Appointment services loaded:', services);
        this.updateServiceSection('appointment', services, false, null);
      },
      error: (error) => {
        console.error('Failed to load appointment services:', error);
        this.updateServiceSection('appointment', [], false, 'Failed to load appointment services');
      }
    });
    
    // Load generic services (using a default service point for now)
    // In a real implementation, this would be determined by the room or service point
    const servicePointId = 'default-service-point';
    const genericLang = this.currentLanguage();
    console.log('CardReaderState: Loading generic services with language:', genericLang);
    this.userServicesService.getGenericServices(servicePointId, genericLang).subscribe({
      next: (services) => {
        console.log('Generic services loaded:', services);
        this.updateServiceSection('generic', services, false, null);
      },
      error: (error) => {
        console.error('Failed to load generic services:', error);
        this.updateServiceSection('generic', [], false, 'Failed to load generic services');
      }
    });
    
    // Also load the original user services as fallback
    const userLang = this.currentLanguage();
    console.log('CardReaderState: Loading user services with language:', userLang);
    this.userServicesService.getUserServices(identifier, userLang).subscribe({
      next: (services) => {
        console.log('User services loaded:', services);
        this.userServices.set(services);
        this.isLoadingServices.set(false);
        this.isManualIdSubmitting.set(false);
        
        // Check if we have any services in any section
        const allSections = this.serviceSections();
        const hasAnyServices = allSections.some(section => section.services.length > 0) || services.length > 0;
        
        if (!hasAnyServices) {
          this.error.set('No services available. Please contact support.');
        }
      },
      error: (error) => {
        console.error('Failed to load user services:', error);
        this.isLoadingServices.set(false);
        this.isManualIdSubmitting.set(false);
        
        // Only show error if no other services loaded
        const allSections = this.serviceSections();
        const hasAnyServices = allSections.some(section => section.services.length > 0);
        
        if (!hasAnyServices) {
          this.error.set('Failed to load available services. Please try again.');
        }
      }
    });
  }
  
  setLanguage(language: string): void {
    this.currentLanguage.set(language);
  }

  // Initialize language synchronization with translation service
  initializeLanguage(translationService: { getCurrentLanguage(): string }): void {
    // Sync the current language from translation service
    const currentLang = translationService.getCurrentLanguage();
    console.log('CardReaderState: Initializing language synchronization. Translation service language:', currentLang, 'CardReaderState language:', this.currentLanguage());
    if (currentLang) {
      this.currentLanguage.set(currentLang);
      console.log('CardReaderState: Language synchronized to:', currentLang);
    }
  }

  loadServices(): void {
    const cardData = this.cardData();
    if (cardData) {
      this.loadUserServices(cardData);
    }
  }

  private updateServiceSection(type: 'appointment' | 'generic', services: UserService[], loading: boolean, error: string | null): void {
    const sections = this.serviceSections();
    const updatedSections = sections.map(section => 
      section.type === type 
        ? { ...section, services, loading, error }
        : section
    );
    console.log(`ServiceSelectionComponent - Updated ${type} section:`, updatedSections);
    this.serviceSections.set(updatedSections);
    
    // Check if all sections are loaded
    const allLoaded = updatedSections.every(section => !section.loading);
    if (allLoaded) {
      this.isLoadingServices.set(false);
      this.isManualIdSubmitting.set(false);
    }
  }

  selectService(service: UserService): void {
    console.log('Service selected:', service);
    this.selectedService.set(service);
  }

  confirmServiceSelection(): void {
    const selectedService = this.selectedService();
    const cardData = this.cardData();
    
    if (selectedService && cardData) {
      console.log('Confirming service selection and generating ticket');
      this.showServiceSelection.set(false);
      this.generateTicket(cardData, selectedService.id);
    }
  }

  retryServiceLoading(): void {
    const cardData = this.cardData();
    if (cardData) {
      this.loadUserServices(cardData);
    }
  }


  submitManualId(idNumber: string): void {
    console.log('Manual ID submitted:', idNumber);
    
    this.isManualIdSubmitting.set(true);
    this.error.set(null);
    
    // Create mock card data from manual ID entry
    const mockCardData: CardData = {
      id_number: idNumber,
      first_name: '', // Will be filled by external API if available
      last_name: '',
      date_of_birth: '',
      gender: '',
      nationality: '',
      address: '',
      issued_date: '',
      expiry_date: '',
      source: 'manual-entry',
      read_time: new Date().toISOString()
    };
    
    // Set the card data and load user services
    this.cardData.set(mockCardData);
    this.loadUserServices(mockCardData);
  }

  private generateTicket(cardData: CardData, serviceId?: string): void {
    console.log('Generating ticket for card data:', cardData);
    
    // Create a mock ID card raw data for the API
    const idCardRaw = JSON.stringify({
      id_number: cardData.id_number,
      first_name: cardData.first_name,
      last_name: cardData.last_name
    });

    // Get service duration from selected service
    const selectedService = this.selectedService();
    const serviceDuration = selectedService?.duration;

    console.log('Calling API with idCardRaw:', idCardRaw, 'serviceId:', serviceId, 'serviceDuration:', serviceDuration);

    this.kioskApiService.generateTicket('triage-1', idCardRaw, serviceId, serviceDuration).subscribe({
      next: (response) => {
        console.log('Ticket generated successfully:', response);
        this.generateQRCode(response.qrUrl).then(qrDataUrl => {
          console.log('QR code generated:', qrDataUrl);
          this.ticketData.set({
            ...response,
            qrCodeDataUrl: qrDataUrl
          });
          // Start countdown timer after ticket is generated
          this.startTicketCountdown();
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

  startTicketCountdown(): void {
    this.ticketCountdown.set(30);
    this.isTicketCountdownActive.set(true);
    
    this.countdownInterval = window.setInterval(() => {
      const current = this.ticketCountdown();
      if (current <= 1) {
        this.resetToMainInterface();
      } else {
        this.ticketCountdown.set(current - 1);
      }
    }, 1000);
  }

  stopTicketCountdown(): void {
    if (this.countdownInterval) {
      clearInterval(this.countdownInterval);
      this.countdownInterval = undefined;
    }
    this.isTicketCountdownActive.set(false);
  }

  resetToMainInterface(): void {
    console.log('Resetting to main interface');
    this.stopTicketCountdown();
    this.ticketData.set(null);
    this.cardData.set(null);
    this.selectedService.set(null);
    this.showServiceSelection.set(false);
    this.userServices.set([]);
    this.error.set(null);
    this.isManualIdSubmitting.set(false); // Clear manual ID submission loading state
    this.cardReaderState.set('waiting');
    this.cardReaderMessage.set('Please insert your ID card');
  }

  returnToMainInterface(): void {
    this.resetToMainInterface();
  }
}
