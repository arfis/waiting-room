import { Component, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { ConfigService } from '../shared/services/config.service';
import { TranslationService, TranslatePipe } from '../../../../src/lib/i18n';

interface GenericService {
  id: string;
  name: string;
  description?: string;
  duration?: number;
  enabled: boolean;
}

interface ExternalAPIConfig {
  appointmentServicesUrl?: string;
  appointmentServicesHttpMethod?: 'GET' | 'POST';
  genericServicesUrl?: string;
  genericServicesHttpMethod?: 'GET' | 'POST';
  genericServices?: GenericService[];
  webhookUrl?: string;
  webhookHttpMethod?: 'GET' | 'POST';
  webhookTimeoutSeconds?: number;
  webhookRetryAttempts?: number;
  timeoutSeconds: number;
  retryAttempts: number;
  headers?: { [key: string]: string };
  // Multilingual configuration
  multilingualSupport?: boolean;
  supportedLanguages?: string[];
  useDeepLTranslation?: boolean;
  // Appointment services language handling
  appointmentServicesLanguageHandling?: 'query_param' | 'header' | 'none';
  appointmentServicesLanguageHeader?: string;
  // Generic services language handling
  genericServicesLanguageHandling?: 'query_param' | 'header' | 'none';
  genericServicesLanguageHeader?: string;
}

interface HeaderEntry {
  key: string;
  value: string;
}

interface ServicePoint {
  id: string;
  name: string;
  description?: string;
  managerId?: string;
  managerName?: string;
}

interface Room {
  id: string;
  name: string;
  description?: string;
  servicePoints: ServicePoint[];
  isDefault: boolean;
}

interface SystemConfiguration {
  id?: string;
  externalAPI: ExternalAPIConfig;
  rooms: Room[];
  defaultRoom: string;
  webSocketPath: string;
  allowWildcard: boolean;
  createdAt?: string;
  updatedAt?: string;
}

@Component({
  selector: 'app-configuration',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslatePipe],
  templateUrl: './configuration.html',
  styleUrl: './configuration.scss'
})
export class ConfigurationComponent {
  constructor(
    private http: HttpClient,
    private configService: ConfigService,
    private translationService: TranslationService
  ) {
    this.loadConfiguration();
  }
  
  isSaving = signal(false);
  lastUpdated = signal('Just now');
  configurationCount = signal(5);
  
  // Language selection for multilingual API
  supportedLanguages = {
    en: true,
    sk: false
  };

  externalAPIConfig: ExternalAPIConfig = {
    timeoutSeconds: 30,
    retryAttempts: 3,
    headers: {},
    multilingualSupport: false,
    supportedLanguages: ['en'],
    useDeepLTranslation: false,
    appointmentServicesLanguageHandling: 'query_param',
    appointmentServicesLanguageHeader: 'Accept-Language',
    appointmentServicesHttpMethod: 'GET',
    genericServicesLanguageHandling: 'query_param',
    genericServicesLanguageHeader: 'Accept-Language',
    genericServicesHttpMethod: 'GET',
    webhookHttpMethod: 'POST',
    genericServices: []
  };
  
  systemConfig: SystemConfiguration = {
    externalAPI: {
      appointmentServicesUrl: 'http://localhost:3001/users/${identifier}/appointments',
      genericServicesUrl: 'http://localhost:3001/service-points/${servicePointId}/services',
      webhookUrl: 'http://localhost:3001/webhooks/ticket-state-changed',
      webhookTimeoutSeconds: 5,
      webhookRetryAttempts: 2,
      timeoutSeconds: 10,
      retryAttempts: 3,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      multilingualSupport: false,
      supportedLanguages: ['en'],
      useDeepLTranslation: false,
      appointmentServicesLanguageHandling: 'query_param',
      appointmentServicesLanguageHeader: 'Accept-Language',
      genericServicesLanguageHandling: 'query_param',
      genericServicesLanguageHeader: 'Accept-Language'
    },
    rooms: [
      {
        id: 'triage-1',
        name: 'Triage Room 1',
        description: 'Main triage room for patient assessment',
        isDefault: true,
        servicePoints: [
          {
            id: 'sp-1',
            name: 'Service Point 1',
            description: 'Main service point',
            managerName: 'Dr. Smith'
          },
          {
            id: 'sp-2',
            name: 'Service Point 2',
            description: 'Secondary service point'
          }
        ]
      },
      {
        id: 'triage-2',
        name: 'Triage Room 2',
        description: 'Secondary triage room',
        isDefault: false,
        servicePoints: [
          {
            id: 'sp-3',
            name: 'Service Point 3',
            description: 'Emergency service point'
          }
        ]
      }
    ],
    defaultRoom: 'triage-1',
    webSocketPath: '/ws/queue',
    allowWildcard: true
  };

