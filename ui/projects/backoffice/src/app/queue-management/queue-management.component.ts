import { Component, inject, OnInit, OnDestroy, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { QueueStateService } from '../core/services/queue-state.service';
import { QueueHeaderComponent } from './components/queue-header/queue-header.component';
import { QueueActionsComponent } from './components/queue-actions/queue-actions.component';
import { CurrentEntryCardComponent } from './components/current-entry-card/current-entry-card.component';
import { QueueStatisticsComponent } from './components/queue-statistics/queue-statistics.component';
import { WaitingQueueListComponent } from './components/waiting-queue-list/waiting-queue-list.component';
import { ActivityLogComponent } from './components/activity-log/activity-log.component';
import { WebSocketQueueEntry } from 'api-client';

@Component({
  selector: 'app-queue-management',
  standalone: true,
  imports: [
    CommonModule,
    QueueHeaderComponent,
    QueueActionsComponent,
    CurrentEntryCardComponent,
    QueueStatisticsComponent,
    WaitingQueueListComponent,
    ActivityLogComponent
  ],
  templateUrl: './queue-management.component.html',
  styleUrl: './queue-management.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class QueueManagementComponent implements OnInit, OnDestroy {
  private readonly queueState = inject(QueueStateService);
  private readonly roomId = 'triage-1';

  // Expose state to template
  protected readonly currentEntry = this.queueState.currentEntry;
  protected readonly waitingEntries = this.queueState.waitingEntries;
  protected readonly isLoading = this.queueState.isLoading;
  protected readonly lastUpdated = this.queueState.lastUpdated;
  protected readonly recentActivity = this.queueState.recentActivity;
  protected readonly estimatedWaitTime = this.queueState.estimatedWaitTime;
  protected readonly isConnected = this.queueState.isConnected;
  protected readonly error = this.queueState.wsError;

  async ngOnInit(): Promise<void> {
    await this.queueState.initialize(this.roomId);
  }

  ngOnDestroy(): void {
    this.queueState.disconnect();
  }

  protected onCallNext(): void {
    this.queueState.callNext(this.roomId);
  }

  protected onFinishCurrent(): void {
    this.queueState.finishCurrent(this.roomId);
  }

  protected onRefresh(): void {
    this.queueState.refreshQueue(this.roomId);
  }

  protected onCallSpecificEntry(entry: WebSocketQueueEntry): void {
    // For now, just call the next person
    // In a real implementation, you might want a specific endpoint for calling a particular entry
    this.queueState.callNext(this.roomId);
  }

  protected getStatusText(status: string): string {
    return this.queueState.getStatusText(status);
  }
}
