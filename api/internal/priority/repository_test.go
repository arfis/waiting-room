package priority

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	// Verify basic structure
	if config == nil {
		t.Fatal("GetDefaultConfig() returned nil")
	}

	if config.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", config.Version)
	}

	// Verify tiers
	if len(config.PriorityModel.Tiers) != 3 {
		t.Errorf("Expected 3 tiers, got %d", len(config.PriorityModel.Tiers))
	}

	// Verify tier IDs and names
	expectedTiers := map[int]string{
		0: "STATIM",
		1: "VIP",
		2: "NORMAL",
	}

	for _, tier := range config.PriorityModel.Tiers {
		if expectedName, ok := expectedTiers[tier.ID]; ok {
			if tier.Name != expectedName {
				t.Errorf("Tier %d: expected name %s, got %s", tier.ID, expectedName, tier.Name)
			}
		} else {
			t.Errorf("Unexpected tier ID: %d", tier.ID)
		}
	}

	// Verify symbol weights
	contrib := config.PriorityModel.Fitness.Contributions
	if contrib.SymbolWeights.Values["STATIM"] != -1000 {
		t.Errorf("Expected STATIM weight -1000, got %f", contrib.SymbolWeights.Values["STATIM"])
	}
	if contrib.SymbolWeights.Values["VIP"] != -500 {
		t.Errorf("Expected VIP weight -500, got %f", contrib.SymbolWeights.Values["VIP"])
	}
	if contrib.SymbolWeights.Values["IMMOBILE"] != -50 {
		t.Errorf("Expected IMMOBILE weight -50, got %f", contrib.SymbolWeights.Values["IMMOBILE"])
	}

	// Verify waiting time
	if contrib.WaitingTime.WeightPerMinute != -1 {
		t.Errorf("Expected waiting time weight -1, got %f", contrib.WaitingTime.WeightPerMinute)
	}

	// Verify appointment deviation
	if contrib.AppointmentDeviation.EarlyPenaltyPerMinute != 2 {
		t.Errorf("Expected early penalty 2, got %f", contrib.AppointmentDeviation.EarlyPenaltyPerMinute)
	}
	if contrib.AppointmentDeviation.LateBonusPerMinute != -3 {
		t.Errorf("Expected late bonus -3, got %f", contrib.AppointmentDeviation.LateBonusPerMinute)
	}

	// Verify age config
	if contrib.Age.Under6PerYearYounger != -5 {
		t.Errorf("Expected under6 weight -5, got %f", contrib.Age.Under6PerYearYounger)
	}
	if contrib.Age.Over65PerYearOlder != -1 {
		t.Errorf("Expected over65 weight -1, got %f", contrib.Age.Over65PerYearOlder)
	}
	if contrib.Age.AgeThresholdSenior != 65 {
		t.Errorf("Expected age threshold 65, got %d", contrib.Age.AgeThresholdSenior)
	}

	// Verify manual override
	if !contrib.ManualOverride.Enabled {
		t.Error("Expected manual override to be enabled")
	}
	if contrib.ManualOverride.Weight != 1 {
		t.Errorf("Expected manual override weight 1, got %f", contrib.ManualOverride.Weight)
	}
}

func TestDefaultConfig_TierConditions(t *testing.T) {
	config := GetDefaultConfig()

	tests := []struct {
		tierID      int
		description string
		checkFunc   func(Condition) bool
	}{
		{
			tierID:      0,
			description: "STATIM tier should require STATIM symbol",
			checkFunc: func(c Condition) bool {
				return len(c.SymbolsAnyOf) == 1 && c.SymbolsAnyOf[0] == "STATIM"
			},
		},
		{
			tierID:      1,
			description: "VIP tier should require VIP but not STATIM",
			checkFunc: func(c Condition) bool {
				hasVIP := false
				for _, s := range c.SymbolsAnyOf {
					if s == "VIP" {
						hasVIP = true
					}
				}
				hasStatimExclusion := false
				for _, s := range c.SymbolsNotAnyOf {
					if s == "STATIM" {
						hasStatimExclusion = true
					}
				}
				return hasVIP && hasStatimExclusion
			},
		},
		{
			tierID:      2,
			description: "NORMAL tier should exclude STATIM and VIP",
			checkFunc: func(c Condition) bool {
				hasStatim := false
				hasVIP := false
				for _, s := range c.SymbolsNotAnyOf {
					if s == "STATIM" {
						hasStatim = true
					}
					if s == "VIP" {
						hasVIP = true
					}
				}
				return hasStatim && hasVIP
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			var tier *Tier
			for i := range config.PriorityModel.Tiers {
				if config.PriorityModel.Tiers[i].ID == tt.tierID {
					tier = &config.PriorityModel.Tiers[i]
					break
				}
			}

			if tier == nil {
				t.Fatalf("Tier %d not found", tt.tierID)
			}

			if !tt.checkFunc(tier.Condition) {
				t.Errorf("Tier %d condition check failed: %s", tt.tierID, tt.description)
			}
		})
	}
}

func TestDefaultConfig_WorksWithCalculator(t *testing.T) {
	config := GetDefaultConfig()
	calculator := NewCalculator(config)

	if calculator == nil {
		t.Fatal("NewCalculator() with default config returned nil")
	}

	// Test basic calculation
	now := time.Now()
	input := CalculationInput{
		Symbols:     []string{"VIP"},
		ArrivalTime: now.Add(-10 * time.Minute),
		CurrentTime: now,
	}

	result := calculator.Calculate(input)

	if result.Tier != 1 {
		t.Errorf("VIP patient should be tier 1, got %d", result.Tier)
	}

	// Should have VIP weight (-500) + 10 minutes waiting (-10)
	expectedScore := -510.0
	if result.FitnessScore != expectedScore {
		t.Errorf("Expected fitness score %f, got %f", expectedScore, result.FitnessScore)
	}
}

