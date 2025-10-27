import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface QueueEntry {
  id: string;
  ticketNumber: string;
  status: 'WAITING' | 'CALLED' | 'IN_SERVICE' | 'COMPLETED' | 'CANCELLED';
  position: number;
  createdAt: string;
  cardData?: {
    firstName: string;
    lastName: string;
    idNumber: string;
  };
}

export interface CallNextResponse {
  entryId: string;
  ticketNumber: string;
  status: string;
}

@Injectable({
  providedIn: 'root'
})
export class QueueApiService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl || 'http://localhost:8080/api';

  getQueue(roomId: string, states?: string[]): Observable<QueueEntry[]> {
    const params: { [key: string]: string | string[] } = {};
    if (states && states.length > 0) {
      params['state'] = states;
    }
    return this.http.get<QueueEntry[]>(`${this.apiUrl}/waiting-rooms/${roomId}/queue`, { params });
  }

  callNext(roomId: string, servicePointId: string): Observable<CallNextResponse> {
    return this.http.post<CallNextResponse>(`${this.apiUrl}/waiting-rooms/${roomId}/service-points/${servicePointId}/next`, {});
  }

  finishCurrent(roomId: string): Observable<CallNextResponse> {
    return this.http.post<CallNextResponse>(`${this.apiUrl}/waiting-rooms/${roomId}/finish`, {});
  }
}
