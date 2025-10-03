package servicepoint

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/arfis/waiting-room/internal/config"
	"github.com/arfis/waiting-room/internal/data/dto"
)

// Service manages service point availability and manager status
type Service struct {
	config        *config.Config
	managerStatus map[string]*dto.ManagerStatus // key: managerID
	mu            sync.RWMutex
}

// NewService creates a new service point service
func NewService(cfg *config.Config) *Service {
	return &Service{
		config:        cfg,
		managerStatus: make(map[string]*dto.ManagerStatus),
	}
}

// SetManagerAvailable marks a manager as available at a service point
func (s *Service) SetManagerAvailable(ctx context.Context, managerID, roomID, servicePointID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get manager name from config
	managerName := s.getManagerName(roomID, servicePointID, managerID)

	status := &dto.ManagerStatus{
		ManagerId:      managerID,
		ManagerName:    managerName,
		ServicePointId: servicePointID,
		RoomId:         roomID,
		IsAvailable:    true,
		LastSeen:       time.Now(),
	}

	s.managerStatus[managerID] = status
	log.Printf("Manager %s (%s) is now available at service point %s in room %s",
		managerName, managerID, servicePointID, roomID)

	return nil
}

// SetManagerUnavailable marks a manager as unavailable
func (s *Service) SetManagerUnavailable(ctx context.Context, managerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if status, exists := s.managerStatus[managerID]; exists {
		status.IsAvailable = false
		status.LastSeen = time.Now()
		log.Printf("Manager %s (%s) is now unavailable", status.ManagerName, managerID)
	}

	return nil
}

// GetAvailableServicePoint returns the first available service point with an active manager for a room
func (s *Service) GetAvailableServicePoint(ctx context.Context, roomID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get service points for this room
	servicePoints := s.config.GetServicePointsForRoom(roomID)

	// Find the first service point that has an available manager
	for _, sp := range servicePoints {
		if sp.ManagerID == "" {
			continue // Skip service points without assigned managers
		}

		if status, exists := s.managerStatus[sp.ManagerID]; exists && status.IsAvailable {
			// Check if manager was seen recently (within last 5 minutes)
			if time.Since(status.LastSeen) < 5*time.Minute {
				log.Printf("Found available service point %s with manager %s for room %s",
					sp.ID, status.ManagerName, roomID)
				return sp.ID, nil
			}
		}
	}

	// If no manager is available, return the first service point as fallback
	if len(servicePoints) > 0 {
		log.Printf("No available managers found for room %s, using fallback service point %s",
			roomID, servicePoints[0].ID)
		return servicePoints[0].ID, nil
	}

	return "", fmt.Errorf("no service points configured for room %s", roomID)
}

// GetManagerStatus returns the current status of all managers
func (s *Service) GetManagerStatus(ctx context.Context) ([]dto.ManagerStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var statuses []dto.ManagerStatus
	for _, status := range s.managerStatus {
		statuses = append(statuses, *status)
	}

	return statuses, nil
}

// GetManagerStatusForRoom returns the status of managers for a specific room
func (s *Service) GetManagerStatusForRoom(ctx context.Context, roomID string) ([]dto.ManagerStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var statuses []dto.ManagerStatus
	for _, status := range s.managerStatus {
		if status.RoomId == roomID {
			statuses = append(statuses, *status)
		}
	}

	return statuses, nil
}

// CleanupInactiveManagers removes managers that haven't been seen for more than 10 minutes
func (s *Service) CleanupInactiveManagers(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	for managerID, status := range s.managerStatus {
		if status.LastSeen.Before(cutoff) {
			log.Printf("Removing inactive manager %s (%s) - last seen: %s",
				status.ManagerName, managerID, status.LastSeen.Format(time.RFC3339))
			delete(s.managerStatus, managerID)
		}
	}
}

// getManagerName retrieves the manager name from config
func (s *Service) getManagerName(roomID, servicePointID, managerID string) string {
	servicePoints := s.config.GetServicePointsForRoom(roomID)
	for _, sp := range servicePoints {
		if sp.ID == servicePointID && sp.ManagerID == managerID {
			return sp.ManagerName
		}
	}
	return "Unknown Manager"
}

// StartCleanupRoutine starts a background routine to clean up inactive managers
func (s *Service) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.CleanupInactiveManagers(ctx)
			}
		}
	}()
}

// ManagerLogin handles manager login to a service point
func (s *Service) ManagerLogin(ctx context.Context, managerID string, req *dto.ManagerLoginRequest) (*dto.ManagerStatus, error) {
	err := s.SetManagerAvailable(ctx, managerID, req.RoomId, req.ServicePointId)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if status, exists := s.managerStatus[managerID]; exists {
		return status, nil
	}

	return nil, fmt.Errorf("manager status not found after login")
}

// ManagerLogout handles manager logout from a service point
func (s *Service) ManagerLogout(ctx context.Context, managerID string) (*dto.ManagerStatus, error) {
	s.mu.RLock()
	var status *dto.ManagerStatus
	if existingStatus, exists := s.managerStatus[managerID]; exists {
		// Create a copy of the status before marking as unavailable
		status = &dto.ManagerStatus{
			ManagerId:      existingStatus.ManagerId,
			ManagerName:    existingStatus.ManagerName,
			ServicePointId: existingStatus.ServicePointId,
			RoomId:         existingStatus.RoomId,
			IsAvailable:    false,
			LastSeen:       time.Now(),
		}
	}
	s.mu.RUnlock()

	if status == nil {
		return nil, fmt.Errorf("manager not found")
	}

	err := s.SetManagerUnavailable(ctx, managerID)
	if err != nil {
		return nil, err
	}

	return status, nil
}
