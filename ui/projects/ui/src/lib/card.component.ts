import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'ui-card',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="rounded-2xl shadow-xl p-8 bg-white" [class]="getVariantClass()">
      @if (title) {
        <div class="mb-6">
          <h2 class="text-2xl font-bold text-gray-900">{{ title }}</h2>
          @if (subtitle) {
            <p class="text-gray-600 mt-1">{{ subtitle }}</p>
          }
        </div>
      }
      <ng-content/>
    </div>
  `,
  styles: [`
    .success {
      border-left: 4px solid #10b981;
    }
    .error {
      border-left: 4px solid #ef4444;
    }
    .warning {
      border-left: 4px solid #f59e0b;
    }
  `]
})
export class CardComponent {
  @Input() title?: string;
  @Input() subtitle?: string;
  @Input() variant: 'default' | 'success' | 'error' | 'warning' = 'default';

  getVariantClass(): string {
    return this.variant === 'default' ? '' : this.variant;
  }
}
