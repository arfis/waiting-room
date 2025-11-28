package queue

import (
	"context"
)

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
