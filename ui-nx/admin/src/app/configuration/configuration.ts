import { Component, signal, computed, inject, OnInit, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { ConfigService } from '../shared/services/config.service';
import { TenantService } from '@lib/tenant';
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
  genericServicesPostBody?: string;
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
  roomId?: string; // Track which room this service point belongs to
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
export class ConfigurationComponent implements OnInit {
  private tenantService = inject(TenantService);
  private currentTenantId = signal<string>('');
  
  constructor(
    private http: HttpClient,
    private configService: ConfigService,
    private translationService: TranslationService
  ) {
    // Watch for tenant selection and load configuration when tenant is selected or changes
    effect(() => {
      const tenantId = this.tenantService.selectedTenantId();
      const previousTenantId = this.currentTenantId();
      
      // Load configuration if:
      // 1. Tenant is selected AND
      // 2. (We haven't loaded yet OR tenant has changed)
      if (tenantId && (tenantId !== previousTenantId)) {
        this.currentTenantId.set(tenantId);
        this.loadConfiguration();
      } else if (!tenantId) {
        // Clear current tenant if none selected
        this.currentTenantId.set('');
      }
    });
  }
  
  ngOnInit(): void {
    // Check if tenant is already selected on init
    const tenantId = this.tenantService.selectedTenantId();
    if (tenantId && tenantId !== this.currentTenantId()) {
      this.currentTenantId.set(tenantId);
      this.loadConfiguration();
    }
  }
  
  isSaving = signal(false);
  lastUpdated = signal('Just now');
  configurationCount = signal(5);
  
  // Track which room each service point belongs to
  private servicePointRoomMap = new Map<string, string>();
  
  // Flattened service points for direct management
  protected readonly servicePoints = computed(() => {
    // Collect all service points from all rooms (use signal for reactivity)
    const allServicePoints: ServicePoint[] = [];
    const rooms = this.roomsSignal();
    rooms.forEach(room => {
      if (room.servicePoints && room.servicePoints.length > 0) {
        // Add room context to service points for tracking
        room.servicePoints.forEach(sp => {
          allServicePoints.push({
            ...sp,
            roomId: room.id // Add roomId for tracking
          });
        });
      }
    });
    return allServicePoints;
  });
  
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
    genericServicesPostBody: '',
    webhookHttpMethod: 'POST',
    genericServices: []
  };
  
  // Use signal for rooms to ensure reactivity
  roomsSignal = signal<Room[]>([]);
  
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
    rooms: [], // Start with empty rooms - will be loaded from backend
    defaultRoom: '',
    webSocketPath: '/ws/queue',
    allowWildcard: true
  };

  get rooms() {
    return this.roomsSignal();
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
    const tenantId = this.tenantService.selectedTenantId();
    console.log(`[ConfigurationComponent] Loading configuration for tenant: ${tenantId || 'none'}`);
    
    // Load external API configuration
    this.http.get(this.configService.adminExternalApiUrl)
      .subscribe({
        next: (response: any) => {
          console.log(`[ConfigurationComponent] Received external API config:`, response);
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
              genericServicesPostBody: response.genericServicesPostBody || '',
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
          } else {
            // Reset to default/empty values when no config is found
            console.log('[ConfigurationComponent] No external API config found, resetting to defaults');
            this.externalAPIConfig = {
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
              genericServicesPostBody: '',
              webhookHttpMethod: 'POST',
              appointmentServicesUrl: '',
              genericServicesUrl: '',
              genericServices: [],
              webhookUrl: '',
              webhookTimeoutSeconds: 5,
              webhookRetryAttempts: 2
            };
            this.systemConfig.externalAPI = {
              timeoutSeconds: 30,
              retryAttempts: 3,
              headers: {},
              multilingualSupport: false,
              supportedLanguages: ['en'],
              useDeepLTranslation: false,
              appointmentServicesLanguageHandling: 'query_param',
              appointmentServicesLanguageHeader: 'Accept-Language',
              genericServicesLanguageHandling: 'query_param',
              genericServicesLanguageHeader: 'Accept-Language'
            };
            this.updateHeaderEntries();
          }
        },
        error: (error) => {
          console.error('Failed to load external API configuration:', error);
          // On error, also reset to defaults
          this.externalAPIConfig = {
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
            genericServicesPostBody: '',
            webhookHttpMethod: 'POST',
            appointmentServicesUrl: '',
            genericServicesUrl: '',
            genericServices: [],
            webhookUrl: '',
            webhookTimeoutSeconds: 5,
            webhookRetryAttempts: 2
          };
          this.updateHeaderEntries();
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
          } else {
            // Reset to default/empty values when no config is found
            console.log('[ConfigurationComponent] No system config found, resetting to defaults');
            this.systemConfig.defaultRoom = '';
            this.systemConfig.webSocketPath = '/ws/queue';
            this.systemConfig.allowWildcard = false;
            this.systemConfig.id = '';
            this.systemConfig.createdAt = undefined;
            this.systemConfig.updatedAt = undefined;
          }
        },
        error: (error) => {
          console.error('Failed to load system configuration:', error);
          // On error, also reset to defaults
          this.systemConfig.defaultRoom = '';
          this.systemConfig.webSocketPath = '/ws/queue';
          this.systemConfig.allowWildcard = false;
          this.systemConfig.id = '';
          this.systemConfig.createdAt = undefined;
          this.systemConfig.updatedAt = undefined;
        }
      });

    // Load rooms configuration
    this.http.get(this.configService.adminRoomsUrl)
      .subscribe({
        next: (response: any) => {
          console.log('[ConfigurationComponent] Received rooms config response:', response);
          if (response && Array.isArray(response) && response.length > 0) {
            // Ensure all rooms have servicePoints array (even if empty)
            const roomsWithServicePoints = response.map((room: Room) => ({
              ...room,
              servicePoints: room.servicePoints || []
            }));
            // Update both the signal and the systemConfig for consistency
            this.roomsSignal.set(roomsWithServicePoints);
            this.systemConfig.rooms = roomsWithServicePoints;
            console.log('[ConfigurationComponent] Loaded rooms configuration:', roomsWithServicePoints);
            console.log('[ConfigurationComponent] Total service points:', this.servicePoints().length);
            
            // Rebuild service point room map
            this.servicePointRoomMap.clear();
            roomsWithServicePoints.forEach((room: Room) => {
              if (room.servicePoints && room.servicePoints.length > 0) {
                room.servicePoints.forEach((sp: ServicePoint) => {
                  this.servicePointRoomMap.set(sp.id, room.id);
                });
              }
            });
          } else {
            // Reset to empty array when no rooms config is found
            console.log('[ConfigurationComponent] No rooms config found or empty response, resetting to empty array');
            console.log('[ConfigurationComponent] Response was:', response);
            this.roomsSignal.set([]);
            this.systemConfig.rooms = [];
            this.servicePointRoomMap.clear();
            console.log('[ConfigurationComponent] Service points after reset:', this.servicePoints().length);
          }
        },
        error: (error) => {
          console.error('[ConfigurationComponent] Failed to load rooms configuration:', error);
          // On error, also reset to empty array
          this.roomsSignal.set([]);
          this.systemConfig.rooms = [];
          this.servicePointRoomMap.clear();
          console.log('[ConfigurationComponent] Service points after error reset:', this.servicePoints().length);
        }
      });
  }

  saveExternalAPIConfig(): void {
    // Validate appointment URL if provided
    if (this.externalAPIConfig.appointmentServicesUrl && !this.externalAPIConfig.appointmentServicesUrl.includes('${identifier}')) {
      alert('Appointment URL must contain ${identifier} placeholder. Example: http://api.example.com/users/${identifier}/appointments');
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
        genericServicesPostBody: this.externalAPIConfig.genericServicesPostBody || '',
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
    
    // If this is the first room, make it default
    if (this.systemConfig.rooms.length === 0) {
      newRoom.isDefault = true;
      this.systemConfig.defaultRoom = newRoom.id;
    }
    
    this.systemConfig.rooms.push(newRoom);
  }

  addServicePoint(): void {
    console.log('[ConfigurationComponent] addServicePoint called');
    const currentRooms = this.roomsSignal();
    
    // Ensure we have at least one room (default room) - created automatically behind the scenes
    let rooms = currentRooms;
    if (rooms.length === 0) {
      const defaultRoom: Room = {
        id: `room-${Date.now()}`,
        name: 'Default Room',
        description: 'Auto-created room for service points',
        isDefault: true,
        servicePoints: []
      };
      rooms = [defaultRoom];
      this.systemConfig.defaultRoom = defaultRoom.id;
      this.roomsSignal.set(rooms);
      this.systemConfig.rooms = rooms;
      console.log('[ConfigurationComponent] Auto-created default room:', defaultRoom.id);
    }
    
    // Add service point to the default room (or first room if no default)
    const defaultRoom = rooms.find(r => r.isDefault) || rooms[0];
    if (!defaultRoom) {
      console.error('[ConfigurationComponent] No default room found');
      return;
    }
    
    if (!defaultRoom.servicePoints) {
      defaultRoom.servicePoints = [];
    }
    
    const newServicePoint: ServicePoint = {
      id: `sp-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      name: 'New Service Point',
      description: '',
      managerId: '',
      managerName: '',
      roomId: defaultRoom.id
    };
    
    // Create a new array with the updated service points to ensure reactivity
    const updatedRooms = rooms.map(room => {
      if (room.id === defaultRoom.id) {
        return {
          ...room,
          servicePoints: [...(room.servicePoints || []), newServicePoint]
        };
      }
      return room;
    });
    
    // Update both the signal and systemConfig
    this.roomsSignal.set(updatedRooms);
    this.systemConfig.rooms = updatedRooms;
    this.servicePointRoomMap.set(newServicePoint.id, defaultRoom.id);
    
    console.log('[ConfigurationComponent] Service point added:', newServicePoint);
    console.log('[ConfigurationComponent] Updated rooms:', updatedRooms);
    console.log('[ConfigurationComponent] Total service points:', this.servicePoints().length);
    console.log('[ConfigurationComponent] Service points signal value:', this.servicePoints());
  }

  deleteServicePointById(servicePointId: string): void {
    if (confirm('Are you sure you want to delete this service point?')) {
      // Find and remove from the room
      const roomId = this.servicePointRoomMap.get(servicePointId);
      if (roomId) {
        const currentRooms = this.roomsSignal();
        const updatedRooms = currentRooms.map(room => {
          if (room.id === roomId && room.servicePoints) {
            return {
              ...room,
              servicePoints: room.servicePoints.filter(sp => sp.id !== servicePointId)
            };
          }
          return room;
        });
        
        // Update both the signal and systemConfig
        this.roomsSignal.set(updatedRooms);
        this.systemConfig.rooms = updatedRooms;
        this.servicePointRoomMap.delete(servicePointId);
        
        console.log('[ConfigurationComponent] Service point deleted:', servicePointId);
        console.log('[ConfigurationComponent] Total service points:', this.servicePoints().length);
      }
    }
  }
  
  trackByServicePointId(index: number, servicePoint: ServicePoint): string {
    return servicePoint.id;
  }
  
  saveServicePointsConfiguration(): void {
    this.isSaving.set(true);
    
    const currentRooms = this.roomsSignal();
    
    // Ensure we have at least one room (default room) - auto-created if needed
    let rooms = currentRooms;
    if (rooms.length === 0) {
      const defaultRoom: Room = {
        id: `room-${Date.now()}`,
        name: 'Default Room',
        description: 'Auto-created room for service points',
        isDefault: true,
        servicePoints: []
      };
      rooms = [defaultRoom];
      this.systemConfig.defaultRoom = defaultRoom.id;
      this.roomsSignal.set(rooms);
      this.systemConfig.rooms = rooms;
      console.log('[ConfigurationComponent] Auto-created default room for saving:', defaultRoom.id);
    }
    
    // Ensure all service points are properly assigned to rooms
    // Service points are stored in the default room (or first room)
    const defaultRoom = rooms.find(r => r.isDefault) || rooms[0];
    if (!defaultRoom) {
      alert('Failed to save: No room available');
      this.isSaving.set(false);
      return;
    }
    
    // Collect all service points from the computed signal
    const servicePoints = this.servicePoints();
    console.log('[ConfigurationComponent] Collecting service points to save:', servicePoints);
    
    // Update the default room's service points (remove roomId before saving)
    const servicePointsToSave = servicePoints.map(sp => {
      const { roomId, ...servicePoint } = sp;
      return servicePoint;
    });
    
    const updatedDefaultRoom = {
      ...defaultRoom,
      servicePoints: servicePointsToSave
    };
    
    // Update the rooms array with the updated default room
    const updatedRooms = rooms.map(room => 
      room.id === defaultRoom.id ? updatedDefaultRoom : { ...room, servicePoints: room.servicePoints || [] }
    );
    
    // Ensure all rooms have servicePoints array (even if empty)
    const roomsToSave = updatedRooms.map(room => ({
      ...room,
      servicePoints: room.servicePoints || []
    }));
    
    // Save the rooms configuration (which includes service points)
    console.log('[ConfigurationComponent] Saving service points:', servicePoints);
    console.log('[ConfigurationComponent] Service points to save (without roomId):', servicePointsToSave);
    console.log('[ConfigurationComponent] Default room service points:', updatedDefaultRoom.servicePoints);
    console.log('[ConfigurationComponent] Rooms to save:', roomsToSave);
    
    this.http.put(this.configService.adminRoomsUrl, roomsToSave)
      .subscribe({
        next: (response) => {
          this.isSaving.set(false);
          this.lastUpdated.set(new Date().toLocaleString());
          console.log('[ConfigurationComponent] Service points configuration saved successfully:', response);
          // Update signal before reloading
          this.roomsSignal.set(updatedRooms);
          this.systemConfig.rooms = updatedRooms;
          // Reload configuration to ensure sync
          this.loadConfiguration();
        },
        error: (error) => {
          this.isSaving.set(false);
          console.error('[ConfigurationComponent] Failed to save service points configuration:', error);
          console.error('[ConfigurationComponent] Error details:', JSON.stringify(error, null, 2));
          const errorMessage = error.error?.message || error.message || 'Unknown error';
          alert(`Failed to save service points configuration: ${errorMessage}`);
        }
      });
  }

  deleteRoom(roomId: string): void {
    if (confirm('Are you sure you want to delete this room?')) {
      const currentRooms = this.roomsSignal();
      const updatedRooms = currentRooms.filter(room => room.id !== roomId);
      
      // If we deleted the default room, set a new default
      if (this.systemConfig.defaultRoom === roomId && updatedRooms.length > 0) {
        this.systemConfig.defaultRoom = updatedRooms[0].id;
      }
      
      this.roomsSignal.set(updatedRooms);
      this.systemConfig.rooms = updatedRooms;
    }
  }

  setDefaultRoom(roomId: string): void {
    this.systemConfig.defaultRoom = roomId;
    
    // Update isDefault flags
    const currentRooms = this.roomsSignal();
    const updatedRooms = currentRooms.map(room => ({
      ...room,
      isDefault: room.id === roomId
    }));
    
    this.roomsSignal.set(updatedRooms);
    this.systemConfig.rooms = updatedRooms;
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