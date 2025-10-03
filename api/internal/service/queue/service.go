package queue

import (
	"context"

	"github.com/arfis/waiting-room/internal/data/dto"
	"github.com/arfis/waiting-room/internal/data/dto/queueentrystatus"
	ngErrors "github.com/arfis/waiting-room/internal/errors"
	"github.com/arfis/waiting-room/internal/queue"
)

type Service struct {
	queueService  *queue.Service
	broadcastFunc func(string) // Function to broadcast queue updates
}

func New(queueService *queue.Service, broadcastFunc func(string)) *Service {
	return &Service{
		queueService:  queueService,
		broadcastFunc: broadcastFunc,
	}
}

func (s *Service) SetBroadcastFunc(f func(string)) {
	s.broadcastFunc = f
}

func (s *Service) GetQueueEntryByToken(ctx context.Context, qrToken string) (*dto.PublicEntry, error) {
	entry, err := s.queueService.GetEntryByQRToken(qrToken)
	if err != nil {
		return nil, ngErrors.New(ngErrors.NotFoundErrorCode, "queue entry not found", 404, nil)
	}

	// Convert to PublicEntry
	publicEntry := &dto.PublicEntry{
		EntryId:      entry.ID,
		TicketNumber: entry.TicketNumber,
		Status:       queueentrystatus.QueueEntryStatus(entry.Status),
		Position:     entry.Position,
		EtaMinutes:   entry.Position * 5, // Simple calculation: 5 minutes per position
		CanCancel:    entry.Status == "WAITING",
	}

	return publicEntry, nil
}

func (s *Service) CallNext(ctx context.Context, roomId string) (*dto.QueueEntry, error) {
	entry, err := s.queueService.CallNext(roomId)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to call next", 500, nil)
	}

	// Convert to QueueEntry
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

	// Broadcast queue update
	if s.broadcastFunc != nil {
		s.broadcastFunc(roomId)
	}

	return queueEntry, nil
}

func (s *Service) FinishCurrent(ctx context.Context, roomId string) (*dto.QueueEntry, error) {
	entry, err := s.queueService.FinishCurrent(roomId)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to finish current", 500, nil)
	}

	if entry == nil {
		return nil, ngErrors.New(ngErrors.NotFoundErrorCode, "no one is currently being served", 404, nil)
	}

	// Convert to QueueEntry
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

	// Broadcast queue update
	if s.broadcastFunc != nil {
		s.broadcastFunc(roomId)
	}

	return queueEntry, nil
}

func (s *Service) GetQueueEntries(ctx context.Context, roomId string) ([]dto.QueueEntry, error) {
	entries, err := s.queueService.GetQueueEntries(roomId)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to get queue entries", 500, nil)
	}

	// Convert to DTOs
	var queueEntries []dto.QueueEntry
	for _, entry := range entries {
		queueEntry := dto.QueueEntry{
			Id:            entry.ID,
			WaitingRoomId: entry.WaitingRoomID,
			TicketNumber:  entry.TicketNumber,
			Status:        queueentrystatus.QueueEntryStatus(entry.Status),
			Position:      entry.Position,
		}
		if entry.ServicePoint != "" {
			queueEntry.ServicePoint = &entry.ServicePoint
		}
		queueEntries = append(queueEntries, queueEntry)
	}

	return queueEntries, nil
}

func (s *Service) GetServicePoints(ctx context.Context, roomId string) ([]dto.ServicePoint, error) {
	return s.queueService.GetServicePoints(ctx, roomId)
}

func (s *Service) CallNextForServicePoint(ctx context.Context, roomId, servicePointId string) (*dto.QueueEntry, error) {
	return s.queueService.CallNextForServicePoint(ctx, roomId, servicePointId)
}

func (s *Service) MarkInRoomForServicePoint(ctx context.Context, roomId, servicePointId string, req *dto.MarkInRoomRequest) (*dto.QueueEntry, error) {
	return s.queueService.MarkInRoomForServicePoint(ctx, roomId, servicePointId, req.EntryId)
}

func (s *Service) FinishCurrentForServicePoint(ctx context.Context, roomId, servicePointId string) (*dto.QueueEntry, error) {
	return s.queueService.FinishCurrentForServicePoint(ctx, roomId, servicePointId)
}
