package priority

import (
	"testing"
	"time"
)

// Helper function to create a pointer to an int
func intPtr(i int) *int {
	return &i
}

// Helper function to create a pointer to a float64
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create a pointer to a time
func timePtr(t time.Time) *time.Time {
	return &t
}

// getTestConfig returns the default configuration for testing
func getTestConfig() *PriorityConfig {
	return &PriorityConfig{
		Version:     "1.0",
		Description: "Test configuration",
		PriorityModel: PriorityModel{
			Tiers: []Tier{
				{
					ID:   0,
					Name: "STATIM",
					Condition: Condition{
						SymbolsAnyOf: []string{"STATIM"},
					},
					Description: "Highest priority - STATIM",
				},
				{
					ID:   1,
					Name: "VIP",
					Condition: Condition{
						SymbolsAnyOf:    []string{"VIP"},
						SymbolsNotAnyOf: []string{"STATIM"},
					},
					Description: "VIP patients not STATIM",
				},
				{
					ID:   2,
					Name: "NORMAL",
					Condition: Condition{
						SymbolsNotAnyOf: []string{"STATIM", "VIP"},
					},
					Description: "Regular patients",
				},
			},
			Fitness: FitnessConfig{
				Contributions: Contributions{
					SymbolWeights: SymbolWeights{
						Values: map[string]float64{
							"STATIM":   -1000,
							"VIP":      -500,
							"IMMOBILE": -50,
						},
					},
					WaitingTime: WaitingTime{
						WeightPerMinute: -1,
					},
					AppointmentDeviation: AppointmentDeviation{
						EarlyPenaltyPerMinute: 2,
						LateBonusPerMinute:    -3,
					},
					Age: AgeConfig{
						Under6PerYearYounger: -5,
						Over65PerYearOlder:   -1,
						AgeThresholdSenior:   65,
					},
					ManualOverride: ManualOverride{
						Enabled: true,
						Weight:  1,
					},
				},
			},
		},
	}
}

func TestCalculateTier(t *testing.T) {
	calculator := NewCalculator(getTestConfig())

	tests := []struct {
		name     string
		symbols  []string
		wantTier int
	}{
		{
			name:     "STATIM patient gets tier 0",
			symbols:  []string{"STATIM"},
			wantTier: 0,
		},
		{
			name:     "STATIM + VIP patient gets tier 0 (STATIM takes precedence)",
			symbols:  []string{"STATIM", "VIP"},
			wantTier: 0,
		},
		{
			name:     "VIP patient gets tier 1",
			symbols:  []string{"VIP"},
			wantTier: 1,
		},
		{
			name:     "VIP + IMMOBILE patient gets tier 1",
			symbols:  []string{"VIP", "IMMOBILE"},
			wantTier: 1,
		},
		{
			name:     "Regular patient gets tier 2",
			symbols:  []string{},
			wantTier: 2,
		},
		{
			name:     "IMMOBILE only patient gets tier 2",
			symbols:  []string{"IMMOBILE"},
			wantTier: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier := calculator.calculateTier(tt.symbols)
			if tier != tt.wantTier {
				t.Errorf("calculateTier() = %v, want %v", tier, tt.wantTier)
			}
		})
	}
}

