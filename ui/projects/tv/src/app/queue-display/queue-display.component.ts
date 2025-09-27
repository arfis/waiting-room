import { Component, signal, inject, OnInit, OnDestroy, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { CardComponent } from 'ui';

interface QueueEntry {
  id: string;
  waitingRoomId: string;
  ticketNumber: string;
  status: string;
  position: number;
  createdAt: string;
  cardData?: {
    idNumber: string;
    firstName: string;
    lastName: string;
    source: string;
  };
}

@Component({
  selector: 'app-queue-display',
  standalone: true,
  imports: [CommonModule, CardComponent],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-blue-900 to-indigo-900 text-white">
      <!-- Header -->
      <div class="text-center py-8">
        <h1 class="text-6xl font-bold mb-4">Waiting Room Queue</h1>
        <p class="text-2xl text-blue-200">Triage Room 1</p>
        <div class="mt-4 text-lg text-blue-300">
          Last updated: {{ lastUpdated() | date:'medium' }}
        </div>
      </div>

      <!-- Current Queue -->
      <div class="max-w-6xl mx-auto px-8">
        <!-- Currently Being Served -->
        @if (currentEntry()) {
          <div class="mb-12">
            <h2 class="text-4xl font-bold text-center mb-8 text-yellow-400">Currently Being Served</h2>
            <ui-card title="Now Serving" variant="success">
              <div class="text-center">
                <div class="text-8xl font-mono font-bold text-green-600 mb-4">
                  {{ currentEntry()?.ticketNumber }}
                </div>
                @if (currentEntry()?.cardData) {
                  <div class="text-2xl text-gray-700">
                    {{ currentEntry()?.cardData?.firstName }} {{ currentEntry()?.cardData?.lastName }}
                  </div>
                }
              </div>
            </ui-card>
          </div>
        }

        <!-- Waiting Queue -->
        <div class="mb-8">
          <h2 class="text-4xl font-bold text-center mb-8 text-blue-200">Waiting Queue</h2>
          
          @if (waitingEntries().length === 0) {
            <ui-card title="No one waiting" variant="default">
              <div class="text-center py-12">
                <svg class="w-24 h-24 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
                <p class="text-2xl text-gray-600">No one is currently waiting</p>
              </div>
            </ui-card>
          } @else {
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              @for (entry of waitingEntries(); track entry.id) {
                <ui-card title="Ticket {{ entry.ticketNumber }}" variant="default">
                  <div class="text-center">
                    <div class="text-4xl font-mono font-bold text-blue-600 mb-3">
                      {{ entry.ticketNumber }}
                    </div>
                    <div class="text-lg text-gray-600 mb-2">
                      Position: <span class="font-bold">{{ entry.position }}</span>
                    </div>
                    @if (entry.cardData) {
                      <div class="text-sm text-gray-500">
                        {{ entry.cardData.firstName }} {{ entry.cardData.lastName }}
                      </div>
                    }
                    <div class="mt-3">
                      <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-yellow-100 text-yellow-800">
                        Waiting
                      </span>
                    </div>
                  </div>
                </ui-card>
              }
            </div>
          }
        </div>

        <!-- Statistics -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <ui-card title="Total Waiting" variant="default">
            <div class="text-center">
              <div class="text-5xl font-bold text-blue-600">{{ waitingEntries().length }}</div>
              <p class="text-gray-600 mt-2">People in queue</p>
            </div>
          </ui-card>
          
          <ui-card title="Estimated Wait" variant="default">
            <div class="text-center">
              <div class="text-5xl font-bold text-green-600">{{ estimatedWaitTime() }}</div>
              <p class="text-gray-600 mt-2">Minutes</p>
            </div>
          </ui-card>
          
          <ui-card title="Average Service Time" variant="default">
            <div class="text-center">
              <div class="text-5xl font-bold text-purple-600">5</div>
              <p class="text-gray-600 mt-2">Minutes per person</p>
            </div>
          </ui-card>
        </div>
      </div>

      <!-- Footer -->
      <div class="text-center py-8 text-blue-300">
        <p class="text-xl">Waiting Room Management System</p>
        <p class="text-sm mt-2">Auto-refreshing every 10 seconds</p>
      </div>
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class QueueDisplayComponent implements OnInit, OnDestroy {
  private http = inject(HttpClient);
  private refreshInterval?: number;

  // State signals
  queueEntries = signal<QueueEntry[]>([]);
  lastUpdated = signal<Date>(new Date());
  isLoading = signal<boolean>(true);

  // Computed signals
  currentEntry = signal<QueueEntry | null>(null);
  waitingEntries = signal<QueueEntry[]>([]);

  ngOnInit() {
    this.loadQueue();
    
    // Auto-refresh every 10 seconds
    this.refreshInterval = window.setInterval(() => {
      this.loadQueue();
    }, 10000);
  }

  ngOnDestroy() {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval);
    }
  }

  private loadQueue() {
    this.isLoading.set(true);

    this.http.get<QueueEntry[]>('http://localhost:8080/api/waiting-rooms/triage-1/queue')
      .subscribe({
        next: (entries) => {
          console.log('Queue entries loaded:', entries);
          this.queueEntries.set(entries);
          this.lastUpdated.set(new Date());
          this.isLoading.set(false);
          
          // Update computed signals
          this.updateComputedSignals(entries);
        },
        error: (error) => {
          console.error('Failed to load queue:', error);
          this.isLoading.set(false);
        }
      });
  }

  private updateComputedSignals(entries: QueueEntry[]) {
    // Find currently being served (CALLED or IN_SERVICE)
    const current = entries.find(entry => 
      entry.status === 'CALLED' || entry.status === 'IN_SERVICE'
    );
    this.currentEntry.set(current || null);

    // Get waiting entries, sorted by position
    const waiting = entries
      .filter(entry => entry.status === 'WAITING')
      .sort((a, b) => a.position - b.position);
    this.waitingEntries.set(waiting);
  }

  estimatedWaitTime(): number {
    const waiting = this.waitingEntries();
    // Simple calculation: 5 minutes per person
    return waiting.length * 5;
  }
}
