package queue

import (
	"context"

	"github.com/arfis/waiting-room/internal/config"
	"github.com/arfis/waiting-room/internal/priority"
	"github.com/arfis/waiting-room/internal/repository"
	"github.com/arfis/waiting-room/internal/service/servicepoint"
	"github.com/arfis/waiting-room/internal/types"
)

// Use types from the types package
type Entry = types.Entry
type CardData = types.CardData

// WaitingQueue manages the queue of patients waiting for service
// Methods are organized across multiple files:
// - entry_creation.go: CreateEntry with priority calculation
// - entry_retrieval.go: GetQueueEntries, GetQueueEntriesWithContext, GetEntryByQRToken
// - entry_management.go: UpdateEntryStatus, DeleteEntry
// - queue_operations.go: CallNext, FinishCurrent
// - servicepoint_operations.go: CallNextForServicePoint, CallSpecificEntryForServicePoint, etc.
// - service_points.go: GetServicePoints
type WaitingQueue struct {
	repo            repository.QueueRepository
	config          *config.Config
	configService   ConfigService
	servicePointSvc *servicepoint.Service
	priorityRepo    *priority.Repository
}

// ConfigService interface for getting tenant-aware configuration
type ConfigService interface {
	GetRoomsConfig(ctx context.Context) ([]types.RoomConfig, error)
}

// NewWaitingQueue creates a new waiting queue instance
func NewWaitingQueue(repo repository.QueueRepository, cfg *config.Config, servicePointSvc *servicepoint.Service, priorityRepo *priority.Repository) *WaitingQueue {
	return &WaitingQueue{
		repo:            repo,
		config:          cfg,
		servicePointSvc: servicePointSvc,
		priorityRepo:    priorityRepo,
	}
}

// SetConfigService sets the tenant-aware config service
func (s *WaitingQueue) SetConfigService(configService ConfigService) {
	s.configService = configService
}

// Close closes the service and its repository
func (s *WaitingQueue) Close() error {
	return s.repo.Close()
}
