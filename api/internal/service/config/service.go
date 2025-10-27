package config

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/arfis/waiting-room/internal/repository"
	"github.com/arfis/waiting-room/internal/types"
)

type Service struct {
	repo  repository.ConfigRepository
	cache *ConfigCache
}

func NewService(repo repository.ConfigRepository) *Service {
	return &Service{
		repo:  repo,
		cache: NewConfigCache(repo),
	}
}

// Stop stops the configuration cache
func (s *Service) Stop() {
	if s.cache != nil {
		s.cache.Stop()
	}
}

// GetSystemConfiguration gets the complete system configuration from cache
func (s *Service) GetSystemConfiguration(ctx context.Context) (*types.SystemConfiguration, error) {
	config := s.cache.GetSystemConfiguration()
	if config != nil {
		return config, nil
	}

	// If no cached config, initialize with environment variables
	log.Println("No cached configuration found, initializing with environment variables")
	envConfig := s.getSystemConfigurationFromEnv()

	// Save the environment config to MongoDB and cache
	if err := s.cache.UpdateConfiguration(ctx, envConfig); err != nil {
		log.Printf("Failed to save environment config: %v", err)
	}

	return envConfig, nil
}

// SetSystemConfiguration sets the complete system configuration in MongoDB and cache
func (s *Service) SetSystemConfiguration(ctx context.Context, config *types.SystemConfiguration) error {
	err := s.repo.SetSystemConfiguration(ctx, config)
	if err != nil {
		return err
	}

	// Update cache immediately
	s.cache.ReloadConfig(ctx)
	return nil
}

// UpdateSystemConfiguration updates specific fields in the system configuration
func (s *Service) UpdateSystemConfiguration(ctx context.Context, updates map[string]interface{}) error {
	err := s.repo.UpdateSystemConfiguration(ctx, updates)
	if err != nil {
		return err
	}

	// Update cache immediately
	s.cache.ReloadConfig(ctx)
	return nil
}

// GetExternalAPIConfig gets external API configuration from cache
func (s *Service) GetExternalAPIConfig(ctx context.Context) (*types.ExternalAPIConfig, error) {
	config := s.cache.GetExternalAPIConfig()
	if config != nil {
		log.Printf("Using cached config - Timeout: %d", config.TimeoutSeconds)
		return config, nil
	}

	// Fallback to environment variables
	log.Println("No cached config found, using environment variables")
	return s.getExternalAPIConfigFromEnv(), nil
}

// SetExternalAPIConfig updates external API configuration
func (s *Service) SetExternalAPIConfig(ctx context.Context, apiConfig *types.ExternalAPIConfig) error {
	log.Printf("Updating external API config - Timeout: %d", apiConfig.TimeoutSeconds)
	return s.cache.UpdateExternalAPIConfiguration(ctx, apiConfig)
}

// GetRoomsConfig gets rooms configuration from cache
func (s *Service) GetRoomsConfig(ctx context.Context) ([]types.RoomConfig, error) {
	rooms := s.cache.GetRoomsConfig()
	if len(rooms) > 0 {
		return rooms, nil
	}

	// Fallback to default rooms
	return s.getDefaultRoomsConfig(), nil
}

// SetRoomsConfig updates rooms configuration
func (s *Service) SetRoomsConfig(ctx context.Context, rooms []types.RoomConfig) error {
	return s.cache.UpdateRoomsConfiguration(ctx, rooms)
}

// GetDefaultRoom gets the default room ID
func (s *Service) GetDefaultRoom(ctx context.Context) (string, error) {
	config, err := s.GetSystemConfiguration(ctx)
	if err != nil {
		return "", err
	}

	if config != nil && config.DefaultRoom != "" {
		return config.DefaultRoom, nil
	}

	// Fallback to environment or default
	return os.Getenv("DEFAULT_ROOM"), nil
}

