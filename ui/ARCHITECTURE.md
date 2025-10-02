# Angular UI Architecture - Best Practices

This document outlines the architectural decisions and best practices implemented in the Angular UI components.

## Project Structure

```
ui/
├── projects/
│   ├── api-client/          # Shared API client library
│   ├── backoffice/          # Backoffice application
│   │   └── src/app/
│   │       ├── core/
│   │       │   └── services/    # Singleton services
│   │       ├── environments/    # Environment configurations
│   │       ├── backoffice/      # Feature module
│   │       └── queue-management/
│   │           └── components/  # Presentational components
│   ├── kiosk/               # Kiosk application
│   │   └── src/app/
│   │       ├── core/
│   │       │   └── services/
│   │       ├── environments/
│   │       └── card-loader/
│   ├── mobile/              # Mobile application
│   ├── tv/                  # TV display application
│   ���── ui/                  # Shared UI component library
```

## Best Practices Implemented

### 1. Component Naming Conventions

- **All components use the `Component` suffix** (e.g., `AppComponent`, `QueueManagementComponent`)
- **File names match component names** in kebab-case (e.g., `app.component.ts`, `queue-management.component.ts`)
- **Selectors use appropriate prefixes**:
  - `app-` for application-specific components
  - `ui-` for shared UI library components

### 2. Smart vs Presentational Components

#### Smart (Container) Components
- Located at feature level (e.g., `QueueManagementComponent`, `CardLoaderComponent`)
- Inject and use services
- Manage state and business logic
- Handle data fetching and side effects
- Pass data down to presentational components via `@Input()`
- Handle events from presentational components via `@Output()`

#### Presentational (Dumb) Components
- Located in `components/` subdirectories
- Receive data via `@Input()` properties
- Emit events via `@Output()` properties
- No service injection (except utility services)
- Focused on UI rendering
- Highly reusable and testable

Examples:
- `QueueHeaderComponent`
- `QueueActionsComponent`
- `CurrentEntryCardComponent`
- `QueueStatisticsComponent`
- `WaitingQueueListComponent`
- `ActivityLogComponent`

### 3. Service Layer Architecture

#### API Services
- Handle HTTP communication
- Located in `core/services/`
- Use `@Injectable({ providedIn: 'root' })`
- Return Observables
- Examples: `QueueApiService`, `KioskApiService`

#### State Services
- Manage application state using signals
- Provide computed values
- Handle business logic
- Coordinate between API services and components
- Examples: `QueueStateService`, `CardReaderStateService`

### 4. Signals for State Management

Using Angular's new signals API for reactive state:

```typescript
// State signals
readonly cardData = signal<CardData | null>(null);
readonly isLoading = signal<boolean>(false);

// Computed signals
readonly currentEntry = computed(() => {
  const entries = this.queueEntries();
  return entries.find(entry => entry.status === 'CALLED') || null;
});
```

Benefits:
- Fine-grained reactivity
- Better performance than Zone.js
- Simpler mental model
- Type-safe

### 5. Change Detection Strategy

All components use `OnPush` change detection:

```typescript
@Component({
  changeDetection: ChangeDetectionStrategy.OnPush
})
```

Benefits:
- Improved performance
- Predictable change detection
- Works perfectly with signals

### 6. Separation of Concerns

#### Templates
- Separated into `.html` files
- Use modern Angular template syntax (`@if`, `@for`, `@else`)
- Minimal logic in templates

#### Styles
- Separated into `.scss` files
- Component-scoped styles
- Tailwind CSS for utility classes

#### TypeScript
- Business logic in services
- Component classes focus on coordination
- Type-safe interfaces for all data structures

### 7. Environment Configuration

Centralized configuration in `environments/`:

```typescript
export const environment = {
  production: false,
  apiUrl: 'http://localhost:8080/api'
};
```

### 8. Dependency Injection

Using modern Angular DI patterns:

```typescript
// Inject function (preferred)
private readonly queueApiService = inject(QueueApiService);

// Constructor injection (when needed)
constructor(private http: HttpClient) {}
```

### 9. Type Safety

- Interfaces for all data structures
- Strict TypeScript configuration
- No `any` types
- Proper generic types for Observables and signals

### 10. Standalone Components

All components are standalone:

```typescript
@Component({
  standalone: true,
  imports: [CommonModule, CardComponent, ...]
})
```

Benefits:
- Simpler mental model
- Better tree-shaking
- Easier lazy loading
- No NgModules needed

### 11. Reactive Programming

- RxJS for async operations
- Proper subscription management
- Unsubscribe in `ngOnDestroy()`
- Use async pipe when possible

### 12. Error Handling

- Centralized error handling in services
- User-friendly error messages
- Console logging for debugging
- Error state in signals

### 13. Accessibility

- Semantic HTML
- Proper ARIA labels
- Keyboard navigation support
- Loading states for async operations

### 14. Performance Optimization

- OnPush change detection
- Computed signals for derived state
- TrackBy functions in `@for` loops
- Lazy loading of routes
- Code splitting

## Component Communication Patterns

### Parent to Child
Use `@Input()` properties:

```typescript
@Input({ required: true }) entries!: WebSocketQueueEntry[];
```

### Child to Parent
Use `@Output()` events:

```typescript
@Output() callEntry = new EventEmitter<WebSocketQueueEntry>();
```

### Sibling Components
Use shared services with signals:

```typescript
readonly queueEntries = this.queueState.queueEntries;
```

## Testing Strategy

- Unit tests for services
- Component tests for presentational components
- Integration tests for smart components
- E2E tests for critical user flows

## Code Style

- Use `readonly` for injected services
- Use `protected` for template-accessible members
- Use `private` for internal implementation
- Prefer `const` over `let`
- Use arrow functions for callbacks
- Use optional chaining (`?.`) and nullish coalescing (`??`)

## Future Improvements

1. Add comprehensive unit tests
2. Implement error boundary components
3. Add loading skeletons
4. Implement optimistic UI updates
5. Add offline support
6. Implement proper logging service
7. Add analytics tracking
8. Implement feature flags
9. Add internationalization (i18n)
10. Implement proper form validation

## Resources

- [Angular Style Guide](https://angular.io/guide/styleguide)
- [Angular Signals](https://angular.io/guide/signals)
- [Angular Best Practices](https://angular.io/guide/best-practices)
- [RxJS Best Practices](https://rxjs.dev/guide/overview)
