import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Api, QueueEntry } from 'api-client';
import { CardComponent } from 'ui';


@Component({
  selector: 'app-backoffice',
  standalone: true,
  imports: [CommonModule, CardComponent],
  template: `
  <div class="min-h-screen bg-slate-50 p-6">
    <ui-card>
      <h1 class="text-xl font-semibold mb-4">Backoffice</h1>
      <div class="flex gap-2">
        <button class="px-4 py-2 rounded bg-black text-white" (click)="callNext()">Call Next</button>
        <button class="px-4 py-2 rounded bg-gray-300" disabled>Skip</button>
        <button class="px-4 py-2 rounded bg-gray-300" disabled>Cancel</button>
      </div>
      <div class="mt-4" *ngIf="last() as l">
        <p>Called ticket: <strong>{{l.ticketNumber}}</strong></p>
      </div>
    </ui-card>
  </div>
  `
})
export class BackofficeComponent {
  private api = new Api();
  last = signal<QueueEntry|null>(null);
  async callNext(){ this.last.set(await this.api.waitingRoom.postWaitingRoomsRoomIdNext('room-1')); }
}