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
}
