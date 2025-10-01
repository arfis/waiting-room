# Quick Reference Guide

## Project Structure

```
ui/projects/
├── backoffice/              ✅ Refactored
│   └── src/app/
│       ├── core/
│       │   └── services/    # API & State services
│       ├── environments/    # Config files
│       ├── backoffice/      # Feature
│       └── queue-management/
│           └── components/  # Presentational components
├── kiosk/                   ✅ Refactored
│   └── src/app/
│       ├── core/
│       │   └── services/
│       ├── environments/
│       └── card-loader/
├── mobile/                  ⏳ To be refactored
└── tv/                      ⏳ To be refactored
```

## Key Patterns

### Component Structure

```typescript
import { Component, inject, OnInit, ChangeDetectionStrategy } from '@angular/core';

@Component({
  selector: 'app-feature',
  standalone: true,
  imports: [CommonModule, ...],
  templateUrl: './feature.component.html',
  styleUrl: './feature.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class FeatureComponent implements OnInit {
  private readonly service = inject(ServiceName);
  
  protected readonly data = this.service.data;
  
  ngOnInit(): void {
    this.service.initialize();
  }
}
```

### Service Structure

```typescript
import { Injectable, signal, computed, inject } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class StateService {
  private readonly apiService = inject(ApiService);
  
  readonly data = signal<Type>(initialValue);
  readonly computed = computed(() => this.data());
  
  loadData(): void {
    this.apiService.getData().subscribe({
      next: (data) => this.data.set(data),
      error: (err) => console.error(err)
    });
  }
}
```

### Presentational Component

```typescript
import { Component, Input, Output, EventEmitter, ChangeDetectionStrategy } from '@angular/core';

@Component({
  selector: 'app-item',
  standalone: true,
  imports: [CommonModule],
  template: `...`,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ItemComponent {
  @Input({ required: true }) data!: Type;
  @Output() action = new EventEmitter<void>();
}
```

## Common Commands

### Build
```bash
# Development build
npx ng build backoffice --configuration development
npx ng build kiosk --configuration development

# Production build
npx ng build backoffice --configuration production
npx ng build kiosk --configuration production
```

### Serve
```bash
# Development server
npx ng serve backoffice
npx ng serve kiosk
```

### Test
```bash
npx ng test backoffice
npx ng test kiosk
```

## Checklist for New Components

- [ ] Use `Component` suffix in class name
- [ ] Add `standalone: true`
- [ ] Add `changeDetection: ChangeDetectionStrategy.OnPush`
- [ ] Separate template into `.html` file
- [ ] Separate styles into `.scss` file
- [ ] Use `inject()` for dependencies
- [ ] Use signals for state
- [ ] Use computed for derived state
- [ ] Add proper TypeScript interfaces
- [ ] Implement lifecycle hooks if needed
- [ ] Use `@Input({ required: true })` for required inputs
- [ ] Use modern template syntax (`@if`, `@for`)

## Checklist for New Services

- [ ] Use `@Injectable({ providedIn: 'root' })`
- [ ] Use `inject()` for dependencies
- [ ] Use signals for state
- [ ] Use computed for derived state
- [ ] Return Observables from API calls
- [ ] Handle errors properly
- [ ] Add TypeScript interfaces
- [ ] Document public methods

## Template Syntax

### Old vs New

```html
<!-- Old -->
<div *ngIf="condition">Content</div>
<div *ngFor="let item of items">{{ item }}</div>

<!-- New -->
@if (condition) {
  <div>Content</div>
}
@for (item of items; track item.id) {
  <div>{{ item }}</div>
}
```

## Signal Usage

```typescript
// Create signal
readonly data = signal<Type>(initialValue);

// Read signal
const value = this.data();

// Update signal
this.data.set(newValue);
this.data.update(current => ({ ...current, field: value }));

// Computed signal
readonly computed = computed(() => {
  return this.data().someProperty;
});
```

## Documentation Files

- **ARCHITECTURE.md** - Detailed architecture and patterns
- **MIGRATION_GUIDE.md** - Step-by-step migration guide
- **REFACTORING_SUMMARY.md** - What was changed
- **CLEANUP_COMPLETE.md** - Completion summary
- **QUICK_REFERENCE.md** - This file

## Need Help?

1. Check `ARCHITECTURE.md` for patterns
2. Look at backoffice or kiosk for examples
3. Follow `MIGRATION_GUIDE.md` for new features
4. Refer to [Angular Documentation](https://angular.io)
