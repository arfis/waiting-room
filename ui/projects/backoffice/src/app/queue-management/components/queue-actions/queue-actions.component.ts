import { Component, Input, Output, EventEmitter, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-queue-actions',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="mb-8 flex gap-4">
      <button 
        class="px-6 py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        (click)="callNextClick.emit()"
        [disabled]="isLoading || !canCallNext">
        @if (isLoading) {
          <span class="flex items-center gap-2">
            <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Processing...
          </span>
        } @else {
          Call Next Person
        }
      </button>
      
      <button 
        class="px-6 py-3 bg-orange-600 text-white rounded-lg font-medium hover:bg-orange-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        (click)="finishCurrentClick.emit()"
        [disabled]="isLoading || !canFinishCurrent">
        @if (isLoading) {
          <span class="flex items-center gap-2">
            <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Processing...
          </span>
        } @else {
          Finish Current Person
        }
      </button>
      
      <button 
        class="px-6 py-3 bg-gray-600 text-white rounded-lg font-medium hover:bg-gray-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        (click)="refreshClick.emit()"
        [disabled]="isLoading">
        Refresh Queue
      </button>
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class QueueActionsComponent {
  @Input({ required: true }) isLoading!: boolean;
  @Input({ required: true }) canCallNext!: boolean;
  @Input({ required: true }) canFinishCurrent!: boolean;
  
  @Output() callNextClick = new EventEmitter<void>();
  @Output() finishCurrentClick = new EventEmitter<void>();
  @Output() refreshClick = new EventEmitter<void>();
}
