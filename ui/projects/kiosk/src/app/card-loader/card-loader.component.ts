import { Component, inject, OnInit, OnDestroy, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardComponent, DataGridComponent } from 'ui';
import { CardReaderStateService } from '../core/services/card-reader-state.service';

@Component({
  selector: 'app-card-loader',
  standalone: true,
  imports: [CommonModule, CardComponent, DataGridComponent],
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

  ngOnInit(): void {
    this.cardReaderState.initialize();
  }

  ngOnDestroy(): void {
    this.cardReaderState.disconnect();
  }

  protected checkCardReaderStatus(): void {
    this.cardReaderState.checkCardReaderStatus();
  }
}
