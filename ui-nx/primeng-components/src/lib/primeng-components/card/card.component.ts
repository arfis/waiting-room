import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'ui-card',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div
      class="bg-white rounded-xl shadow-lg border overflow-hidden transition-all duration-300 hover:shadow-2xl"
      [class]="getVariantClass()"
      [class.hover:-translate-y-1]="hoverable">
      @if (title || subtitle) {
        <div class="px-6 py-4" [class]="getHeaderClass()">
          @if (title) {
            <h3 class="text-xl font-bold" [class]="getTitleColorClass()">{{ title }}</h3>
          }
          @if (subtitle) {
            <p class="text-sm mt-1" [class]="getSubtitleColorClass()">{{ subtitle }}</p>
          }
        </div>
      }
      <div class="p-6" [class.pt-6]="!title && !subtitle">
        <ng-content />
      </div>
    </div>
  `,
  styles: []
})
export class CardComponent {
  @Input() title?: string;
  @Input() subtitle?: string;
  @Input() variant: 'default' | 'success' | 'error' | 'warning' | 'info' = 'default';
  @Input() hoverable = true; // Enable hover lift effect by default

  getVariantClass(): string {
    switch (this.variant) {
      case 'success':
        return 'border-emerald-200';
      case 'error':
        return 'border-red-200';
      case 'warning':
        return 'border-amber-200';
      case 'info':
        return 'border-blue-200';
      default:
        return 'border-gray-200';
    }
  }

  getHeaderClass(): string {
    switch (this.variant) {
      case 'success':
        return 'bg-gradient-to-r from-emerald-50 to-green-50 border-b border-emerald-200';
      case 'error':
        return 'bg-gradient-to-r from-red-50 to-rose-50 border-b border-red-200';
      case 'warning':
        return 'bg-gradient-to-r from-amber-50 to-yellow-50 border-b border-amber-200';
      case 'info':
        return 'bg-gradient-to-r from-blue-50 to-indigo-50 border-b border-blue-200';
      default:
        return 'bg-gradient-to-r from-gray-50 to-slate-50 border-b border-gray-200';
    }
  }

  getTitleColorClass(): string {
    switch (this.variant) {
      case 'success':
        return 'text-emerald-900';
      case 'error':
        return 'text-red-900';
      case 'warning':
        return 'text-amber-900';
      case 'info':
        return 'text-blue-900';
      default:
        return 'text-gray-900';
    }
  }

  getSubtitleColorClass(): string {
    switch (this.variant) {
      case 'success':
        return 'text-emerald-700';
      case 'error':
        return 'text-red-700';
      case 'warning':
        return 'text-amber-700';
      case 'info':
        return 'text-blue-700';
      default:
        return 'text-gray-600';
    }
  }
}
