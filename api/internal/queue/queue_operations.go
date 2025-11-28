package queue

import (
	"context"
	"fmt"
	"log"
)

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
