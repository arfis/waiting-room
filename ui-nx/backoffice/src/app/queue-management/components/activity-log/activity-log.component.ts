import { Component, Input, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from '@waiting-room/primeng-components';
import { ActivityLog } from '../../../core/services/queue-state.service';
import { TranslatePipe } from '../../../../../../src/lib/i18n';

@Component({
  selector: 'app-activity-log',
  standalone: true,
  imports: [CommonModule, CardComponent, TranslatePipe],
  template: `
    <div class="mb-8">
      <h2 class="text-2xl font-bold text-gray-900 mb-4">{{ 'backoffice.recentActivity' | translate }}</h2>
      <ui-card [title]="'backoffice.activityLog' | translate" variant="default">
        <div class="space-y-3">
          @if (activities.length === 0) {
            <p class="text-gray-500 text-center py-4">{{ 'backoffice.noRecentActivity' | translate }}</p>
          } @else {
            @for (activity of activities; track activity.id) {
              <div class="flex items-center justify-between py-2 border-b border-gray-100 last:border-b-0">
                <div>
                  <span class="font-medium">{{ activity.ticketNumber }}</span>
                  <span class="text-gray-600 ml-2">{{ activity.action }}</span>
                </div>
                <div class="text-sm text-gray-500">
                  {{ activity.timestamp | date:'short' }}
                </div>
              </div>
            }
          }
        </div>
      </ui-card>
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ActivityLogComponent {
  @Input({ required: true }) activities!: ActivityLog[];
}
