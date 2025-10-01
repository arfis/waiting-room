import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

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

  getQueue(roomId: string): Observable<QueueEntry[]> {
    return this.http.get<QueueEntry[]>(`${this.apiUrl}/waiting-rooms/${roomId}/queue`);
  }

  callNext(roomId: string): Observable<CallNextResponse> {
    return this.http.post<CallNextResponse>(`${this.apiUrl}/waiting-rooms/${roomId}/next`, {});
  }

  finishCurrent(roomId: string): Observable<CallNextResponse> {
    return this.http.post<CallNextResponse>(`${this.apiUrl}/waiting-rooms/${roomId}/finish`, {});
  }
}
