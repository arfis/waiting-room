package kiosk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arfis/waiting-room/internal/config"
	"github.com/arfis/waiting-room/internal/data/dto"
	ngErrors "github.com/arfis/waiting-room/internal/errors"
	"github.com/arfis/waiting-room/internal/queue"
	configService "github.com/arfis/waiting-room/internal/service/config"
	"github.com/arfis/waiting-room/internal/service/webhook"
)

type Service struct {
	queueService   *queue.WaitingQueue
	broadcastFunc  func(string) // Function to broadcast queue updates
	config         *config.Config
	configService  *configService.Service
	webhookService *webhook.Service
}

func New(queueService *queue.WaitingQueue, broadcastFunc func(string), config *config.Config, configService *configService.Service, webhookService *webhook.Service) *Service {
	return &Service{
		queueService:   queueService,
		broadcastFunc:  broadcastFunc,
		config:         config,
		configService:  configService,
		webhookService: webhookService,
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

	// Use service duration from request, fallback to 5 minutes if not provided
	approximateDurationMinutes := req.GetServiceDuration()
	if approximateDurationMinutes == 0 {
		approximateDurationMinutes = 5 // Default fallback
	}

	// Get service name if service ID is provided
	serviceName := ""
	if req.ServiceId != nil && *req.ServiceId != "" {
		// Get service name by calling the external API with the same identifier
		services, err := s.GetUserServices(ctx, cardData.IDNumber)
		if err == nil {
			for _, service := range services {
				if service.Id == *req.ServiceId {
					serviceName = service.ServiceName
					break
				}
			}
		}
	}

	// Create queue entry using the existing queue service
	entry, err := s.queueService.CreateEntry(roomId, cardData, approximateDurationMinutes, serviceName)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to create queue entry", 500, nil)
	}

	// Generate QR URL
	qrUrl := "http://localhost:4204/q/" + entry.QRToken

	// Broadcast queue update
	if s.broadcastFunc != nil {
		s.broadcastFunc(roomId)
	}

	// Send webhook notification for service selected (if service was selected)
	if s.webhookService != nil && req.ServiceId != nil && *req.ServiceId != "" {
		go func() {
			if err := s.webhookService.SendServiceSelectedWebhook(ctx, entry.ID, *req.ServiceId, roomId, "", cardData.IDNumber); err != nil {
				log.Printf("Failed to send webhook notification for service selected: %v", err)
			}
		}()
	}

	// Return the join result
	result := &dto.JoinResult{
		EntryID:      entry.ID,
		TicketNumber: entry.TicketNumber,
		QrUrl:        qrUrl,
	}

	// Add service duration if provided
	if approximateDurationMinutes > 0 {
		result.ServiceDuration = &approximateDurationMinutes
	}

	// Add service name if service ID is provided
	if req.ServiceId != nil && *req.ServiceId != "" {
		// Get service name by calling the external API with the same identifier
		// This is a bit inefficient, but ensures we get the current service name
		services, err := s.GetUserServices(ctx, cardData.IDNumber)
		if err == nil {
			for _, service := range services {
				if service.Id == *req.ServiceId {
					result.ServiceName = &service.ServiceName
					break
				}
			}
		}
		// Fallback if we can't get the service name
		if result.ServiceName == nil {
			serviceName := "Selected Service"
			result.ServiceName = &serviceName
		}
	}

	return result, nil
}

func (s *Service) GetUserServices(ctx context.Context, identifier string) ([]dto.UserService, error) {
	// Get external API configuration from cache
	apiConfig, err := s.configService.GetExternalAPIConfig(ctx)
	if err != nil {
		// Fallback to environment variables if cache fails
		externalAPIURL := s.config.GetExternalAPIUserServicesURL()
		timeoutSeconds := s.config.GetExternalAPITimeout()
		log.Printf("Using fallback config - URL: %s, Timeout: %d, Error: %v", externalAPIURL, timeoutSeconds, err)
		return s.makeExternalAPICall(ctx, externalAPIURL, timeoutSeconds, nil, identifier)
	}

	// Replace ${identifier} placeholder with actual identifier value
	actualURL := s.replaceIdentifierInURL(apiConfig.AppointmentServicesURL, identifier)

	return s.makeExternalAPICall(ctx, actualURL, apiConfig.TimeoutSeconds, apiConfig.Headers, identifier)
}