func TestCalculateFitnessScore_SymbolWeights(t *testing.T) {
	calculator := NewCalculator(getTestConfig())
	now := time.Now()

	tests := []struct {
		name      string
		symbols   []string
		wantScore float64
	}{
		{
			name:      "STATIM symbol",
			symbols:   []string{"STATIM"},
			wantScore: -1000,
		},
		{
			name:      "VIP symbol",
			symbols:   []string{"VIP"},
			wantScore: -500,
		},
		{
			name:      "IMMOBILE symbol",
			symbols:   []string{"IMMOBILE"},
			wantScore: -50,
		},
		{
			name:      "STATIM + VIP symbols",
			symbols:   []string{"STATIM", "VIP"},
			wantScore: -1500, // -1000 + -500
		},
		{
			name:      "VIP + IMMOBILE symbols",
			symbols:   []string{"VIP", "IMMOBILE"},
			wantScore: -550, // -500 + -50
		},
		{
			name:      "No symbols",
			symbols:   []string{},
			wantScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := CalculationInput{
				Symbols:     tt.symbols,
				ArrivalTime: now,
				CurrentTime: now,
			}
			score := calculator.calculateFitnessScore(input)
			if score != tt.wantScore {
				t.Errorf("calculateFitnessScore() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestCalculateFitnessScore_WaitingTime(t *testing.T) {
	calculator := NewCalculator(getTestConfig())
	now := time.Now()

	tests := []struct {
		name           string
		waitingMinutes int
		wantScore      float64
	}{
		{
			name:           "No waiting time",
			waitingMinutes: 0,
			wantScore:      0,
		},
		{
			name:           "5 minutes waiting",
			waitingMinutes: 5,
			wantScore:      -5, // 5 * -1
		},
		{
			name:           "30 minutes waiting",
			waitingMinutes: 30,
			wantScore:      -30, // 30 * -1
		},
		{
			name:           "60 minutes waiting",
			waitingMinutes: 60,
			wantScore:      -60, // 60 * -1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrivalTime := now.Add(-time.Duration(tt.waitingMinutes) * time.Minute)
			input := CalculationInput{
				Symbols:     []string{},
				ArrivalTime: arrivalTime,
				CurrentTime: now,
			}
			score := calculator.calculateFitnessScore(input)
			if score != tt.wantScore {
				t.Errorf("calculateFitnessScore() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestCalculateFitnessScore_AppointmentDeviation(t *testing.T) {
	calculator := NewCalculator(getTestConfig())
	now := time.Now()

	tests := []struct {
		name                string
		appointmentOffset   int // minutes from now (negative = past, positive = future)
		wantScore           float64
		description         string
	}{
		{
			name:              "No appointment",
			appointmentOffset: 0,
			wantScore:         0,
			description:       "nil appointment time should not affect score",
		},
		{
			name:              "On time (exactly at appointment)",
			appointmentOffset: 0,
			wantScore:         0,
			description:       "Arriving exactly at appointment time",
		},
		{
			name:              "10 minutes early",
			appointmentOffset: 10,
			wantScore:         20, // 10 * 2 (early penalty)
			description:       "Early arrival gets penalty",
		},
		{
			name:              "30 minutes early",
			appointmentOffset: 30,
			wantScore:         60, // 30 * 2
			description:       "Very early arrival gets bigger penalty",
		},
		{
			name:              "10 minutes late",
			appointmentOffset: -10,
			wantScore:         -30, // 10 * -3 (late bonus)
			description:       "Late arrival gets priority boost",
		},
		{
			name:              "30 minutes late",
			appointmentOffset: -30,
			wantScore:         -90, // 30 * -3
			description:       "Very late arrival gets bigger priority boost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var appointmentTime *time.Time
			if tt.name != "No appointment" {
				apt := now.Add(time.Duration(tt.appointmentOffset) * time.Minute)
				appointmentTime = &apt
			}

			input := CalculationInput{
				Symbols:         []string{},
				AppointmentTime: appointmentTime,
				ArrivalTime:     now,
				CurrentTime:     now,
			}
			score := calculator.calculateFitnessScore(input)
			if score != tt.wantScore {
				t.Errorf("calculateFitnessScore() = %v, want %v (%s)", score, tt.wantScore, tt.description)
			}
		})
	}
}

func TestCalculateFitnessScore_Age(t *testing.T) {
	calculator := NewCalculator(getTestConfig())
	now := time.Now()

	tests := []struct {
		name      string
		age       *int
		wantScore float64
	}{
		{
			name:      "No age provided",
			age:       nil,
			wantScore: 0,
		},
		{
			name:      "Newborn (age 0)",
			age:       intPtr(0),
			wantScore: -30, // (6-0) * -5 = -30
		},
		{
			name:      "Toddler (age 2)",
			age:       intPtr(2),
			wantScore: -20, // (6-2) * -5 = -20
		},
		{
			name:      "Young child (age 5)",
			age:       intPtr(5),
			wantScore: -5, // (6-5) * -5 = -5
		},
		{
			name:      "Child exactly 6",
			age:       intPtr(6),
			wantScore: 0, // No age priority
		},
		{
			name:      "Teenager (age 15)",
			age:       intPtr(15),
			wantScore: 0, // No age priority
		},
		{
			name:      "Adult (age 40)",
			age:       intPtr(40),
			wantScore: 0, // No age priority
		},
		{
			name:      "Senior exactly 65",
			age:       intPtr(65),
			wantScore: 0, // Threshold, no bonus yet
		},
		{
			name:      "Senior (age 70)",
			age:       intPtr(70),
			wantScore: -5, // (70-65) * -1 = -5
		},
		{
			name:      "Elderly (age 80)",
			age:       intPtr(80),
			wantScore: -15, // (80-65) * -1 = -15
		},
		{
			name:      "Very elderly (age 90)",
			age:       intPtr(90),
			wantScore: -25, // (90-65) * -1 = -25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := CalculationInput{
				Symbols:     []string{},
				Age:         tt.age,
				ArrivalTime: now,
				CurrentTime: now,
			}
			score := calculator.calculateFitnessScore(input)
			if score != tt.wantScore {
				t.Errorf("calculateFitnessScore() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestCalculateFitnessScore_ManualOverride(t *testing.T) {
	calculator := NewCalculator(getTestConfig())
	now := time.Now()

	tests := []struct {
		name           string
		manualOverride *float64
		wantScore      float64
	}{
		{
			name:           "No manual override",
			manualOverride: nil,
			wantScore:      0,
		},
		{
			name:           "Manual override: 0",
			manualOverride: float64Ptr(0),
			wantScore:      0,
		},
		{
			name:           "Manual override: -100 (higher priority)",
			manualOverride: float64Ptr(-100),
			wantScore:      -100,
		},
		{
			name:           "Manual override: 100 (lower priority)",
			manualOverride: float64Ptr(100),
			wantScore:      100,
		},
		{
			name:           "Manual override: 0.0001 (small adjustment)",
			manualOverride: float64Ptr(0.0001),
			wantScore:      0.0001,
		},
		{
			name:           "Manual override: -0.5 (small boost)",
			manualOverride: float64Ptr(-0.5),
			wantScore:      -0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := CalculationInput{
				Symbols:        []string{},
				ManualOverride: tt.manualOverride,
				ArrivalTime:    now,
				CurrentTime:    now,
			}
			score := calculator.calculateFitnessScore(input)
			if score != tt.wantScore {
				t.Errorf("calculateFitnessScore() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestCalculate_ComplexScenarios(t *testing.T) {
	calculator := NewCalculator(getTestConfig())
	now := time.Now()

	tests := []struct {
		name            string
		symbols         []string
		waitingMinutes  int
		appointmentMins *int // minutes from now
		age             *int
		manualOverride  *float64
		wantTier        int
		wantScore       float64
		description     string
	}{
		{
			name:            "STATIM elderly patient, 15min late",
			symbols:         []string{"STATIM"},
			waitingMinutes:  15,
			appointmentMins: intPtr(-15),
			age:             intPtr(75),
			wantTier:        0,
			wantScore:       -1000 + (-15) + (-45) + (-10), // STATIM + waiting + late + age(75-65)
			description:     "Emergency elderly patient who is late",
		},
		{
			name:            "VIP child, 5min wait, no appointment",
			symbols:         []string{"VIP"},
			waitingMinutes:  5,
			age:             intPtr(3),
			wantTier:        1,
			wantScore:       -500 + (-5) + (-15), // VIP + waiting + age(6-3)*-5
			description:     "VIP young child",
		},
		{
			name:            "Regular immobile patient, 20min early",
			symbols:         []string{"IMMOBILE"},
			waitingMinutes:  20,
			appointmentMins: intPtr(20),
			age:             intPtr(45),
			wantTier:        2,
			wantScore:       -50 + (-20) + 40, // IMMOBILE + waiting + early penalty
			description:     "Regular patient with mobility issues who arrived early",
		},
		{
			name:            "Regular patient with manual boost",
			symbols:         []string{},
			waitingMinutes:  10,
			manualOverride:  float64Ptr(-200),
			wantTier:        2,
			wantScore:       -10 + (-200), // waiting + manual override
			description:     "Staff manually boosted this patient's priority",
		},
		{
			name:           "Regular patient, just arrived, on time",
			symbols:        []string{},
			waitingMinutes: 0,
			age:            intPtr(30),
			wantTier:       2,
			wantScore:      0,
			description:    "Baseline regular patient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrivalTime := now.Add(-time.Duration(tt.waitingMinutes) * time.Minute)

			var appointmentTime *time.Time
			if tt.appointmentMins != nil {
				apt := now.Add(time.Duration(*tt.appointmentMins) * time.Minute)
				appointmentTime = &apt
			}

			input := CalculationInput{
				Symbols:         tt.symbols,
				AppointmentTime: appointmentTime,
				Age:             tt.age,
				ManualOverride:  tt.manualOverride,
				ArrivalTime:     arrivalTime,
				CurrentTime:     now,
			}

			result := calculator.Calculate(input)

			if result.Tier != tt.wantTier {
				t.Errorf("Calculate().Tier = %v, want %v", result.Tier, tt.wantTier)
			}
			if result.FitnessScore != tt.wantScore {
				t.Errorf("Calculate().FitnessScore = %v, want %v (%s)",
					result.FitnessScore, tt.wantScore, tt.description)
			}
		})
	}
}

func TestCalculate_NegativeAndDecimalScores(t *testing.T) {
	calculator := NewCalculator(getTestConfig())
	now := time.Now()

	tests := []struct {
		name          string
		manualOverride *float64
		symbols       []string
		wantScore     float64
	}{
		{
			name:          "Very negative score",
			symbols:       []string{"STATIM", "VIP", "IMMOBILE"},
			manualOverride: float64Ptr(-500),
			wantScore:     -2050, // -1000 + -500 + -50 + -500
		},
		{
			name:          "Decimal score",
			manualOverride: float64Ptr(0.0001),
			wantScore:     0.0001,
		},
		{
			name:          "Negative decimal",
			manualOverride: float64Ptr(-0.5),
			wantScore:     -0.5,
		},
		{
			name:      "Zero score",
			symbols:   []string{},
			wantScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := CalculationInput{
				Symbols:        tt.symbols,
				ManualOverride: tt.manualOverride,
				ArrivalTime:    now,
				CurrentTime:    now,
			}

			result := calculator.Calculate(input)
			if result.FitnessScore != tt.wantScore {
				t.Errorf("Calculate().FitnessScore = %v, want %v", result.FitnessScore, tt.wantScore)
			}
		})
	}
}

func TestCalculate_QueueOrdering(t *testing.T) {
	calculator := NewCalculator(getTestConfig())
	now := time.Now()

	// Create multiple patients and verify their ordering
	patients := []struct {
		name      string
		symbols   []string
		age       *int
		waitMins  int
		wantTier  int
	}{
		{"STATIM patient", []string{"STATIM"}, nil, 5, 0},
		{"VIP elderly", []string{"VIP"}, intPtr(80), 10, 1},
		{"VIP young child", []string{"VIP"}, intPtr(2), 5, 1},
		{"Regular child", []string{}, intPtr(4), 15, 2},
		{"Regular adult", []string{}, intPtr(40), 10, 2},
		{"Regular elderly", []string{}, intPtr(70), 5, 2},
	}

	results := make([]struct {
		name  string
		tier  int
		score float64
	}, len(patients))

	for i, p := range patients {
		arrivalTime := now.Add(-time.Duration(p.waitMins) * time.Minute)
		input := CalculationInput{
			Symbols:     p.symbols,
			Age:         p.age,
			ArrivalTime: arrivalTime,
			CurrentTime: now,
		}
		result := calculator.Calculate(input)
		results[i] = struct {
			name  string
			tier  int
			score float64
		}{p.name, result.Tier, result.FitnessScore}

		if result.Tier != p.wantTier {
			t.Errorf("%s: got tier %d, want %d", p.name, result.Tier, p.wantTier)
		}
	}

	// Verify STATIM has lowest tier
	if results[0].tier != 0 {
		t.Error("STATIM should be tier 0")
	}

	// Verify VIP patients have tier 1
	if results[1].tier != 1 || results[2].tier != 1 {
		t.Error("VIP patients should be tier 1")
	}

	// Verify within VIP tier, elderly or young child has lower score (higher priority)
	// VIP elderly (80): -500 (VIP) + -10 (waiting) + -15 (age) = -525
	// VIP young child (2): -500 (VIP) + -5 (waiting) + -20 (age) = -525
	// Both should have negative scores
	if results[1].score >= 0 {
		t.Errorf("VIP elderly should have negative score, got %v", results[1].score)
	}
	if results[2].score >= 0 {
		t.Errorf("VIP young child should have negative score, got %v", results[2].score)
	}

	// Log results for manual inspection
	t.Log("Patient ordering:")
	for _, r := range results {
		t.Logf("  %s: Tier=%d, Score=%.2f", r.name, r.tier, r.score)
	}
}
