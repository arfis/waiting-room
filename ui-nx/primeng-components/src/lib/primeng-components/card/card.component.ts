import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'ui-card',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="bg-white rounded-lg shadow-md border border-gray-200 p-6" [class]="getVariantClass()">
      @if (title) {
        <div class="mb-4">
          <h3 class="text-lg font-semibold text-gray-900">{{ title }}</h3>
          @if (subtitle) {
            <p class="text-sm text-gray-600 mt-1">{{ subtitle }}</p>
          }
        </div>
      }
      <ng-content />
    </div>
  `,
  styles: []
})
export class CardComponent {
  @Input() title?: string;
  @Input() subtitle?: string;
  @Input() variant: 'default' | 'success' | 'error' | 'warning' = 'default';

  getVariantClass(): string {
    switch (this.variant) {
      case 'success':
        return 'border-green-200 bg-green-50';
      case 'error':
        return 'border-red-200 bg-red-50';
      case 'warning':
        return 'border-yellow-200 bg-yellow-50';
      default:
        return 'border-gray-200 bg-white';
    }
  }
}
