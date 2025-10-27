package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/arfis/waiting-room/internal/data/dto"
	"github.com/arfis/waiting-room/internal/service/config"
	"github.com/arfis/waiting-room/internal/types"
)

type Service struct {
	configService *config.Service
}

func NewService(configService *config.Service) *Service {
	return &Service{
		configService: configService,
	}
}

// System Configuration methods
func (s *Service) GetSystemConfiguration(ctx context.Context) (*dto.SystemConfiguration, error) {
	config, err := s.configService.GetSystemConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	// Convert types to DTOs
	return s.convertSystemConfigurationToDTO(config), nil
}

func (s *Service) UpdateSystemConfiguration(ctx context.Context, config *dto.SystemConfiguration) (*dto.SystemConfiguration, error) {
	// Convert DTO to types
	systemConfig := s.convertDTOToSystemConfiguration(config)
	err := s.configService.SetSystemConfiguration(ctx, systemConfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// External API Configuration methods
func (s *Service) GetExternalAPIConfiguration(ctx context.Context) (*dto.ExternalAPIConfig, error) {
	config, err := s.configService.GetExternalAPIConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	// Convert types to DTO
	externalAPIConfig := &dto.ExternalAPIConfig{
		TimeoutSeconds: int64(config.TimeoutSeconds),
		RetryAttempts:  int64(config.RetryAttempts),
		Headers:        config.Headers,
	}

	// Add optional URLs if they exist
	if config.AppointmentServicesURL != "" {
		externalAPIConfig.AppointmentServicesUrl = &config.AppointmentServicesURL
	}
	if config.GenericServicesURL != "" {
		externalAPIConfig.GenericServicesUrl = &config.GenericServicesURL
	}

	// Convert GenericServices from types to DTO
	if len(config.GenericServices) > 0 {
		genericServices := make([]dto.GenericService, len(config.GenericServices))
		for i, service := range config.GenericServices {
			genericServices[i] = dto.GenericService{
				Id:          service.ID,
				Name:        service.Name,
				Description: service.Description,
				Duration:    service.Duration,
				Enabled:     service.Enabled,
			}
		}
		externalAPIConfig.GenericServices = genericServices
	}

	if config.WebhookURL != "" {
		externalAPIConfig.WebhookUrl = &config.WebhookURL
	}
	if config.WebhookTimeoutSeconds > 0 {
		timeout := int64(config.WebhookTimeoutSeconds)
		externalAPIConfig.WebhookTimeoutSeconds = &timeout
	}
	if config.WebhookRetryAttempts > 0 {
		retries := int64(config.WebhookRetryAttempts)
		externalAPIConfig.WebhookRetryAttempts = &retries
	}

	return externalAPIConfig, nil
}

func (s *Service) UpdateExternalAPIConfiguration(ctx context.Context, config *dto.ExternalAPIConfig) (*dto.ExternalAPIConfig, error) {
	// Convert DTO to types
	externalAPIConfig := &types.ExternalAPIConfig{
		TimeoutSeconds: int(config.TimeoutSeconds),
		RetryAttempts:  int(config.RetryAttempts),
		Headers:        config.Headers,
	}

	// Add optional URLs if they exist
	if config.AppointmentServicesUrl != nil && *config.AppointmentServicesUrl != "" {
		// Validate that the URL contains the ${identifier} placeholder
		if !strings.Contains(*config.AppointmentServicesUrl, "${identifier}") {
			return nil, fmt.Errorf("URL must contain ${identifier} placeholder. Example: http://api.example.com/users/${identifier}/services")
		}
		externalAPIConfig.AppointmentServicesURL = *config.AppointmentServicesUrl
	}
	if config.GenericServicesUrl != nil && *config.GenericServicesUrl != "" {
		externalAPIConfig.GenericServicesURL = *config.GenericServicesUrl
	}

	// Convert GenericServices from DTO to types
	if len(config.GenericServices) > 0 {
		genericServices := make([]types.GenericService, len(config.GenericServices))
		for i, service := range config.GenericServices {
			genericServices[i] = types.GenericService{
				ID:          service.Id,
				Name:        service.Name,
				Description: service.Description,
				Duration:    service.Duration,
				Enabled:     service.Enabled,
			}
		}
		externalAPIConfig.GenericServices = genericServices
	}

	if config.WebhookUrl != nil && *config.WebhookUrl != "" {
		externalAPIConfig.WebhookURL = *config.WebhookUrl
	}
	if config.WebhookTimeoutSeconds != nil && *config.WebhookTimeoutSeconds > 0 {
		externalAPIConfig.WebhookTimeoutSeconds = int(*config.WebhookTimeoutSeconds)
	}
	if config.WebhookRetryAttempts != nil && *config.WebhookRetryAttempts > 0 {
		externalAPIConfig.WebhookRetryAttempts = int(*config.WebhookRetryAttempts)
	}

	err := s.configService.UpdateExternalAPIConfiguration(ctx, externalAPIConfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// Rooms Configuration methods
func (s *Service) GetRoomsConfiguration(ctx context.Context) ([]dto.RoomConfig, error) {
	rooms, err := s.configService.GetRoomsConfiguration(ctx)
	if err != nil {
		return nil, err
	}

	// Convert types to DTOs
	var dtoRooms []dto.RoomConfig
	for _, room := range rooms {
		dtoRooms = append(dtoRooms, s.convertRoomConfigToDTO(room))
	}
	return dtoRooms, nil
}

func (s *Service) UpdateRoomsConfiguration(ctx context.Context, rooms []dto.RoomConfig) ([]dto.RoomConfig, error) {
	// Convert DTOs to types
	var typeRooms []types.RoomConfig
	for _, room := range rooms {
		typeRooms = append(typeRooms, s.convertDTOToRoomConfig(room))
	}

	err := s.configService.UpdateRoomsConfiguration(ctx, typeRooms)
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

// Card Reader methods
func (s *Service) GetCardReaders(ctx context.Context) ([]dto.CardReaderStatus, error) {
	readers, err := s.configService.GetCardReaders(ctx)
	if err != nil {
		return nil, err
	}

	// Convert types to DTOs
	var dtoReaders []dto.CardReaderStatus
	for _, reader := range readers {
		dtoReaders = append(dtoReaders, s.convertCardReaderStatusToDTO(reader))
	}
	return dtoReaders, nil
}

func (s *Service) RestartCardReader(ctx context.Context, id string) (*dto.RestartResponse, error) {
	result, err := s.configService.RestartCardReader(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert map to DTO
	success, ok := result["success"].(bool)
	if !ok {
		success = false
	}

	message, ok := result["message"].(string)
	if !ok {
		message = "Unknown error"
	}

	return &dto.RestartResponse{
		Success: success,
		Message: message,
	}, nil
}

// Helper conversion methods
func (s *Service) convertSystemConfigurationToDTO(config *types.SystemConfiguration) *dto.SystemConfiguration {
	if config == nil {
		return nil
	}

	// Convert ExternalAPI
	externalAPI := dto.ExternalAPIConfig{
		TimeoutSeconds: int64(config.ExternalAPI.TimeoutSeconds),
		RetryAttempts:  int64(config.ExternalAPI.RetryAttempts),
		Headers:        config.ExternalAPI.Headers,
	}

	// Add optional URLs if they exist
	if config.ExternalAPI.AppointmentServicesURL != "" {
		externalAPI.AppointmentServicesUrl = &config.ExternalAPI.AppointmentServicesURL
	}
	if config.ExternalAPI.GenericServicesURL != "" {
		externalAPI.GenericServicesUrl = &config.ExternalAPI.GenericServicesURL
	}
	if config.ExternalAPI.WebhookURL != "" {
		externalAPI.WebhookUrl = &config.ExternalAPI.WebhookURL
	}
	if config.ExternalAPI.WebhookTimeoutSeconds > 0 {
		timeout := int64(config.ExternalAPI.WebhookTimeoutSeconds)
		externalAPI.WebhookTimeoutSeconds = &timeout
	}
	if config.ExternalAPI.WebhookRetryAttempts > 0 {
		retries := int64(config.ExternalAPI.WebhookRetryAttempts)
		externalAPI.WebhookRetryAttempts = &retries
	}

	// Convert Rooms
	var dtoRooms []dto.RoomConfig
	for _, room := range config.Rooms {
		dtoRooms = append(dtoRooms, s.convertRoomConfigToDTO(room))
	}

	return &dto.SystemConfiguration{
		Id:            &config.ID,
		ExternalAPI:   &externalAPI,
		Rooms:         dtoRooms,
		DefaultRoom:   config.DefaultRoom,
		WebSocketPath: config.WebSocketPath,
		AllowWildcard: config.AllowWildcard,
		CreatedAt:     &config.CreatedAt,
		UpdatedAt:     &config.UpdatedAt,
	}
}

func (s *Service) convertDTOToSystemConfiguration(dtoConfig *dto.SystemConfiguration) *types.SystemConfiguration {
	if dtoConfig == nil {
		return nil
	}

	// Convert ExternalAPI
	externalAPI := types.ExternalAPIConfig{
		TimeoutSeconds: int(dtoConfig.ExternalAPI.TimeoutSeconds),
		RetryAttempts:  int(dtoConfig.ExternalAPI.RetryAttempts),
		Headers:        dtoConfig.ExternalAPI.Headers,
	}

	// Add optional URLs if they exist
	if dtoConfig.ExternalAPI.AppointmentServicesUrl != nil && *dtoConfig.ExternalAPI.AppointmentServicesUrl != "" {
		externalAPI.AppointmentServicesURL = *dtoConfig.ExternalAPI.AppointmentServicesUrl
	}
	if dtoConfig.ExternalAPI.GenericServicesUrl != nil && *dtoConfig.ExternalAPI.GenericServicesUrl != "" {
		externalAPI.GenericServicesURL = *dtoConfig.ExternalAPI.GenericServicesUrl
	}
	if dtoConfig.ExternalAPI.WebhookUrl != nil && *dtoConfig.ExternalAPI.WebhookUrl != "" {
		externalAPI.WebhookURL = *dtoConfig.ExternalAPI.WebhookUrl
	}
	if dtoConfig.ExternalAPI.WebhookTimeoutSeconds != nil && *dtoConfig.ExternalAPI.WebhookTimeoutSeconds > 0 {
		externalAPI.WebhookTimeoutSeconds = int(*dtoConfig.ExternalAPI.WebhookTimeoutSeconds)
	}
	if dtoConfig.ExternalAPI.WebhookRetryAttempts != nil && *dtoConfig.ExternalAPI.WebhookRetryAttempts > 0 {
		externalAPI.WebhookRetryAttempts = int(*dtoConfig.ExternalAPI.WebhookRetryAttempts)
	}

	// Convert Rooms
	var typeRooms []types.RoomConfig
	for _, room := range dtoConfig.Rooms {
		typeRooms = append(typeRooms, s.convertDTOToRoomConfig(room))
	}

	config := &types.SystemConfiguration{
		ExternalAPI:   externalAPI,
		Rooms:         typeRooms,
		DefaultRoom:   dtoConfig.DefaultRoom,
		WebSocketPath: dtoConfig.WebSocketPath,
		AllowWildcard: dtoConfig.AllowWildcard,
	}

	if dtoConfig.Id != nil {
		config.ID = *dtoConfig.Id
	}
	if dtoConfig.CreatedAt != nil {
		config.CreatedAt = *dtoConfig.CreatedAt
	}
	if dtoConfig.UpdatedAt != nil {
		config.UpdatedAt = *dtoConfig.UpdatedAt
	}

	return config
}

func (s *Service) convertRoomConfigToDTO(room types.RoomConfig) dto.RoomConfig {
	// Convert ServicePoints
	var dtoServicePoints []dto.ServicePointConfig
	for _, sp := range room.ServicePoints {
		spConfig := dto.ServicePointConfig{
			Id:   sp.ID,
			Name: sp.Name,
		}
		if sp.Description != "" {
			spConfig.Description = &sp.Description
		}
		if sp.ManagerID != "" {
			spConfig.ManagerId = &sp.ManagerID
		}
		if sp.ManagerName != "" {
			spConfig.ManagerName = &sp.ManagerName
		}
		dtoServicePoints = append(dtoServicePoints, spConfig)
	}

	roomConfig := dto.RoomConfig{
		Id:            room.ID,
		Name:          room.Name,
		ServicePoints: dtoServicePoints,
		IsDefault:     room.IsDefault,
	}

	if room.Description != "" {
		roomConfig.Description = &room.Description
	}

	return roomConfig
}

func (s *Service) convertDTOToRoomConfig(dtoRoom dto.RoomConfig) types.RoomConfig {
	// Convert ServicePoints
	var typeServicePoints []types.ServicePointConfig
	for _, sp := range dtoRoom.ServicePoints {
		spConfig := types.ServicePointConfig{
			ID:   sp.Id,
			Name: sp.Name,
		}
		if sp.Description != nil {
			spConfig.Description = *sp.Description
		}
		if sp.ManagerId != nil {
			spConfig.ManagerID = *sp.ManagerId
		}
		if sp.ManagerName != nil {
			spConfig.ManagerName = *sp.ManagerName
		}
		typeServicePoints = append(typeServicePoints, spConfig)
	}

	return types.RoomConfig{
		ID:            dtoRoom.Id,
		Name:          dtoRoom.Name,
		ServicePoints: typeServicePoints,
		IsDefault:     dtoRoom.IsDefault,
		Description:   getStringValue(dtoRoom.Description),
	}
}

func (s *Service) convertCardReaderStatusToDTO(reader types.CardReaderStatus) dto.CardReaderStatus {
	cardReader := dto.CardReaderStatus{
		Id:     reader.ID,
		Name:   reader.Name,
		Status: reader.Status,
	}

	// Handle optional fields with pointers
	if !reader.LastSeen.IsZero() {
		cardReader.LastSeen = &reader.LastSeen
	}
	if reader.IPAddress != "" {
		cardReader.IpAddress = &reader.IPAddress
	}
	if reader.Version != "" {
		cardReader.Version = &reader.Version
	}
	if reader.LastError != "" {
		cardReader.LastError = &reader.LastError
	}
	if !reader.CreatedAt.IsZero() {
		cardReader.CreatedAt = &reader.CreatedAt
	}
	if !reader.UpdatedAt.IsZero() {
		cardReader.UpdatedAt = &reader.UpdatedAt
	}

	return cardReader
}

// Helper function to get string value from pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
