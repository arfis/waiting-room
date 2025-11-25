import { Component, Input, Output, EventEmitter, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from '@waiting-room/primeng-components';
import { WebSocketQueueEntry } from '@waiting-room/api-client';
import { TranslatePipe } from '../../../../../../src/lib/i18n';

@Component({
  selector: 'app-waiting-queue-list',
  standalone: true,
  imports: [CommonModule, CardComponent, TranslatePipe],
  template: `
    <div class="mb-8">
      <div class="flex items-center justify-between mb-6">
        <div>
          <h2 class="text-3xl font-bold text-gray-900 flex items-center gap-3">
            <svg class="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
            </svg>
            {{ 'backoffice.waitingQueue' | translate }}
          </h2>
          <p class="text-gray-600 mt-1">{{ entries.length }} {{ (entries.length === 1 ? 'backoffice.personWaiting' : 'backoffice.peopleWaiting') | translate }}</p>
        </div>
      </div>

      @if (entries.length === 0) {
        <div class="bg-white rounded-2xl shadow-lg border border-gray-200 overflow-hidden">
          <div class="text-center py-16 px-6">
            <div class="inline-flex items-center justify-center w-20 h-20 rounded-full bg-gray-100 mb-6">
              <svg class="w-10 h-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
            </div>
            <h3 class="text-2xl font-bold text-gray-900 mb-2">{{ 'backoffice.queueIsEmpty' | translate }}</h3>
            <p class="text-gray-600">{{ 'backoffice.noOneWaiting' | translate }}</p>
          </div>
        </div>
      } @else {
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          @for (entry of entries; track entry.id) {
            <div class="bg-white rounded-xl shadow-lg border border-gray-200 overflow-hidden transition-all duration-300 hover:shadow-2xl hover:-translate-y-1 hover:border-blue-300 group">
              <!-- Header with Position Badge -->
              <div class="bg-gradient-to-r from-slate-50 to-gray-50 px-5 py-3 border-b border-gray-200">
                <div class="flex items-center justify-between">
                  <div class="text-3xl font-mono font-black text-blue-600 group-hover:text-blue-700 transition-colors">
                    {{ entry.ticketNumber }}
                  </div>
                  <div class="flex items-center gap-2">
                    <span class="inline-flex items-center px-3 py-1.5 rounded-full text-sm font-bold bg-gradient-to-r from-amber-400 to-orange-400 text-white shadow-md">
                      <svg class="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M10 3.5a1.5 1.5 0 013 0V4a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-.5a1.5 1.5 0 000 3h.5a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-.5a1.5 1.5 0 00-3 0v.5a1 1 0 01-1 1H6a1 1 0 01-1-1v-3a1 1 0 00-1-1h-.5a1.5 1.5 0 010-3H4a1 1 0 001-1V6a1 1 0 011-1h3a1 1 0 001-1v-.5z"/>
                      </svg>
                      {{ 'backoffice.position' | translate }} {{ entry.position }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Customer Information -->
              <div class="p-5 space-y-4">
                @if (entry.cardData) {
                  <div class="bg-gradient-to-br from-blue-50 to-indigo-50 rounded-lg p-4 border border-blue-100">
                    <div class="flex items-start gap-3">
                      <div class="flex-shrink-0">
                        <div class="w-12 h-12 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center shadow-md">
                          <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
                          </svg>
                        </div>
                      </div>
                      <div class="flex-1 min-w-0">
                        <div class="text-base font-bold text-gray-900 truncate">
                          {{ getCardDataField(entry, 'firstName') }} {{ getCardDataField(entry, 'lastName') }}
                        </div>
                        <div class="text-sm font-mono text-gray-600 mt-1">
                          ID: {{ getCardDataField(entry, 'idNumber') }}
                        </div>
                      </div>
                    </div>
                  </div>
                }

                <!-- Service Details -->
                <div class="space-y-2">
                  @if (entry.servicePoint) {
                    <div class="flex items-center gap-2 text-sm">
                      <div class="w-8 h-8 rounded-lg bg-green-100 flex items-center justify-center flex-shrink-0">
                        <svg class="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
                        </svg>
                      </div>
                      <div class="flex-1">
                        <div class="text-xs text-gray-500 font-semibold uppercase tracking-wide">{{ 'common.servicePoint' | translate }}</div>
                        <div class="font-bold text-gray-900">{{ getServicePointName(entry.servicePoint) }}</div>
                      </div>
                    </div>
                  }

                  @if (entry.serviceName) {
                    <div class="flex items-center gap-2 text-sm">
                      <div class="w-8 h-8 rounded-lg bg-purple-100 flex items-center justify-center flex-shrink-0">
                        <svg class="w-4 h-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                        </svg>
                      </div>
                      <div class="flex-1">
                        <div class="text-xs text-gray-500 font-semibold uppercase tracking-wide">{{ 'backoffice.serviceType' | translate }}</div>
                        <div class="font-bold text-gray-900">{{ entry.serviceName }}</div>
                      </div>
                    </div>
                  }

                  @if (entry.serviceDuration) {
                    <div class="flex items-center gap-2 text-sm">
                      <div class="w-8 h-8 rounded-lg bg-orange-100 flex items-center justify-center flex-shrink-0">
                        <svg class="w-4 h-4 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                        </svg>
                      </div>
                      <div class="flex-1">
                        <div class="text-xs text-gray-500 font-semibold uppercase tracking-wide">{{ 'backoffice.estDuration' | translate }}</div>
                        <div class="font-bold text-gray-900">{{ entry.serviceDuration }} {{ 'backoffice.minutes' | translate }}</div>
                      </div>
                    </div>
                  }
                </div>

                <!-- Timestamp -->
                <div class="flex items-center gap-2 text-xs text-gray-500 pt-2 border-t border-gray-100">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                  </svg>
                  <span>{{ 'backoffice.joined' | translate }} {{ entry.createdAt | date:'short' }}</span>
                </div>

                <!-- Call Button -->
                <button
                  class="w-full px-4 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-lg text-base font-bold hover:from-blue-700 hover:to-indigo-700 transition-all duration-200 shadow-md hover:shadow-xl transform hover:scale-105 flex items-center justify-center gap-2"
                  (click)="callEntry.emit(entry)">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"></path>
                  </svg>
                  {{ 'backoffice.callThisPerson' | translate }}
                </button>
              </div>
            </div>
          }
        </div>
      }
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class WaitingQueueListComponent {
  @Input({ required: true }) entries!: WebSocketQueueEntry[];
  @Output() callEntry = new EventEmitter<WebSocketQueueEntry>();

  getServicePointName(servicePointId: string): string {
    // Return the service point ID as-is since names come from configuration
    // If needed, we can inject the configuration service here to get the actual name
    return servicePointId;
  }

  // Safe method to get card data field
  getCardDataField(entry: WebSocketQueueEntry, field: 'idNumber' | 'firstName' | 'lastName'): string {
    if (!entry?.cardData) {
      return '';
    }

    // If cardData is a string, try to parse it
    let cardData = entry.cardData;
    if (typeof entry.cardData === 'string') {
      try {
        cardData = JSON.parse(entry.cardData as any);
      } catch (e) {
        console.error('[WaitingQueueList] Failed to parse cardData:', e);
        return '';
      }
    }

    // Handle both snake_case and camelCase
    const fieldMap = {
      idNumber: ['idNumber', 'id_number'],
      firstName: ['firstName', 'first_name'],
      lastName: ['lastName', 'last_name']
    };

    const possibleFields = fieldMap[field];
    for (const possibleField of possibleFields) {
      if (cardData && typeof cardData === 'object' && possibleField in cardData) {
        const value = (cardData as any)[possibleField];
        return value || '';
      }
    }

    return '';
  }
}
