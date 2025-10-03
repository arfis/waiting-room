import { Component, Input, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from 'ui';
import { WebSocketQueueEntry } from 'api-client';

@Component({
  selector: 'app-current-entry-card',
  standalone: true,
  imports: [CommonModule, CardComponent],
  template: `
    @if (entry) {
      <div class="mb-8">
        <ui-card title="Currently Being Served" variant="success">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-3xl font-mono font-bold text-green-600 mb-2">
                {{ entry.ticketNumber }}
              </div>
              @if (entry.cardData) {
                <div class="text-lg text-gray-700">
                  {{ entry.cardData.firstName }} {{ entry.cardData.lastName }}
                </div>
                <div class="text-sm text-gray-500">
                  ID: {{ entry.cardData.idNumber }}
                </div>
              }
              @if (entry.servicePoint) {
                <div class="mt-2">
                  <span class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-700">
                    <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
                    </svg>
                    {{ getServicePointName(entry.servicePoint) }}
                  </span>
                </div>
              }
            </div>
            <div class="text-right">
              <div class="text-sm text-gray-500 mb-1">Status</div>
              <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
                {{ statusText }}
              </span>
            </div>
          </div>
        </ui-card>
      </div>
    }
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class CurrentEntryCardComponent {
  @Input() entry: WebSocketQueueEntry | null = null;
  @Input({ required: true }) statusText!: string;

  getServicePointName(servicePointId: string): string {
    // Map service point IDs to display names
    const servicePointNames: { [key: string]: string } = {
      'window-1': 'Window 1',
      'window-2': 'Window 2',
      'door-1': 'Door 1',
      'door-2': 'Door 2',
      'counter-1': 'Counter 1',
      'counter-2': 'Counter 2'
    };
    
    return servicePointNames[servicePointId] || servicePointId;
  }
}
