import { Injectable, signal, computed } from '@angular/core';

export interface Translation {
  [key: string]: string | Translation;
}

export interface LanguageConfig {
  code: string;
  name: string;
  flag: string;
  direction: 'ltr' | 'rtl';
}

@Injectable({
  providedIn: 'root'
})
export class TranslationService {
  private readonly translations = signal<Record<string, Translation>>({});
  private readonly currentLanguage = signal<string>('en');
  
  // Available languages
  readonly availableLanguages: LanguageConfig[] = [
    { code: 'en', name: 'English', flag: '🇺🇸', direction: 'ltr' },
    { code: 'sk', name: 'Slovenčina', flag: '🇸🇰', direction: 'ltr' }
  ];

  // Computed properties
  readonly currentLanguageConfig = computed(() => 
    this.availableLanguages.find(lang => lang.code === this.currentLanguage()) || this.availableLanguages[0]
  );

  readonly isRTL = computed(() => this.currentLanguageConfig().direction === 'rtl');

  constructor() {
    this.loadTranslations();
    this.loadSavedLanguage();
  }

  /**
   * Get translation for a key
   */
  translate(key: string, params?: Record<string, string | number>): string {
    const translation = this.getNestedTranslation(key);
    if (!translation) {
      console.warn(`Translation missing for key: ${key}`);
      return key;
    }

    return this.interpolateParams(translation, params);
  }

  /**
   * Get translation for a key (shorthand method)
   */
  t(key: string, params?: Record<string, string | number>): string {
    return this.translate(key, params);
  }

  /**
   * Set current language
   */
  setLanguage(languageCode: string): void {
    if (this.availableLanguages.some(lang => lang.code === languageCode)) {
      this.currentLanguage.set(languageCode);
      this.saveLanguagePreference(languageCode);
      this.updateDocumentLanguage();
    }
  }

  /**
   * Get current language code
   */
  getCurrentLanguage(): string {
    return this.currentLanguage();
  }

  /**
   * Get all available languages
   */
  getAvailableLanguages(): LanguageConfig[] {
    return this.availableLanguages;
  }

  /**
   * Add or update translations for a language
   */
  addTranslations(languageCode: string, translations: Translation): void {
    const currentTranslations = this.translations();
    this.translations.set({
      ...currentTranslations,
      [languageCode]: translations
    });
  }

  /**
   * Get nested translation value
   */
  private getNestedTranslation(key: string): string | null {
    const keys = key.split('.');
    const currentTranslations = this.translations();
    const languageTranslations = currentTranslations[this.currentLanguage()];
    
    if (!languageTranslations) {
      return null;
    }

    let value: any = languageTranslations;
    for (const k of keys) {
      if (value && typeof value === 'object' && k in value) {
        value = value[k];
      } else {
        return null;
      }
    }

    return typeof value === 'string' ? value : null;
  }

  /**
   * Interpolate parameters in translation string
   */
  private interpolateParams(text: string, params?: Record<string, string | number>): string {
    if (!params) return text;

    return text.replace(/\{\{(\w+)\}\}/g, (match, key) => {
      return params[key]?.toString() || match;
    });
  }

  /**
   * Load translations from external JSON files
   */
  private async loadTranslations(): Promise<void> {
    try {
      // Load English translations
      const enResponse = await fetch('/assets/i18n/en.json');
      if (enResponse.ok) {
        const enTranslations = await enResponse.json();
        this.addTranslations('en', enTranslations);
      } else {
        console.warn('Failed to load English translations, using fallback');
        this.loadFallbackTranslations();
      }

      // Load Slovak translations
      const skResponse = await fetch('/assets/i18n/sk.json');
      if (skResponse.ok) {
        const skTranslations = await skResponse.json();
        this.addTranslations('sk', skTranslations);
      } else {
        console.warn('Failed to load Slovak translations');
      }
    } catch (error) {
      console.error('Error loading translations:', error);
      this.loadFallbackTranslations();
    }
  }

