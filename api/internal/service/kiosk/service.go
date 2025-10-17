package kiosk

import (
	"context"

	"github.com/arfis/waiting-room/internal/data/dto"
	ngErrors "github.com/arfis/waiting-room/internal/errors"
	"github.com/arfis/waiting-room/internal/queue"
)

type Service struct {
	queueService  *queue.WaitingQueue
	broadcastFunc func(string) // Function to broadcast queue updates
}

func New(queueService *queue.WaitingQueue, broadcastFunc func(string)) *Service {
	return &Service{
		queueService:  queueService,
		broadcastFunc: broadcastFunc,
	}
}

func (s *Service) SetBroadcastFunc(f func(string)) {
	s.broadcastFunc = f
}

func (s *Service) SwipeCard(ctx context.Context, roomId string, req *dto.SwipeRequest) (*dto.JoinResult, error) {
	// Create CardData from the raw card data
	cardData := queue.CardData{
		IDNumber: *req.IdCardRaw,
		Source:   "card-reader",
	}

	// this will be later replaced with the actual duration
	approximateDurationMinutes := int64(5)
	// Create queue entry using the existing queue service
	entry, err := s.queueService.CreateEntry(roomId, cardData, approximateDurationMinutes)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to create queue entry", 500, nil)
	}

	// Generate QR URL
	qrUrl := "http://localhost:4204/q/" + entry.QRToken

	// Broadcast queue update
	if s.broadcastFunc != nil {
		s.broadcastFunc(roomId)
	}

	// Return the join result
	return &dto.JoinResult{
		EntryID:      entry.ID,
		TicketNumber: entry.TicketNumber,
		QrUrl:        qrUrl,
	}, nil
}
