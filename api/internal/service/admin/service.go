package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/arfis/waiting-room/internal/data/dto"
	"github.com/arfis/waiting-room/internal/service/config"
	tenantService "github.com/arfis/waiting-room/internal/service/tenant"
	"github.com/arfis/waiting-room/internal/service/translation"
	"github.com/arfis/waiting-room/internal/types"
)

type Service struct {
	configService      *config.Service
	translationService *translation.DeepLTranslationService
	tenantService      *tenantService.Service
}

func NewService(configService *config.Service, translationService *translation.DeepLTranslationService, tenantService *tenantService.Service) *Service {
	return &Service{
		configService:      configService,
		translationService: translationService,
		tenantService:      tenantService,
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
	if config.AppointmentServicesHttpMethod != nil {
		externalAPIConfig.AppointmentServicesHttpMethod = config.AppointmentServicesHttpMethod
	}
	if config.GenericServicesURL != "" {
		externalAPIConfig.GenericServicesUrl = &config.GenericServicesURL
	}
	if config.GenericServicesHttpMethod != nil {
		externalAPIConfig.GenericServicesHttpMethod = config.GenericServicesHttpMethod
	}
	// Note: GenericServicesPostBody removed from DTO

	// Convert GenericServices from types to DTO
	if len(config.GenericServices) > 0 {
		genericServices := make([]dto.GenericService, len(config.GenericServices))
		for i, service := range config.GenericServices {
			var duration *int64
			if service.Duration > 0 {
				d := int64(service.Duration)
				duration = &d
			}
			var description *string
			if service.Description != "" {
				description = &service.Description
			}
			genericServices[i] = dto.GenericService{
				Id:          service.ID,
				Name:        service.Name,
				Description: description,
				Duration:    duration,
				Enabled:     service.Enabled,
			}
		}
		externalAPIConfig.GenericServices = genericServices
	}

	if config.WebhookURL != "" {
		externalAPIConfig.WebhookUrl = &config.WebhookURL
	}
	if config.WebhookHttpMethod != nil {
		externalAPIConfig.WebhookHttpMethod = config.WebhookHttpMethod
	}
	if config.WebhookTimeoutSeconds > 0 {
		timeout := int64(config.WebhookTimeoutSeconds)
		externalAPIConfig.WebhookTimeoutSeconds = &timeout
	}
	if config.WebhookRetryAttempts > 0 {
		retries := int64(config.WebhookRetryAttempts)
		externalAPIConfig.WebhookRetryAttempts = &retries
	}

	// Add multilingual configuration
	if config.MultilingualSupport != nil {
		externalAPIConfig.MultilingualSupport = config.MultilingualSupport
	}
	if len(config.SupportedLanguages) > 0 {
		externalAPIConfig.SupportedLanguages = config.SupportedLanguages
	}
	if config.UseDeepLTranslation != nil {
		externalAPIConfig.UseDeepLTranslation = config.UseDeepLTranslation
	}
	if config.AppointmentServicesLanguageHandling != nil {
		externalAPIConfig.AppointmentServicesLanguageHandling = config.AppointmentServicesLanguageHandling
	}
	if config.AppointmentServicesLanguageHeader != nil {
		externalAPIConfig.AppointmentServicesLanguageHeader = config.AppointmentServicesLanguageHeader
	}
	if config.GenericServicesLanguageHandling != nil {
		externalAPIConfig.GenericServicesLanguageHandling = config.GenericServicesLanguageHandling
	}
	if config.GenericServicesLanguageHeader != nil {
		externalAPIConfig.GenericServicesLanguageHeader = config.GenericServicesLanguageHeader
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
	if config.AppointmentServicesHttpMethod != nil {
		externalAPIConfig.AppointmentServicesHttpMethod = config.AppointmentServicesHttpMethod
	}
	if config.GenericServicesUrl != nil && *config.GenericServicesUrl != "" {
		externalAPIConfig.GenericServicesURL = *config.GenericServicesUrl
	}
	if config.GenericServicesHttpMethod != nil {
		externalAPIConfig.GenericServicesHttpMethod = config.GenericServicesHttpMethod
	}
	// Note: GenericServicesPostBody removed from DTO

	// Convert GenericServices from DTO to types
	if len(config.GenericServices) > 0 {
		genericServices := make([]types.GenericService, len(config.GenericServices))
		for i, service := range config.GenericServices {
			var duration int
			if service.Duration != nil {
				duration = int(*service.Duration)
			}
			var description string
			if service.Description != nil {
				description = *service.Description
			}
			genericServices[i] = types.GenericService{
				ID:          service.Id,
				Name:        service.Name,
				Description: description,
				Duration:    duration,
				Enabled:     service.Enabled,
			}
		}
		externalAPIConfig.GenericServices = genericServices
	}

	if config.WebhookUrl != nil && *config.WebhookUrl != "" {
		externalAPIConfig.WebhookURL = *config.WebhookUrl
	}
	if config.WebhookHttpMethod != nil {
		externalAPIConfig.WebhookHttpMethod = config.WebhookHttpMethod
	}
	if config.WebhookTimeoutSeconds != nil && *config.WebhookTimeoutSeconds > 0 {
		externalAPIConfig.WebhookTimeoutSeconds = int(*config.WebhookTimeoutSeconds)
	}
	if config.WebhookRetryAttempts != nil && *config.WebhookRetryAttempts > 0 {
		externalAPIConfig.WebhookRetryAttempts = int(*config.WebhookRetryAttempts)
	}

	// Add multilingual configuration
	if config.MultilingualSupport != nil {
		externalAPIConfig.MultilingualSupport = config.MultilingualSupport
	}
	if len(config.SupportedLanguages) > 0 {
		externalAPIConfig.SupportedLanguages = config.SupportedLanguages
	}
	if config.UseDeepLTranslation != nil {
		externalAPIConfig.UseDeepLTranslation = config.UseDeepLTranslation
	}
	if config.AppointmentServicesLanguageHandling != nil {
		externalAPIConfig.AppointmentServicesLanguageHandling = config.AppointmentServicesLanguageHandling
	}
	if config.AppointmentServicesLanguageHeader != nil {
		externalAPIConfig.AppointmentServicesLanguageHeader = config.AppointmentServicesLanguageHeader
	}
	if config.GenericServicesLanguageHandling != nil {
		externalAPIConfig.GenericServicesLanguageHandling = config.GenericServicesLanguageHandling
	}
	if config.GenericServicesLanguageHeader != nil {
		externalAPIConfig.GenericServicesLanguageHeader = config.GenericServicesLanguageHeader
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

// Helper function to get int64 value from map
func getInt64Value(m map[string]interface{}, key string) *int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			i := int64(v)
			return &i
		case int64:
			return &v
		}
	}
	return nil
}

// Helper function to get string value from map
func getStringValueFromMap(m map[string]interface{}, key string) *string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return &str
		}
	}
	return nil
}

