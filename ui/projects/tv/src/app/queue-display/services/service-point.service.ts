import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ServicePointService {
  private readonly servicePointNames: { [key: string]: string } = {
    'window-1': 'Window 1',
    'window-2': 'Window 2',
    'door-1': 'Door 1',
    'door-2': 'Door 2',
    'counter-1': 'Counter 1',
    'counter-2': 'Counter 2'
  };

  getServicePointName(servicePointId: string): string {
    return this.servicePointNames[servicePointId] || servicePointId;
  }
}
