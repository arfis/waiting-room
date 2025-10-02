# Migration Guide - Angular Best Practices

This guide helps migrate the remaining applications (mobile, tv) to follow the same best practices as backoffice and kiosk.

## Checklist for Each Application

### 1. Component Naming

- [ ] Rename `App` to `AppComponent`
- [ ] Update all component class names to include `Component` suffix
- [ ] Update imports in `main.ts`

**Before:**
```typescript
export class App {
  protected readonly title = signal('app-name');
}
```

**After:**
```typescript
export class AppComponent {
  protected readonly title = 'app-name';
}
```

### 2. File Structure

Create the following directory structure:

```
src/app/
├── core/
│   └── services/          # Singleton services
│       ├── api.service.ts
│       └── state.service.ts
├── environments/
│   ├── environment.ts
│   └── environment.prod.ts
├── features/              # Feature modules
│   └── feature-name/
│       ├── components/    # Presentational components
│       ├── feature-name.component.ts
│       ├── feature-name.component.html
│       └── feature-name.component.scss
└── shared/                # Shared components (if needed)
```

### 3. Create Environment Files

**environment.ts:**
```typescript
export const environment = {
  production: false,
  apiUrl: 'http://localhost:8080/api'
};
```

**environment.prod.ts:**
```typescript
export const environment = {
  production: true,
  apiUrl: '/api'
};
```

### 4. Extract API Services

Create API service for HTTP calls:

```typescript
import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl;

  // Add your API methods here
  getData(): Observable<any> {
    return this.http.get(`${this.apiUrl}/endpoint`);
  }
}
```

### 5. Create State Services

Create state service for managing application state:

```typescript
import { Injectable, signal, computed, inject } from '@angular/core';
import { ApiService } from './api.service';

@Injectable({
  providedIn: 'root'
})
export class StateService {
  private readonly apiService = inject(ApiService);
  
  // State signals
  readonly data = signal<any>(null);
  readonly isLoading = signal<boolean>(false);
  readonly error = signal<string | null>(null);
  
  // Computed signals
  readonly processedData = computed(() => {
    const data = this.data();
    // Process data here
    return data;
  });

  loadData(): void {
    this.isLoading.set(true);
    this.apiService.getData().subscribe({
      next: (data) => {
        this.data.set(data);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.error.set('Failed to load data');
        this.isLoading.set(false);
      }
    });
  }
}
```

### 6. Refactor Components

#### Smart Component (Container)

```typescript
import { Component, inject, OnInit, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StateService } from '../core/services/state.service';

@Component({
  selector: 'app-feature',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './feature.component.html',
  styleUrl: './feature.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class FeatureComponent implements OnInit {
  private readonly stateService = inject(StateService);

  // Expose state to template
  protected readonly data = this.stateService.data;
  protected readonly isLoading = this.stateService.isLoading;
  protected readonly error = this.stateService.error;

  ngOnInit(): void {
    this.stateService.loadData();
  }

  protected onAction(): void {
    // Handle user actions
  }
}
```

#### Presentational Component

```typescript
import { Component, Input, Output, EventEmitter, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-feature-item',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="item">
      <h3>{{ title }}</h3>
      <button (click)="action.emit()">Action</button>
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class FeatureItemComponent {
  @Input({ required: true }) title!: string;
  @Output() action = new EventEmitter<void>();
}
```

### 7. Separate Templates and Styles

Move inline templates to `.html` files:

**Before:**
```typescript
@Component({
  template: `<div>...</div>`
})
```

**After:**
```typescript
@Component({
  templateUrl: './component.html'
})
```

### 8. Use Modern Template Syntax

Replace `*ngIf` and `*ngFor` with `@if` and `@for`:

**Before:**
```html
<div *ngIf="data">{{ data }}</div>
<div *ngFor="let item of items">{{ item }}</div>
```

**After:**
```html
@if (data) {
  <div>{{ data }}</div>
}
@for (item of items; track item.id) {
  <div>{{ item }}</div>
}
```

### 9. Add Change Detection Strategy

Add to all components:

```typescript
@Component({
  changeDetection: ChangeDetectionStrategy.OnPush
})
```

### 10. Use Signals Instead of BehaviorSubject

**Before:**
```typescript
private dataSubject = new BehaviorSubject<any>(null);
data$ = this.dataSubject.asObservable();
```

**After:**
```typescript
readonly data = signal<any>(null);
```

### 11. Implement Proper Lifecycle Hooks

```typescript
export class Component implements OnInit, OnDestroy {
  private subscription?: Subscription;

  ngOnInit(): void {
    // Initialize
  }

  ngOnDestroy(): void {
    // Cleanup
    this.subscription?.unsubscribe();
  }
}
```

### 12. Add Type Safety

Define interfaces for all data:

```typescript
export interface DataModel {
  id: string;
  name: string;
  value: number;
}
```

## Step-by-Step Migration Process

1. **Create directory structure**
   - Create `core/services/`
   - Create `environments/`
   - Create feature directories

2. **Create environment files**
   - Add `environment.ts`
   - Add `environment.prod.ts`

3. **Extract services**
   - Create API service
   - Create state service
   - Move HTTP calls from components to services

4. **Refactor components**
   - Rename component classes
   - Add `Component` suffix
   - Update imports

5. **Separate templates**
   - Move inline templates to `.html` files
   - Move inline styles to `.scss` files

6. **Add change detection**
   - Add `OnPush` to all components

7. **Convert to signals**
   - Replace BehaviorSubject with signals
   - Add computed signals for derived state

8. **Create presentational components**
   - Extract UI logic to separate components
   - Use `@Input()` and `@Output()`

9. **Update templates**
   - Use `@if` and `@for` syntax
   - Remove unnecessary logic

10. **Test thoroughly**
    - Test all functionality
    - Check for console errors
    - Verify change detection works

## Common Pitfalls

1. **Forgetting to update main.ts** - Update the bootstrap component name
2. **Not using OnPush** - Always add change detection strategy
3. **Mixing concerns** - Keep smart and presentational components separate
4. **Not unsubscribing** - Always clean up subscriptions
5. **Using any type** - Always define proper interfaces
6. **Inline templates for large components** - Separate into files
7. **Not using signals** - Prefer signals over BehaviorSubject
8. **Not using computed** - Use computed for derived state

## Testing After Migration

- [ ] Application starts without errors
- [ ] All features work as before
- [ ] No console errors or warnings
- [ ] Change detection works properly
- [ ] API calls are successful
- [ ] State updates correctly
- [ ] UI responds to user actions
- [ ] Loading states display correctly
- [ ] Error handling works

## Benefits After Migration

1. **Better Performance** - OnPush change detection and signals
2. **Easier Testing** - Separated concerns and dependency injection
3. **Better Maintainability** - Clear structure and separation
4. **Type Safety** - Proper interfaces and types
5. **Reusability** - Presentational components can be reused
6. **Scalability** - Clear patterns for adding features
7. **Developer Experience** - Easier to understand and modify

## Need Help?

Refer to:
- `ARCHITECTURE.md` for architectural decisions
- Backoffice app for complete example
- Kiosk app for another example
- Angular documentation for specific features
