import { Component, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { ConfigService } from '../shared/services/config.service';

interface ExternalAPIConfig {
  userServicesUrl: string;
  timeoutSeconds: number;
  retryAttempts: number;
  headers?: { [key: string]: string };
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
  imports: [CommonModule, FormsModule],
  templateUrl: './configuration.html',
  styleUrl: './configuration.scss'
})
export class ConfigurationComponent {
  constructor(
    private http: HttpClient,
    private configService: ConfigService
  ) {
    this.loadConfiguration();
  }
  
  isSaving = signal(false);
  lastUpdated = signal('Just now');
  configurationCount = signal(5);
  
  systemConfig: SystemConfiguration = {
    externalAPI: {
      userServicesUrl: 'http://localhost:3001/users/${identifier}/services',
      timeoutSeconds: 10,
      retryAttempts: 3,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      }
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

  get externalAPIConfig() {
    return this.systemConfig.externalAPI;
  }

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
            this.updateHeaderEntries();
            console.log('Loaded external API configuration:', response);
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
    // Validate URL contains ${identifier} placeholder
    if (!this.externalAPIConfig.userServicesUrl.includes('${identifier}')) {
      alert('URL must contain ${identifier} placeholder. Example: http://api.example.com/users/${identifier}/services');
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
      userServicesUrl: this.externalAPIConfig.userServicesUrl,
      timeoutSeconds: this.externalAPIConfig.timeoutSeconds,
      retryAttempts: this.externalAPIConfig.retryAttempts,
      headers: headersObject
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
}