import { Component, Input, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from 'ui';

@Component({
  selector: 'app-queue-statistics',
  standalone: true,
  imports: [CommonModule, CardComponent],
  template: `
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
      <ui-card title="Total Waiting" variant="default">
        <div class="text-center">
          <div class="text-3xl font-bold text-blue-600">{{ totalWaiting }}</div>
          <p class="text-gray-600 mt-1">People in queue</p>
        </div>
      </ui-card>
      
      <ui-card title="Estimated Wait" variant="default">
        <div class="text-center">
          <div class="text-3xl font-bold text-orange-600">{{ estimatedWaitTime }}</div>
          <p class="text-gray-600 mt-1">Minutes</p>
        </div>
      </ui-card>
      
      <ui-card title="Average Service" variant="default">
        <div class="text-center">
          <div class="text-3xl font-bold text-purple-600">{{ averageServiceTime }}</div>
          <p class="text-gray-600 mt-1">Minutes per person</p>
        </div>
      </ui-card>
      
      <ui-card title="Queue Status" variant="default">
        <div class="text-center">
          <div class="text-3xl font-bold" [class]="totalWaiting > 0 ? 'text-red-600' : 'text-green-600'">
            {{ totalWaiting > 0 ? 'Active' : 'Empty' }}
          </div>
          <p class="text-gray-600 mt-1">Current status</p>
        </div>
      </ui-card>
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class QueueStatisticsComponent {
  @Input({ required: true }) totalWaiting!: number;
  @Input({ required: true }) estimatedWaitTime!: number;
  @Input() averageServiceTime = 5;
}