// GetTranslationCacheStats returns statistics about the translation cache
func (s *Service) GetTranslationCacheStats(ctx context.Context) (*dto.TranslationCacheStats, error) {
	if s.translationService == nil {
		return &dto.TranslationCacheStats{}, fmt.Errorf("translation service not configured")
	}
	stats := s.translationService.GetCacheStats()

	// Convert map to DTO
	result := &dto.TranslationCacheStats{
		Cache_size:      getInt64Value(stats, "cache_size"),
		Max_cache_size:  getInt64Value(stats, "max_cache_size"),
		Hits:            getInt64Value(stats, "hits"),
		Misses:          getInt64Value(stats, "misses"),
		Total_requests:  getInt64Value(stats, "total_requests"),
		Hit_rate:        getStringValueFromMap(stats, "hit_rate"),
		Api_calls_saved: getInt64Value(stats, "api_calls_saved"),
		Expiration_time: getStringValueFromMap(stats, "expiration_time"),
	}
	return result, nil
}

// ClearTranslationCache clears the translation cache
func (s *Service) ClearTranslationCache(ctx context.Context) (*dto.CacheClearResponse, error) {
	if s.translationService == nil {
		return nil, fmt.Errorf("translation service not configured")
	}
	s.translationService.ClearCache()
	msg := "Cache cleared successfully"
	return &dto.CacheClearResponse{Message: &msg}, nil
}

// Tenant methods
func (s *Service) GetAllTenants(ctx context.Context) ([]dto.Tenant, error) {
	return s.tenantService.GetAllTenants(ctx)
}

func (s *Service) GetTenant(ctx context.Context, id string) (*dto.Tenant, error) {
	return s.tenantService.GetTenant(ctx, id)
}

func (s *Service) CreateTenant(ctx context.Context, req *dto.CreateTenantRequest) (*dto.Tenant, error) {
	// Convert CreateTenantRequest to Tenant DTO
	tenantDTO := &dto.Tenant{
		BuildingId:  req.BuildingId,
		SectionId:   req.SectionId,
		Name:        req.Name,
		Description: req.Description,
	}
	return s.tenantService.CreateTenant(ctx, tenantDTO)
}

func (s *Service) UpdateTenant(ctx context.Context, req *dto.CreateTenantRequest) (*dto.Tenant, error) {
	// Convert CreateTenantRequest to Tenant DTO
	tenantDTO := &dto.Tenant{
		BuildingId:  req.BuildingId,
		SectionId:   req.SectionId,
		Name:        req.Name,
		Description: req.Description,
	}
	return s.tenantService.UpdateTenant(ctx, tenantDTO)
}

func (s *Service) DeleteTenant(ctx context.Context, id string) error {
	return s.tenantService.DeleteTenant(ctx, id)
}