  get rooms() {
    return this.systemConfig.rooms;
  }

  get headers() {
    return this.systemConfig.externalAPI.headers || {};
  }

  // Signal for header entries to ensure proper reactivity
  headerEntries = signal<HeaderEntry[]>([]);

  // Update header entries from the headers object
  private updateHeaderEntries(): void {
    const headers = this.headers;
    const entries = Object.keys(headers).map(key => ({
      key: key,
      value: headers[key]
    }));
    this.headerEntries.set(entries);
  }

  // Sync header entries back to the headers object when they change
  syncHeadersFromEntries(): void {
    if (!this.systemConfig.externalAPI.headers) {
      this.systemConfig.externalAPI.headers = {};
    }
    
    // Clear existing headers
    this.systemConfig.externalAPI.headers = {};
    
    // Add all entries back to headers object
    this.headerEntries().forEach(entry => {
      if (entry.key.trim() !== '') {
        this.systemConfig.externalAPI.headers![entry.key] = entry.value;
      }
    });
  }

  loadConfiguration(): void {
    // Load external API configuration
    this.http.get(this.configService.adminExternalApiUrl)
      .subscribe({
        next: (response: any) => {
          if (response) {
            this.systemConfig.externalAPI = response;
            // Sync with externalAPIConfig for form binding, ensuring all fields are present
            this.externalAPIConfig = {
              timeoutSeconds: response.timeoutSeconds || 30,
              retryAttempts: response.retryAttempts || 3,
              headers: response.headers || {},
              multilingualSupport: response.multilingualSupport || false,
              supportedLanguages: response.supportedLanguages || ['en'],
              useDeepLTranslation: response.useDeepLTranslation || false,
              appointmentServicesLanguageHandling: response.appointmentServicesLanguageHandling || 'query_param',
              appointmentServicesLanguageHeader: response.appointmentServicesLanguageHeader || 'Accept-Language',
              appointmentServicesHttpMethod: response.appointmentServicesHttpMethod || 'GET',
              genericServicesLanguageHandling: response.genericServicesLanguageHandling || 'query_param',
              genericServicesLanguageHeader: response.genericServicesLanguageHeader || 'Accept-Language',
              genericServicesHttpMethod: response.genericServicesHttpMethod || 'GET',
              webhookHttpMethod: response.webhookHttpMethod || 'POST',
              appointmentServicesUrl: response.appointmentServicesUrl,
              genericServicesUrl: response.genericServicesUrl,
              genericServices: response.genericServices || [],
              webhookUrl: response.webhookUrl,
              webhookTimeoutSeconds: response.webhookTimeoutSeconds,
              webhookRetryAttempts: response.webhookRetryAttempts
            };
            this.updateHeaderEntries();
            console.log('Loaded external API configuration:', response);
            console.log('Synced externalAPIConfig:', this.externalAPIConfig);
          }
        },
        error: (error) => {
          console.error('Failed to load external API configuration:', error);
        }
      });

    // Load system configuration
    this.http.get(this.configService.adminConfigUrl)
      .subscribe({
        next: (response: any) => {
          if (response) {
            // Update all system configuration fields
            this.systemConfig.id = response.id;
            this.systemConfig.defaultRoom = response.defaultRoom || this.systemConfig.defaultRoom;
            this.systemConfig.webSocketPath = response.webSocketPath || this.systemConfig.webSocketPath;
            this.systemConfig.allowWildcard = response.allowWildcard !== undefined ? response.allowWildcard : this.systemConfig.allowWildcard;
            this.systemConfig.createdAt = response.createdAt;
            this.systemConfig.updatedAt = response.updatedAt;
            console.log('Loaded system configuration:', response);
          }
        },
        error: (error) => {
          console.error('Failed to load system configuration:', error);
        }
      });

    // Load rooms configuration
    this.http.get(this.configService.adminRoomsUrl)
      .subscribe({
        next: (response: any) => {
          if (response && Array.isArray(response)) {
            this.systemConfig.rooms = response;
            console.log('Loaded rooms configuration:', response);
          }
        },
        error: (error) => {
          console.error('Failed to load rooms configuration:', error);
        }
      });
  }

