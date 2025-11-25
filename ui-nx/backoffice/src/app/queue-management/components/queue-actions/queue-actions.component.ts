import { Component, Input, Output, EventEmitter, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslatePipe } from '../../../../../../src/lib/i18n';

@Component({
  selector: 'app-queue-actions',
  standalone: true,
  imports: [CommonModule, TranslatePipe],
  template: `
    <div class="mb-6">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
        <!-- Call Next Button -->
        <button
          class="group relative px-6 py-4 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl font-bold hover:from-blue-700 hover:to-indigo-700 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:from-blue-600 disabled:hover:to-indigo-600 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 disabled:transform-none"
          (click)="callNextClick.emit()"
          [disabled]="isLoading || !canCallNext">
          @if (isLoading) {
            <span class="flex items-center justify-center gap-3">
              <svg class="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>{{ 'backoffice.processing' | translate }}</span>
            </span>
          } @else {
            <span class="flex items-center justify-center gap-3">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 9l3 3m0 0l-3 3m3-3H8m13 0a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
              <span>{{ 'backoffice.callNextPerson' | translate }}</span>
            </span>
          }
        </button>

        <!-- Finish Current Button -->
        <button
          class="group relative px-6 py-4 bg-gradient-to-r from-emerald-600 to-green-600 text-white rounded-xl font-bold hover:from-emerald-700 hover:to-green-700 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:from-emerald-600 disabled:hover:to-green-600 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 disabled:transform-none"
          (click)="finishCurrentClick.emit()"
          [disabled]="isLoading || !canFinishCurrent">
          @if (isLoading) {
            <span class="flex items-center justify-center gap-3">
              <svg class="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>{{ 'backoffice.processing' | translate }}</span>
            </span>
          } @else {
            <span class="flex items-center justify-center gap-3">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
              <span>{{ 'backoffice.finishCurrentPerson' | translate }}</span>
            </span>
          }
        </button>

        <!-- Refresh Button -->
        <button
          class="group relative px-6 py-4 bg-gradient-to-r from-gray-600 to-slate-700 text-white rounded-xl font-bold hover:from-gray-700 hover:to-slate-800 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:from-gray-600 disabled:hover:to-slate-700 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 disabled:transform-none"
          (click)="refreshClick.emit()"
          [disabled]="isLoading">
          <span class="flex items-center justify-center gap-3">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            <span>{{ 'backoffice.refreshQueue' | translate }}</span>
          </span>
        </button>
      </div>
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
