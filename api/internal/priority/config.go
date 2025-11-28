package priority

// PriorityConfig represents the configuration for the priority calculation system
type PriorityConfig struct {
	Version       string        `json:"version" bson:"version"`
	Description   string        `json:"description" bson:"description"`
	PriorityModel PriorityModel `json:"priorityModel" bson:"priorityModel"`
}

// PriorityModel defines the algorithm and rules for priority calculation
type PriorityModel struct {
	Algorithm Algorithm        `json:"algorithm" bson:"algorithm"`
	Tiers     []Tier           `json:"tiers" bson:"tiers"`
	Fitness   FitnessConfig    `json:"fitness" bson:"fitness"`
}

// Algorithm describes the ordering logic
type Algorithm struct {
	Explanation    []string `json:"explanation" bson:"explanation"`
	OrderingFields []string `json:"orderingFields" bson:"orderingFields"`
}

// Tier represents a priority tier with conditions
type Tier struct {
	ID          int       `json:"id" bson:"id"`
	Name        string    `json:"name" bson:"name"`
	Condition   Condition `json:"condition" bson:"condition"`
	Description string    `json:"description" bson:"description"`
}

// Condition defines the matching rules for a tier
type Condition struct {
	SymbolsAnyOf    []string `json:"symbolsAnyOf,omitempty" bson:"symbolsAnyOf,omitempty"`
	SymbolsNotAnyOf []string `json:"symbolsNotAnyOf,omitempty" bson:"symbolsNotAnyOf,omitempty"`
}

// FitnessConfig defines the fitness score calculation rules
type FitnessConfig struct {
	Explanation   string        `json:"explanation" bson:"explanation"`
	Contributions Contributions `json:"contributions" bson:"contributions"`
}

// Contributions defines all the factors that contribute to the fitness score
type Contributions struct {
	SymbolWeights        SymbolWeights        `json:"symbolWeights" bson:"symbolWeights"`
	WaitingTime          WaitingTime          `json:"waitingTime" bson:"waitingTime"`
	AppointmentDeviation AppointmentDeviation `json:"appointmentDeviation" bson:"appointmentDeviation"`
	Age                  AgeConfig            `json:"age" bson:"age"`
	ManualOverride       ManualOverride       `json:"manualOverride" bson:"manualOverride"`
}

// SymbolWeights defines the weight for each symbol
type SymbolWeights struct {
	Description string             `json:"description" bson:"description"`
	Values      map[string]float64 `json:"values" bson:"values"`
}

// WaitingTime defines how waiting time affects fitness
type WaitingTime struct {
	Description      string  `json:"description" bson:"description"`
	WeightPerMinute  float64 `json:"weightPerMinute" bson:"weightPerMinute"`
}

// AppointmentDeviation defines how appointment timing affects fitness
type AppointmentDeviation struct {
	Description           string  `json:"description" bson:"description"`
	EarlyPenaltyPerMinute float64 `json:"earlyPenaltyPerMinute" bson:"earlyPenaltyPerMinute"`
	LateBonusPerMinute    float64 `json:"lateBonusPerMinute" bson:"lateBonusPerMinute"`
}

// AgeConfig defines how age affects fitness
type AgeConfig struct {
	Description            string `json:"description" bson:"description"`
	Under6PerYearYounger   float64 `json:"under6PerYearYounger" bson:"under6PerYearYounger"`
	Over65PerYearOlder     float64 `json:"over65PerYearOlder" bson:"over65PerYearOlder"`
	AgeThresholdSenior     int     `json:"ageThresholdSenior" bson:"ageThresholdSenior"`
}

// ManualOverride defines how manual overrides work
type ManualOverride struct {
	Description string  `json:"description" bson:"description"`
	Enabled     bool    `json:"enabled" bson:"enabled"`
	Weight      float64 `json:"weight" bson:"weight"`
}
