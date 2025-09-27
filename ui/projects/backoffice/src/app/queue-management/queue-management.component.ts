import { Component, signal, inject, OnInit, OnDestroy, ChangeDetectionStrategy, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { CardComponent } from 'ui';
import { QueueWebSocketService, WebSocketQueueEntry } from 'api-client';

// Using WebSocketQueueEntry from api-client

@Component({
  selector: 'app-queue-management',
  standalone: true,
  imports: [CommonModule, CardComponent],
  template: `
    <div class="min-h-screen bg-gray-50 p-6">
      <div class="max-w-7xl mx-auto">
        <!-- Header -->
        <div class="mb-8">
          <h1 class="text-4xl font-bold text-gray-900 mb-2">Queue Management</h1>
          <p class="text-lg text-gray-600">Manage the waiting room queue for Triage Room 1</p>
          <div class="mt-2 text-sm text-gray-500">
            Last updated: {{ lastUpdated() | date:'medium' }}
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="mb-8 flex gap-4">
          <button 
            class="px-6 py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            (click)="callNext()"
            [disabled]="isLoading() || waitingEntries().length === 0">
            @if (isLoading()) {
              <span class="flex items-center gap-2">
                <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Processing...
              </span>
            } @else {
              Call Next Person
            }
          </button>
          
          <button 
            class="px-6 py-3 bg-gray-600 text-white rounded-lg font-medium hover:bg-gray-700 transition-colors"
            (click)="refreshQueue()"
            [disabled]="isLoading()">
            Refresh Queue
          </button>
        </div>

        <!-- Currently Being Served -->
        @if (currentEntry()) {
          <div class="mb-8">
            <ui-card title="Currently Being Served" variant="success">
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-3xl font-mono font-bold text-green-600 mb-2">
                    {{ currentEntry()?.ticketNumber }}
                  </div>
                  @if (currentEntry()?.cardData) {
                    <div class="text-lg text-gray-700">
                      {{ currentEntry()?.cardData?.firstName }} {{ currentEntry()?.cardData?.lastName }}
                    </div>
                    <div class="text-sm text-gray-500">
                      ID: {{ currentEntry()?.cardData?.idNumber }}
                    </div>
                  }
                </div>
                <div class="text-right">
                  <div class="text-sm text-gray-500 mb-1">Status</div>
                  <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
                    {{ getStatusText(currentEntry()?.status || '') }}
                  </span>
                </div>
              </div>
            </ui-card>
          </div>
        }

        <!-- Queue Statistics -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <ui-card title="Total Waiting" variant="default">
            <div class="text-center">
              <div class="text-3xl font-bold text-blue-600">{{ waitingEntries().length }}</div>
              <p class="text-gray-600 mt-1">People in queue</p>
            </div>
          </ui-card>
          
          <ui-card title="Estimated Wait" variant="default">
            <div class="text-center">
              <div class="text-3xl font-bold text-orange-600">{{ estimatedWaitTime() }}</div>
              <p class="text-gray-600 mt-1">Minutes</p>
            </div>
          </ui-card>
          
          <ui-card title="Average Service" variant="default">
            <div class="text-center">
              <div class="text-3xl font-bold text-purple-600">5</div>
              <p class="text-gray-600 mt-1">Minutes per person</p>
            </div>
          </ui-card>
          
          <ui-card title="Queue Status" variant="default">
            <div class="text-center">
              <div class="text-3xl font-bold" [class]="waitingEntries().length > 0 ? 'text-red-600' : 'text-green-600'">
                {{ waitingEntries().length > 0 ? 'Active' : 'Empty' }}
              </div>
              <p class="text-gray-600 mt-1">Current status</p>
            </div>
          </ui-card>
        </div>

        <!-- Waiting Queue -->
        <div class="mb-8">
          <h2 class="text-2xl font-bold text-gray-900 mb-4">Waiting Queue</h2>
          
          @if (waitingEntries().length === 0) {
            <ui-card title="No one waiting" variant="default">
              <div class="text-center py-12">
                <svg class="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
                <p class="text-xl text-gray-600">No one is currently waiting</p>
                <p class="text-sm text-gray-500 mt-2">The queue is empty</p>
              </div>
            </ui-card>
          } @else {
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              @for (entry of waitingEntries(); track entry.id) {
                <ui-card title="Ticket {{ entry.ticketNumber }}" variant="default">
                  <div class="space-y-3">
                    <div class="flex items-center justify-between">
                      <div class="text-2xl font-mono font-bold text-blue-600">
                        {{ entry.ticketNumber }}
                      </div>
                      <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                        Position {{ entry.position }}
                      </span>
                    </div>
                    
                    @if (entry.cardData) {
                      <div class="text-sm text-gray-600">
                        <div class="font-medium">{{ entry.cardData.firstName }} {{ entry.cardData.lastName }}</div>
                        <div class="text-xs text-gray-500">ID: {{ entry.cardData.idNumber }}</div>
                      </div>
                    }
                    
                    <div class="text-xs text-gray-500">
                      Joined: {{ entry.createdAt | date:'short' }}
                    </div>
                    
                    <div class="pt-2 border-t">
                      <button 
                        class="w-full px-3 py-2 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700 transition-colors"
                        (click)="callSpecificEntry(entry)">
                        Call This Person
                      </button>
                    </div>
                  </div>
                </ui-card>
              }
            </div>
          }
        </div>

        <!-- Recent Activity -->
        <div class="mb-8">
          <h2 class="text-2xl font-bold text-gray-900 mb-4">Recent Activity</h2>
          <ui-card title="Activity Log" variant="default">
            <div class="space-y-3">
              @if (recentActivity().length === 0) {
                <p class="text-gray-500 text-center py-4">No recent activity</p>
              } @else {
                @for (activity of recentActivity(); track activity.id) {
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
      </div>
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class QueueManagementComponent implements OnInit, OnDestroy {
  private http = inject(HttpClient);
  private queueWebSocket = inject(QueueWebSocketService);
  lastUpdated = signal<Date>(new Date());

  // Use WebSocket service signals directly
  queueEntries = this.queueWebSocket.queueEntries;
  isConnected = this.queueWebSocket.isConnected;
  error = this.queueWebSocket.error;
  isLoading = signal<boolean>(false); // Only for API calls
  recentActivity = signal<Array<{id: string, ticketNumber: string, action: string, timestamp: Date}>>([]);

  // Computed signals using WebSocket service methods
  currentEntry = signal<WebSocketQueueEntry | null>(null);
  waitingEntries = signal<WebSocketQueueEntry[]>([]);

  constructor() {
    // Update computed signals when queue entries change
    // Use effect to watch for changes in the signal - must be in constructor
    effect(() => {
      const entries = this.queueEntries();
      this.lastUpdated.set(new Date());
      this.updateComputedSignals(entries);
    });
  }

  async ngOnInit() {
    // Initialize with HTTP API first, then connect WebSocket
    await this.queueWebSocket.initialize('triage-1');
  }

  ngOnDestroy() {
    // Disconnect from WebSocket
    this.queueWebSocket.disconnect();
  }

  private loadQueue() {
    this.isLoading.set(true);

    this.http.get<WebSocketQueueEntry[]>('http://localhost:8080/api/waiting-rooms/triage-1/queue')
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

  private updateComputedSignals(entries: WebSocketQueueEntry[]) {
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

  callNext() {
    this.isLoading.set(true);
    
    this.http.post<any>('http://localhost:8080/waiting-rooms/triage-1/next', {})
      .subscribe({
        next: (response) => {
          console.log('Next person called:', response);
          this.addActivity(response.ticketNumber, 'Called to service');
          this.loadQueue(); // Refresh the queue
        },
        error: (error) => {
          console.error('Failed to call next person:', error);
          this.isLoading.set(false);
        }
      });
  }

  callSpecificEntry(entry: WebSocketQueueEntry) {
    // For now, just call the next person
    // In a real implementation, you might want a specific endpoint for calling a particular entry
    this.callNext();
  }

  refreshQueue() {
    this.loadQueue();
  }

  estimatedWaitTime(): number {
    const waiting = this.waitingEntries();
    // Simple calculation: 5 minutes per person
    return waiting.length * 5;
  }

  getStatusText(status: string): string {
    switch (status) {
      case 'WAITING':
        return 'Waiting';
      case 'CALLED':
        return 'Called';
      case 'IN_SERVICE':
        return 'In Service';
      case 'COMPLETED':
        return 'Completed';
      case 'CANCELLED':
        return 'Cancelled';
      default:
        return status;
    }
  }

  private addActivity(ticketNumber: string, action: string) {
    const activity = {
      id: Date.now().toString(),
      ticketNumber,
      action,
      timestamp: new Date()
    };
    
    this.recentActivity.update(activities => {
      const newActivities = [activity, ...activities].slice(0, 10); // Keep only last 10
      return newActivities;
    });
  }
}
