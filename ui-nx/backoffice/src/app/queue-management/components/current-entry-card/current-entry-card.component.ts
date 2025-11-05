import { Component, Input, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from '@waiting-room/primeng-components';
import { WebSocketQueueEntry } from '@waiting-room/api-client';
import { TranslatePipe } from '../../../../../../src/lib/i18n';

@Component({
  selector: 'app-current-entry-card',
  standalone: true,
  imports: [CommonModule, CardComponent, TranslatePipe],
  template: `
    @if (entry) {
      <div class="mb-8">
        <ui-card [title]="'backoffice.currentEntry' | translate" variant="success">
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
              @if (entry.serviceName || entry.serviceDuration) {
                <div class="mt-2 space-y-1">
                  @if (entry.serviceName) {
                    <div class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-purple-100 text-purple-700">
                      <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                      </svg>
                      {{ entry.serviceName }}
                    </div>
                  }
                  @if (entry.serviceDuration) {
                    <div class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-orange-100 text-orange-700">
                      <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                      </svg>
                      {{ entry.serviceDuration }} min
                    </div>
                  }
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
    // Return the service point ID as-is since names come from configuration
    // If needed, we can inject the configuration service here to get the actual name
    return servicePointId;
  }
}
