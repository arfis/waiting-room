import { Component, Input, computed } from '@angular/core';
import { CommonModule } from '@angular/common';

export interface DataField {
  label: string;
  value: string | number | null | undefined;
  type?: 'text' | 'date' | 'datetime' | 'image';
  imageAlt?: string;
  formattedValue?: string; // Pre-formatted value for better performance
}

@Component({
  selector: 'ui-data-grid',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="data-grid">
      @for (field of processedFields(); track field.label) {
        <div class="data-item">
          <span class="font-medium">{{ field.label }}:</span>
          @if (field.type === 'image') {
            <div class="mt-2">
              <img [src]="field.value" [alt]="field.imageAlt || field.label" class="w-24 h-32 object-cover rounded border">
            </div>
          } @else {
            <span class="ml-2">{{ field.formattedValue || field.value }}</span>
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .data-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 1rem;
    }
    
    .data-item {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }
    
    @media (max-width: 640px) {
      .data-grid {
        grid-template-columns: 1fr;
      }
    }
  `]
})
export class DataGridComponent {
  @Input() fields: DataField[] = [];

  // Computed signal for better performance - only recalculates when fields change
  processedFields = computed(() => {
    return this.fields
      .filter(field => field.value !== null && field.value !== undefined)
      .map(field => ({
        ...field,
        formattedValue: this.formatValue(field.value, field.type)
      }));
  });

  private formatValue(value: string | number | null | undefined, type?: string): string {
    if (!value) return '';
    
    if (type === 'date') {
      try {
        const date = new Date(value);
        return date.toLocaleDateString();
      } catch {
        return String(value);
      }
    }
    
    if (type === 'datetime') {
      try {
        const date = new Date(value);
        return date.toLocaleString();
      } catch {
        return String(value);
      }
    }
    
    return String(value);
  }
}
