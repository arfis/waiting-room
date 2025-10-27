import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface UserService {
  serviceName: string;
  duration: number;
  id: string;
}

export interface ServiceSection {
  title: string;
  services: UserService[];
  type: 'appointment' | 'generic';
  loading: boolean;
  error: string | null;
}

@Injectable({
  providedIn: 'root'
})
export class UserServicesService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl || 'http://localhost:8080/api';

  getUserServices(identifier: string): Observable<UserService[]> {
    // Call backend API which will then call external API
    return this.http.get<UserService[]>(`${this.apiUrl}/user-services`, {
      params: { identifier }
    });
  }

  getAppointmentServices(identifier: string): Observable<UserService[]> {
    // Call backend API for appointment-specific services
    return this.http.get<UserService[]>(`${this.apiUrl}/appointment-services`, {
      params: { identifier }
    });
  }

  getGenericServices(servicePointId: string): Observable<UserService[]> {
    // Call backend API for generic services available at service point
    return this.http.get<UserService[]>(`${this.apiUrl}/generic-services`, {
      params: { servicePointId }
    });
  }
}
