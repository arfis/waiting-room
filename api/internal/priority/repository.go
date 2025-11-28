package priority

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository handles loading and storing priority configurations
type Repository struct {
	collection *mongo.Collection
}

// NewRepository creates a new priority configuration repository
func NewRepository(database *mongo.Database) *Repository {
	return &Repository{
		collection: database.Collection("priority_configs"),
	}
}

// GetConfig retrieves the active priority configuration
// If no config exists, it returns the default configuration
func (r *Repository) GetConfig(ctx context.Context, tenantID, sectionID string) (*PriorityConfig, error) {
	// Try to find a config for the specific tenant+section first
	filter := bson.M{
		"tenantId":  tenantID,
		"sectionId": sectionID,
	}

	var config PriorityConfig
	err := r.collection.FindOne(ctx, filter).Decode(&config)
	if err == nil {
		log.Printf("[PriorityRepository] Found config for tenant %s, section %s", tenantID, sectionID)
		return &config, nil
	}

	// If not found, try tenant-level config (no section)
	if err == mongo.ErrNoDocuments && sectionID != "" {
		filter = bson.M{
			"tenantId":  tenantID,
			"sectionId": bson.M{"$exists": false},
		}
		err = r.collection.FindOne(ctx, filter).Decode(&config)
		if err == nil {
			log.Printf("[PriorityRepository] Found tenant-level config for tenant %s", tenantID)
			return &config, nil
		}
	}

	// If still not found, return default config
	if err == mongo.ErrNoDocuments {
		log.Printf("[PriorityRepository] No config found for tenant %s, section %s. Using default config", tenantID, sectionID)
		return GetDefaultConfig(), nil
	}

	return nil, fmt.Errorf("failed to get priority config: %w", err)
}

// SaveConfig saves a priority configuration
func (r *Repository) SaveConfig(ctx context.Context, config *PriorityConfig, tenantID, sectionID string) error {
	// Add metadata
	doc := bson.M{
		"tenantId":  tenantID,
		"sectionId": sectionID,
		"config":    config,
		"updatedAt": time.Now(),
	}

	// Upsert: update if exists, insert if not
	filter := bson.M{
		"tenantId":  tenantID,
		"sectionId": sectionID,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := r.collection.ReplaceOne(ctx, filter, doc, opts)
	if err != nil {
		return fmt.Errorf("failed to save priority config: %w", err)
	}

	log.Printf("[PriorityRepository] Saved config for tenant %s, section %s", tenantID, sectionID)
	return nil
}

// GetDefaultConfig returns the default priority configuration
func GetDefaultConfig() *PriorityConfig {
	// Parse the default configuration from JSON
	defaultJSON := `{
  "version": "1.0",
  "description": "Configuration for waiting-list priority calculation using tiers + fitness score.",
  "priorityModel": {
    "algorithm": {
      "explanation": [
        "1) Compute a TIER for each ticket.",
        "2) Within the same tier, compute a numeric SCORE (fitness). Lower is higher priority.",
        "3) Order tickets by: tier ASC, score ASC, arrivalTime ASC, ticketNumber ASC."
      ],
      "orderingFields": [
        "tierAsc",
        "scoreAsc",
        "arrivalTimeAsc",
        "ticketNumberAsc"
      ]
    },
    "tiers": [
      {
        "id": 0,
        "name": "STATIM",
        "condition": {
          "symbolsAnyOf": ["STATIM"]
        },
        "description": "Highest priority. Any ticket with STATIM symbol, regardless of other symbols."
      },
      {
        "id": 1,
        "name": "VIP",
        "condition": {
          "symbolsAnyOf": ["VIP"],
          "symbolsNotAnyOf": ["STATIM"]
        },
        "description": "VIP patients that are not STATIM."
      },
      {
        "id": 2,
        "name": "NORMAL",
        "condition": {
          "symbolsNotAnyOf": ["STATIM", "VIP"]
        },
        "description": "Regular patients."
      }
    ],
    "fitness": {
      "explanation": "Final score = sum of contributions. Lower score = higher priority.",
      "contributions": {
        "symbolWeights": {
          "description": "Extra score for symbols (within same tier). Negative = more priority.",
          "values": {
            "STATIM": -1000,
            "VIP": -500,
            "IMMOBILE": -50
          }
        },
        "waitingTime": {
          "description": "Per-minute effect of waiting time (now - arrivalTime). Negative = longer wait => more priority.",
          "weightPerMinute": -1
        },
        "appointmentDeviation": {
          "description": "Effect of being early/late vs appointmentTime (if present). Minutes = now - appointmentTime.",
          "earlyPenaltyPerMinute": 2,
          "lateBonusPerMinute": -3
        },
        "age": {
          "description": "Age preference. Under 6: younger first. Over 65: older first. Otherwise neutral.",
          "under6PerYearYounger": -5,
          "over65PerYearOlder": -1,
          "ageThresholdSenior": 65
        },
        "manualOverride": {
          "description": "Optional manual override field on ticket. Lower is more priority.",
          "enabled": true,
          "weight": 1
        }
      }
    }
  }
}`

	var config PriorityConfig
	if err := json.Unmarshal([]byte(defaultJSON), &config); err != nil {
		log.Printf("[PriorityRepository] ERROR: Failed to parse default config: %v", err)
		// Return a minimal safe config
		return &PriorityConfig{
			Version:     "1.0",
			Description: "Minimal fallback configuration",
			PriorityModel: PriorityModel{
				Tiers: []Tier{
					{ID: 0, Name: "DEFAULT", Condition: Condition{}, Description: "Default tier"},
				},
				Fitness: FitnessConfig{
					Contributions: Contributions{
						WaitingTime: WaitingTime{WeightPerMinute: -1},
					},
				},
			},
		}
	}

	return &config
}
