import { Component, input, ChangeDetectionStrategy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from '@waiting-room/primeng-components';
import { WebSocketQueueEntry } from '@waiting-room/api-client';
import { ServicePointService } from '../../services/service-point.service';

@Component({
  selector: 'app-called-ticket',
  standalone: true,
  imports: [CommonModule, CardComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <ui-card>
      <div class="text-center">
        <div class="text-3xl font-mono font-bold text-orange-600 mb-3">
          {{ ticket().ticketNumber }}
        </div>
        @if (ticket().cardData) {
          <div class="text-lg text-gray-700 mb-3">
            {{ ticket().cardData?.firstName }} {{ ticket().cardData?.lastName }}
          </div>
        }
        <div class="mb-3">
        @if (ticket().servicePoint) {
          <div class="mb-3">
            <span class="inline-flex items-center px-4 py-2 rounded-lg text-lg font-medium bg-orange-100 text-orange-700 border border-orange-200">
              <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
              </svg>
              {{ getServicePointName(ticket().servicePoint || '') }}
            </span>
          </div>
        }
        <div class="mt-3">
          <span class="inline-flex items-center px-4 py-2 rounded-full text-lg font-medium bg-orange-100 text-orange-800 border border-orange-200">
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"></path>
            </svg>
            Called
          </span>
        </div>
        </div>
      </div>
    </ui-card>
  `
})
export class CalledTicketComponent {
  private readonly servicePointService = inject(ServicePointService);
  ticket = input.required<WebSocketQueueEntry>();

  getServicePointName(servicePointId: string): string {
    return this.servicePointService.getServicePointName(servicePointId);
  }
}
