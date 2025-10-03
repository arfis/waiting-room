import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface CardData {
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

export interface TicketResponse {
  entryId: string;
  ticketNumber: string;
  qrUrl: string;
  servicePoint?: string;
}

export interface CardReaderStatus {
  connected: boolean;
  status: string;
}

@Injectable({
  providedIn: 'root'
})
export class KioskApiService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl || 'http://localhost:8080/api';

  generateTicket(roomId: string, idCardRaw: string): Observable<TicketResponse> {
    return this.http.post<TicketResponse>(
      `${this.apiUrl}/waiting-rooms/${roomId}/swipe`,
      { idCardRaw }
    );
  }

  getCardReaderStatus(): Observable<CardReaderStatus> {
    return this.http.get<CardReaderStatus>(`${this.apiUrl}/card-reader/status`);
  }
}
