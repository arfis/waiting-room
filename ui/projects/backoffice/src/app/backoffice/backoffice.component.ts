import { Component, signal, inject, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent } from 'ui';
import { QueueApiService, CallNextResponse } from '../core/services/queue-api.service';

@Component({
  selector: 'app-backoffice',
  standalone: true,
  imports: [CommonModule, CardComponent],
  templateUrl: './backoffice.component.html',
  styleUrl: './backoffice.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class BackofficeComponent {
  private readonly queueApiService = inject(QueueApiService);
  private readonly roomId = 'room-1';

  protected readonly lastCalled = signal<CallNextResponse | null>(null);
  protected readonly isLoading = signal<boolean>(false);
  protected readonly error = signal<string | null>(null);

  protected callNext(): void {
    this.isLoading.set(true);
    this.error.set(null);

    this.queueApiService.callNext(this.roomId).subscribe({
      next: (response) => {
        this.lastCalled.set(response);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to call next:', err);
        this.error.set('Failed to call next person. Please try again.');
        this.isLoading.set(false);
      }
    });
  }
}
