package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arfis/waiting-room/internal/priority"
	"github.com/arfis/waiting-room/internal/service"
	"github.com/arfis/waiting-room/internal/types"
)

// CreateEntry creates a new queue entry with priority calculation
// approximateDurationSeconds, symbols, appointmentTime, age, manualOverride are used for priority calculation
func (s *WaitingQueue) CreateEntry(ctx context.Context, roomId string, cardData CardData,
	approximateDurationSeconds int64, serviceName string, symbols []string,
	appointmentTime *time.Time, age *int, manualOverride *float64) (*Entry, error) {

	// Extract tenant ID from context (format: "buildingId:sectionId")
	tenantIDHeader := service.GetTenantID(ctx)

	// Parse tenant ID to extract buildingId and sectionID
	buildingID, sectionID, _ := types.ParseTenantID(tenantIDHeader)

	log.Printf("[WaitingQueue] Creating entry for room %s, buildingId: %s, sectionId: %s", roomId, buildingID, sectionID)

	// Load priority configuration
	var priorityConfig *priority.PriorityConfig
	if s.priorityRepo != nil {
		var err error
		priorityConfig, err = s.priorityRepo.GetConfig(ctx, buildingID, sectionID)
		if err != nil {
			log.Printf("Warning: Failed to load priority config, using default: %v", err)
			priorityConfig = priority.GetDefaultConfig()
		}
	} else {
		// If priority repo is nil (e.g., in tests), use default config
		priorityConfig = priority.GetDefaultConfig()
	}

	// Calculate tier and fitness score
	calculator := priority.NewCalculator(priorityConfig)
	now := time.Now()

	calcInput := priority.CalculationInput{
		Symbols:         symbols,
		AppointmentTime: appointmentTime,
		Age:             age,
		ManualOverride:  manualOverride,
		ArrivalTime:     now,
		CurrentTime:     now,
	}

	result := calculator.Calculate(calcInput)

	log.Printf("[WaitingQueue] Calculated priority - Tier: %d, FitnessScore: %.2f", result.Tier, result.FitnessScore)

	// Get current WAITING entries to determine initial position (filtered by tenant and section)
	entries, err := s.repo.GetQueueEntries(ctx, roomId, []string{"WAITING"})
	if err != nil {
		log.Printf("Failed to get queue entries: %v", err)
		// Continue with position 1 if we can't get current entries
	}

	// Calculate next position based only on WAITING entries
	// This is a temporary position - it will be recalculated based on priority
	nextPosition := len(entries) + 1

	// Create new entry with priority metadata
	entry := &Entry{
		WaitingRoomID:              roomId,
		TenantID:                   buildingID,
		SectionID:                  sectionID,
		TicketNumber:               "", // Will be set by repository
		QRToken:                    "", // Will be set by repository
		Status:                     "WAITING",
		Position:                   int64(nextPosition),
		CardData:                   cardData,
		ApproximateDurationSeconds: approximateDurationSeconds,
		ServiceName:                serviceName,
		Symbols:                    symbols,
		AppointmentTime:            appointmentTime,
		Age:                        age,
		ManualOverride:             manualOverride,
		FitnessScore:               result.FitnessScore,
		Tier:                       result.Tier,
	}

	// Save to repository
	if err := s.repo.CreateEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create queue entry: %w", err)
	}

	// Recalculate positions based on priority (tier, fitness score, arrival time)
	if err := s.repo.RecalculatePositions(ctx, roomId); err != nil {
		log.Printf("Warning: Failed to recalculate positions after creating entry: %v", err)
	}

	log.Printf("Created queue entry %s with ticket %s for room %s (tier: %d, fitness: %.2f)",
		entry.ID, entry.TicketNumber, roomId, entry.Tier, entry.FitnessScore)
	return entry, nil
}
