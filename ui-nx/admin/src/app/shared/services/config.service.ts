import { Injectable } from '@angular/core';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  private _apiUrl = environment.apiUrl;
  private _wsUrl = environment.wsUrl;
  private _appName = environment.appName;

  get apiUrl(): string {
    return this._apiUrl;
  }

  get wsUrl(): string {
    return this._wsUrl;
  }

  get appName(): string {
    return this._appName;
  }

  // Admin specific endpoints
  get adminConfigUrl(): string {
    return `${this._apiUrl}/admin/configuration`;
  }

  get adminExternalApiUrl(): string {
    return `${this._apiUrl}/admin/configuration/external-api`;
  }

  get adminRoomsUrl(): string {
    return `${this._apiUrl}/admin/configuration/rooms`;
  }

  get adminCardReadersUrl(): string {
    return `${this._apiUrl}/admin/card-readers`;
  }

  // Queue endpoints
  get queueUrl(): string {
    return `${this._apiUrl}/waiting-rooms`;
  }

  // Kiosk endpoints
  get kioskSwipeUrl(): string {
    return `${this._apiUrl}/waiting-rooms`;
  }

  // WebSocket endpoints
  get queueWebSocketUrl(): string {
    return `${this._wsUrl}/queue`;
  }
}
