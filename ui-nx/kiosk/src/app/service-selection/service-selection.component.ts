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
    <div class="space-y-6 md:space-y-8">
      @if (isLoading()) {
        <div class="text-center py-8 md:py-12">
          <div class="inline-block animate-spin rounded-full h-10 w-10 md:h-12 md:w-12 border-b-2 border-blue-600"></div>
          <p class="mt-4 md:mt-6 text-gray-600 text-base md:text-lg">{{ 'kiosk.services.loadingServices' | translate }}</p>
        </div>
      } @else if (error()) {
        <div class="text-center py-8 md:py-12">
          <div class="text-red-600 mb-6 md:mb-8">
            <svg class="w-16 h-16 md:w-20 md:h-20 mx-auto mb-3 md:mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <p class="text-xl md:text-2xl font-medium">{{ error() }}</p>
          </div>
          <button 
            class="px-6 md:px-8 py-3 md:py-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-base md:text-lg font-medium"
            (click)="retry.emit()">
            {{ 'common.retry' | translate }}
          </button>
        </div>
      } @else {
        <!-- Service Sections -->
        @if (serviceSections().length > 0) {
          @for (section of serviceSections(); track section.type) {
          <div class="space-y-4 md:space-y-6">
            <div class="flex items-center space-x-3 md:space-x-4">
              <h2 class="text-xl md:text-2xl font-semibold text-gray-900">{{ section.title | translate }}</h2>
              @if (section.type === 'appointment') {
                <span class="px-3 md:px-4 py-1 md:py-2 text-sm md:text-base bg-blue-100 text-blue-800 rounded-full font-medium">{{ 'kiosk.services.personal' | translate }}</span>
              } @else {
                <span class="px-3 md:px-4 py-1 md:py-2 text-sm md:text-base bg-green-100 text-green-800 rounded-full font-medium">{{ 'kiosk.services.general' | translate }}</span>
              }
            </div>
            
            @if (section.loading) {
              <div class="text-center py-6 md:py-8">
                <div class="inline-block animate-spin rounded-full h-8 w-8 md:h-10 md:w-10 border-b-2 border-gray-400"></div>
                <p class="mt-3 md:mt-4 text-base md:text-lg text-gray-500">{{ 'common.loadingServices' | translate }}</p>
              </div>
            } @else if (section.error) {
              <div class="text-center py-6 md:py-8">
                <div class="text-red-500 mb-3 md:mb-4">
                  <svg class="w-12 h-12 md:w-16 md:h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                  </svg>
                </div>
                <p class="text-base md:text-lg text-red-600">{{ section.error }}</p>
              </div>
            } @else if (section.services.length === 0) {
              <div class="text-center py-6 md:py-8">
                <div class="text-gray-400 mb-3 md:mb-4">
                  <svg class="w-12 h-12 md:w-16 md:h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 12h6m-6-4h6m2 5.291A7.962 7.962 0 0112 15c-2.34 0-4.29-1.009-5.824-2.5M15 6.75a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  </svg>
                </div>
                <p class="text-base md:text-lg text-gray-500">{{ 'common.noServicesAvailable' | translate }}</p>
              </div>
            } @else {
              <div class="grid grid-cols-1 md:grid-cols-2 gap-3 md:gap-4">
                @for (service of section.services; track service.id) {
                  <button 
                    class="w-full p-4 md:p-6 text-left border border-gray-200 rounded-lg hover:border-blue-300 hover:bg-blue-50 transition-colors min-h-[80px] md:min-h-[100px]"
                    [class]="selectedServiceId() === service.id ? 'border-blue-500 bg-blue-50 ring-2 ring-blue-200' : 'border-gray-200'"
                    (click)="selectService(service)">
                    <div class="flex items-center justify-between h-full">
                      <div class="flex-1">
                        <h3 class="font-semibold text-gray-900 text-base md:text-lg mb-1 md:mb-2">{{ service.serviceName }}</h3>
                        <!-- TODO needs to be removed the / 60 part  -->
                        <p class="text-sm md:text-base text-gray-600">{{ 'common.estimatedDuration' | translate }}: {{ service.duration / 60}} {{ 'common.minutes' | translate }}</p>
                      </div>
                      <div class="flex items-center ml-3 md:ml-4">
                        @if (selectedServiceId() === service.id) {
                          <div class="w-8 h-8 md:w-10 md:h-10 bg-blue-600 rounded-full flex items-center justify-center">
                            <svg class="w-5 h-5 md:w-6 md:h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                            </svg>
                          </div>
                        } @else {
                          <div class="w-8 h-8 md:w-10 md:h-10 border-2 border-gray-300 rounded-full"></div>
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
            <div class="text-center py-8 md:py-12">
              <div class="text-gray-400 mb-6 md:mb-8">
                <svg class="w-16 h-16 md:w-20 md:h-20 mx-auto mb-3 md:mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 12h6m-6-4h6m2 5.291A7.962 7.962 0 0112 15c-2.34 0-4.29-1.009-5.824-2.5M15 6.75a3 3 0 11-6 0 3 3 0 016 0z"></path>
                </svg>
                <p class="text-xl md:text-2xl font-medium">{{ 'common.noServicesAvailable' | translate }}</p>
                <p class="text-gray-500 text-base md:text-lg">{{ 'common.noServicesDescription' | translate }}</p>
              </div>
            </div>
          } @else {
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 md:gap-4">
              @for (service of services(); track service.id) {
                <button 
                  class="w-full p-4 md:p-6 text-left border border-gray-200 rounded-lg hover:border-blue-300 hover:bg-blue-50 transition-colors min-h-[80px] md:min-h-[100px]"
                  [class]="selectedServiceId() === service.id ? 'border-blue-500 bg-blue-50 ring-2 ring-blue-200' : 'border-gray-200'"
                  (click)="selectService(service)">
                  <div class="flex items-center justify-between h-full">
                    <div class="flex-1">
                      <h3 class="font-semibold text-gray-900 text-base md:text-lg mb-1 md:mb-2">{{ service.serviceName }}</h3>
                      <p class="text-sm md:text-base text-gray-600">{{ 'common.estimatedDuration' | translate }}: {{ service.duration }} {{ 'common.minutes' | translate }}</p>
                    </div>
                    <div class="flex items-center ml-3 md:ml-4">
                      @if (selectedServiceId() === service.id) {
                        <div class="w-8 h-8 md:w-10 md:h-10 bg-blue-600 rounded-full flex items-center justify-center">
                          <svg class="w-5 h-5 md:w-6 md:h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                          </svg>
                        </div>
                      } @else {
                        <div class="w-8 h-8 md:w-10 md:h-10 border-2 border-gray-300 rounded-full"></div>
                      }
                    </div>
                  </div>
                </button>
              }
            </div>
          }
        }

        @if (selectedServiceId()) {
          <div class="mt-8 md:mt-12 pt-6 md:pt-8 border-t border-gray-200">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 md:gap-4">
              <button 
                class="py-4 md:py-5 px-6 md:px-8 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors font-medium text-lg md:text-xl"
                (click)="cancel.emit()">
                {{ 'common.cancel' | translate }}
              </button>
              <button 
                class="py-4 md:py-5 px-6 md:px-8 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-lg md:text-xl"
                (click)="confirmSelection()">
                {{ 'common.continueWithSelectedService' | translate }}
              </button>
            </div>
          </div>
        } @else {
          <div class="mt-8 md:mt-12 pt-6 md:pt-8 border-t border-gray-200">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 md:gap-4">
              <button 
                class="py-4 md:py-5 px-6 md:px-8 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors font-medium text-lg md:text-xl"
                (click)="cancel.emit()">
                {{ 'common.cancel' | translate }}
              </button>
              <button 
                class="py-4 md:py-5 px-6 md:px-8 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-lg md:text-xl"
                (click)="proceedWithoutService.emit()">
                {{ 'kiosk.services.proceedWithoutService' | translate }}
              </button>
            </div>
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
  cancel = output<void>();
  proceedWithoutService = output<void>();

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