// GetGenericServices returns generic services available at a service point
func (s *Service) GetGenericServices(ctx context.Context, servicePointId string) ([]dto.UserService, error) {
	// Get external API configuration from cache
	apiConfig, err := s.configService.GetExternalAPIConfig(ctx)
	if err != nil {
		log.Printf("Failed to get external API config for generic services: %v", err)
		return []dto.UserService{}, nil // Return empty list if config fails
	}

	var services []dto.UserService

	// First, try to get admin-created generic services
	if len(apiConfig.GenericServices) > 0 {
		log.Printf("Found %d admin-created generic services", len(apiConfig.GenericServices))
		for _, service := range apiConfig.GenericServices {
			if service.Enabled {
				userService := dto.UserService{
					Id:          service.ID,
					ServiceName: service.Name,
					Duration:    int64(service.Duration),
				}
				services = append(services, userService)
			}
		}
	}

	// If external URL is configured, also fetch from external API
	if apiConfig.GenericServicesURL != "" {
		log.Printf("Fetching generic services from external URL: %s", apiConfig.GenericServicesURL)
		// Replace ${servicePointId} placeholder with actual service point ID
		actualURL := s.replaceServicePointIdInURL(apiConfig.GenericServicesURL, servicePointId)

		externalServices, err := s.makeExternalAPICall(ctx, actualURL, apiConfig.TimeoutSeconds, apiConfig.Headers, "")
		if err != nil {
			log.Printf("Failed to fetch external generic services: %v", err)
		} else {
			// Append external services to admin-created services
			services = append(services, externalServices...)
		}
	}

	// If no admin-created services and no external URL, return empty list
	if len(apiConfig.GenericServices) == 0 && apiConfig.GenericServicesURL == "" {
		log.Printf("No generic services configured (neither admin-created nor external URL)")
		return []dto.UserService{}, nil
	}

	log.Printf("Returning %d total generic services", len(services))
	return services, nil
}

// GetAppointmentServices returns appointment-specific services for a user
func (s *Service) GetAppointmentServices(ctx context.Context, identifier string) ([]dto.UserService, error) {
	// Get external API configuration from cache
	apiConfig, err := s.configService.GetExternalAPIConfig(ctx)
	if err != nil {
		log.Printf("Failed to get external API config for appointment services: %v", err)
		return []dto.UserService{}, nil // Return empty list if config fails
	}

	// Check if appointment services URL is configured
	if apiConfig.AppointmentServicesURL == "" {
		log.Printf("Appointment services URL not configured")
		return []dto.UserService{}, nil // Return empty list if not configured
	}

	// Replace ${identifier} placeholder with actual identifier value
	actualURL := s.replaceIdentifierInURL(apiConfig.AppointmentServicesURL, identifier)

	return s.makeExternalAPICall(ctx, actualURL, apiConfig.TimeoutSeconds, apiConfig.Headers, identifier)
}

// replaceIdentifierInURL replaces ${identifier} placeholder with the actual identifier value
func (s *Service) replaceIdentifierInURL(urlTemplate, identifier string) string {
	return strings.ReplaceAll(urlTemplate, "${identifier}", identifier)
}

// replaceServicePointIdInURL replaces ${servicePointId} placeholder with the actual service point ID
func (s *Service) replaceServicePointIdInURL(urlTemplate, servicePointId string) string {
	return strings.ReplaceAll(urlTemplate, "${servicePointId}", servicePointId)
}

// makeExternalAPICall makes the actual HTTP call to the external API
func (s *Service) makeExternalAPICall(ctx context.Context, externalAPIURL string, timeoutSeconds int, headers map[string]string, identifier string) ([]dto.UserService, error) {
	// Create HTTP client with configured timeout
	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", externalAPIURL, nil)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to create request", 500, nil)
	}

	// Add custom headers if configured
	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// Add query parameter
	q := req.URL.Query()
	q.Add("identifier", identifier)
	req.URL.RawQuery = q.Encode()

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, fmt.Sprintf("failed to call external API: %s", externalAPIURL), 500, nil)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, fmt.Sprintf("external API returned status %d", resp.StatusCode), 500, nil)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to read response body", 500, nil)
	}

	// Parse JSON response
	var services []dto.UserService
	if err := json.Unmarshal(body, &services); err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to parse response", 500, nil)
	}

	return services, nil
}
