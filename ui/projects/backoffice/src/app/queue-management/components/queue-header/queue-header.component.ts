import { Component, Input, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-queue-header',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="mb-8">
      <h1 class="text-4xl font-bold text-gray-900 mb-2">{{ title }}</h1>
      <p class="text-lg text-gray-600">{{ subtitle }}</p>
      @if (lastUpdated) {
        <div class="mt-2 text-sm text-gray-500">
          Last updated: {{ lastUpdated | date:'medium' }}
        </div>
      }
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class QueueHeaderComponent {
  @Input({ required: true }) title!: string;
  @Input({ required: true }) subtitle!: string;
  @Input() lastUpdated?: Date;
}
