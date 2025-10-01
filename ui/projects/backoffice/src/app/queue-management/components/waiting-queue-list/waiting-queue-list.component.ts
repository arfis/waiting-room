import { Component, Input, Output, EventEmitter, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from 'ui';
import { WebSocketQueueEntry } from 'api-client';

@Component({
  selector: 'app-waiting-queue-list',
  standalone: true,
  imports: [CommonModule, CardComponent],
  template: `
    <div class="mb-8">
      <h2 class="text-2xl font-bold text-gray-900 mb-4">Waiting Queue</h2>
      
      @if (entries.length === 0) {
        <ui-card title="No one waiting" variant="default">
          <div class="text-center py-12">
            <svg class="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <p class="text-xl text-gray-600">No one is currently waiting</p>
            <p class="text-sm text-gray-500 mt-2">The queue is empty</p>
          </div>
        </ui-card>
      } @else {
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          @for (entry of entries; track entry.id) {
            <ui-card title="Ticket {{ entry.ticketNumber }}" variant="default">
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <div class="text-2xl font-mono font-bold text-blue-600">
                    {{ entry.ticketNumber }}
                  </div>
                  <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                    Position {{ entry.position }}
                  </span>
                </div>
                
                @if (entry.cardData) {
                  <div class="text-sm text-gray-600">
                    <div class="font-medium">{{ entry.cardData.firstName }} {{ entry.cardData.lastName }}</div>
                    <div class="text-xs text-gray-500">ID: {{ entry.cardData.idNumber }}</div>
                  </div>
                }
                
                <div class="text-xs text-gray-500">
                  Joined: {{ entry.createdAt | date:'short' }}
                </div>
                
                <div class="pt-2 border-t">
                  <button 
                    class="w-full px-3 py-2 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700 transition-colors"
                    (click)="callEntry.emit(entry)">
                    Call This Person
                  </button>
                </div>
              </div>
            </ui-card>
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
}