// SetDefaultRoom updates the default room
func (s *Service) SetDefaultRoom(ctx context.Context, roomId string) error {
	updates := map[string]interface{}{
		"defaultRoom": roomId,
	}
	return s.repo.UpdateSystemConfiguration(ctx, updates)
}

// GetCardReaders gets all card reader statuses
func (s *Service) GetCardReaders(ctx context.Context) ([]types.CardReaderStatus, error) {
	return s.repo.GetAllCardReaders(ctx)
}

// UpdateCardReaderStatus updates or creates a card reader status
func (s *Service) UpdateCardReaderStatus(ctx context.Context, status *types.CardReaderStatus) error {
	return s.repo.SetCardReaderStatus(ctx, status)
}

// UpdateCardReaderLastSeen updates the last seen timestamp for a card reader
func (s *Service) UpdateCardReaderLastSeen(ctx context.Context, id string) error {
	return s.repo.UpdateCardReaderLastSeen(ctx, id)
}

// DeleteCardReader removes a card reader from monitoring
func (s *Service) DeleteCardReader(ctx context.Context, id string) error {
	return s.repo.DeleteCardReader(ctx, id)
}

// RestartCardReader sends a restart signal to a card reader
func (s *Service) RestartCardReader(ctx context.Context, id string) (map[string]interface{}, error) {
	// In a real implementation, this would send a restart signal to the card reader
	// For now, we'll just return a success response
	return map[string]interface{}{
		"success": true,
		"message": "Restart signal sent to card reader " + id,
	}, nil
}

// GetExternalAPIConfiguration gets external API configuration
func (s *Service) GetExternalAPIConfiguration(ctx context.Context) (*types.ExternalAPIConfig, error) {
	return s.GetExternalAPIConfig(ctx)
}

// UpdateExternalAPIConfiguration updates external API configuration
func (s *Service) UpdateExternalAPIConfiguration(ctx context.Context, config *types.ExternalAPIConfig) error {
	return s.SetExternalAPIConfig(ctx, config)
}

// GetRoomsConfiguration gets rooms configuration
func (s *Service) GetRoomsConfiguration(ctx context.Context) ([]types.RoomConfig, error) {
	return s.GetRoomsConfig(ctx)
}

// UpdateRoomsConfiguration updates rooms configuration
func (s *Service) UpdateRoomsConfiguration(ctx context.Context, rooms []types.RoomConfig) error {
	return s.SetRoomsConfig(ctx, rooms)
}

// Helper methods for environment fallbacks
func (s *Service) getSystemConfigurationFromEnv() *types.SystemConfiguration {
	return &types.SystemConfiguration{
		ExternalAPI:   *s.getExternalAPIConfigFromEnv(),
		Rooms:         s.getDefaultRoomsConfig(),
		DefaultRoom:   os.Getenv("DEFAULT_ROOM"),
		WebSocketPath: "/ws/queue",
		AllowWildcard: true,
	}
}

func (s *Service) getExternalAPIConfigFromEnv() *types.ExternalAPIConfig {
	url := os.Getenv("EXTERNAL_API_URL")
	if url == "" {
		url = "http://localhost:3001/user-services"
	}

	timeoutStr := os.Getenv("EXTERNAL_API_TIMEOUT")
	timeout := 10
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = t
		}
	}

	retryStr := os.Getenv("EXTERNAL_API_RETRY")
	retry := 3
	if retryStr != "" {
		if r, err := strconv.Atoi(retryStr); err == nil {
			retry = r
		}
	}

	return &types.ExternalAPIConfig{
		TimeoutSeconds: timeout,
		RetryAttempts:  retry,
	}
}

func (s *Service) getDefaultRoomsConfig() []types.RoomConfig {
	return []types.RoomConfig{
		{
			ID:          "triage-1",
			Name:        "Triage Room 1",
			Description: "Main triage room",
			IsDefault:   true,
			ServicePoints: []types.ServicePointConfig{
				{
					ID:          "sp-1",
					Name:        "Service Point 1",
					Description: "Main service point",
				},
			},
		},
	}
}
