import { Component, signal, inject, OnInit, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { ConfigService } from '../shared/services/config.service';
import { TenantService } from '@lib/tenant';
import { TranslatePipe } from '../../../../src/lib/i18n';

// TypeScript interfaces matching the Go structs
interface PriorityConfig {
  version: string;
  description: string;
  priorityModel: PriorityModel;
}

interface PriorityModel {
  algorithm: Algorithm;
  tiers: Tier[];
  fitness: FitnessConfig;
}

interface Algorithm {
  explanation: string[];
  orderingFields: string[];
}

interface Tier {
  id: number;
  name: string;
  condition: TierCondition;
  description: string;
}

interface TierCondition {
  symbolsAnyOf?: string[];
  symbolsNotAnyOf?: string[];
}

interface FitnessConfig {
  explanation: string;
  contributions: Contributions;
}

interface Contributions {
  symbolWeights: SymbolWeights;
  waitingTime: WaitingTime;
  appointmentDeviation: AppointmentDeviation;
  age: AgeConfig;
  manualOverride: ManualOverride;
}

interface SymbolWeights {
  description: string;
  values: { [key: string]: number };
}

interface WaitingTime {
  description: string;
  weightPerMinute: number;
}

interface AppointmentDeviation {
  description: string;
  earlyPenaltyPerMinute: number;
  lateBonusPerMinute: number;
}

interface AgeConfig {
  description: string;
  under6PerYearYounger: number;
  over65PerYearOlder: number;
  ageThresholdSenior: number;
}

interface ManualOverride {
  description: string;
  enabled: boolean;
  weight: number;
}

@Component({
  selector: 'app-priority-configuration',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslatePipe],
  templateUrl: './priority-configuration.html',
  styleUrl: './priority-configuration.scss'
})
export class PriorityConfigurationComponent implements OnInit {
  private tenantService = inject(TenantService);
  private currentTenantId = '';

  constructor(
    private http: HttpClient,
    private configService: ConfigService
  ) {
    // Watch for tenant selection and load configuration when tenant is selected or changes
    effect(() => {
      const tenantId = this.tenantService.selectedTenantId();

      if (tenantId && tenantId !== this.currentTenantId) {
        this.currentTenantId = tenantId;
        this.loadConfiguration();
      } else if (!tenantId) {
        this.currentTenantId = '';
      }
    });
  }

  ngOnInit(): void {
    const tenantId = this.tenantService.selectedTenantId();
    if (tenantId && tenantId !== this.currentTenantId) {
      this.currentTenantId = tenantId;
      this.loadConfiguration();
    }
  }

  isSaving = signal(false);
  isLoading = signal(false);
  lastUpdated = signal('Never');

  // Use regular object for form binding
  priorityConfig: PriorityConfig = {
    version: '1.0',
    description: '',
    priorityModel: {
      algorithm: {
        explanation: [],
        orderingFields: ['tier', 'fitnessScore']
      },
      tiers: [],
      fitness: {
        explanation: '',
        contributions: {
          symbolWeights: {
            description: '',
            values: {}
          },
          waitingTime: {
            description: '',
            weightPerMinute: 0
          },
          appointmentDeviation: {
            description: '',
            earlyPenaltyPerMinute: 0,
            lateBonusPerMinute: 0
          },
          age: {
            description: '',
            under6PerYearYounger: 0,
            over65PerYearOlder: 0,
            ageThresholdSenior: 65
          },
          manualOverride: {
            description: '',
            enabled: false,
            weight: 0
          }
        }
      }
    }
  };

  loadConfiguration(): void {
    const tenantId = this.tenantService.selectedTenantId();
    console.log(`[PriorityConfigurationComponent] Loading configuration for tenant: ${tenantId || 'none'}`);

    this.isLoading.set(true);

    this.http.get<PriorityConfig>(this.configService.adminPriorityConfigUrl)
      .subscribe({
        next: (response) => {
          console.log('[PriorityConfigurationComponent] Received priority config:', response);
          this.isLoading.set(false);

          if (response && response.priorityModel) {
            // Ensure all nested objects exist
            if (!response.priorityModel.algorithm) {
              response.priorityModel.algorithm = { explanation: [], orderingFields: ['tier', 'fitnessScore'] };
            }
            if (!response.priorityModel.fitness) {
              response.priorityModel.fitness = this.priorityConfig.priorityModel.fitness;
            }
            this.priorityConfig = response;
            this.lastUpdated.set(new Date().toLocaleString());
          } else {
            console.log('[PriorityConfigurationComponent] No config found, using defaults');
          }
        },
        error: (error) => {
          console.error('[PriorityConfigurationComponent] Failed to load configuration:', error);
          this.isLoading.set(false);

          // If it's a 404, that's okay - no config exists yet
          if (error.status === 404) {
            console.log('[PriorityConfigurationComponent] No config exists yet, using defaults');
          } else {
            alert('Failed to load priority configuration. Please try again.');
          }
        }
      });
  }

  loadDefaultConfiguration(): void {
    if (!confirm('This will replace your current configuration with the default. Continue?')) {
      return;
    }

    this.isLoading.set(true);

    this.http.get<PriorityConfig>(this.configService.adminPriorityConfigDefaultUrl)
      .subscribe({
        next: (response) => {
          console.log('[PriorityConfigurationComponent] Received default priority config:', response);
          if (response) {
            this.priorityConfig = response;
          }
          this.isLoading.set(false);
        },
        error: (error) => {
          console.error('[PriorityConfigurationComponent] Failed to load default configuration:', error);
          alert('Failed to load default configuration');
          this.isLoading.set(false);
        }
      });
  }

