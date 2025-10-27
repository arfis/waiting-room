import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ButtonModule } from 'primeng/button';

@Component({
  selector: 'ui-button',
  standalone: true,
  imports: [CommonModule, ButtonModule],
  template: `
    <p-button
      [label]="label"
      [icon]="icon"
      [iconPos]="iconPos"
      [loading]="loading"
      [disabled]="disabled"
      [severity]="severity"
      [size]="size"
      [outlined]="outlined"
      [text]="text"
      [raised]="raised"
      [rounded]="rounded"
      [class]="customClass"
      (onClick)="onClick.emit($event)"
    />
  `,
  styles: []
})
export class ButtonComponent {
  @Input() label?: string;
  @Input() icon?: string;
  @Input() iconPos: 'left' | 'right' | 'top' | 'bottom' = 'left';
  @Input() loading = false;
  @Input() disabled = false;
  @Input() severity: 'primary' | 'secondary' | 'success' | 'info' | 'help' | 'danger' | 'contrast' = 'primary';
  @Input() size: 'small' | 'large' = 'small';
  @Input() outlined = false;
  @Input() text = false;
  @Input() raised = false;
  @Input() rounded = false;
  @Input() customClass = '';
  @Output() onClick = new EventEmitter<any>();
}
