package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arfis/waiting-room/internal/data/dto"
	"github.com/arfis/waiting-room/internal/data/dto/queueentrystatus"
)

// CallNextForServicePoint calls the next person for a specific service point
func (s *WaitingQueue) CallNextForServicePoint(ctx context.Context, roomId, servicePointId string) (*Entry, error) {
	log.Printf("CallNextForServicePoint: Starting for room %s, service point %s", roomId, servicePointId)

	// First, complete any currently served person for this service point
	currentEntry, err := s.repo.GetCurrentServedEntryForServicePoint(ctx, roomId, servicePointId)
	if err != nil {
		log.Printf("CallNextForServicePoint: Failed to get current served entry for service point: %v", err)
		// Continue anyway, as there might not be a current entry
	}

	if currentEntry != nil {
		log.Printf("CallNextForServicePoint: Found current entry %s for service point %s, completing it", currentEntry.ID, servicePointId)
		// Complete the current person
		if err := s.repo.UpdateEntryStatus(ctx, currentEntry.ID, "COMPLETED"); err != nil {
			log.Printf("CallNextForServicePoint: Failed to complete current entry: %v", err)
			return nil, fmt.Errorf("failed to complete current entry: %w", err)
		}
		log.Printf("CallNextForServicePoint: Completed current entry %s", currentEntry.ID)
	} else {
		log.Printf("CallNextForServicePoint: No current entry found for service point %s", servicePointId)
	}

	// Get the next waiting entry for this specific service point
	entry, err := s.repo.GetNextWaitingEntry(ctx, roomId)
	if err != nil {
		return nil, fmt.Errorf("failed to get next waiting entry for service point %s: %w", servicePointId, err)
	}

	if entry == nil {
		return nil, fmt.Errorf("no waiting entries found for service point %s", servicePointId)
	}

	log.Printf("CallNextForServicePoint: Found next entry %s, calling them for service point %s", entry.ID, servicePointId)

	// Update status to CALLED and set service point
	entry.Status = "CALLED"
	entry.UpdatedAt = time.Now()
	entry.ServicePoint = servicePointId

	if err := s.repo.UpdateEntryStatus(ctx, entry.ID, "CALLED"); err != nil {
		return nil, fmt.Errorf("failed to update entry status: %w", err)
	}

	// Also update the service point in the database
	if err := s.repo.UpdateEntryServicePoint(ctx, entry.ID, servicePointId); err != nil {
		log.Printf("Warning: Failed to update service point: %v", err)
	}

	// Recalculate positions
	if err := s.repo.RecalculatePositions(ctx, roomId); err != nil {
		log.Printf("Warning: Failed to recalculate positions after calling next: %v", err)
	}

	log.Printf("CallNextForServicePoint: Successfully called entry %s (ticket %s) for service point %s in room %s",
		entry.ID, entry.TicketNumber, servicePointId, roomId)

	return entry, nil
}

// CallSpecificEntryForServicePoint calls a specific entry by ID for a service point
func (s *WaitingQueue) CallSpecificEntryForServicePoint(ctx context.Context, roomId, servicePointId, entryId string) (*Entry, error) {
	log.Printf("CallSpecificEntryForServicePoint: Starting for room %s, service point %s, entry %s", roomId, servicePointId, entryId)

	// Get the entry
	entry, err := s.repo.GetEntryByID(ctx, entryId)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	if entry == nil {
		return nil, fmt.Errorf("entry not found")
	}

	// Verify the entry is for the correct room
	if entry.WaitingRoomID != roomId {
		return nil, fmt.Errorf("entry is not for room %s", roomId)
	}

	// Verify the entry is WAITING
	if entry.Status != "WAITING" {
		return nil, fmt.Errorf("entry is not in WAITING status (current status: %s)", entry.Status)
	}

	// First, complete any currently served person for this service point
	currentEntry, err := s.repo.GetCurrentServedEntryForServicePoint(ctx, roomId, servicePointId)
	if err != nil {
		log.Printf("CallSpecificEntryForServicePoint: Failed to get current served entry for service point: %v", err)
		// Continue anyway, as there might not be a current entry
	}

	if currentEntry != nil {
		log.Printf("CallSpecificEntryForServicePoint: Found current entry %s for service point %s, completing it", currentEntry.ID, servicePointId)
		// Complete the current person
		if err := s.repo.UpdateEntryStatus(ctx, currentEntry.ID, "COMPLETED"); err != nil {
			log.Printf("CallSpecificEntryForServicePoint: Failed to complete current entry: %v", err)
			return nil, fmt.Errorf("failed to complete current entry: %w", err)
		}
		log.Printf("CallSpecificEntryForServicePoint: Completed current entry %s", currentEntry.ID)
	}

	log.Printf("CallSpecificEntryForServicePoint: Calling specific entry %s for service point %s", entry.ID, servicePointId)

	// Update status to CALLED and set service point
	entry.Status = "CALLED"
	entry.UpdatedAt = time.Now()
	entry.ServicePoint = servicePointId

	if err := s.repo.UpdateEntryStatus(ctx, entry.ID, "CALLED"); err != nil {
		return nil, fmt.Errorf("failed to update entry status: %w", err)
	}

	// Also update the service point in the database
	if err := s.repo.UpdateEntryServicePoint(ctx, entry.ID, servicePointId); err != nil {
		log.Printf("Warning: Failed to update service point: %v", err)
	}

	// Recalculate positions
	if err := s.repo.RecalculatePositions(ctx, roomId); err != nil {
		log.Printf("Warning: Failed to recalculate positions after calling specific entry: %v", err)
	}

	log.Printf("CallSpecificEntryForServicePoint: Successfully called entry %s (ticket %s) for service point %s in room %s",
		entry.ID, entry.TicketNumber, servicePointId, roomId)

	return entry, nil
}