  saveExternalAPIConfig(): void {
    // Validate appointment URL if provided
    if (this.externalAPIConfig.appointmentServicesUrl && !this.externalAPIConfig.appointmentServicesUrl.includes('${identifier}')) {
      alert('Appointment URL must contain ${identifier} placeholder. Example: http://api.example.com/users/${identifier}/appointments');
      return;
    }
    
    // Validate generic URL if provided
    if (this.externalAPIConfig.genericServicesUrl && !this.externalAPIConfig.genericServicesUrl.includes('${servicePointId}')) {
      alert('Generic URL must contain ${servicePointId} placeholder. Example: http://api.example.com/service-points/${servicePointId}/services');
      return;
    }

    // Sync headers from entries before saving
    this.syncHeadersFromEntries();

    this.isSaving.set(true);
    
    // Convert headers from array format to flat object as required by API
    const headersObject: { [key: string]: string } = {};
    this.headerEntries().forEach(entry => {
      if (entry.key.trim() !== '') {
        headersObject[entry.key] = entry.value;
      }
    });
    
      const config = {
        appointmentServicesUrl: this.externalAPIConfig.appointmentServicesUrl,
        appointmentServicesHttpMethod: this.externalAPIConfig.appointmentServicesHttpMethod || 'GET',
        genericServicesUrl: this.externalAPIConfig.genericServicesUrl,
        genericServicesHttpMethod: this.externalAPIConfig.genericServicesHttpMethod || 'GET',
        genericServices: this.externalAPIConfig.genericServices || [],
        webhookUrl: this.externalAPIConfig.webhookUrl,
        webhookHttpMethod: this.externalAPIConfig.webhookHttpMethod || 'POST',
        webhookTimeoutSeconds: this.externalAPIConfig.webhookTimeoutSeconds,
        webhookRetryAttempts: this.externalAPIConfig.webhookRetryAttempts,
        timeoutSeconds: this.externalAPIConfig.timeoutSeconds,
        retryAttempts: this.externalAPIConfig.retryAttempts,
        headers: headersObject,
        multilingualSupport: this.externalAPIConfig.multilingualSupport || false,
        supportedLanguages: this.externalAPIConfig.supportedLanguages || ['en'],
        useDeepLTranslation: this.externalAPIConfig.useDeepLTranslation || false,
        appointmentServicesLanguageHandling: this.externalAPIConfig.appointmentServicesLanguageHandling || 'query_param',
        appointmentServicesLanguageHeader: this.externalAPIConfig.appointmentServicesLanguageHeader || 'Accept-Language',
        genericServicesLanguageHandling: this.externalAPIConfig.genericServicesLanguageHandling || 'query_param',
        genericServicesLanguageHeader: this.externalAPIConfig.genericServicesLanguageHeader || 'Accept-Language'
      };
    
    this.http.put(this.configService.adminExternalApiUrl, config)
      .subscribe({
        next: (response) => {
          this.isSaving.set(false);
          this.lastUpdated.set(new Date().toLocaleString());
          console.log('External API configuration saved:', response);
          alert('External API configuration saved successfully!');
        },
        error: (error) => {
          this.isSaving.set(false);
          console.error('Failed to save external API configuration:', error);
          const errorMessage = error.error?.message || error.message || 'Unknown error';
          alert(`Failed to save external API configuration: ${errorMessage}`);
        }
      });
  }

