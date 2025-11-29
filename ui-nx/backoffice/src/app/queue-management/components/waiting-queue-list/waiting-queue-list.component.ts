import { Component, Input, Output, EventEmitter, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { WebSocketQueueEntry } from '@waiting-room/api-client';
import { TranslatePipe } from '../../../../../../src/lib/i18n';
import { QueueList } from './container/queue-list';

@Component({
  selector: 'app-waiting-queue-list',
  standalone: true,
  imports: [CommonModule, TranslatePipe, QueueList],
  templateUrl: './waiting-queue-list.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class WaitingQueueListComponent {
  @Input({ required: true }) entries!: WebSocketQueueEntry[];
  @Output() callEntry = new EventEmitter<WebSocketQueueEntry>();
}
