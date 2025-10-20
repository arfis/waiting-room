package repository

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/arfis/waiting-room/internal/types"
)

// MockQueueRepository implements QueueRepository using in-memory storage
type MockQueueRepository struct {
	entries map[string]*types.Entry
	mutex   sync.RWMutex
	counter int
}

// NewMockQueueRepository creates a new mock queue repository
func NewMockQueueRepository() *MockQueueRepository {
	return &MockQueueRepository{
		entries: make(map[string]*types.Entry),
		counter: 0,
	}
}

// CreateEntry creates a new queue entry
func (r *MockQueueRepository) CreateEntry(ctx context.Context, entry *types.Entry) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.counter++
	entry.ID = fmt.Sprintf("mock-%d", r.counter)
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	// Generate ticket number
	entry.TicketNumber = fmt.Sprintf("A-%03d", r.counter)

	// Generate QR token
	entry.QRToken = fmt.Sprintf("qr-token-%d", r.counter)

	r.entries[entry.ID] = entry
	log.Printf("Mock: Created queue entry %s with ticket %s", entry.ID, entry.TicketNumber)

	return nil
}

// GetQueueEntries retrieves all queue entries for a room
func (r *MockQueueRepository) GetQueueEntries(ctx context.Context, roomId string, states []string) ([]*types.Entry, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var entries []*types.Entry
	for _, entry := range r.entries {
		if entry.WaitingRoomID == roomId {
			// If no states specified, include all entries
			if len(states) == 0 {
				entries = append(entries, entry)
			} else {
				// Check if entry status is in the states array
				for _, state := range states {
					if entry.Status == state {
						entries = append(entries, entry)
						break
					}
				}
			}
		}
	}

	// Sort by position
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].Position > entries[j].Position {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	return entries, nil
}

// GetEntryByID retrieves a queue entry by ID
func (r *MockQueueRepository) GetEntryByID(ctx context.Context, id string) (*types.Entry, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	entry, exists := r.entries[id]
	if !exists {
		return nil, fmt.Errorf("queue entry not found")
	}

	return entry, nil
}

// GetEntryByQRToken retrieves a queue entry by QR token
func (r *MockQueueRepository) GetEntryByQRToken(ctx context.Context, qrToken string) (*types.Entry, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, entry := range r.entries {
		if entry.QRToken == qrToken {
			return entry, nil
		}
	}

	return nil, fmt.Errorf("queue entry not found")
}

// UpdateEntryStatus updates the status of a queue entry
func (r *MockQueueRepository) UpdateEntryStatus(ctx context.Context, id string, status string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	entry, exists := r.entries[id]
	if !exists {
		return fmt.Errorf("queue entry not found")
	}

	entry.Status = status
	entry.UpdatedAt = time.Now()

	log.Printf("Mock: Updated entry %s status to %s", id, status)
	return nil
}

// UpdateEntryPosition updates the position of a queue entry
func (r *MockQueueRepository) UpdateEntryPosition(ctx context.Context, id string, position int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	entry, exists := r.entries[id]
	if !exists {
		return fmt.Errorf("queue entry not found")
	}

	entry.Position = int64(position)
	entry.UpdatedAt = time.Now()

	return nil
}

// UpdateEntryServicePoint updates the service point of a queue entry
func (r *MockQueueRepository) UpdateEntryServicePoint(ctx context.Context, id string, servicePoint string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	entry, exists := r.entries[id]
	if !exists {
		return fmt.Errorf("queue entry not found")
	}

	entry.ServicePoint = servicePoint
	entry.UpdatedAt = time.Now()

	log.Printf("Mock: Updated entry %s service point to %s", id, servicePoint)
	return nil
}

// GetNextWaitingEntry gets the next waiting entry for a room
func (r *MockQueueRepository) GetNextWaitingEntry(ctx context.Context, roomId string) (*types.Entry, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var nextEntry *types.Entry
	minPosition := int(^uint(0) >> 1) // Max int

	for _, entry := range r.entries {
		if entry.WaitingRoomID == roomId && entry.Status == "WAITING" {
			if entry.Position < int64(minPosition) {
				minPosition = int(entry.Position)
				nextEntry = entry
			}
		}
	}

	if nextEntry == nil {
		return nil, fmt.Errorf("no waiting entries found")
	}

	return nextEntry, nil
}

// GetCurrentServedEntry gets the currently served entry for a room
func (r *MockQueueRepository) GetCurrentServedEntry(ctx context.Context, roomId string) (*types.Entry, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, entry := range r.entries {
		if entry.WaitingRoomID == roomId && (entry.Status == "CALLED" || entry.Status == "IN_SERVICE") {
			return entry, nil
		}
	}

	return nil, nil // No one currently being served
}

// RecalculatePositions recalculates positions for all waiting entries in a room
func (r *MockQueueRepository) RecalculatePositions(ctx context.Context, roomId string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var waitingEntries []*types.Entry
	for _, entry := range r.entries {
		if entry.WaitingRoomID == roomId && entry.Status == "WAITING" {
			waitingEntries = append(waitingEntries, entry)
		}
	}

	// Sort by creation time
	for i := 0; i < len(waitingEntries)-1; i++ {
		for j := i + 1; j < len(waitingEntries); j++ {
			if waitingEntries[i].CreatedAt.After(waitingEntries[j].CreatedAt) {
				waitingEntries[i], waitingEntries[j] = waitingEntries[j], waitingEntries[i]
			}
		}
	}

	// Update positions
	for i, entry := range waitingEntries {
		entry.Position = int64(i + 1)
		entry.UpdatedAt = time.Now()
	}

	log.Printf("Mock: Recalculated positions for %d waiting entries in room %s", len(waitingEntries), roomId)
	return nil
}

// DeleteEntry deletes a queue entry
func (r *MockQueueRepository) DeleteEntry(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, exists := r.entries[id]
	if !exists {
		return fmt.Errorf("queue entry not found")
	}

	delete(r.entries, id)
	log.Printf("Mock: Deleted queue entry %s", id)
	return nil
}

// GetNextWaitingEntryForServicePoint gets the next waiting entry for a specific service point
func (r *MockQueueRepository) GetNextWaitingEntryForServicePoint(ctx context.Context, roomId, servicePointId string) (*types.Entry, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, entry := range r.entries {
		if entry.WaitingRoomID == roomId && entry.ServicePoint == servicePointId && entry.Status == "WAITING" {
			return entry, nil
		}
	}
	return nil, nil
}

// GetCurrentServedEntryForServicePoint gets the currently served entry for a specific service point
func (r *MockQueueRepository) GetCurrentServedEntryForServicePoint(ctx context.Context, roomId, servicePointId string) (*types.Entry, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, entry := range r.entries {
		if entry.WaitingRoomID == roomId && entry.ServicePoint == servicePointId &&
			(entry.Status == "CALLED" || entry.Status == "IN_ROOM" || entry.Status == "IN_SERVICE") {
			return entry, nil
		}
	}
	return nil, nil
}

// Close closes the repository connection (no-op for mock)
func (r *MockQueueRepository) Close() error {
	return nil
}
