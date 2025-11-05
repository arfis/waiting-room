package configuration

import (
	"context"
	"log"

	"github.com/arfis/waiting-room/internal/config"
	"github.com/arfis/waiting-room/internal/data/dto"
	"github.com/arfis/waiting-room/internal/types"
)

type ConfigService interface {
	GetRoomsConfig(ctx context.Context) ([]types.RoomConfig, error)
	GetSystemConfiguration(ctx context.Context) (*types.SystemConfiguration, error)
}

type Service struct {
	cfg          *config.Config
	configService ConfigService
}

func New(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
	}
}

// SetConfigService sets the tenant-aware config service
func (s *Service) SetConfigService(configService ConfigService) {
	s.configService = configService
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (s *Service) GetConfiguration(ctx context.Context) (*dto.ConfigurationResponse, error) {
	// Try to get tenant-aware configuration first
	if s.configService != nil {
		systemConfig, err := s.configService.GetSystemConfiguration(ctx)
		if err == nil && systemConfig != nil {
			log.Printf("[ConfigurationService] Using tenant-aware configuration")
			
			rooms, err := s.configService.GetRoomsConfig(ctx)
			if err == nil && len(rooms) > 0 {
				response := dto.ConfigurationResponse{
					DefaultRoom:   systemConfig.DefaultRoom,
					AllowWildcard: systemConfig.AllowWildcard,
					WebSocketPath: systemConfig.WebSocketPath,
					Rooms:         make([]dto.RoomConfiguration, 0, len(rooms)),
				}

				for _, room := range rooms {
					roomDetails := dto.RoomConfiguration{
						ID:            room.ID,
						Name:          room.Name,
						ServicePoints: make([]dto.ServicePointConfiguration, 0, len(room.ServicePoints)),
					}

					for _, sp := range room.ServicePoints {
						roomDetails.ServicePoints = append(roomDetails.ServicePoints, dto.ServicePointConfiguration{
							ID:          sp.ID,
							Name:        sp.Name,
							Description: optionalString(sp.Description),
							ManagerID:   optionalString(sp.ManagerID),
							ManagerName: optionalString(sp.ManagerName),
						})
					}

					response.Rooms = append(response.Rooms, roomDetails)
				}

				log.Printf("[ConfigurationService] Returning %d rooms from tenant-aware config", len(response.Rooms))
				return &response, nil
			} else {
				log.Printf("[ConfigurationService] No rooms found in tenant-aware config, falling back to static config")
			}
		} else {
			log.Printf("[ConfigurationService] Failed to get tenant-aware config, falling back to static config: %v", err)
		}
	}

	// Fallback to static config
	log.Printf("[ConfigurationService] Using static configuration")
	response := dto.ConfigurationResponse{
		DefaultRoom:   s.cfg.Rooms.DefaultRoom,
		AllowWildcard: s.cfg.Rooms.AllowWildcard,
		WebSocketPath: s.cfg.WebSocket.Path,
		Rooms:         make([]dto.RoomConfiguration, 0, len(s.cfg.Rooms.Rooms)),
	}

	for _, room := range s.cfg.Rooms.Rooms {
		roomDetails := dto.RoomConfiguration{
			ID:            room.ID,
			Name:          room.Name,
			ServicePoints: make([]dto.ServicePointConfiguration, 0, len(room.ServicePoints)),
		}

		for _, sp := range room.ServicePoints {
			roomDetails.ServicePoints = append(roomDetails.ServicePoints, dto.ServicePointConfiguration{
				ID:          sp.ID,
				Name:        sp.Name,
				Description: optionalString(sp.Description),
				ManagerID:   optionalString(sp.ManagerID),
				ManagerName: optionalString(sp.ManagerName),
			})
		}

		response.Rooms = append(response.Rooms, roomDetails)
	}

	// If no explicit rooms are configured, expose at least the default room with fallback service points.
	if len(response.Rooms) == 0 {
		servicePoints := s.cfg.GetServicePointsForRoom(s.cfg.Rooms.DefaultRoom)
		roomDetails := dto.RoomConfiguration{
			ID:            s.cfg.Rooms.DefaultRoom,
			Name:          s.cfg.Rooms.DefaultRoom,
			ServicePoints: make([]dto.ServicePointConfiguration, 0, len(servicePoints)),
		}
		for _, sp := range servicePoints {
			roomDetails.ServicePoints = append(roomDetails.ServicePoints, dto.ServicePointConfiguration{
				ID:          sp.ID,
				Name:        sp.Name,
				Description: optionalString(sp.Description),
				ManagerID:   optionalString(sp.ManagerID),
				ManagerName: optionalString(sp.ManagerName),
			})
		}
		response.Rooms = append(response.Rooms, roomDetails)
	}

	return &response, nil
}