  saveConfiguration(): void {
    this.isSaving.set(true);

    this.http.put<PriorityConfig>(this.configService.adminPriorityConfigUrl, this.priorityConfig)
      .subscribe({
        next: (response) => {
          this.isSaving.set(false);
          this.lastUpdated.set(new Date().toLocaleString());
          console.log('[PriorityConfigurationComponent] Configuration saved successfully:', response);
          alert('Priority configuration saved successfully!');
        },
        error: (error) => {
          this.isSaving.set(false);
          console.error('[PriorityConfigurationComponent] Failed to save configuration:', error);
          const errorMessage = error.error?.message || error.message || 'Unknown error';
          alert(`Failed to save priority configuration: ${errorMessage}`);
        }
      });
  }

  // Tier management methods
  addTier(): void {
    const newTier: Tier = {
      id: this.priorityConfig.priorityModel.tiers.length,
      name: `Tier ${this.priorityConfig.priorityModel.tiers.length}`,
      description: '',
      condition: {
        symbolsAnyOf: [],
        symbolsNotAnyOf: []
      }
    };
    this.priorityConfig.priorityModel.tiers.push(newTier);
  }

  removeTier(index: number): void {
    if (!confirm('Are you sure you want to remove this tier?')) {
      return;
    }

    this.priorityConfig.priorityModel.tiers.splice(index, 1);

    // Reindex tiers
    this.priorityConfig.priorityModel.tiers.forEach((tier, idx) => {
      tier.id = idx;
    });
  }

  moveTierUp(index: number): void {
    if (index === 0) return;

    const tiers = this.priorityConfig.priorityModel.tiers;
    [tiers[index - 1], tiers[index]] = [tiers[index], tiers[index - 1]];

    // Reindex tiers
    tiers.forEach((tier, idx) => {
      tier.id = idx;
    });
  }

  moveTierDown(index: number): void {
    const tiers = this.priorityConfig.priorityModel.tiers;
    if (index === tiers.length - 1) return;

    [tiers[index], tiers[index + 1]] = [tiers[index + 1], tiers[index]];

    // Reindex tiers
    tiers.forEach((tier, idx) => {
      tier.id = idx;
    });
  }

  // Symbol management methods
  addSymbolToTier(tierIndex: number, listType: 'anyOf' | 'notAnyOf'): void {
    const symbol = prompt(`Enter symbol to add to ${listType === 'anyOf' ? 'Any Of' : 'Not Any Of'} list:`);
    if (!symbol || symbol.trim() === '') return;

    const tier = this.priorityConfig.priorityModel.tiers[tierIndex];
    const condition = tier.condition;

    if (listType === 'anyOf') {
      if (!condition.symbolsAnyOf) {
        condition.symbolsAnyOf = [];
      }
      condition.symbolsAnyOf.push(symbol.trim());
    } else {
      if (!condition.symbolsNotAnyOf) {
        condition.symbolsNotAnyOf = [];
      }
      condition.symbolsNotAnyOf.push(symbol.trim());
    }
  }

  removeSymbolFromTier(tierIndex: number, listType: 'anyOf' | 'notAnyOf', symbolIndex: number): void {
    const tier = this.priorityConfig.priorityModel.tiers[tierIndex];
    const condition = tier.condition;

    if (listType === 'anyOf' && condition.symbolsAnyOf) {
      condition.symbolsAnyOf.splice(symbolIndex, 1);
    } else if (listType === 'notAnyOf' && condition.symbolsNotAnyOf) {
      condition.symbolsNotAnyOf.splice(symbolIndex, 1);
    }
  }

  // Symbol weights management - use getter to avoid creating new arrays on every change detection
  get symbolWeightEntries(): Array<{key: string, value: number}> {
    const values = this.priorityConfig.priorityModel.fitness.contributions.symbolWeights.values;
    if (!values) {
      return [];
    }
    return Object.keys(values).map(key => ({ key, value: values[key] }));
  }

  addSymbolWeight(): void {
    const symbol = prompt('Enter symbol name:');
    if (!symbol || symbol.trim() === '') return;

    const weight = prompt('Enter weight (number):');
    if (!weight) return;

    const weightNum = parseFloat(weight);
    if (isNaN(weightNum)) {
      alert('Weight must be a number');
      return;
    }

    this.priorityConfig.priorityModel.fitness.contributions.symbolWeights.values[symbol.trim()] = weightNum;
  }

  removeSymbolWeight(key: string): void {
    delete this.priorityConfig.priorityModel.fitness.contributions.symbolWeights.values[key];
  }

  updateSymbolWeight(key: string, newValue: number): void {
    this.priorityConfig.priorityModel.fitness.contributions.symbolWeights.values[key] = newValue;
  }

  // Helper method to track tiers by ID
  trackByTierId(index: number, tier: Tier): number {
    return tier.id;
  }

  // Helper method to track symbol weight entries by key
  trackByKey(index: number, item: {key: string, value: number}): string {
    return item.key;
  }
}
