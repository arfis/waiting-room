import { Component, Input, Output, EventEmitter, ChangeDetectionStrategy } from '@angular/core';
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
      <div class="mb-4">
        <ui-card [title]="'backoffice.currentEntry' | translate" variant="success" [hoverable]="false">
          <!-- Main content row -->
          <div class="flex items-center gap-4">
            <!-- Ticket Number -->
            <div class="flex-shrink-0">
              <div class="text-xs text-gray-500 uppercase tracking-wide mb-1">{{ 'backoffice.ticketNumber' | translate }}</div>
              <div class="text-2xl font-mono font-black text-emerald-700">
                {{ entry.ticketNumber }}
              </div>
            </div>

            <!-- Divider -->
            <div class="h-16 w-px bg-gray-300"></div>

            <!-- Customer Info -->
            <div class="flex-1">
              @if (entry.cardData) {
                <div class="space-y-1">
                  <div class="flex items-center gap-2">
                    <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
                    </svg>
                    <span class="font-bold text-gray-900">
                      {{ getCardDataField(entry, 'firstName') }} {{ getCardDataField(entry, 'lastName') }}
                    </span>
                  </div>
                  <div class="flex items-center gap-2 text-sm">
                    <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V8a2 2 0 00-2-2h-5m-4 0V5a2 2 0 114 0v1m-4 0a2 2 0 104 0m-5 8a2 2 0 100-4 2 2 0 000 4zm0 0c1.306 0 2.417.835 2.83 2M9 14a3.001 3.001 0 00-2.83 2M15 11h3m-3 4h2"></path>
                    </svg>
                    <span class="text-gray-600 font-mono">{{ getCardDataField(entry, 'idNumber') }}</span>
                  </div>
                </div>
              }

              <!-- Service badges -->
              <div class="flex flex-wrap gap-2 mt-2">
                @if (entry.serviceName) {
                  <span class="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-purple-50 text-purple-700 border border-purple-200">
                    {{ entry.serviceName }}
                  </span>
                }
                @if (entry.serviceDuration) {
                  <span class="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-amber-50 text-amber-700 border border-amber-200">
                    {{ entry.serviceDuration }} {{ 'backoffice.minutes' | translate }}
                  </span>
                }
              </div>
            </div>

            <!-- Finish Button -->
            <div class="flex-shrink-0">
              <button
                class="px-6 py-3 bg-gradient-to-r from-emerald-600 to-green-600 text-white rounded-lg font-bold hover:from-emerald-700 hover:to-green-700 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed shadow-md hover:shadow-lg"
                (click)="finishClick.emit()"
                [disabled]="isLoading">
                @if (isLoading) {
                  <span class="flex items-center gap-2">
                    <svg class="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span class="text-sm">{{ 'backoffice.processing' | translate }}</span>
                  </span>
                } @else {
                  <span class="flex items-center gap-2">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                    </svg>
                    <span>{{ 'backoffice.finishCurrentPerson' | translate }}</span>
                  </span>
                }
              </button>
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
  @Input() isLoading = false;
  @Output() finishClick = new EventEmitter<void>();

  getServicePointName(servicePointId: string): string {
    // Return the service point ID as-is since names come from configuration
    // If needed, we can inject the configuration service here to get the actual name
    return servicePointId;
  }

  // Safe method to get card data field
  getCardDataField(entry: WebSocketQueueEntry | null, field: 'idNumber' | 'firstName' | 'lastName'): string {
    if (!entry?.cardData) {
      console.log('[CurrentEntryCard] No cardData found');
      return '';
    }

    console.log('[CurrentEntryCard] cardData:', entry.cardData, 'Type:', typeof entry.cardData);

    // If cardData is a string, try to parse it
    let cardData = entry.cardData;
    if (typeof entry.cardData === 'string') {
      try {
        cardData = JSON.parse(entry.cardData as any);
        console.log('[CurrentEntryCard] Parsed cardData:', cardData);
      } catch (e) {
        console.error('[CurrentEntryCard] Failed to parse cardData:', e);
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
        console.log(`[CurrentEntryCard] Found ${field} as ${possibleField}:`, value);
        return value || '';
      }
    }

    console.log('[CurrentEntryCard] Field not found:', field);
    return '';
  }
}
