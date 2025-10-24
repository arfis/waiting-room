import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  private _apiUrl = 'http://localhost:8080/api';
  private _wsUrl = 'ws://localhost:8080/ws';
  private _appName = 'Backoffice';

  get apiUrl(): string {
    return this._apiUrl;
  }

  get wsUrl(): string {
    return this._wsUrl;
  }

  get appName(): string {
    return this._appName;
  }

  // Queue endpoints
  get queueUrl(): string {
    return `${this._apiUrl}/waiting-rooms`;
  }

  get queueEntriesUrl(): string {
    return `${this._apiUrl}/waiting-rooms`;
  }

  // WebSocket endpoints
  get queueWebSocketUrl(): string {
    return `${this._wsUrl}/queue`;
  }
}
