package priority

import (
	"time"
)

// Calculator calculates tier and fitness score for queue entries
type Calculator struct {
	config *PriorityConfig
}

// NewCalculator creates a new priority calculator
func NewCalculator(config *PriorityConfig) *Calculator {
	return &Calculator{
		config: config,
	}
}

// CalculationInput contains all the data needed for priority calculation
type CalculationInput struct {
	Symbols         []string
	AppointmentTime *time.Time
	Age             *int
	ManualOverride  *float64
	ArrivalTime     time.Time
	CurrentTime     time.Time
}

// CalculationResult contains the calculated tier and fitness score
type CalculationResult struct {
	Tier         int
	FitnessScore float64
}

// Calculate determines the tier and fitness score for an entry
func (c *Calculator) Calculate(input CalculationInput) CalculationResult {
	// 1. Calculate tier
	tier := c.calculateTier(input.Symbols)

	// 2. Calculate fitness score
	fitnessScore := c.calculateFitnessScore(input)

	return CalculationResult{
		Tier:         tier,
		FitnessScore: fitnessScore,
	}
}

// calculateTier determines which tier an entry belongs to based on symbols
func (c *Calculator) calculateTier(symbols []string) int {
	symbolSet := make(map[string]bool)
	for _, symbol := range symbols {
		symbolSet[symbol] = true
	}

	// Check each tier in order
	for _, tier := range c.config.PriorityModel.Tiers {
		if c.matchesTierCondition(symbolSet, tier.Condition) {
			return tier.ID
		}
	}

	// If no tier matches, return the highest tier number (lowest priority)
	if len(c.config.PriorityModel.Tiers) > 0 {
		return c.config.PriorityModel.Tiers[len(c.config.PriorityModel.Tiers)-1].ID
	}
	return 0
}

// matchesTierCondition checks if the symbols match a tier's condition
func (c *Calculator) matchesTierCondition(symbolSet map[string]bool, condition Condition) bool {
	// Check SymbolsAnyOf - must have at least one of these symbols
	if len(condition.SymbolsAnyOf) > 0 {
		hasAny := false
		for _, symbol := range condition.SymbolsAnyOf {
			if symbolSet[symbol] {
				hasAny = true
				break
			}
		}
		if !hasAny {
			return false
		}
	}

	// Check SymbolsNotAnyOf - must NOT have any of these symbols
	if len(condition.SymbolsNotAnyOf) > 0 {
		for _, symbol := range condition.SymbolsNotAnyOf {
			if symbolSet[symbol] {
				return false
			}
		}
	}

	return true
}

// calculateFitnessScore computes the fitness score based on all factors
func (c *Calculator) calculateFitnessScore(input CalculationInput) float64 {
	score := 0.0
	contrib := c.config.PriorityModel.Fitness.Contributions

	// 1. Symbol weights
	for _, symbol := range input.Symbols {
		if weight, ok := contrib.SymbolWeights.Values[symbol]; ok {
			score += weight
		}
	}

	// 2. Waiting time (now - arrivalTime)
	waitingMinutes := input.CurrentTime.Sub(input.ArrivalTime).Minutes()
	score += waitingMinutes * contrib.WaitingTime.WeightPerMinute

	// 3. Appointment deviation (if appointment time is set)
	if input.AppointmentTime != nil {
		deviationMinutes := input.CurrentTime.Sub(*input.AppointmentTime).Minutes()
		if deviationMinutes < 0 {
			// Early (before appointment time) - penalty
			score += (-deviationMinutes) * contrib.AppointmentDeviation.EarlyPenaltyPerMinute
		} else {
			// Late (after appointment time) - bonus (negative score)
			score += deviationMinutes * contrib.AppointmentDeviation.LateBonusPerMinute
		}
	}

	// 4. Age-based prioritization
	if input.Age != nil {
		age := *input.Age
		if age < 6 {
			// Children under 6: younger is higher priority
			yearsYounger := 6 - age
			score += float64(yearsYounger) * contrib.Age.Under6PerYearYounger
		} else if age >= contrib.Age.AgeThresholdSenior {
			// Seniors: older is higher priority
			yearsOlder := age - contrib.Age.AgeThresholdSenior
			score += float64(yearsOlder) * contrib.Age.Over65PerYearOlder
		}
	}

	// 5. Manual override
	if contrib.ManualOverride.Enabled && input.ManualOverride != nil {
		score += (*input.ManualOverride) * contrib.ManualOverride.Weight
	}

	return score
}
