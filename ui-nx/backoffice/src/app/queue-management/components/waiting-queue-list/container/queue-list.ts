import { Component, input, output } from '@angular/core';
import { WebSocketQueueEntry } from '@waiting-room/api-client';
import { QueueItem } from '../components/queue-item';
import { TranslatePipe } from '../../../../../../../src/lib/i18n';

@Component({
  selector: 'app-queue-list',
  imports: [QueueItem, TranslatePipe],
  templateUrl: './queue-list.html',
  styleUrl: './queue-list.css',
})
export class QueueList {
  readonly entries = input.required<WebSocketQueueEntry[]>();
  readonly callEntry = output<WebSocketQueueEntry>();
}
