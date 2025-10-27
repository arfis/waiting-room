import { Component, inject, signal, computed, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

export interface LanguageConfig {
  code: string;
  name: string;
  flag: string;
  direction: 'ltr' | 'rtl';
}

@Component({
  selector: 'ui-language-selector',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="relative">
      <button 
        (click)="toggleDropdown()"
        class="flex items-center gap-2 px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors">
        <span class="text-lg">{{ currentLanguageConfig().flag }}</span>
        <span class="text-sm font-medium text-gray-700">{{ currentLanguageConfig().name }}</span>
        <svg class="w-4 h-4 text-gray-500 transition-transform" [class.rotate-180]="isDropdownOpen()" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
        </svg>
      </button>

      @if (isDropdownOpen()) {
        <div class="absolute top-full left-0 mt-1 w-48 bg-white border border-gray-200 rounded-md shadow-lg z-50">
          <div class="py-1">
            @for (language of availableLanguages; track language.code) {
              <button
                (click)="selectLanguage(language.code)"
                class="w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-gray-100 transition-colors"
                [class.bg-blue-50]="language.code === currentLanguage()">
                <span class="text-lg">{{ language.flag }}</span>
                <span class="text-sm text-gray-700">{{ language.name }}</span>
                @if (language.code === currentLanguage()) {
                  <svg class="w-4 h-4 text-blue-600 ml-auto" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"></path>
                  </svg>
                }
              </button>
            }
          </div>
        </div>
      }
    </div>
  `,
  styles: [`
    .rotate-180 {
      transform: rotate(180deg);
    }
  `]
})
export class LanguageSelectorComponent {
  readonly isDropdownOpen = signal(false);
  readonly currentLanguage = signal<string>('en');
  
  // Available languages
  readonly availableLanguages: LanguageConfig[] = [
    { code: 'en', name: 'English', flag: '🇺🇸', direction: 'ltr' },
    { code: 'sk', name: 'Slovenčina', flag: '🇸🇰', direction: 'ltr' }
  ];

  // Computed properties
  readonly currentLanguageConfig = computed(() => 
    this.availableLanguages.find(lang => lang.code === this.currentLanguage()) || this.availableLanguages[0]
  );

  // Output for language change events
  languageChanged = output<string>();

  constructor() {
    // Load saved language preference from localStorage
    const savedLanguage = localStorage.getItem('preferred-language');
    if (savedLanguage && this.availableLanguages.some(lang => lang.code === savedLanguage)) {
      this.currentLanguage.set(savedLanguage);
    }
  }

  toggleDropdown(): void {
    this.isDropdownOpen.set(!this.isDropdownOpen());
  }

  selectLanguage(languageCode: string): void {
    this.currentLanguage.set(languageCode);
    localStorage.setItem('preferred-language', languageCode);
    this.languageChanged.emit(languageCode);
    this.isDropdownOpen.set(false);
    
    // Update document language and direction
    const config = this.currentLanguageConfig();
    document.documentElement.lang = config.code;
    document.documentElement.dir = config.direction;
  }
}
