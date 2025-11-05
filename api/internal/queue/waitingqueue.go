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
	"github.com/arfis/waiting-room/internal/service"
	"github.com/arfis/waiting-room/internal/service/servicepoint"
	"github.com/arfis/waiting-room/internal/types"
)

// Use types from the types package
type Entry = types.Entry
type CardData = types.CardData

type WaitingQueue struct {
	repo            repository.QueueRepository
	config          *config.Config
	configService   ConfigService
	servicePointSvc *servicepoint.Service
}

// ConfigService interface for getting tenant-aware configuration
type ConfigService interface {
	GetRoomsConfig(ctx context.Context) ([]types.RoomConfig, error)
}

func NewWaitingQueue(repo repository.QueueRepository, cfg *config.Config, servicePointSvc *servicepoint.Service) *WaitingQueue {
	return &WaitingQueue{
		repo:            repo,
		config:          cfg,
		servicePointSvc: servicePointSvc,
	}
}

// SetConfigService sets the tenant-aware config service
func (s *WaitingQueue) SetConfigService(configService ConfigService) {
	s.configService = configService
}

// CreateEntry creates a new queue entry
func (s *WaitingQueue) CreateEntry(ctx context.Context, roomId string, cardData CardData, approximateDurationSeconds int64, serviceName string) (*Entry, error) {
	// Extract tenant ID from context (format: "buildingId:sectionId")
	tenantIDHeader := service.GetTenantID(ctx)
	
	// Parse tenant ID to extract buildingId and sectionId
	buildingID, sectionID, _ := types.ParseTenantID(tenantIDHeader)
	
	log.Printf("[WaitingQueue] Creating entry for room %s, buildingId: %s, sectionId: %s", roomId, buildingID, sectionID)

	// Get current WAITING entries to determine position (filtered by tenant and section)
	entries, err := s.repo.GetQueueEntries(ctx, roomId, []string{"WAITING"})
	if err != nil {
		log.Printf("Failed to get queue entries: %v", err)
		// Continue with position 1 if we can't get current entries
	}

	// Calculate next position based only on WAITING entries
	nextPosition := 1
	nextPosition = len(entries) + 1

	// Create new entry
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
	}

	// Save to repository
	if err := s.repo.CreateEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create queue entry: %w", err)
	}

	// Recalculate positions to ensure consistency (filtered by tenant and section)
	if err := s.repo.RecalculatePositions(ctx, roomId); err != nil {
		log.Printf("Warning: Failed to recalculate positions after creating entry: %v", err)
	}

	log.Printf("Created queue entry %s with ticket %s for room %s, buildingId: %s, sectionId: %s", entry.ID, entry.TicketNumber, roomId, buildingID, sectionID)
	return entry, nil
}

// GetServicePoints returns the configured service points for a room
func (s *WaitingQueue) GetServicePoints(ctx context.Context, roomId string) ([]dto.ServicePoint, error) {
	var servicePoints []dto.ServicePoint

	// Try to get service points from tenant-aware config service first
	if s.configService != nil {
		rooms, err := s.configService.GetRoomsConfig(ctx)
		if err == nil && len(rooms) > 0 {
			// Find the room that matches roomId
			for _, room := range rooms {
				if room.ID == roomId {
					// Convert service points from room config to DTO
					for _, sp := range room.ServicePoints {
						servicePoint := dto.ServicePoint{
							ID:   sp.ID,
							Name: sp.Name,
						}
						if sp.Description != "" {
							servicePoint.Description = &sp.Description
						}
						servicePoints = append(servicePoints, servicePoint)
					}
					log.Printf("[WaitingQueue] Retrieved %d service points for room %s from tenant-aware config", len(servicePoints), roomId)
					return servicePoints, nil
				}
			}
			log.Printf("[WaitingQueue] Room %s not found in tenant-aware config, falling back to static config", roomId)
		} else {
			log.Printf("[WaitingQueue] Failed to get tenant-aware config or no rooms found, falling back to static config: %v", err)
		}
	}

	// Fallback to static config if tenant-aware config is not available or room not found
	servicePointConfigs := s.config.GetServicePointsForRoom(roomId)

	// Convert config to DTO
	for _, config := range servicePointConfigs {
		servicePoint := dto.ServicePoint{
			ID:   config.ID,
			Name: config.Name,
		}
		if config.Description != "" {
			servicePoint.Description = &config.Description
		}
		servicePoints = append(servicePoints, servicePoint)
	}

	log.Printf("[WaitingQueue] Retrieved %d service points for room %s from static config", len(servicePoints), roomId)
	return servicePoints, nil
}

// GetQueueEntries retrieves all queue entries for a room
func (s *WaitingQueue) GetQueueEntries(roomId string, states []string) ([]*Entry, error) {
	ctx := context.Background()
	entries, err := s.repo.GetQueueEntries(ctx, roomId, states)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue entries: %w", err)
	}
	return entries, nil
}

// GetQueueEntriesWithContext retrieves all queue entries for a room with a specific context (for tenant filtering)
func (s *WaitingQueue) GetQueueEntriesWithContext(ctx context.Context, roomId string, states []string) ([]*Entry, error) {
	entries, err := s.repo.GetQueueEntries(ctx, roomId, states)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue entries: %w", err)
	}
	return entries, nil
}

// GetEntryByQRToken retrieves a queue entry by QR token
func (s *WaitingQueue) GetEntryByQRToken(qrToken string) (*Entry, error) {
	ctx := context.Background()
	entry, err := s.repo.GetEntryByQRToken(ctx, qrToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry by QR token: %w", err)
	}
	return entry, nil
}

// CallNext calls the next person in the queue
func (s *WaitingQueue) CallNext(ctx context.Context, roomId string) (*Entry, error) {
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
func (s *WaitingQueue) FinishCurrent(roomId string) (*Entry, error) {
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
func (s *WaitingQueue) UpdateEntryStatus(id string, status string) error {
	ctx := context.Background()
	return s.repo.UpdateEntryStatus(ctx, id, status)
}

// DeleteEntry deletes a queue entry
func (s *WaitingQueue) DeleteEntry(id string) error {
	ctx := context.Background()
	return s.repo.DeleteEntry(ctx, id)
}

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

// Close closes the service and its repository
func (s *WaitingQueue) Close() error {
	return s.repo.Close()
}
