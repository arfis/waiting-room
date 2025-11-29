import { Component, input, signal, computed, DestroyRef, inject } from '@angular/core';
import { DatePipe } from '@angular/common';
import { TranslatePipe } from '../../../../../../../src/lib/i18n';
import { WebSocketQueueEntry } from '@waiting-room/api-client';

export interface AppointmentTimeInfo {
  formattedTime: string;
  diffMinutes: number;
  status: 'early' | 'on-time' | 'late';
  displayText: string;
}

// needs refactor not to use methods in HTML!
@Component({
  selector: 'app-queue-item',
  imports: [DatePipe, TranslatePipe],
  templateUrl: './queue-item.html',
  styleUrl: './queue-item.scss',
})
export class QueueItem {
  private readonly destroyRef = inject(DestroyRef);
  readonly entry = input.required<WebSocketQueueEntry>();

  // Signal that updates every minute to trigger recalculation
  private readonly currentTime = signal(new Date());

  // Computed signal for waiting time that auto-updates
  readonly waitingTime = computed(() => {
    const now = this.currentTime(); // Subscribe to time updates
    const createdAt = this.entry().createdAt;

    if (!createdAt) {
      return 'N/A';
    }

    const created = new Date(createdAt);

    // Check if date parsing failed
    if (isNaN(created.getTime())) {
      console.error('[WaitingTime] Invalid date format:', createdAt);
      return 'Invalid';
    }

    const diffMs = now.getTime() - created.getTime();

    if (diffMs < 0) {
      console.warn('[WaitingTime] Future createdAt detected:', createdAt);
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
  });

  // Computed signal for appointment time info that auto-updates
  readonly appointmentInfo = computed((): AppointmentTimeInfo | null => {
    const now = this.currentTime(); // Subscribe to time updates
    const appointmentTime = this.entry().appointmentTime;

    if (!appointmentTime) {
      return null;
    }

    const appointment = new Date(appointmentTime);

    // Calculate difference in minutes (negative = appointment in the past/late)
    const diffMs = appointment.getTime() - now.getTime();
    const diffMinutes = Math.round(diffMs / 60000);

    // Format the appointment time (HH:MM)
    const hours = appointment.getHours().toString().padStart(2, '0');
    const minutes = appointment.getMinutes().toString().padStart(2, '0');
    const formattedTime = `${hours}:${minutes}`;

    // Determine status based on time difference
    let status: 'early' | 'on-time' | 'late';
    let displayText: string;

    if (diffMinutes > 15) {
      status = 'early';
      const absMinutes = Math.abs(diffMinutes);
      const hours = Math.floor(absMinutes / 60);
      const mins = absMinutes % 60;
      if (hours > 0) {
        displayText = `${hours}h ${mins}m early`;
      } else {
        displayText = `${mins}m early`;
      }
    } else if (diffMinutes >= -5) {
      status = 'on-time';
      displayText = 'On time';
    } else {
      status = 'late';
      const absMinutes = Math.abs(diffMinutes);
      const hours = Math.floor(absMinutes / 60);
      const mins = absMinutes % 60;
      if (hours > 0) {
        displayText = `${hours}h ${mins}m late`;
      } else {
        displayText = `${mins}m late`;
      }
    }

    return {
      formattedTime,
      diffMinutes,
      status,
      displayText,
    };
  });

  constructor() {
    // Update time every minute (60 seconds)
    const intervalId = setInterval(() => {
      this.currentTime.set(new Date());
    }, 60000);

    // Cleanup interval on component destroy
    this.destroyRef.onDestroy(() => {
      clearInterval(intervalId);
    });
  }

  // Safe method to get card data field
  getCardDataField(
    entry: WebSocketQueueEntry,
    field: 'idNumber' | 'firstName' | 'lastName'
  ): string {
    if (!entry?.cardData) {
      return '';
    }

    // If cardData is a string, try to parse it
    let cardData = entry.cardData;
    if (typeof entry.cardData === 'string') {
      try {
        cardData = JSON.parse(entry.cardData as any);
      } catch (e) {
        console.error('[WaitingQueueList] Failed to parse cardData:', e);
        return '';
      }
    }

    // Handle both snake_case and camelCase
    const fieldMap = {
      idNumber: ['idNumber', 'id_number'],
      firstName: ['firstName', 'first_name'],
      lastName: ['lastName', 'last_name'],
    };

    const possibleFields = fieldMap[field];
    for (const possibleField of possibleFields) {
      if (
        cardData &&
        typeof cardData === 'object' &&
        possibleField in cardData
      ) {
        const value = (cardData as any)[possibleField];
        return value || '';
      }
    }

    return '';
  }

  // Get appropriate CSS class for priority symbols
  getSymbolClass(symbol: string): string {
    const upperSymbol = symbol.toUpperCase();
    switch (upperSymbol) {
      case 'STATIM':
        return 'bg-red-600 text-white';
      case 'VIP':
        return 'bg-purple-600 text-white';
      case 'IMMOBILE':
        return 'bg-orange-600 text-white';
      default:
        return 'bg-gray-600 text-white';
    }
  }

  // Get appropriate CSS class for appointment status
  getAppointmentStatusClass(status: 'early' | 'on-time' | 'late'): string {
    switch (status) {
      case 'early':
        return 'bg-blue-100 text-blue-700 border-blue-200';
      case 'on-time':
        return 'bg-green-100 text-green-700 border-green-200';
      case 'late':
        return 'bg-red-100 text-red-700 border-red-200';
    }
  }

  // Calculate waiting time status based on duration
  private getWaitingTimeStatus(): 'good' | 'moderate' | 'long' {
    const createdAt = this.entry().createdAt;
    if (!createdAt) return 'good';

    const now = new Date();
    const created = new Date(createdAt);
    const diffMinutes = Math.floor((now.getTime() - created.getTime()) / 60000);

    if (diffMinutes < 15) return 'good';
    if (diffMinutes < 30) return 'moderate';
    return 'long';
  }

  // Get background color class for waiting time card
  getWaitingTimeColorClass(): string {
    const status = this.getWaitingTimeStatus();
    switch (status) {
      case 'good':
        return 'from-green-50 to-emerald-50 border-green-200';
      case 'moderate':
        return 'from-yellow-50 to-amber-50 border-yellow-200';
      case 'long':
        return 'from-red-50 to-rose-50 border-red-200';
    }
  }

  // Get icon background class for waiting time
  getWaitingTimeIconClass(): string {
    const status = this.getWaitingTimeStatus();
    switch (status) {
      case 'good':
        return 'bg-green-500';
      case 'moderate':
        return 'bg-yellow-500';
      case 'long':
        return 'bg-red-500';
    }
  }

  // Get label text color class for waiting time
  getWaitingTimeLabelClass(): string {
    const status = this.getWaitingTimeStatus();
    switch (status) {
      case 'good':
        return 'text-green-600';
      case 'moderate':
        return 'text-yellow-600';
      case 'long':
        return 'text-red-600';
    }
  }

  // Get value text color class for waiting time
  getWaitingTimeTextClass(): string {
    const status = this.getWaitingTimeStatus();
    switch (status) {
      case 'good':
        return 'text-green-900';
      case 'moderate':
        return 'text-yellow-900';
      case 'long':
        return 'text-red-900';
    }
  }
}
