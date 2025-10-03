package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arfis/waiting-room/internal/config"
	"github.com/arfis/waiting-room/internal/data/dto"
	"github.com/arfis/waiting-room/internal/data/dto/queueentrystatus"
	"github.com/arfis/waiting-room/internal/repository"
	"github.com/arfis/waiting-room/internal/service/servicepoint"
	"github.com/arfis/waiting-room/internal/types"
)

// Use types from the types package
type Entry = types.Entry
type CardData = types.CardData

type Service struct {
	repo            repository.QueueRepository
	config          *config.Config
	servicePointSvc *servicepoint.Service
}

func NewService(repo repository.QueueRepository, cfg *config.Config, servicePointSvc *servicepoint.Service) *Service {
	return &Service{
		repo:            repo,
		config:          cfg,
		servicePointSvc: servicePointSvc,
	}
}

// CreateEntry creates a new queue entry
func (s *Service) CreateEntry(roomId string, cardData CardData) (*Entry, error) {
	ctx := context.Background()

	// Get current WAITING entries to determine position
	entries, err := s.repo.GetQueueEntries(ctx, roomId)
	if err != nil {
		log.Printf("Failed to get queue entries: %v", err)
		// Continue with position 1 if we can't get current entries
	}

	// Calculate next position based only on WAITING entries
	nextPosition := 1
	waitingCount := 0
	for _, entry := range entries {
		if entry.Status == "WAITING" {
			waitingCount++
		}
	}
	nextPosition = waitingCount + 1

	// Get available service point for this room (with active manager)
	servicePoint, err := s.servicePointSvc.GetAvailableServicePoint(ctx, roomId)
	if err != nil {
		log.Printf("Warning: Failed to get available service point for room %s: %v", roomId, err)
		// Fallback to default service point
		servicePoint = s.config.GetDefaultServicePoint(roomId)
	}

	// Create new entry
	entry := &Entry{
		WaitingRoomID: roomId,
		TicketNumber:  "", // Will be set by repository
		QRToken:       "", // Will be set by repository
		Status:        "WAITING",
		Position:      nextPosition,
		ServicePoint:  servicePoint,
		CardData:      cardData,
	}

	// Save to repository
	if err := s.repo.CreateEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create queue entry: %w", err)
	}

	// Recalculate positions to ensure consistency
	if err := s.repo.RecalculatePositions(ctx, roomId); err != nil {
		log.Printf("Warning: Failed to recalculate positions after creating entry: %v", err)
	}

	log.Printf("Created queue entry %s with ticket %s for room %s", entry.ID, entry.TicketNumber, roomId)
	return entry, nil
}

// GetServicePoints returns the configured service points for a room
func (s *Service) GetServicePoints(ctx context.Context, roomId string) ([]dto.ServicePoint, error) {
	servicePointConfigs := s.config.GetServicePointsForRoom(roomId)

	// Convert config to DTO
	var servicePoints []dto.ServicePoint
	for _, config := range servicePointConfigs {
		servicePoint := dto.ServicePoint{
			Id:   config.ID,
			Name: config.Name,
		}
		if config.Description != "" {
			servicePoint.Description = &config.Description
		}
		servicePoints = append(servicePoints, servicePoint)
	}

	log.Printf("Retrieved %d service points for room %s", len(servicePoints), roomId)
	return servicePoints, nil
}

// GetQueueEntries retrieves all queue entries for a room
func (s *Service) GetQueueEntries(roomId string) ([]*Entry, error) {
	ctx := context.Background()
	entries, err := s.repo.GetQueueEntries(ctx, roomId)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue entries: %w", err)
	}
	return entries, nil
}

// GetEntryByQRToken retrieves a queue entry by QR token
func (s *Service) GetEntryByQRToken(qrToken string) (*Entry, error) {
	ctx := context.Background()
	entry, err := s.repo.GetEntryByQRToken(ctx, qrToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry by QR token: %w", err)
	}
	return entry, nil
}

// CallNext calls the next person in the queue
func (s *Service) CallNext(roomId string) (*Entry, error) {
	ctx := context.Background()
	log.Printf("CallNext: Starting for room %s", roomId)

	// First, complete any currently served person
	currentEntry, err := s.repo.GetCurrentServedEntry(ctx, roomId)
	if err != nil {
		log.Printf("CallNext: Failed to get current served entry: %v", err)
		return nil, fmt.Errorf("failed to get current served entry: %w", err)
	}

	if currentEntry != nil {
		log.Printf("CallNext: Found current entry %s, completing it", currentEntry.ID)
		// Complete the current person
		if err := s.repo.UpdateEntryStatus(ctx, currentEntry.ID, "COMPLETED"); err != nil {
			log.Printf("CallNext: Failed to complete current entry: %v", err)
			return nil, fmt.Errorf("failed to complete current entry: %w", err)
		}
		log.Printf("Completed current entry %s", currentEntry.ID)
	} else {
		log.Printf("CallNext: No current entry found")
	}

	// Get the next waiting person
	log.Printf("CallNext: Getting next waiting entry")
	nextEntry, err := s.repo.GetNextWaitingEntry(ctx, roomId)
	if err != nil {
		log.Printf("CallNext: Failed to get next waiting entry: %v", err)
		return nil, fmt.Errorf("failed to get next waiting entry: %w", err)
	}

	if nextEntry == nil {
		log.Printf("CallNext: No waiting entries found")
		return nil, fmt.Errorf("no waiting entries found")
	}

	log.Printf("CallNext: Found next entry %s, calling them", nextEntry.ID)

	// Call the next person
	if err := s.repo.UpdateEntryStatus(ctx, nextEntry.ID, "CALLED"); err != nil {
		log.Printf("CallNext: Failed to update entry status: %v", err)
		return nil, fmt.Errorf("failed to call next entry: %w", err)
	}

	log.Printf("CallNext: Successfully called entry %s", nextEntry.ID)

	// Recalculate positions for remaining waiting entries
	if err := s.repo.RecalculatePositions(ctx, roomId); err != nil {
		log.Printf("Warning: Failed to recalculate positions: %v", err)
	}

	log.Printf("Called next entry %s with ticket %s", nextEntry.ID, nextEntry.TicketNumber)
	return nextEntry, nil
}

