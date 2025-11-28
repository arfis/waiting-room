package queue

import (
	"context"
	"log"

	"github.com/arfis/waiting-room/internal/data/dto"
)

// GetServicePoints returns the configured service points for a room
func (s *WaitingQueue) GetServicePoints(ctx context.Context, roomId string) ([]dto.ServicePoint, error) {
	var servicePoints []dto.ServicePoint

	// Try to get service points from tenant-aware config service first
	if s.configService != nil {
		rooms, err := s.configService.GetRoomsConfig(ctx)
		if err == nil && len(rooms) > 0 {
			// Find the room that matches roomId
			for _, room := range rooms {
				if room.ID == roomId {
					// Convert service points from room config to DTO
					for _, sp := range room.ServicePoints {
						servicePoint := dto.ServicePoint{
							ID:   sp.ID,
							Name: sp.Name,
						}
						if sp.Description != "" {
							servicePoint.Description = &sp.Description
						}
						servicePoints = append(servicePoints, servicePoint)
					}
					log.Printf("[WaitingQueue] Retrieved %d service points for room %s from tenant-aware config", len(servicePoints), roomId)
					return servicePoints, nil
				}
			}
			log.Printf("[WaitingQueue] Room %s not found in tenant-aware config, falling back to static config", roomId)
		} else {
			log.Printf("[WaitingQueue] Failed to get tenant-aware config or no rooms found, falling back to static config: %v", err)
		}
	}

	// Fallback to static config if tenant-aware config is not available or room not found
	servicePointConfigs := s.config.GetServicePointsForRoom(roomId)

	// Convert config to DTO
	for _, config := range servicePointConfigs {
		servicePoint := dto.ServicePoint{
			ID:   config.ID,
			Name: config.Name,
		}
		if config.Description != "" {
			servicePoint.Description = &config.Description
		}
		servicePoints = append(servicePoints, servicePoint)
	}

	log.Printf("[WaitingQueue] Retrieved %d service points for room %s from static config", len(servicePoints), roomId)
	return servicePoints, nil
}