// MarkInRoomForServicePoint marks a person as in room for a specific service point
func (s *WaitingQueue) MarkInRoomForServicePoint(ctx context.Context, roomId, servicePointId, entryId string) (*dto.QueueEntry, error) {
	// Get the entry
	entry, err := s.repo.GetEntryByID(ctx, entryId)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	if entry == nil {
		return nil, fmt.Errorf("entry not found")
	}

	// Verify the entry is for the correct service point
	if entry.ServicePoint != servicePointId {
		return nil, fmt.Errorf("entry is not assigned to service point %s", servicePointId)
	}

	// Update status to IN_ROOM
	entry.Status = "IN_ROOM"
	entry.UpdatedAt = time.Now()

	if err := s.repo.UpdateEntryStatus(ctx, entry.ID, "IN_ROOM"); err != nil {
		return nil, fmt.Errorf("failed to update entry status: %w", err)
	}

	// Convert to DTO
	queueEntry := &dto.QueueEntry{
		ID:            entry.ID,
		WaitingRoomID: entry.WaitingRoomID,
		TicketNumber:  entry.TicketNumber,
		Status:        queueentrystatus.QueueEntryStatus(entry.Status),
		Position:      entry.Position,
	}
	if entry.ServicePoint != "" {
		queueEntry.ServicePoint = &entry.ServicePoint
	}
	if entry.ServiceName != "" {
		queueEntry.ServiceName = &entry.ServiceName
	}
	if entry.ApproximateDurationSeconds > 0 {
		durationMinutes := entry.ApproximateDurationSeconds / 60 // Convert seconds to minutes for API
		queueEntry.ServiceDuration = &durationMinutes
	}

	log.Printf("Marked person %s (ticket %s) as in room for service point %s",
		entry.ID, entry.TicketNumber, servicePointId)

	return queueEntry, nil
}

// FinishCurrentForServicePoint finishes the current person for a specific service point
func (s *WaitingQueue) FinishCurrentForServicePoint(ctx context.Context, roomId, servicePointId string) (*dto.QueueEntry, error) {
	// Get the current served entry for this service point
	entry, err := s.repo.GetCurrentServedEntryForServicePoint(ctx, roomId, servicePointId)
	if err != nil {
		return nil, fmt.Errorf("failed to get current served entry for service point %s: %w", servicePointId, err)
	}

	if entry == nil {
		return nil, fmt.Errorf("no current served entry found for service point %s", servicePointId)
	}

	// Update status to COMPLETED
	entry.Status = "COMPLETED"
	entry.UpdatedAt = time.Now()

	if err := s.repo.UpdateEntryStatus(ctx, entry.ID, "COMPLETED"); err != nil {
		return nil, fmt.Errorf("failed to update entry status: %w", err)
	}

	// Recalculate positions
	if err := s.repo.RecalculatePositions(ctx, roomId); err != nil {
		log.Printf("Warning: Failed to recalculate positions after finishing current: %v", err)
	}

	// Convert to DTO
	queueEntry := &dto.QueueEntry{
		ID:            entry.ID,
		WaitingRoomID: entry.WaitingRoomID,
		TicketNumber:  entry.TicketNumber,
		Status:        queueentrystatus.QueueEntryStatus(entry.Status),
		Position:      entry.Position,
	}
	if entry.ServicePoint != "" {
		queueEntry.ServicePoint = &entry.ServicePoint
	}
	if entry.ServiceName != "" {
		queueEntry.ServiceName = &entry.ServiceName
	}
	if entry.ApproximateDurationSeconds > 0 {
		durationMinutes := entry.ApproximateDurationSeconds / 60 // Convert seconds to minutes for API
		queueEntry.ServiceDuration = &durationMinutes
	}

	log.Printf("Finished current person %s (ticket %s) for service point %s in room %s",
		entry.ID, entry.TicketNumber, servicePointId, roomId)

	return queueEntry, nil
}
