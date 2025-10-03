package repository

import (
	"context"

	"github.com/arfis/waiting-room/internal/types"
)

// QueueRepository defines the interface for queue data operations
type QueueRepository interface {
	// CreateEntry creates a new queue entry
	CreateEntry(ctx context.Context, entry *types.Entry) error

	// GetQueueEntries retrieves all queue entries for a room
	GetQueueEntries(ctx context.Context, roomId string) ([]*types.Entry, error)

	// GetEntryByID retrieves a queue entry by ID
	GetEntryByID(ctx context.Context, id string) (*types.Entry, error)

	// GetEntryByQRToken retrieves a queue entry by QR token
	GetEntryByQRToken(ctx context.Context, qrToken string) (*types.Entry, error)

	// UpdateEntryStatus updates the status of a queue entry
	UpdateEntryStatus(ctx context.Context, id string, status string) error

	// UpdateEntryPosition updates the position of a queue entry
	UpdateEntryPosition(ctx context.Context, id string, position int) error

	// GetNextWaitingEntry gets the next waiting entry for a room
	GetNextWaitingEntry(ctx context.Context, roomId string) (*types.Entry, error)

	// GetCurrentServedEntry gets the currently served entry for a room
	GetCurrentServedEntry(ctx context.Context, roomId string) (*types.Entry, error)

	// GetNextWaitingEntryForServicePoint gets the next waiting entry for a specific service point
	GetNextWaitingEntryForServicePoint(ctx context.Context, roomId, servicePointId string) (*types.Entry, error)

	// GetCurrentServedEntryForServicePoint gets the currently served entry for a specific service point
	GetCurrentServedEntryForServicePoint(ctx context.Context, roomId, servicePointId string) (*types.Entry, error)

	// RecalculatePositions recalculates positions for all waiting entries in a room
	RecalculatePositions(ctx context.Context, roomId string) error

	// DeleteEntry deletes a queue entry
	DeleteEntry(ctx context.Context, id string) error

	// Close closes the repository connection
	Close() error
}
