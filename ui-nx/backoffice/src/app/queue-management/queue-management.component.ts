import { Component, inject, OnInit, OnDestroy, ChangeDetectionStrategy, signal, computed, effect, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { QueueStateService } from '../core/services/queue-state.service';
import { QueueActionsComponent } from './components/queue-actions/queue-actions.component';
import { CurrentEntryCardComponent } from './components/current-entry-card/current-entry-card.component';
import { WaitingQueueListComponent } from './components/waiting-queue-list/waiting-queue-list.component';
import { ActivityLogComponent } from './components/activity-log/activity-log.component';
import { WebSocketQueueEntry, ConfigurationResponse, ServicePointConfiguration, QueueEntryStatus } from '@waiting-room/api-client';
import { TenantSelectorComponent, TenantService } from '@lib/tenant';
import { LanguageSelectorComponent, TranslatePipe } from '../../../../src/lib/i18n';
import { environment } from '../../environments/environment';

@Component({
  selector: 'app-queue-management',
  standalone: true,
  imports: [
    CommonModule,
    QueueActionsComponent,
    CurrentEntryCardComponent,
    WaitingQueueListComponent,
    ActivityLogComponent,
    TenantSelectorComponent,
    LanguageSelectorComponent,
    TranslatePipe,
  ],
  templateUrl: './queue-management.component.html',
  styleUrl: './queue-management.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class QueueManagementComponent implements OnInit, OnDestroy {
  private readonly queueState = inject(QueueStateService);
  private readonly tenantService = inject(TenantService);
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl || 'http://localhost:8080/api';
  private readonly roomId = 'triage-1';

  // Configuration state
  private readonly configuration = signal<ConfigurationResponse | null>(null);
  protected readonly isConfigLoading = signal(true);
  protected readonly configError = signal<string | null>(null);
  protected readonly selectedServicePoint = signal<ServicePointConfiguration | null>(null);

  // Expose state to template
  protected readonly currentEntry = this.queueState.currentEntry;
  protected readonly waitingEntries = this.queueState.waitingEntries;
  protected readonly calledEntries = this.queueState.calledEntries;
  protected readonly isLoading = this.queueState.isLoading;
  protected readonly lastUpdated = this.queueState.lastUpdated;
  protected readonly recentActivity = this.queueState.recentActivity;
  protected readonly estimatedWaitTime = this.queueState.estimatedWaitTime;
  protected readonly isConnected = this.queueState.isConnected;
  protected readonly error = this.queueState.wsError;

  // Scroll state
  protected readonly isScrolled = signal(false);
  // Listen to window scroll events
  @HostListener('window:scroll', ['$event'])
  onWindowScroll(event: Event) {
    this.isScrolled.set(window.scrollY > 50);
  }

  // Configuration-related computed properties
  protected readonly availableServicePoints = computed(() => {
    const config = this.configuration();
    if (!config || !config.rooms.length) {
      // If no rooms configured, return empty array (no service points)
      return [];
    }

    // Get service points from the default room or first room
    const defaultRoom = config.rooms.find((room: any) => room.id === config.defaultRoom) || config.rooms[0];
    const servicePoints = defaultRoom?.servicePoints || [];

    // Return service points as-is (empty if none configured)
    return servicePoints;
  });

  // Check if there are any configured service points (not default/auto-generated)
  protected readonly hasConfiguredServicePoints = computed(() => {
    const servicePoints = this.availableServicePoints();
    return servicePoints.length > 0;
  });

  protected readonly showServicePointSelection = computed(() => {
    // Only show service point selection if:
    // 1. A section is selected (not "All Sections" which is 0), OR
    // 2. The tenant has no sections at all (sections array is empty after loading)
    // AND there are 2+ service points configured

    // Must not be loading or have error first
    if (this.isConfigLoading() || this.configError()) {
      return false;
    }

    const selectedSectionId = this.tenantService.selectedSectionId();
    const sections = this.tenantService.sections();
    const tenantLoading = this.tenantService.loading();
    const hasSectionSelected = selectedSectionId > 0;
    const hasNoSections = sections.length === 0 && !tenantLoading;

    // Don't show if "All Sections" (0) is selected and sections exist
    if (selectedSectionId === 0 && sections.length > 0) {
      return false;
    }

    // Only proceed if section selected OR no sections exist (and not loading)
    if (!hasSectionSelected && !hasNoSections) {
      return false;
    }

    const servicePoints = this.availableServicePoints();
    if (servicePoints.length < 2) {
      return false; // Don't show selection if 0 or 1 service point
    }

    return !this.selectedServicePoint() && selectedSectionId > 0;
  });

  protected readonly showQueueManagement = computed(() => {
    // Show queue management if:
    // 1. Not loading config
    // 2. No config error
    // 3. Section is selected OR no sections exist
    // 4. Service point is selected OR there's 0-1 service points (implicit/auto-selected)

    // Must not be loading or have error first
    if (this.isConfigLoading() || this.configError()) {
      return false;
    }

    const selectedSectionId = this.tenantService.selectedSectionId();
    const sections = this.tenantService.sections();
    const tenantLoading = this.tenantService.loading();
    const hasSectionSelected = selectedSectionId > 0;
    const hasNoSections = sections.length === 0 && !tenantLoading;

    // Don't show if "All Sections" (0) is selected and sections exist
    // if (selectedSectionId === 0 && sections.length > 0) {
    //   console.log('[QueueManagement] Not showing - All Sections selected but sections exist');
    //   return false;
    // }

    console.log('THIS: ', hasSectionSelected, hasNoSections);
    // Must have section selected OR no sections exist (and not loading)
    // if (!hasSectionSelected && !hasNoSections) {
    //   console.log('[QueueManagement] Not showing - No section selected and sections exist or still loading', {
    //     hasSectionSelected,
    //     hasNoSections,
    //     sectionsLength: sections.length,
    //     tenantLoading
    //   });
    //   return false;
    // }

    const servicePoints = this.availableServicePoints();
    const hasImplicitOrAutoSelected = servicePoints.length <= 1; // 0 or 1 service point = no selection needed
    const hasSelectedServicePoint = !!this.selectedServicePoint();

    return true //hasSelectedServicePoint || hasImplicitOrAutoSelected;
  });

  private lastTenantId: number | null = null;
  private lastCompositeId: string | null = null;

  constructor() {
    // Single effect to watch for tenant/section changes and reload everything
    effect(() => {
      const tenantId = this.tenantService.selectedTenantId();
      const compositeId = this.tenantService.getSelectedTenantIdSync();

      // Handle tenant/section changes
      if (compositeId && compositeId !== this.lastCompositeId && this.lastCompositeId !== null) {
        console.log('[QueueManagement] Tenant/section changed from', this.lastCompositeId, 'to', compositeId);
        this.handleTenantChange(tenantId, compositeId);
      } else if (compositeId && this.lastCompositeId === null) {
        // First initialization
        this.lastTenantId = tenantId;
        this.lastCompositeId = compositeId;
      }
    });
  }

  private async handleTenantChange(tenantId: number, compositeId: string): Promise<void> {
    this.lastTenantId = tenantId;
    this.lastCompositeId = compositeId;

    // Disconnect from WebSocket and reset state
    this.queueState.disconnect();
    this.selectedServicePoint.set(null);
    this.isConfigLoading.set(true);
    this.configError.set(null);

    // Reload configuration and initialize queue
    await this.loadConfiguration();
    await this.initializeQueueIfReady();
  }

  private async initializeQueueIfReady(): Promise<void> {
    const tenantId = this.tenantService.selectedTenantId();
    const configLoaded = !this.isConfigLoading() && !this.configError();
    const servicePoints = this.availableServicePoints();
    const hasServicePointContext = this.selectedServicePoint() || servicePoints.length <= 1;

    if (tenantId && configLoaded && hasServicePointContext) {
      const servicePointName = this.selectedServicePoint()?.name || 'implicit';
      console.log('[QueueManagement] Initializing queue state with tenant:', tenantId, 'service point:', servicePointName);
      await this.queueState.initialize(this.roomId, ['WAITING', 'IN_SERVICE']);
    }
  }

  async ngOnInit(): Promise<void> {
    // Check if tenant is selected before proceeding
    const compositeId = this.tenantService.getSelectedTenantIdSync();
    const tenantId = this.tenantService.selectedTenantId();
    if (!compositeId || !tenantId) {
      console.warn('Cannot initialize queue management: No tenant selected');
      return;
    }

    // Track the initial composite ID
    this.lastTenantId = tenantId;
    this.lastCompositeId = compositeId;
    console.log('[QueueManagement] ngOnInit - Initializing with tenant/section:', compositeId);

    // Load configuration and initialize queue
    await this.loadConfiguration();
    await this.initializeQueueIfReady();
  }

  ngOnDestroy(): void {
    this.queueState.disconnect();
  }

  protected onCallNext(): void {
    // Get service point name: use selected if available, otherwise use empty string (implicit single service point)
    const servicePointName = this.selectedServicePoint()?.name || '';
    this.queueState.callNext(this.roomId, servicePointName);
  }

  protected onFinishCurrent(): void {
    this.queueState.finishCurrent(this.roomId);
  }

  protected onPlaceBack(): void {
    this.queueState.placeBack(this.roomId);
  }

  protected onRefresh(): void {
    // Refresh both waiting entries and current entry
    this.queueState.refreshWaitingEntries(this.roomId);
    this.queueState.refreshCurrentEntry(this.roomId);
  }

  protected onCallSpecificEntry(entry: WebSocketQueueEntry): void {
    // Get service point name: use selected if available, otherwise use empty string (implicit single service point)
    // Use NAME not ID, consistent with onCallNext() - backend stores service point by name
    const servicePointName = this.selectedServicePoint()?.name || '';

    // Call the specific entry by ID
    // If no service point selected and entry has a service point, use that (for consistency)
    // Otherwise use the selected service point name or empty string
    const targetServicePointName = servicePointName || entry.servicePoint || '';
    console.log('[QueueManagement] Calling specific entry:', {
      entryId: entry.id,
      ticketNumber: entry.ticketNumber,
      selectedServicePointName: servicePointName,
      entryServicePoint: entry.servicePoint,
      targetServicePointName: targetServicePointName
    });
    this.queueState.callSpecificEntry(this.roomId, targetServicePointName, entry.id);
  }

  protected getStatusText(status: string): string {
    return this.queueState.getStatusText(status);
  }

  protected getCalledEntryForSelectedServicePoint(): WebSocketQueueEntry | null {
    const selectedServicePoint = this.selectedServicePoint();
    const servicePointName = selectedServicePoint?.name || '';

    // If no service point selected (0 service points = implicit), try to get current entry without service point filter
    // Otherwise, get the current entry for the selected service point
    if (!selectedServicePoint) {
      // For implicit single service point (0 configured), get any current entry
      const entries = this.queueState.queueEntries();
      return entries.find(entry =>
        entry.status === 'CALLED' || entry.status === 'IN_SERVICE'
      ) || null;
    }

    // Get the current entry (CALLED or IN_SERVICE) for the selected service point
    return this.queueState.getCurrentEntryForServicePoint(servicePointName);
  }

  // Configuration methods
  private async loadConfiguration(): Promise<void> {
    this.isConfigLoading.set(true);
    this.configError.set(null);

    try {
      const compositeId = this.tenantService.getSelectedTenantIdSync();
      const sectionId = this.tenantService.selectedSectionId();
      console.log('[QueueManagement] Loading configuration for tenant/section:', compositeId, '(section:', sectionId || 'none', ')');

      // Use HttpClient directly so it goes through the tenant interceptor
      const config = await this.http.get<ConfigurationResponse>(`${this.apiUrl}/config`).toPromise();
      if (config) {
        console.log('[QueueManagement] Configuration loaded:', config);
        console.log('[QueueManagement] Rooms:', config.rooms?.length || 0);
        if (config.rooms && config.rooms.length > 0) {
          const defaultRoom = config.rooms.find((room: any) => room.id === config.defaultRoom) || config.rooms[0];
          console.log('[QueueManagement] Default room:', defaultRoom?.id, defaultRoom?.name);
          console.log('[QueueManagement] Service points in default room:', defaultRoom?.servicePoints?.length || 0);
          if (defaultRoom?.servicePoints) {
            defaultRoom.servicePoints.forEach((sp: any, idx: number) => {
              console.log(`[QueueManagement] Service point ${idx}:`, sp.id, sp.name);
            });
          }
        }

        this.configuration.set(config);

        // Handle service points: 0 = implicit single, 1 = auto-select, 2+ = user selects
        const servicePoints = this.availableServicePoints();
        console.log('[QueueManagement] Available service points (after processing):', servicePoints.length);
        if (servicePoints.length === 1) {
          // One service point - auto-select it
          console.log('[QueueManagement] Auto-selecting single service point:', servicePoints[0]);
          this.selectedServicePoint.set(servicePoints[0]);
          // Initialize queue state immediately after auto-selecting
          if (compositeId) {
            console.log('[QueueManagement] Initializing queue state with tenant/section:', compositeId);
            this.queueState.initialize(this.roomId, ['WAITING', 'IN_SERVICE']);
          } else {
            console.warn('[QueueManagement] Cannot initialize queue state: No tenant selected');
          }
        } else if (servicePoints.length === 0) {
          // No service points configured = implicit single service point (like a single barber)
          // No selection needed, just initialize queue state
          console.log('[QueueManagement] No service points configured - using implicit single service point');
          if (compositeId) {
            console.log('[QueueManagement] Initializing queue state with tenant/section (implicit service point):', compositeId);
            this.queueState.initialize(this.roomId, ['WAITING', 'IN_SERVICE']);
          } else {
            console.warn('[QueueManagement] Cannot initialize queue state: No tenant selected');
          }
        }
        // If servicePoints.length >= 2, user will select from the selection screen
      } else {
        console.warn('[QueueManagement] Configuration is null or undefined');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load configuration';
      this.configError.set(errorMessage);
      console.error('[QueueManagement] Configuration loading failed:', err);
    } finally {
      this.isConfigLoading.set(false);
    }
  }

  protected async selectServicePoint(servicePoint: ServicePointConfiguration): Promise<void> {
    this.selectedServicePoint.set(servicePoint);
    console.log('[QueueManagement] Service point selected:', servicePoint.name);
    await this.initializeQueueIfReady();
  }

  protected clearServicePoint(): void {
    this.selectedServicePoint.set(null);
    this.queueState.disconnect();
  }

  protected retryConfiguration(): void {
    this.loadConfiguration();
  }

  protected readonly screen = screen;
}
