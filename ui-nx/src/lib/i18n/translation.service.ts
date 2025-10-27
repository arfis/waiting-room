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
   * Load default translations
   */
  private loadTranslations(): void {
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
        enabledServices: 'enabled'
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
        enabledServices: 'povolených'
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
