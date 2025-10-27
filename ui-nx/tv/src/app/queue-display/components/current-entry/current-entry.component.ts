import { Component, input, ChangeDetectionStrategy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from '@waiting-room/primeng-components';
import { WebSocketQueueEntry } from '@waiting-room/api-client';
import { ServicePointService } from '../../services/service-point.service';
import { TranslatePipe } from '../../../../../../src/lib/i18n';

@Component({
  selector: 'app-current-entry',
  standalone: true,
  imports: [CommonModule, CardComponent, TranslatePipe],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <ui-card [title]="'tv.nowServing' | translate" variant="success">
      <div class="text-center">
        <div class="text-8xl font-mono font-bold text-green-600 mb-4">
          {{ entry().ticketNumber }}
        </div>
        @if (entry().cardData) {
          <div class="text-2xl text-gray-700 mb-4">
            {{ entry().cardData?.firstName }} {{ entry().cardData?.lastName }}
          </div>
        }
        @if (entry().servicePoint) {
          <div class="mt-4">
            <span class="inline-flex items-center px-4 py-2 rounded-lg text-lg font-medium bg-green-100 text-green-700 border border-green-200">
              <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
              </svg>
              {{ getServicePointName(entry().servicePoint || '') }}
            </span>
          </div>
        }
      </div>
    </ui-card>
  `
})
export class CurrentEntryComponent {
  private readonly servicePointService = inject(ServicePointService);
  entry = input.required<WebSocketQueueEntry>();

  getServicePointName(servicePointId: string): string {
    return this.servicePointService.getServicePointName(servicePointId);
  }
}