  saveSystemConfiguration(): void {
    this.isSaving.set(true);
    this.updateHeaderEntries()

    // Send complete SystemConfiguration object as required by API
    const config = {
      id: this.systemConfig.id,
      externalAPI: this.systemConfig.externalAPI,
      rooms: this.systemConfig.rooms,
      defaultRoom: this.systemConfig.defaultRoom,
      webSocketPath: this.systemConfig.webSocketPath,
      allowWildcard: this.systemConfig.allowWildcard,
      createdAt: this.systemConfig.createdAt,
      updatedAt: this.systemConfig.updatedAt
    };
    
    this.http.put(this.configService.adminConfigUrl, config)
      .subscribe({
        next: (response) => {
          this.isSaving.set(false);
          this.lastUpdated.set(new Date().toLocaleString());
        },
        error: (error) => {
          this.isSaving.set(false);
        }
      });
  }

  addRoom(): void {
    const newRoom: Room = {
      id: `room-${Date.now()}`,
      name: 'New Room',
      description: '',
      isDefault: false,
      servicePoints: []
    };
    
    this.systemConfig.rooms.push(newRoom);
  }

  editRoom(room: Room): void {
    // In a real implementation, this would open an edit dialog
    console.log('Edit room:', room);
  }

  deleteRoom(roomId: string): void {
    if (confirm('Are you sure you want to delete this room?')) {
      this.systemConfig.rooms = this.systemConfig.rooms.filter(room => room.id !== roomId);
      
      // If we deleted the default room, set a new default
      if (this.systemConfig.defaultRoom === roomId && this.systemConfig.rooms.length > 0) {
        this.systemConfig.defaultRoom = this.systemConfig.rooms[0].id;
      }
    }
  }

  setDefaultRoom(roomId: string): void {
    this.systemConfig.defaultRoom = roomId;
    
    // Update isDefault flags
    this.systemConfig.rooms.forEach(room => {
      room.isDefault = room.id === roomId;
    });
  }

  addHeader(): void {
    if (!this.systemConfig.externalAPI.headers) {
      this.systemConfig.externalAPI.headers = {};
    }

    this.systemConfig.externalAPI.headers['header_'+Object.keys(this.systemConfig.externalAPI.headers).length] = 'value_'+Object.keys(this.systemConfig.externalAPI.headers).length;
    this.updateHeaderEntries();
  }

  removeHeader(key: string): void {
    if (this.systemConfig.externalAPI.headers) {
      delete this.systemConfig.externalAPI.headers[key];
    }
    this.updateHeaderEntries();
  }

  trackByKey(index: number, item: HeaderEntry): string {
    return item.key;
  }

  saveRoomsConfiguration(): void {
    this.isSaving.set(true);
    
    // Send rooms array as required by API
    this.http.put(this.configService.adminRoomsUrl, this.systemConfig.rooms)
      .subscribe({
        next: (response) => {
          this.isSaving.set(false);
          this.lastUpdated.set(new Date().toLocaleString());
          console.log('Rooms configuration saved:', response);
        },
        error: (error) => {
          this.isSaving.set(false);
          console.error('Failed to save rooms configuration:', error);
          const errorMessage = error.error?.message || error.message || 'Unknown error';
          alert(`Failed to save rooms configuration: ${errorMessage}`);
        }
      });
  }

  // Generic Services Management Methods
  addGenericService(): void {
    if (!this.externalAPIConfig.genericServices) {
      this.externalAPIConfig.genericServices = [];
    }
    
    const newService: GenericService = {
      id: this.generateServiceId(),
      name: '',
      description: '',
      duration: 30,
      enabled: true
    };
    
    this.externalAPIConfig.genericServices.push(newService);
  }

  removeGenericService(index: number): void {
    if (this.externalAPIConfig.genericServices && index >= 0 && index < this.externalAPIConfig.genericServices.length) {
      this.externalAPIConfig.genericServices.splice(index, 1);
    }
  }

  toggleGenericService(index: number): void {
    if (this.externalAPIConfig.genericServices && index >= 0 && index < this.externalAPIConfig.genericServices.length) {
      this.externalAPIConfig.genericServices[index].enabled = !this.externalAPIConfig.genericServices[index].enabled;
    }
  }

  private generateServiceId(): string {
    return 'service-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
  }

  get genericServicesCount(): number {
    return this.externalAPIConfig.genericServices?.length || 0;
  }

  get enabledGenericServicesCount(): number {
    if (!this.externalAPIConfig.genericServices) {
      return 0;
    }
    return this.externalAPIConfig.genericServices.filter(service => service.enabled).length;
  }
}