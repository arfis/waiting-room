import { Component, inject, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CardReaderStateService } from '../core/services/card-reader-state.service';
import { PatientInformation, KioskApiService } from '../core/services/kiosk-api.service';

@Component({
  selector: 'app-debug-panel',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './debug-panel.component.html',
  styleUrls: ['./debug-panel.component.scss']
})
export class DebugPanelComponent {
  private readonly cardReaderState = inject(CardReaderStateService);
  private readonly kioskApiService = inject(KioskApiService);

  // Debug mode state
  protected readonly debugMode = this.cardReaderState.debugMode;

  // Patient information form fields
  protected readonly symbols = signal<string>('');
  protected readonly appointmentTime = signal<string>('');
  protected readonly age = signal<number | null>(null);
  protected readonly manualOverride = signal<number | null>(null);

  // Manual submission fields
  protected readonly manualIdNumber = signal<string>('');
  protected readonly manualServiceId = signal<string>('');
  protected readonly manualDuration = signal<number | null>(null);
  protected readonly isSubmitting = signal<boolean>(false);

  constructor() {
    // Update patient information whenever any field changes
    effect(() => {
      if (this.debugMode()) {
        const info: PatientInformation = {};

        // Parse symbols (comma-separated)
        const symbolsStr = this.symbols();
        if (symbolsStr && symbolsStr.trim()) {
          info.symbols = symbolsStr.split(',').map(s => s.trim()).filter(s => s.length > 0);
        }

        // Set appointment time if provided
        const appointmentStr = this.appointmentTime();
        if (appointmentStr) {
          info.appointmentTime = appointmentStr;
        }

        // Set age if provided
        const ageValue = this.age();
        if (ageValue !== null && !isNaN(ageValue)) {
          info.age = ageValue;
        }

        // Set manual override if provided
        const overrideValue = this.manualOverride();
        if (overrideValue !== null && !isNaN(overrideValue)) {
          info.manualOverride = overrideValue;
        }

        this.cardReaderState.setPatientInformation(Object.keys(info).length > 0 ? info : null);
      } else {
        this.cardReaderState.setPatientInformation(null);
      }
    });
  }

  protected toggleDebugMode(): void {
    this.cardReaderState.toggleDebugMode();

    // Clear fields when debug mode is turned off
    if (!this.debugMode()) {
      this.clearFields();
    }
  }

  protected clearFields(): void {
    this.symbols.set('');
    this.appointmentTime.set('');
    this.age.set(null);
    this.manualOverride.set(null);
  }

  protected onSymbolsChange(value: string): void {
    this.symbols.set(value);
  }

  protected onAppointmentTimeChange(value: string): void {
    this.appointmentTime.set(value);
  }

  protected onAgeChange(value: string): void {
    const numValue = value ? parseInt(value, 10) : null;
    this.age.set(numValue);
  }

  protected onManualOverrideChange(value: string): void {
    const numValue = value ? parseFloat(value) : null;
    this.manualOverride.set(numValue);
  }

  protected onManualIdNumberChange(value: string): void {
    this.manualIdNumber.set(value);
  }

  protected onManualServiceIdChange(value: string): void {
    this.manualServiceId.set(value);
  }

  protected onManualDurationChange(value: string): void {
    const numValue = value ? parseInt(value, 10) : null;
    this.manualDuration.set(numValue);
  }

  protected generateRandomId(): void {
    // Generate a random ID number (9 digits)
    const randomId = Math.floor(100000000 + Math.random() * 900000000).toString();
    this.manualIdNumber.set(randomId);
  }

  protected submitManualEntry(): void {
    const idNumber = this.manualIdNumber();

    if (!idNumber || idNumber.trim() === '') {
      alert('ID Number is required');
      return;
    }

    this.isSubmitting.set(true);

    // Build patient information from debug panel fields
    const patientInfo: PatientInformation = {};

    const symbolsStr = this.symbols();
    if (symbolsStr && symbolsStr.trim()) {
      patientInfo.symbols = symbolsStr.split(',').map(s => s.trim()).filter(s => s.length > 0);
    }

    const appointmentStr = this.appointmentTime();
    if (appointmentStr) {
      patientInfo.appointmentTime = appointmentStr;
    }

    const ageValue = this.age();
    if (ageValue !== null && !isNaN(ageValue)) {
      patientInfo.age = ageValue;
    }

    const overrideValue = this.manualOverride();
    if (overrideValue !== null && !isNaN(overrideValue)) {
      patientInfo.manualOverride = overrideValue;
    }

    // Call the API
    const roomId = 'triage-1'; // Default room ID
    const serviceId = this.manualServiceId() || undefined;
    const duration = this.manualDuration() || undefined;

    this.kioskApiService.generateTicket(
      roomId,
      idNumber,
      serviceId,
      duration,
      Object.keys(patientInfo).length > 0 ? patientInfo : undefined
    ).subscribe({
      next: (response) => {
        console.log('Manual entry submitted successfully:', response);
        alert(`Ticket generated: ${response.ticketNumber}`);
        this.isSubmitting.set(false);

        // Clear the form and generate a new random ID for next entry
        this.clearManualSubmissionFields();
        this.generateRandomId();
      },
      error: (error) => {
        console.error('Failed to submit manual entry:', error);
        alert(`Failed to submit: ${error.message || 'Unknown error'}`);
        this.isSubmitting.set(false);
      }
    });
  }

  protected clearManualSubmissionFields(): void {
    this.manualIdNumber.set('');
    this.manualServiceId.set('');
    this.manualDuration.set(null);
  }
}
