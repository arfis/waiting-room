import { Component, input, ChangeDetectionStrategy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from '@waiting-room/primeng-components';
import { WebSocketQueueEntry } from '@waiting-room/api-client';
import { ServicePointService } from '../../services/service-point.service';

@Component({
  selector: 'app-waiting-ticket',
  standalone: true,
  imports: [CommonModule, CardComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <ui-card title="Ticket {{ ticket().ticketNumber }}" variant="default">
      <div class="text-center">
        <div class="text-4xl font-mono font-bold text-blue-600 mb-3">
          {{ ticket().ticketNumber }}
        </div>
        <div class="text-lg text-gray-600 mb-2">
          Position: <span class="font-bold">{{ ticket().position }}</span>
        </div>
        @if (ticket().cardData) {
          <div class="text-sm text-gray-500 mb-2">
            {{ ticket().cardData?.firstName }} {{ ticket().cardData?.lastName }}
          </div>
        }
        @if (ticket().servicePoint) {
          <div class="mt-2 mb-3">
            <span class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-700">
              <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
              </svg>
              {{ getServicePointName(ticket().servicePoint || '') }}
            </span>
          </div>
        }
        <div class="mt-3">
          <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-yellow-100 text-yellow-800">
            Waiting
          </span>
        </div>
      </div>
    </ui-card>
  `
})
export class WaitingTicketComponent {
  private readonly servicePointService = inject(ServicePointService);
  ticket = input.required<WebSocketQueueEntry>();

  getServicePointName(servicePointId: string): string {
    return this.servicePointService.getServicePointName(servicePointId);
  }
}