// Integration test with MongoDB (requires MongoDB running)
// Skipped by default, run with: go test -tags=integration
func TestRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Try to connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test:", err)
		return
	}
	defer client.Disconnect(ctx)

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		t.Skip("Cannot ping MongoDB, skipping integration test:", err)
		return
	}

	// Create test database and repository
	testDB := client.Database("test_priority_" + time.Now().Format("20060102150405"))
	defer testDB.Drop(ctx)

	repo := NewRepository(testDB)

	t.Run("GetConfig returns default when none exists", func(t *testing.T) {
		config, err := repo.GetConfig(ctx, "test-tenant", "test-section")
		if err != nil {
			t.Fatalf("GetConfig() error = %v", err)
		}
		if config == nil {
			t.Fatal("GetConfig() returned nil config")
		}
		if config.Version != "1.0" {
			t.Errorf("Expected default config version 1.0, got %s", config.Version)
		}
	})

	t.Run("SaveConfig and GetConfig roundtrip", func(t *testing.T) {
		customConfig := &PriorityConfig{
			Version:     "2.0",
			Description: "Test config",
			PriorityModel: PriorityModel{
				Tiers: []Tier{
					{ID: 0, Name: "URGENT", Condition: Condition{SymbolsAnyOf: []string{"URGENT"}}},
					{ID: 1, Name: "NORMAL", Condition: Condition{}},
				},
			},
		}

		// Save config
		err := repo.SaveConfig(ctx, customConfig, "tenant1", "section1")
		if err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		// Retrieve config
		retrieved, err := repo.GetConfig(ctx, "tenant1", "section1")
		if err != nil {
			t.Fatalf("GetConfig() error = %v", err)
		}

		if retrieved.Version != "2.0" {
			t.Errorf("Expected version 2.0, got %s", retrieved.Version)
		}
		if len(retrieved.PriorityModel.Tiers) != 2 {
			t.Errorf("Expected 2 tiers, got %d", len(retrieved.PriorityModel.Tiers))
		}
	})

	t.Run("Tenant-level config fallback", func(t *testing.T) {
		tenantConfig := &PriorityConfig{
			Version:     "3.0",
			Description: "Tenant-level config",
			PriorityModel: PriorityModel{
				Tiers: []Tier{{ID: 0, Name: "DEFAULT"}},
			},
		}

		// Save tenant-level config (no section)
		err := repo.SaveConfig(ctx, tenantConfig, "tenant2", "")
		if err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		// Try to get section-specific config (should fall back to tenant-level)
		retrieved, err := repo.GetConfig(ctx, "tenant2", "some-section")
		if err != nil {
			t.Fatalf("GetConfig() error = %v", err)
		}

		// Should get tenant-level config
		if retrieved.Version != "3.0" {
			t.Errorf("Expected tenant-level config version 3.0, got %s", retrieved.Version)
		}
	})

	t.Run("Section-specific config overrides tenant-level", func(t *testing.T) {
		// Save tenant-level config
		tenantConfig := &PriorityConfig{
			Version:     "4.0",
			Description: "Tenant config",
		}
		err := repo.SaveConfig(ctx, tenantConfig, "tenant3", "")
		if err != nil {
			t.Fatalf("SaveConfig() tenant error = %v", err)
		}

		// Save section-specific config
		sectionConfig := &PriorityConfig{
			Version:     "5.0",
			Description: "Section config",
		}
		err = repo.SaveConfig(ctx, sectionConfig, "tenant3", "special-section")
		if err != nil {
			t.Fatalf("SaveConfig() section error = %v", err)
		}

		// Get section-specific config
		retrieved, err := repo.GetConfig(ctx, "tenant3", "special-section")
		if err != nil {
			t.Fatalf("GetConfig() error = %v", err)
		}

		// Should get section-specific config, not tenant-level
		if retrieved.Version != "5.0" {
			t.Errorf("Expected section config version 5.0, got %s", retrieved.Version)
		}
	})
}

func BenchmarkCalculate(b *testing.B) {
	config := GetDefaultConfig()
	calculator := NewCalculator(config)
	now := time.Now()
	arrivalTime := now.Add(-15 * time.Minute)
	appointmentTime := now.Add(-5 * time.Minute)
	age := 70
	manualOverride := -10.0

	input := CalculationInput{
		Symbols:         []string{"VIP", "IMMOBILE"},
		AppointmentTime: &appointmentTime,
		Age:             &age,
		ManualOverride:  &manualOverride,
		ArrivalTime:     arrivalTime,
		CurrentTime:     now,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculator.Calculate(input)
	}
}

func BenchmarkCalculateTier(b *testing.B) {
	calculator := NewCalculator(GetDefaultConfig())
	symbols := []string{"VIP", "IMMOBILE"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculator.calculateTier(symbols)
	}
}

func BenchmarkCalculateFitnessScore(b *testing.B) {
	calculator := NewCalculator(GetDefaultConfig())
	now := time.Now()
	arrivalTime := now.Add(-15 * time.Minute)
	appointmentTime := now.Add(-5 * time.Minute)
	age := 70
	manualOverride := -10.0

	input := CalculationInput{
		Symbols:         []string{"VIP", "IMMOBILE"},
		AppointmentTime: &appointmentTime,
		Age:             &age,
		ManualOverride:  &manualOverride,
		ArrivalTime:     arrivalTime,
		CurrentTime:     now,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculator.calculateFitnessScore(input)
	}
}
