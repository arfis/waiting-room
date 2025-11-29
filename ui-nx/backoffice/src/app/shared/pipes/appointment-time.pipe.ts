import { Pipe, PipeTransform } from '@angular/core';

export interface AppointmentTimeInfo {
  formattedTime: string;
  diffMinutes: number;
  status: 'early' | 'on-time' | 'late';
  displayText: string;
}

/**
 * Pipe to format appointment time and calculate difference from current time
 * Returns formatted time, difference in minutes, and status (early/on-time/late)
 */
@Pipe({
  name: 'appointmentTime',
  standalone: true,
})
export class AppointmentTimePipe implements PipeTransform {
  transform(appointmentTime: string | Date | null | undefined): AppointmentTimeInfo | null {
    if (!appointmentTime) {
      return null;
    }

    const now = new Date();
    const appointment = new Date(appointmentTime);

    // Calculate difference in minutes (negative = appointment in the past/late)
    const diffMs = appointment.getTime() - now.getTime();
    const diffMinutes = Math.round(diffMs / 60000);

    // Format the appointment time (HH:MM)
    const hours = appointment.getHours().toString().padStart(2, '0');
    const minutes = appointment.getMinutes().toString().padStart(2, '0');
    const formattedTime = `${hours}:${minutes}`;

    // Determine status based on time difference
    // early: more than 15 minutes before appointment
    // on-time: within 15 minutes before or 5 minutes after
    // late: more than 5 minutes after appointment time
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
  }
}
