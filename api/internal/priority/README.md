# Priority Calculation Module

This module implements a sophisticated tier-based priority calculation system for the waiting room queue.

## Overview

The priority system uses a two-level approach:
1. **Tier-based Classification** - Assigns patients to priority tiers (0 = highest)
2. **Fitness Score** - Within each tier, calculates a numeric score (lower = higher priority)

## Components

### Calculator (`calculator.go`)
Core calculation engine that determines tier and fitness score based on patient metadata.

### Configuration (`config.go`)
Defines the structure for priority rules, including:
- Tier definitions with conditions
- Symbol weights (STATIM, VIP, IMMOBILE, etc.)
- Time-based factors (waiting time, appointment deviation)
- Age-based prioritization
- Manual override support

### Repository (`repository.go`)
MongoDB-backed storage for priority configurations with:
- Tenant-specific configurations
- Section-specific configurations
- Fallback to default configuration

## Default Priority Configuration

### Tiers
- **Tier 0 (STATIM)**: Emergency/urgent cases - highest priority
- **Tier 1 (VIP)**: VIP patients (excluding STATIM)
- **Tier 2 (NORMAL)**: Regular patients

### Fitness Score Factors

| Factor | Weight | Description |
|--------|--------|-------------|
| STATIM symbol | -1000 | Emergency priority |
| VIP symbol | -500 | VIP priority |
| IMMOBILE symbol | -50 | Mobility issues |
| Waiting time | -1 per minute | Longer wait = more priority |
| Early arrival | +2 per minute | Penalty for arriving too early |
| Late arrival | -3 per minute | Bonus for being late |
| Young children (<6) | -5 per year younger | Younger = more priority |
| Seniors (>65) | -1 per year older | Older = more priority |
| Manual override | ±1 * value | Staff adjustments |

**Note**: Negative scores = higher priority

## Usage

### Basic Calculation

```go
// Load configuration
priorityRepo := priority.NewRepository(database)
config, _ := priorityRepo.GetConfig(ctx, tenantID, sectionID)

// Create calculator
calculator := priority.NewCalculator(config)

// Calculate priority
input := priority.CalculationInput{
    Symbols:         []string{"VIP", "IMMOBILE"},
    AppointmentTime: &appointmentTime,
    Age:             &age,
    ManualOverride:  &manualOverride,
    ArrivalTime:     arrivalTime,
    CurrentTime:     time.Now(),
}

result := calculator.Calculate(input)
// result.Tier = 1 (VIP tier)
// result.FitnessScore = -565.0 (VIP: -500, IMMOBILE: -50, waiting: -15)
```

### Queue Ordering

Entries are sorted by:
1. **Tier** (ascending) - Lower tier numbers first
2. **Fitness Score** (ascending) - Lower scores first
3. **Arrival Time** (ascending) - Earlier arrivals first
4. **Ticket Number** (ascending) - Tiebreaker

## Examples

### Example 1: STATIM Elderly Patient (Late for Appointment)
```go
Symbols: ["STATIM"]
Age: 75
Waiting: 15 minutes
Appointment: 15 minutes ago (late)

Result:
  Tier: 0
  Score: -1000 (STATIM) + -15 (waiting) + -45 (late) + -10 (age) = -1070
```

### Example 2: VIP Young Child
```go
Symbols: ["VIP"]
Age: 3
Waiting: 5 minutes

Result:
  Tier: 1
  Score: -500 (VIP) + -5 (waiting) + -15 (age: 6-3=3 * -5) = -520
```

### Example 3: Regular Patient with Manual Boost
```go
Symbols: []
Waiting: 10 minutes
ManualOverride: -200

Result:
  Tier: 2
  Score: -10 (waiting) + -200 (manual) = -210
```

## Test Coverage

### Unit Tests
The module includes comprehensive unit tests covering:
- ✅ Tier calculation with various symbol combinations
- ✅ Fitness score factors (symbols, waiting time, appointments, age, manual override)
- ✅ Complex multi-factor scenarios
- ✅ Negative and decimal scores
- ✅ Queue ordering verification
- ✅ Default configuration validation

**Coverage**: 100% for calculator functions

### Running Tests

```bash
# Run all unit tests
go test -short -v ./internal/priority/...

# Run with coverage
go test -short -coverprofile=coverage.out ./internal/priority/...
go tool cover -html=coverage.out

# Run integration tests (requires MongoDB)
go test -v ./internal/priority/...

# Run benchmarks
go test -bench=. -benchmem ./internal/priority/...
```

## Performance

Benchmark results on Apple M2 Pro:

| Operation | Time/op | Memory | Allocations |
|-----------|---------|--------|-------------|
| Full Calculate | 89.89 ns | 0 B | 0 |
| Tier Calculation | 52.17 ns | 0 B | 0 |
| Fitness Score | 24.36 ns | 0 B | 0 |

**Result**: The calculator is extremely fast with zero memory allocations, suitable for high-throughput scenarios.

## Configuration Management

### Loading Configuration

Configurations are loaded in this order:
1. Section-specific config (tenant + section)
2. Tenant-level config (tenant only)
3. Default hardcoded config

```go
// Get config for specific section
config, err := repo.GetConfig(ctx, "hospital-1", "cardiology")

// If not found, falls back to tenant-level
config, err := repo.GetConfig(ctx, "hospital-1", "")

// If still not found, returns default config
```

### Saving Custom Configuration

```go
customConfig := &priority.PriorityConfig{
    Version: "2.0",
    PriorityModel: priority.PriorityModel{
        Tiers: []priority.Tier{
            {
                ID: 0,
                Name: "URGENT",
                Condition: priority.Condition{
                    SymbolsAnyOf: []string{"URGENT"},
                },
            },
            // ... more tiers
        },
        Fitness: priority.FitnessConfig{
            // ... custom weights
        },
    },
}

err := repo.SaveConfig(ctx, customConfig, "tenant-id", "section-id")
```

## Future Enhancements

Potential improvements:
- [ ] Time-of-day based adjustments
- [ ] Dynamic weight adjustment based on queue length
- [ ] Predicted service time integration
- [ ] Historical wait time analysis
- [ ] Multi-service coordination
- [ ] Real-time recalculation triggers

## API Integration

The priority system is automatically used when creating queue entries:

```http
POST /api/waiting-rooms/{roomId}/swipe
Content-Type: application/json

{
  "idCardRaw": "123456789",
  "serviceId": "cardiology-consultation",
  "symbols": ["VIP", "IMMOBILE"],
  "appointmentTime": "2025-11-27T14:00:00Z",
  "age": 75,
  "manualOverride": -50
}
```

The system will automatically:
1. Calculate tier and fitness score
2. Insert entry at correct position in queue
3. Recalculate all positions if needed
4. Broadcast updates via WebSocket

## Notes

- Fitness scores can be negative, positive, or decimal (0.0001)
- All priority fields are optional (backward compatible)
- Without priority data, entries default to Tier 2, Score 0
- Recalculation happens automatically on every entry creation
- Configuration changes don't automatically recalculate existing entries
