package queue

import (
	"context"
	"fmt"
	"log"

	"github.com/arfis/waiting-room/internal/repository"
	"github.com/arfis/waiting-room/internal/types"
)

// Use types from the types package
type Entry = types.Entry
type CardData = types.CardData

type Service struct {
	repo repository.QueueRepository
}

func NewService(repo repository.QueueRepository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateEntry creates a new queue entry
func (s *Service) CreateEntry(roomId string, cardData CardData) (*Entry, error) {
	ctx := context.Background()

	// Get current queue to determine position
	entries, err := s.repo.GetQueueEntries(ctx, roomId)
	if err != nil {
		log.Printf("Failed to get queue entries: %v", err)
		// Continue with position 1 if we can't get current entries
	}

	// Calculate next position
	nextPosition := 1
	if len(entries) > 0 {
		maxPosition := 0
		for _, entry := range entries {
			if entry.Position > maxPosition {
				maxPosition = entry.Position
			}
		}
		nextPosition = maxPosition + 1
	}

	// Create new entry
	entry := &Entry{
		WaitingRoomID: roomId,
		TicketNumber:  "", // Will be set by repository
		QRToken:       "", // Will be set by repository
		Status:        "WAITING",
		Position:      nextPosition,
		CardData:      cardData,
	}

	// Save to repository
	if err := s.repo.CreateEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create queue entry: %w", err)
	}

	log.Printf("Created queue entry %s with ticket %s for room %s", entry.ID, entry.TicketNumber, roomId)
	return entry, nil
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

// Close closes the service and its repository
func (s *Service) Close() error {
	return s.repo.Close()
}
