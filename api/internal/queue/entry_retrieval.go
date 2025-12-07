package queue

import (
	"context"
	"fmt"
)

// GetQueueEntriesWithContext retrieves all queue entries for a room with a specific context (for tenant filtering)
func (s *WaitingQueue) GetQueueEntriesWithContext(ctx context.Context, roomId string, states []string) ([]*Entry, error) {
	entries, err := s.repo.GetQueueEntries(ctx, roomId, states)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue entries: %w", err)
	}
	return entries, nil
}

// GetEntryByQRToken retrieves a queue entry by QR token
func (s *WaitingQueue) GetEntryByQRToken(ctx context.Context, qrToken string) (*Entry, error) {
	entry, err := s.repo.GetEntryByQRToken(ctx, qrToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry by QR token: %w", err)
	}
	return entry, nil
}