  /**
   * Load fallback translations (hardcoded) when external files fail
   */
  private loadFallbackTranslations(): void {
    // English (default)
    this.addTranslations('en', {
      common: {
        loading: 'Loading...',
        error: 'Error',
        success: 'Success',
        warning: 'Warning',
        info: 'Information',
        cancel: 'Cancel',
        confirm: 'Confirm',
        save: 'Save',
        edit: 'Edit',
        delete: 'Delete',
        close: 'Close',
        back: 'Back',
        next: 'Next',
        previous: 'Previous',
        submit: 'Submit',
        reset: 'Reset',
        search: 'Search',
        filter: 'Filter',
        sort: 'Sort',
        refresh: 'Refresh',
        retry: 'Retry',
        yes: 'Yes',
        no: 'No',
        ok: 'OK'
      },
      kiosk: {
        checkin: 'Check In',
        title: 'Waiting Room Kiosk',
        welcomeMessage: 'Welcome to our waiting room system',
        insertCard: 'Please insert your ID card',
        readingCard: 'Reading card...',
        cardRead: 'Card read successfully',
        cardError: 'Card reading failed',
        selectService: 'Select a service',
        yourTicket: 'Your Ticket',
        ticketNumber: 'Ticket Number',
        estimatedWait: 'Estimated Wait Time',
        minutes: 'minutes',
        cardInformation: 'Card Information',
        services: {
          appointments: 'Your Appointments',
          generic: 'General Services',
          personal: 'Personal',
          general: 'General',
          noServices: 'No services available',
          loadingServices: 'Loading services...',
          serviceError: 'Failed to load services',
          selectService: 'Please select the service you need',
          success: 'Thank you for checking in!'
        },
        connection: {
          connected: 'Connected',
          connecting: 'Connecting...',
          disconnected: 'Disconnected',
          connectionError: 'Connection error'
        }
      },
      admin: {
        title: 'Admin Panel',
        configuration: 'Configuration',
        externalAPI: 'External API',
        rooms: 'Rooms',
        servicePoints: 'Service Points',
        genericServices: 'Generic Services',
        webhook: 'Webhook',
        appointmentServices: 'Appointment Services',
        genericServicesUrl: 'Generic Services URL',
        webhookUrl: 'Webhook URL',
        timeout: 'Timeout (seconds)',
        retryAttempts: 'Retry Attempts',
        headers: 'Headers',
        addService: 'Add Service',
        serviceName: 'Service Name',
        serviceDescription: 'Service Description',
        duration: 'Duration (minutes)',
        enabled: 'Enabled',
        remove: 'Remove',
        totalServices: 'total services',
        enabledServices: 'enabled',
        multilingualAPI: 'Multilingual API Configuration',
        multilingualSupport: 'External API supports multiple languages',
        supportedLanguages: 'Supported Languages',
        multilingualHelp: 'If enabled, the API will receive ?lang=EN or ?lang=SK parameter',
        useDeepLTranslation: 'Use DeepL for automatic translation',
        deepLHelp: 'When enabled, external API responses will be automatically translated using DeepL',
        appointmentServicesLanguage: 'Appointment Services Language Configuration',
        languageHandlingMethod: 'Language Handling Method',
        queryParam: 'Query Parameter (?lang=EN)',
        header: 'HTTP Header',
        none: 'No Language Handling',
        languageHandlingHelp: 'Choose how the appointment services API should receive language information',
        languageHeaderName: 'Language Header Name',
        languageHeaderHelp: 'Name of the HTTP header to send language information (e.g., Accept-Language)',
        translationBehavior: 'Translation Behavior',
        queryParamBehavior: 'API receives ?lang=EN parameter. If API returns English, DeepL will translate.',
        headerBehavior: 'API receives language in HTTP header. If API returns English, DeepL will translate.',
        noneBehavior: 'API receives no language info. DeepL will always translate responses to target language.',
        genericServicesLanguage: 'Generic Services Language Configuration',
        genericLanguageHandlingHelp: 'Choose how the generic services API should receive language information',
        genericQueryParamBehavior: 'API receives ?lang=EN parameter. If API returns English, DeepL will translate.',
        genericHeaderBehavior: 'API receives language in HTTP header. If API returns English, DeepL will translate.',
        genericNoneBehavior: 'API receives no language info. DeepL will always translate responses to target language.'
      },
      backoffice: {
        title: 'Backoffice',
        queueManagement: 'Queue Management',
        currentEntry: 'Current Entry',
        waitingQueue: 'Waiting Queue',
        activityLog: 'Activity Log',
        statistics: 'Statistics',
        callNext: 'Call Next',
        complete: 'Complete',
        skip: 'Skip',
        recall: 'Recall'
      },
      tv: {
        title: 'Queue Display',
        nowServing: 'Now Serving',
        nextInLine: 'Next in Line',
        waiting: 'Waiting',
        pleaseWait: 'Please wait for your turn',
        yourTurn: 'Your turn is next!',
        called: 'Called'
      },
      mobile: {
        title: 'Mobile Queue',
        yourPosition: 'Your Position',
        estimatedWait: 'Estimated Wait',
        status: 'Status',
        waiting: 'Waiting',
        called: 'Called',
        completed: 'Completed'
      }
    });

    // Spanish
    this.addTranslations('es', {
      common: {
        loading: 'Cargando...',
        error: 'Error',
        success: 'Éxito',
        warning: 'Advertencia',
        info: 'Información',
        cancel: 'Cancelar',
        confirm: 'Confirmar',
        save: 'Guardar',
        edit: 'Editar',
        delete: 'Eliminar',
        close: 'Cerrar',
        back: 'Atrás',
        next: 'Siguiente',
        previous: 'Anterior',
        submit: 'Enviar',
        reset: 'Restablecer',
        search: 'Buscar',
        filter: 'Filtrar',
        sort: 'Ordenar',
        refresh: 'Actualizar',
        retry: 'Reintentar',
        yes: 'Sí',
        no: 'No',
        ok: 'OK'
      },
      kiosk: {
        title: 'Quiosco de Sala de Espera',
        welcomeMessage: 'Bienvenido a nuestro sistema de sala de espera',
        insertCard: 'Por favor inserte su tarjeta de identificación',
        readingCard: 'Leyendo tarjeta...',
        cardRead: 'Tarjeta leída exitosamente',
        cardError: 'Error al leer la tarjeta',
        selectService: 'Seleccione un servicio',
        yourTicket: 'Su Ticket',
        ticketNumber: 'Número de Ticket',
        estimatedWait: 'Tiempo de Espera Estimado',
        minutes: 'minutos',
        services: {
          appointments: 'Sus Citas',
          generic: 'Servicios Generales',
          noServices: 'No hay servicios disponibles',
          loadingServices: 'Cargando servicios...',
          serviceError: 'Error al cargar servicios'
        },
        connection: {
          connected: 'Conectado',
          connecting: 'Conectando...',
          disconnected: 'Desconectado',
          connectionError: 'Error de conexión'
        }
      },
      admin: {
        title: 'Panel de Administración',
        configuration: 'Configuración',
        externalAPI: 'API Externa',
        rooms: 'Salas',
        servicePoints: 'Puntos de Servicio',
        genericServices: 'Servicios Genéricos',
        webhook: 'Webhook',
        appointmentServices: 'Servicios de Citas',
        genericServicesUrl: 'URL de Servicios Genéricos',
        webhookUrl: 'URL de Webhook',
        timeout: 'Tiempo de Espera (segundos)',
        retryAttempts: 'Intentos de Reintento',
        headers: 'Encabezados',
        addService: 'Agregar Servicio',
        serviceName: 'Nombre del Servicio',
        serviceDescription: 'Descripción del Servicio',
        duration: 'Duración (minutos)',
        enabled: 'Habilitado',
        remove: 'Eliminar',
        totalServices: 'servicios totales',
        enabledServices: 'habilitados'
      }
    });

    // Slovak
    this.addTranslations('sk', {
      common: {
        loading: 'Načítava sa...',
        error: 'Chyba',
        success: 'Úspech',
        warning: 'Upozornenie',
        info: 'Informácia',
        cancel: 'Zrušiť',
        confirm: 'Potvrdiť',
        save: 'Uložiť',
        edit: 'Upraviť',
        delete: 'Vymazať',
        close: 'Zavrieť',
        back: 'Späť',
        next: 'Ďalej',
        previous: 'Predchádzajúci',
        submit: 'Odoslať',
        reset: 'Resetovať',
        search: 'Hľadať',
        filter: 'Filtrovať',
        sort: 'Zoradiť',
        refresh: 'Obnoviť',
        retry: 'Skúsiť znova',
        yes: 'Áno',
        no: 'Nie',
        ok: 'OK'
      },
      kiosk: {
        checkin: 'Prihláste sa',
        title: 'Kiosk Čakárne',
        welcomeMessage: 'Vitajte v našom systéme čakárne',
        insertCard: 'Vložte prosím svoju ID kartu',
        readingCard: 'Čítam kartu...',
        cardRead: 'Karta úspešne prečítaná',
        cardError: 'Chyba pri čítaní karty',
        selectService: 'Vyberte službu',
        yourTicket: 'Váš Lístok',
        ticketNumber: 'Číslo Lístka',
        estimatedWait: 'Odhadovaný Čas Čakania',
        minutes: 'minút',
        cardInformation: 'Informácie o Karte',
        services: {
          appointments: 'Vaše Termíny',
          generic: 'Všeobecné Služby',
          personal: 'Osobné',
          general: 'Všeobecné',
          noServices: 'Žiadne služby nie sú dostupné',
          loadingServices: 'Načítavajú sa služby...',
          serviceError: 'Chyba pri načítavaní služieb',
          selectService: 'Vyberte službu, ktorú potrebujete',
          success: 'Ďakujeme za prihlásenie!'
        },
        connection: {
          connected: 'Pripojené',
          connecting: 'Pripája sa...',
          disconnected: 'Odpojené',
          connectionError: 'Chyba pripojenia'
        }
      },
      admin: {
        title: 'Administračný Panel',
        configuration: 'Konfigurácia',
        externalAPI: 'Externé API',
        rooms: 'Miestnosti',
        servicePoints: 'Servisné Body',
        genericServices: 'Všeobecné Služby',
        webhook: 'Webhook',
        appointmentServices: 'Služby Termínov',
        genericServicesUrl: 'URL Všeobecných Služieb',
        webhookUrl: 'Webhook URL',
        timeout: 'Časový Limit (sekundy)',
        retryAttempts: 'Počet Pokusov',
        headers: 'Hlavičky',
        addService: 'Pridať Službu',
        serviceName: 'Názov Služby',
        serviceDescription: 'Popis Služby',
        duration: 'Trvanie (minúty)',
        enabled: 'Povolené',
        remove: 'Odstrániť',
        totalServices: 'celkovo služieb',
        enabledServices: 'povolených',
        multilingualAPI: 'Konfigurácia Viacjazyčného API',
        multilingualSupport: 'Externé API podporuje viacero jazykov',
        supportedLanguages: 'Podporované Jazyky',
        multilingualHelp: 'Ak je povolené, API dostane parameter ?lang=EN alebo ?lang=SK',
        useDeepLTranslation: 'Použiť DeepL na automatický preklad',
        deepLHelp: 'Ak je povolené, odpovede externého API budú automaticky preložené pomocou DeepL',
        appointmentServicesLanguage: 'Konfigurácia Jazyka Služieb Termínov',
        languageHandlingMethod: 'Spôsob Spracovania Jazyka',
        queryParam: 'Query Parameter (?lang=EN)',
        header: 'HTTP Hlavička',
        none: 'Žiadne Spracovanie Jazyka',
        languageHandlingHelp: 'Vyberte, ako má API služieb termínov dostávať informácie o jazyku',
        languageHeaderName: 'Názov Hlavičky Jazyka',
        languageHeaderHelp: 'Názov HTTP hlavičky na odoslanie informácií o jazyku (napr. Accept-Language)',
        translationBehavior: 'Správanie Prekladu',
        queryParamBehavior: 'API dostane parameter ?lang=EN. Ak API vráti angličtinu, DeepL preloží.',
        headerBehavior: 'API dostane jazyk v HTTP hlavičke. Ak API vráti angličtinu, DeepL preloží.',
        noneBehavior: 'API nedostane žiadne informácie o jazyku. DeepL bude vždy prekladať odpovede do cieľového jazyka.',
        genericServicesLanguage: 'Konfigurácia Jazyka Generických Služieb',
        genericLanguageHandlingHelp: 'Vyberte, ako má API generických služieb dostávať informácie o jazyku',
        genericQueryParamBehavior: 'API dostane parameter ?lang=EN. Ak API vráti angličtinu, DeepL preloží.',
        genericHeaderBehavior: 'API dostane jazyk v HTTP hlavičke. Ak API vráti angličtinu, DeepL preloží.',
        genericNoneBehavior: 'API nedostane žiadne informácie o jazyku. DeepL bude vždy prekladať odpovede do cieľového jazyka.'
      },
      backoffice: {
        title: 'Backoffice',
        queueManagement: 'Správa Fronty',
        currentEntry: 'Aktuálny Záznam',
        waitingQueue: 'Fronta Čakajúcich',
        activityLog: 'Záznam Aktivity',
        statistics: 'Štatistiky',
        callNext: 'Zavolať Ďalšieho',
        complete: 'Dokončiť',
        skip: 'Preskočiť',
        recall: 'Zavolať Znova'
      },
      tv: {
        title: 'Zobrazenie Fronty',
        nowServing: 'Práve Obsluhuje',
        nextInLine: 'Ďalší v Rade',
        waiting: 'Čaká',
        pleaseWait: 'Prosím čakajte na svoj rad',
        yourTurn: 'Váš rad je ďalší!',
        called: 'Zavolaný'
      },
      mobile: {
        title: 'Mobilná Fronta',
        yourPosition: 'Vaša Pozícia',
        estimatedWait: 'Odhadované Čakanie',
        status: 'Stav',
        waiting: 'Čaká',
        called: 'Zavolaný',
        completed: 'Dokončené'
      }
    });
  }

  /**
   * Load saved language preference from localStorage
   */
  private loadSavedLanguage(): void {
    const savedLanguage = localStorage.getItem('preferred-language');
    if (savedLanguage && this.availableLanguages.some(lang => lang.code === savedLanguage)) {
      this.currentLanguage.set(savedLanguage);
    }
    this.updateDocumentLanguage();
  }

  /**
   * Save language preference to localStorage
   */
  private saveLanguagePreference(languageCode: string): void {
    localStorage.setItem('preferred-language', languageCode);
  }

  /**
   * Update document language and direction
   */
  private updateDocumentLanguage(): void {
    const config = this.currentLanguageConfig();
    document.documentElement.lang = config.code;
    document.documentElement.dir = config.direction;
  }
}
