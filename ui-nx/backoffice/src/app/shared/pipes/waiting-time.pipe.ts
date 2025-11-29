import { Pipe, PipeTransform } from '@angular/core';

/**
 * Pipe to calculate and format waiting time duration
 * Shows how long someone has been waiting since their createdAt timestamp
 */
@Pipe({
  name: 'waitingTime',
  standalone: true,
})
export class WaitingTimePipe implements PipeTransform {
  transform(createdAt: string | Date): string {
    if (!createdAt) {
      return '';
    }

    const now = new Date();
    const created = new Date(createdAt);
    const diffMs = now.getTime() - created.getTime();

    if (diffMs < 0) {
      return '0m';
    }

    const diffMinutes = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMinutes / 60);
    const remainingMinutes = diffMinutes % 60;

    if (diffHours > 0) {
      return `${diffHours}h ${remainingMinutes}m`;
    } else {
      return `${diffMinutes}m`;
    }
  }
}
