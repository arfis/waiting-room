package queue

import (
	"context"
	"log"

	"github.com/arfis/waiting-room/internal/data/dto"
	"github.com/arfis/waiting-room/internal/data/dto/queueentrystatus"
	ngErrors "github.com/arfis/waiting-room/internal/errors"
	"github.com/arfis/waiting-room/internal/middleware"
	"github.com/arfis/waiting-room/internal/queue"
	"github.com/arfis/waiting-room/internal/service"
	"github.com/arfis/waiting-room/internal/service/webhook"
)

type Service struct {
	queueService   *queue.WaitingQueue
	broadcastFunc  func(string, string) // Function to broadcast queue updates (roomId, tenantID)
	webhookService *webhook.Service
}

func New(queueService *queue.WaitingQueue, broadcastFunc func(string, string), webhookService *webhook.Service) *Service {
	return &Service{
		queueService:   queueService,
		broadcastFunc:  broadcastFunc,
		webhookService: webhookService,
	}
}

func (s *Service) SetBroadcastFunc(f func(string, string)) {
	s.broadcastFunc = f
}

func (s *Service) GetQueueEntryByToken(ctx context.Context, qrToken string) (*dto.PublicEntry, error) {
	entry, err := s.queueService.GetEntryByQRToken(qrToken)
	if err != nil {
		return nil, ngErrors.New(ngErrors.NotFoundErrorCode, "queue entry not found", 404, nil)
	}

	// Convert to PublicEntry
	publicEntry := &dto.PublicEntry{
		EntryID:      entry.ID,
		TicketNumber: entry.TicketNumber,
		Status:       queueentrystatus.QueueEntryStatus(entry.Status),
		Position:     entry.Position,
		EtaMinutes:   entry.Position * 5, // Simple calculation: 5 minutes per position
		CanCancel:    entry.Status == "WAITING",
	}

	return publicEntry, nil
}

func (s *Service) CallNext(ctx context.Context, roomId string, servicePointId string) (*dto.QueueEntry, error) {

	entry, err := s.queueService.CallNextForServicePoint(ctx, roomId, servicePointId)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to call next", 500, nil)
	}

	// Convert to QueueEntry
	queueEntry := &dto.QueueEntry{
		ID:            entry.ID,
		WaitingRoomID: entry.WaitingRoomID,
		TicketNumber:  entry.TicketNumber,
		Status:        queueentrystatus.QueueEntryStatus(entry.Status),
		Position:      int64(entry.Position),
		ServicePoint:  &entry.ServicePoint,
	}
	if entry.ServicePoint != "" {
		queueEntry.ServicePoint = &entry.ServicePoint
	}

	// Broadcast queue update - only to the tenant that changed
	// Extract tenant ID from context (format: "buildingId:sectionId")
	if s.broadcastFunc != nil {
		tenantID := service.GetTenantID(ctx)
		log.Printf("[QueueService] Broadcasting queue update for room %s, tenantID: '%s' (length: %d)", roomId, tenantID, len(tenantID))
		if tenantID == "" {
			log.Printf("[QueueService] WARNING: tenantID is empty, broadcasting to all clients")
		}
		s.broadcastFunc(roomId, tenantID)
	}

	// Send webhook notification for ticket called
	if s.webhookService != nil {
		go func() {
			if err := s.webhookService.SendTicketCalledWebhook(ctx, entry.ID, roomId, servicePointId, ""); err != nil {
				log.Printf("Failed to send webhook notification for ticket called: %v", err)
			}
		}()
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
		ID:            entry.ID,
		WaitingRoomID: entry.WaitingRoomID,
		TicketNumber:  entry.TicketNumber,
		Status:        queueentrystatus.QueueEntryStatus(entry.Status),
		Position:      entry.Position,
	}
	if entry.ServicePoint != "" {
		queueEntry.ServicePoint = &entry.ServicePoint
	}

	// Broadcast queue update - only to the tenant that changed
	// Extract tenant ID from context (format: "buildingId:sectionId")
	if s.broadcastFunc != nil {
		tenantID := service.GetTenantID(ctx)
		log.Printf("[QueueService] Broadcasting queue update for room %s, tenantID: '%s' (length: %d)", roomId, tenantID, len(tenantID))
		if tenantID == "" {
			log.Printf("[QueueService] WARNING: tenantID is empty, broadcasting to all clients")
		}
		s.broadcastFunc(roomId, tenantID)
	}

	// Send webhook notification for ticket completed
	if s.webhookService != nil {
		go func() {
			if err := s.webhookService.SendTicketCompletedWebhook(ctx, entry.ID, roomId, entry.ServicePoint, ""); err != nil {
				log.Printf("Failed to send webhook notification for ticket completed: %v", err)
			}
		}()
	}

	return queueEntry, nil
}

func (s *Service) GetQueueEntries(ctx context.Context, roomId string, states []string) ([]dto.QueueEntry, error) {
	// Log tenant ID from context for debugging
	tenantID := ctx.Value(middleware.TENANT)
	if tenantID != nil {
		log.Printf("[QueueService] GetQueueEntries for room %s with tenant ID: %s", roomId, tenantID.(string))
	} else {
		log.Printf("[QueueService] GetQueueEntries for room %s without tenant ID", roomId)
	}
	
	// Use GetQueueEntriesWithContext to preserve tenant ID from context
	entries, err := s.queueService.GetQueueEntriesWithContext(ctx, roomId, states)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to get queue entries", 500, nil)
	}
	
	log.Printf("[QueueService] GetQueueEntries returned %d entries for room %s", len(entries), roomId)

	// Convert to DTOs
	var queueEntries []dto.QueueEntry
	for _, entry := range entries {
		queueEntry := dto.QueueEntry{
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
		if entry.ApproximateDurationMinutes > 0 {
			queueEntry.ServiceDuration = &entry.ApproximateDurationMinutes
		}
		queueEntries = append(queueEntries, queueEntry)
	}

	return queueEntries, nil
}

func (s *Service) GetServicePoints(ctx context.Context, roomId string) ([]dto.ServicePoint, error) {
	return s.queueService.GetServicePoints(ctx, roomId)
}

func (s *Service) MarkInRoomForServicePoint(ctx context.Context, roomId, servicePointId string, req *dto.MarkInRoomRequest) (*dto.QueueEntry, error) {
	return s.queueService.MarkInRoomForServicePoint(ctx, roomId, servicePointId, req.EntryID)
}

func (s *Service) FinishCurrentForServicePoint(ctx context.Context, roomId, servicePointId string) (*dto.QueueEntry, error) {
	entry, err := s.queueService.FinishCurrentForServicePoint(ctx, roomId, servicePointId)
	if err != nil {
		return nil, err
	}

	// Send webhook notification for ticket completed
	if s.webhookService != nil {
		go func() {
			if err := s.webhookService.SendTicketCompletedWebhook(ctx, entry.ID, roomId, servicePointId, ""); err != nil {
				log.Printf("Failed to send webhook notification for ticket completed: %v", err)
			}
		}()
	}

	return entry, nil
}
