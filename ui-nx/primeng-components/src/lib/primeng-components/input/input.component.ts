import { Component, Input, Output, EventEmitter, forwardRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import { InputTextModule } from 'primeng/inputtext';

@Component({
  selector: 'ui-input',
  standalone: true,
  imports: [CommonModule, FormsModule, InputTextModule],
  template: `
    <div class="flex flex-col gap-2">
      <label *ngIf="label" [for]="inputId" class="text-sm font-medium text-gray-700">
        {{ label }}
        <span *ngIf="required" class="text-red-500">*</span>
      </label>
      <input
        pInputText
        [id]="inputId"
        [value]="value"
        [placeholder]="placeholder"
        [disabled]="disabled"
        [readonly]="readonly"
        [class]="inputClass"
        (input)="onInput($event)"
        (blur)="onBlurHandler()"
        (focus)="onFocusHandler()"
      />
      <small *ngIf="helpText" class="text-gray-500">{{ helpText }}</small>
      <small *ngIf="errorText" class="text-red-500">{{ errorText }}</small>
    </div>
  `,
  styles: [],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => InputComponent),
      multi: true
    }
  ]
})
export class InputComponent implements ControlValueAccessor {
  @Input() label?: string;
  @Input() placeholder = '';
  @Input() helpText?: string;
  @Input() errorText?: string;
  @Input() required = false;
  @Input() disabled = false;
  @Input() readonly = false;
  @Input() inputClass = '';
  @Input() inputId = '';
  @Output() onFocusEvent = new EventEmitter<FocusEvent>();
  @Output() onBlurEvent = new EventEmitter<FocusEvent>();

  value = '';
  private onChange = (value: string) => {};
  private onTouched = () => {};

  onInput(event: Event): void {
    const target = event.target as HTMLInputElement;
    this.value = target.value;
    this.onChange(this.value);
  }

  onFocusHandler(): void {
    this.onFocusEvent.emit();
  }

  onBlurHandler(): void {
    this.onTouched();
    this.onBlurEvent.emit();
  }

  writeValue(value: string): void {
    this.value = value || '';
  }

  registerOnChange(fn: (value: string) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled = isDisabled;
  }
}
