import { Component, inject, OnInit, OnDestroy, ChangeDetectionStrategy, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CardComponent } from '@waiting-room/primeng-components';
import { CardReaderStateService } from '../core/services/card-reader-state.service';
import { ServiceSelectionComponent } from '../service-selection/service-selection.component';
import { UserService } from '../core/services/user-services.service';

@Component({
  selector: 'app-card-loader',
  standalone: true,
  imports: [CommonModule, FormsModule, CardComponent, ServiceSelectionComponent],
  templateUrl: './card-loader.component.html',
  styleUrls: ['./card-loader.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class CardLoaderComponent implements OnInit, OnDestroy {
  private readonly cardReaderState = inject(CardReaderStateService);

  // Expose state to template
  protected readonly cardData = this.cardReaderState.cardData;
  protected readonly ticketData = this.cardReaderState.ticketData;
  protected readonly error = this.cardReaderState.error;
  protected readonly isReading = this.cardReaderState.isReading;
  protected readonly cardReaderStatus = this.cardReaderState.cardReaderStatus;
  protected readonly wsConnectionStatus = this.cardReaderState.wsConnectionStatus;
  protected readonly readerState = this.cardReaderState.cardReaderState;
  protected readonly cardReaderMessage = this.cardReaderState.cardReaderMessage;
  protected readonly cardDataFields = this.cardReaderState.cardDataFields;
  
  // Service selection state
  protected readonly userServices = this.cardReaderState.userServices;
  protected readonly serviceSections = this.cardReaderState.serviceSections;
  protected readonly isLoadingServices = this.cardReaderState.isLoadingServices;
  protected readonly selectedService = this.cardReaderState.selectedService;
  protected readonly showServiceSelection = this.cardReaderState.showServiceSelection;
  
  // Manual ID entry state
  protected readonly isManualIdSubmitting = this.cardReaderState.isManualIdSubmitting;
  protected readonly manualIdNumber = signal<string>('');
  
  // Ticket display state
  protected readonly ticketCountdown = this.cardReaderState.ticketCountdown;
  protected readonly isTicketCountdownActive = this.cardReaderState.isTicketCountdownActive;

  // Computed properties for dynamic UI
  protected readonly cardTitle = computed(() => {
    if (this.ticketData()) return 'Your Ticket';
    if (this.showServiceSelection()) return 'Service Selection';
    return 'Check In';
  });
  
  protected readonly cardSubtitle = computed(() => {
    if (this.ticketData()) return 'Your ticket has been generated successfully';
    if (this.showServiceSelection()) return 'Choose the service you need';
    return 'Scan your ID card or enter your ID number';
  });
  
  protected readonly headerMessage = computed(() => {
    if (this.ticketData()) return 'Your ticket is ready!';
    if (this.showServiceSelection()) return 'Please select the service you need';
    return 'Please scan your ID card or enter your ID manually to check in';
  });

  ngOnInit(): void {
    this.cardReaderState.initialize();
  }

  ngOnDestroy(): void {
    this.cardReaderState.disconnect();
  }

  protected checkCardReaderStatus(): void {
    this.cardReaderState.checkCardReaderStatus();
  }

  protected getServicePointName(servicePointId: string): string {
    // Map service point IDs to display names
    const servicePointNames: { [key: string]: string } = {
      'window-1': 'Window 1',
      'window-2': 'Window 2',
      'door-1': 'Door 1',
      'door-2': 'Door 2',
      'counter-1': 'Counter 1',
      'counter-2': 'Counter 2'
    };
    
    return servicePointNames[servicePointId] || servicePointId;
  }

  protected onServiceSelected(service: UserService): void {
    this.cardReaderState.selectService(service);
    this.cardReaderState.confirmServiceSelection();
  }

  protected onConfirmServiceSelection(): void {
    this.cardReaderState.confirmServiceSelection();
  }

  protected onRetryServiceLoading(): void {
    this.cardReaderState.retryServiceLoading();
  }

  protected returnToMainInterface(): void {
    this.cardReaderState.returnToMainInterface();
    // Clear the manual ID input
    this.manualIdNumber.set('');
  }

  protected onManualIdSubmitted(idNumber: string): void {
    this.cardReaderState.submitManualId(idNumber);
  }
}
