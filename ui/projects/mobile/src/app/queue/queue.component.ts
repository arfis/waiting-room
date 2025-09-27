import { Component, signal, inject, OnInit, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { ActivatedRoute } from '@angular/router';
import { CardComponent } from 'ui';

interface QueueEntry {
  entryId: string;
  ticketNumber: string;
  status: string;
  position: number;
  etaMinutes: number;
  canCancel: boolean;
}

@Component({
  selector: 'app-queue',
  standalone: true,
  imports: [CommonModule, CardComponent],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 p-4">
      <div class="max-w-md mx-auto">
        <!-- Header -->
        <div class="text-center mb-6">
          <h1 class="text-2xl font-bold text-gray-900 mb-2">Queue Status</h1>
          <p class="text-gray-600">Track your position in the waiting room</p>
        </div>

        <!-- Loading State -->
        @if (isLoading()) {
          <ui-card title="Loading..." variant="default">
            <div class="text-center py-8">
              <svg class="animate-spin w-8 h-8 text-blue-500 mx-auto mb-4" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <p class="text-gray-600">Loading your queue information...</p>
            </div>
          </ui-card>
        }

        <!-- Queue Information -->
        @if (queueEntry() && !isLoading()) {
          <ui-card title="Your Queue Status" variant="success">
            <div class="text-center">
              <!-- Ticket Number -->
              <div class="mb-6">
                <h3 class="text-lg font-semibold text-gray-700 mb-2">Ticket Number</h3>
                <div class="text-4xl font-mono font-bold text-blue-600 bg-blue-50 px-6 py-3 rounded-lg inline-block">
                  {{ queueEntry()?.ticketNumber }}
                </div>
              </div>

              <!-- Status -->
              <div class="mb-6">
                <h3 class="text-lg font-semibold text-gray-700 mb-2">Status</h3>
                <div class="inline-flex items-center px-4 py-2 rounded-full text-sm font-medium"
                     [class]="getStatusClass(queueEntry()?.status || '')">
                  {{ getStatusText(queueEntry()?.status || '') }}
                </div>
              </div>

              <!-- Position and ETA -->
              @if (queueEntry()?.status === 'WAITING') {
                <div class="mb-6">
                  <div class="grid grid-cols-2 gap-4">
                    <div class="text-center">
                      <h4 class="text-sm font-medium text-gray-600 mb-1">Position</h4>
                      <div class="text-2xl font-bold text-gray-900">{{ queueEntry()?.position }}</div>
                    </div>
                    <div class="text-center">
                      <h4 class="text-sm font-medium text-gray-600 mb-1">Est. Wait Time</h4>
                      <div class="text-2xl font-bold text-gray-900">{{ queueEntry()?.etaMinutes }} min</div>
                    </div>
                  </div>
                </div>
              }

              <!-- Progress Bar -->
              @if (queueEntry()?.status === 'WAITING') {
                <div class="mb-6">
                  <h4 class="text-sm font-medium text-gray-600 mb-2">Queue Progress</h4>
                  <div class="w-full bg-gray-200 rounded-full h-3">
                    <div class="bg-blue-600 h-3 rounded-full transition-all duration-500"
                         [style.width.%]="getProgressPercentage()"></div>
                  </div>
                  <p class="text-xs text-gray-500 mt-1">{{ queueEntry()?.position }} people ahead of you</p>
                </div>
              }

              <!-- Cancel Button -->
              @if (queueEntry()?.canCancel) {
                <button 
                  class="w-full px-4 py-2 bg-red-600 text-white rounded-lg font-medium hover:bg-red-700 transition-colors"
                  (click)="cancelTicket()">
                  Cancel Ticket
                </button>
              }
            </div>
          </ui-card>
        }

        <!-- Error State -->
        @if (error() && !isLoading()) {
          <ui-card title="Error" variant="error">
            <div class="text-center">
              <svg class="w-12 h-12 text-red-500 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
              <h3 class="text-lg font-semibold text-red-800 mb-2">Ticket Not Found</h3>
              <p class="text-red-700 mb-4">{{ error() }}</p>
              <p class="text-sm text-gray-600">Please check your QR code or contact support.</p>
            </div>
          </ui-card>
        }

        <!-- Footer -->
        <div class="text-center mt-8 text-sm text-gray-500">
          <p>Waiting Room Management System</p>
        </div>
      </div>
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class QueueComponent implements OnInit {
  private http = inject(HttpClient);
  private route = inject(ActivatedRoute);

  // State signals
  queueEntry = signal<QueueEntry | null>(null);
  isLoading = signal<boolean>(true);
  error = signal<string | null>(null);

  ngOnInit() {
    // Get QR token from route parameters
    this.route.params.subscribe(params => {
      const qrToken = params['token'];
      if (qrToken) {
        this.loadQueueEntry(qrToken);
      } else {
        this.error.set('Invalid QR code');
        this.isLoading.set(false);
      }
    });
  }

  private loadQueueEntry(qrToken: string) {
    this.isLoading.set(true);
    this.error.set(null);

    this.http.get<QueueEntry>(`http://localhost:8080/queue-entries/token/${qrToken}`)
      .subscribe({
        next: (entry) => {
          console.log('Queue entry loaded:', entry);
          this.queueEntry.set(entry);
          this.isLoading.set(false);
          
          // Auto-refresh every 30 seconds if still waiting
          if (entry.status === 'WAITING') {
            setTimeout(() => this.loadQueueEntry(qrToken), 30000);
          }
        },
        error: (error) => {
          console.error('Failed to load queue entry:', error);
          this.error.set('Failed to load queue information');
          this.isLoading.set(false);
        }
      });
  }

  getStatusClass(status: string): string {
    switch (status) {
      case 'WAITING':
        return 'bg-yellow-100 text-yellow-800';
      case 'CALLED':
        return 'bg-blue-100 text-blue-800';
      case 'IN_SERVICE':
        return 'bg-green-100 text-green-800';
      case 'COMPLETED':
        return 'bg-gray-100 text-gray-800';
      case 'CANCELLED':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  }

  getStatusText(status: string): string {
    switch (status) {
      case 'WAITING':
        return 'Waiting in Queue';
      case 'CALLED':
        return 'Your Turn - Please Come Forward';
      case 'IN_SERVICE':
        return 'Currently Being Served';
      case 'COMPLETED':
        return 'Service Completed';
      case 'CANCELLED':
        return 'Cancelled';
      default:
        return status;
    }
  }

  getProgressPercentage(): number {
    const entry = this.queueEntry();
    if (!entry || entry.position <= 1) return 100;
    
    // Simple progress calculation - you might want to make this more sophisticated
    const maxPosition = Math.max(entry.position + 5, 10); // Assume at least 10 people max
    return Math.max(0, Math.min(100, ((maxPosition - entry.position) / maxPosition) * 100));
  }

  cancelTicket() {
    // TODO: Implement cancel functionality
    console.log('Cancel ticket requested');
  }
}
