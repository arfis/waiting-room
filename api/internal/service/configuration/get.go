package configuration

import (
	"context"

	"github.com/arfis/waiting-room/internal/config"
	"github.com/arfis/waiting-room/internal/data/dto"
)

type Service struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (s *Service) GetConfiguration(_ context.Context) (*dto.ConfigurationResponse, error) {
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
