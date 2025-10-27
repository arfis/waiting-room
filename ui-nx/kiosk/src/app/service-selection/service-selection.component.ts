import { Component, input, output, signal, computed, ChangeDetectionStrategy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { UserServicesService, UserService, ServiceSection } from '../core/services/user-services.service';
import { TranslationService, TranslatePipe } from '../../../../src/lib/i18n';

@Component({
  selector: 'app-service-selection',
  standalone: true,
  imports: [CommonModule, TranslatePipe],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="space-y-6">
      @if (isLoading()) {
        <div class="text-center py-8">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p class="mt-4 text-gray-600">{{ 'kiosk.services.loadingServices' | translate }}</p>
        </div>
      } @else if (error()) {
        <div class="text-center py-8">
          <div class="text-red-600 mb-4">
            <svg class="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <p class="text-lg font-medium">{{ error() }}</p>
          </div>
          <button 
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            (click)="retry.emit()">
            {{ 'common.retry' | translate }}
          </button>
        </div>
      } @else {
        <!-- Service Sections -->
        @if (serviceSections().length > 0) {
          @for (section of serviceSections(); track section.type) {
          <div class="space-y-4">
            <div class="flex items-center space-x-2">
              <h2 class="text-lg font-semibold text-gray-900">{{ section.title | translate }}</h2>
              @if (section.type === 'appointment') {
                <span class="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">{{ 'kiosk.services.personal' | translate }}</span>
              } @else {
                <span class="px-2 py-1 text-xs bg-green-100 text-green-800 rounded-full">{{ 'kiosk.services.general' | translate }}</span>
              }
            </div>
            
            @if (section.loading) {
              <div class="text-center py-4">
                <div class="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-gray-400"></div>
                <p class="mt-2 text-sm text-gray-500">{{ 'common.loadingServices' | translate }}</p>
              </div>
            } @else if (section.error) {
              <div class="text-center py-4">
                <div class="text-red-500 mb-2">
                  <svg class="w-8 h-8 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                  </svg>
                </div>
                <p class="text-sm text-red-600">{{ section.error }}</p>
              </div>
            } @else if (section.services.length === 0) {
              <div class="text-center py-4">
                <div class="text-gray-400 mb-2">
                  <svg class="w-8 h-8 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 12h6m-6-4h6m2 5.291A7.962 7.962 0 0112 15c-2.34 0-4.29-1.009-5.824-2.5M15 6.75a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  </svg>
                </div>
                <p class="text-sm text-gray-500">{{ 'common.noServicesAvailable' | translate }}</p>
              </div>
            } @else {
              <div class="space-y-2">
                @for (service of section.services; track service.id) {
                  <button 
                    class="w-full p-4 text-left border border-gray-200 rounded-lg hover:border-blue-300 hover:bg-blue-50 transition-colors"
                    [class]="selectedServiceId() === service.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200'"
                    (click)="selectService(service)">
                    <div class="flex items-center justify-between">
                      <div>
                        <h3 class="font-semibold text-gray-900">{{ service.serviceName }}</h3>
                        <p class="text-sm text-gray-600">{{ 'common.estimatedDuration' | translate }}: {{ service.duration }} {{ 'common.minutes' | translate }}</p>
                      </div>
                      <div class="flex items-center">
                        @if (selectedServiceId() === service.id) {
                          <div class="w-6 h-6 bg-blue-600 rounded-full flex items-center justify-center">
                            <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                            </svg>
                          </div>
                        } @else {
                          <div class="w-6 h-6 border-2 border-gray-300 rounded-full"></div>
                        }
                      </div>
                    </div>
                  </button>
                }
              </div>
            }
          </div>
          }
        } @else {
          <!-- Fallback to old services array -->
          @if (services().length === 0) {
            <div class="text-center py-8">
              <div class="text-gray-400 mb-4">
                <svg class="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 12h6m-6-4h6m2 5.291A7.962 7.962 0 0112 15c-2.34 0-4.29-1.009-5.824-2.5M15 6.75a3 3 0 11-6 0 3 3 0 016 0z"></path>
                </svg>
                <p class="text-lg font-medium">{{ 'common.noServicesAvailable' | translate }}</p>
                <p class="text-gray-500">{{ 'common.noServicesDescription' | translate }}</p>
              </div>
            </div>
          } @else {
            <div class="space-y-3">
              @for (service of services(); track service.id) {
                <button 
                  class="w-full p-4 text-left border border-gray-200 rounded-lg hover:border-blue-300 hover:bg-blue-50 transition-colors"
                  [class]="selectedServiceId() === service.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200'"
                  (click)="selectService(service)">
                  <div class="flex items-center justify-between">
                    <div>
                      <h3 class="font-semibold text-gray-900">{{ service.serviceName }}</h3>
                      <p class="text-sm text-gray-600">{{ 'common.estimatedDuration' | translate }}: {{ service.duration }} {{ 'common.minutes' | translate }}</p>
                    </div>
                    <div class="flex items-center">
                      @if (selectedServiceId() === service.id) {
                        <div class="w-6 h-6 bg-blue-600 rounded-full flex items-center justify-center">
                          <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                          </svg>
                        </div>
                      } @else {
                        <div class="w-6 h-6 border-2 border-gray-300 rounded-full"></div>
                      }
                    </div>
                  </div>
                </button>
              }
            </div>
          }
        }

        @if (selectedServiceId()) {
          <div class="mt-6 pt-6 border-t border-gray-200">
            <button 
              class="w-full py-3 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
              (click)="confirmSelection()">
              {{ 'common.continueWithSelectedService' | translate }}
            </button>
          </div>
        }
      }
    </div>
  `
})
export class ServiceSelectionComponent {
  private readonly userServicesService = inject(UserServicesService);
  private readonly translationService = inject(TranslationService);

  // Inputs
  services = input<UserService[]>([]);
  serviceSections = input<ServiceSection[]>([]);
  isLoading = input.required<boolean>();
  error = input<string | null>(null);

  constructor() {
    console.log('ServiceSelectionComponent - NEW VERSION LOADED!');
    console.log('Service sections:', this.serviceSections());
    console.log('Services:', this.services());
  }


  // Outputs
  serviceSelected = output<UserService>();
  retry = output<void>();

  // Local state
  selectedServiceId = signal<string | null>(null);

  selectService(service: UserService): void {
    console.log('Service clicked:', service);
    this.selectedServiceId.set(service.id);
  }

  confirmSelection(): void {
    // todo: needs a bit of rework
    const selectedService = this.services().find(s => s.id === this.selectedServiceId()) || this.serviceSections().find(s => s.services.find(service => service.id === this.selectedServiceId()))?.services.find(service => service.id === this.selectedServiceId()) as UserService;
    if (selectedService) {
      console.log('Confirming service selection:', selectedService);
      this.serviceSelected.emit(selectedService);
    } else {
      console.error('No service selected');
    }
  }
}