// FinishCurrent finishes the current person without calling the next
func (s *Service) FinishCurrent(roomId string) (*Entry, error) {
	ctx := context.Background()

	// Get the currently served person
	currentEntry, err := s.repo.GetCurrentServedEntry(ctx, roomId)
	if err != nil {
		return nil, fmt.Errorf("failed to get current served entry: %w", err)
	}

	if currentEntry == nil {
		return nil, fmt.Errorf("no one is currently being served")
	}

	// Complete the current person
	if err := s.repo.UpdateEntryStatus(ctx, currentEntry.ID, "COMPLETED"); err != nil {
		return nil, fmt.Errorf("failed to complete current entry: %w", err)
	}

	// Recalculate positions for remaining waiting entries
	if err := s.repo.RecalculatePositions(ctx, roomId); err != nil {
		log.Printf("Warning: Failed to recalculate positions: %v", err)
	}

	log.Printf("Finished current entry %s with ticket %s", currentEntry.ID, currentEntry.TicketNumber)
	return currentEntry, nil
}

// UpdateEntryStatus updates the status of a queue entry
func (s *Service) UpdateEntryStatus(id string, status string) error {
	ctx := context.Background()
	return s.repo.UpdateEntryStatus(ctx, id, status)
}

// DeleteEntry deletes a queue entry
func (s *Service) DeleteEntry(id string) error {
	ctx := context.Background()
	return s.repo.DeleteEntry(ctx, id)
}

// CallNextForServicePoint calls the next person for a specific service point
func (s *Service) CallNextForServicePoint(ctx context.Context, roomId, servicePointId string) (*dto.QueueEntry, error) {
	// Get the next waiting entry for this specific service point
	entry, err := s.repo.GetNextWaitingEntryForServicePoint(ctx, roomId, servicePointId)
	if err != nil {
		return nil, fmt.Errorf("failed to get next waiting entry for service point %s: %w", servicePointId, err)
	}

	if entry == nil {
		return nil, fmt.Errorf("no waiting entries found for service point %s", servicePointId)
	}

	// Update status to CALLED
	entry.Status = "CALLED"
	entry.UpdatedAt = time.Now()

	if err := s.repo.UpdateEntryStatus(ctx, entry.ID, "CALLED"); err != nil {
		return nil, fmt.Errorf("failed to update entry status: %w", err)
	}

	// Recalculate positions
	if err := s.repo.RecalculatePositions(ctx, roomId); err != nil {
		log.Printf("Warning: Failed to recalculate positions after calling next: %v", err)
	}

	// Convert to DTO
	queueEntry := &dto.QueueEntry{
		Id:            entry.ID,
		WaitingRoomId: entry.WaitingRoomID,
		TicketNumber:  entry.TicketNumber,
		Status:        queueentrystatus.QueueEntryStatus(entry.Status),
		Position:      entry.Position,
	}
	if entry.ServicePoint != "" {
		queueEntry.ServicePoint = &entry.ServicePoint
	}

	log.Printf("Called next person %s (ticket %s) for service point %s in room %s",
		entry.ID, entry.TicketNumber, servicePointId, roomId)

	return queueEntry, nil
}

// MarkInRoomForServicePoint marks a person as in room for a specific service point
func (s *Service) MarkInRoomForServicePoint(ctx context.Context, roomId, servicePointId, entryId string) (*dto.QueueEntry, error) {
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
		Id:            entry.ID,
		WaitingRoomId: entry.WaitingRoomID,
		TicketNumber:  entry.TicketNumber,
		Status:        queueentrystatus.QueueEntryStatus(entry.Status),
		Position:      entry.Position,
	}
	if entry.ServicePoint != "" {
		queueEntry.ServicePoint = &entry.ServicePoint
	}

	log.Printf("Marked person %s (ticket %s) as in room for service point %s",
		entry.ID, entry.TicketNumber, servicePointId)

	return queueEntry, nil
}

// FinishCurrentForServicePoint finishes the current person for a specific service point
func (s *Service) FinishCurrentForServicePoint(ctx context.Context, roomId, servicePointId string) (*dto.QueueEntry, error) {
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
		Id:            entry.ID,
		WaitingRoomId: entry.WaitingRoomID,
		TicketNumber:  entry.TicketNumber,
		Status:        queueentrystatus.QueueEntryStatus(entry.Status),
		Position:      entry.Position,
	}
	if entry.ServicePoint != "" {
		queueEntry.ServicePoint = &entry.ServicePoint
	}

	log.Printf("Finished current person %s (ticket %s) for service point %s in room %s",
		entry.ID, entry.TicketNumber, servicePointId, roomId)

	return queueEntry, nil
}

// Close closes the service and its repository
func (s *Service) Close() error {
	return s.repo.Close()
}
