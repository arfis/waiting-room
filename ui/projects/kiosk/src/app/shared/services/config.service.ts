import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  private _apiUrl = 'http://localhost:8080/api';
  private _wsUrl = 'ws://localhost:8080/ws';
  private _appName = 'Kiosk';

  get apiUrl(): string {
    return this._apiUrl;
  }

  get wsUrl(): string {
    return this._wsUrl;
  }

  get appName(): string {
    return this._appName;
  }

  // Kiosk endpoints
  get kioskSwipeUrl(): string {
    return `${this._apiUrl}/waiting-rooms`;
  }

  get kioskManualIdUrl(): string {
    return `${this._apiUrl}/waiting-rooms`;
  }

  // WebSocket endpoints
  get queueWebSocketUrl(): string {
    return `${this._wsUrl}/queue`;
  }
}
