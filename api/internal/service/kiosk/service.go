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
	"github.com/arfis/waiting-room/internal/service/translation"
	"github.com/arfis/waiting-room/internal/service/webhook"
)

type Service struct {
	queueService       *queue.WaitingQueue
	broadcastFunc      func(string) // Function to broadcast queue updates
	config             *config.Config
	configService      *configService.Service
	webhookService     *webhook.Service
	translationService *translation.DeepLTranslationService
}

func New(queueService *queue.WaitingQueue, broadcastFunc func(string), config *config.Config, configService *configService.Service, webhookService *webhook.Service, translationService *translation.DeepLTranslationService) *Service {
	return &Service{
		queueService:       queueService,
		broadcastFunc:      broadcastFunc,
		config:             config,
		configService:      configService,
		webhookService:     webhookService,
		translationService: translationService,
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
		defaultLang := "en"
		services, err := s.GetUserServices(ctx, cardData.IDNumber, &defaultLang)
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
		defaultLang := "en"
		services, err := s.GetUserServices(ctx, cardData.IDNumber, &defaultLang)
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

func (s *Service) GetUserServices(ctx context.Context, identifier string, language *string) ([]dto.UserService, error) {
	// Default language to English if not provided
	lang := "en"
	if language != nil {
		lang = *language
	}

	// Get external API configuration from cache
	apiConfig, err := s.configService.GetExternalAPIConfig(ctx)
	if err != nil {
		// Fallback to environment variables if cache fails
		externalAPIURL := s.config.GetExternalAPIUserServicesURL()
		timeoutSeconds := s.config.GetExternalAPITimeout()
		log.Printf("Using fallback config - URL: %s, Timeout: %d, Error: %v", externalAPIURL, timeoutSeconds, err)
		return s.makeExternalAPICall(ctx, externalAPIURL, timeoutSeconds, nil, identifier, lang, true)
	}

	// Replace ${identifier} placeholder with actual identifier value
	actualURL := s.replaceIdentifierInURL(apiConfig.AppointmentServicesURL, identifier)

	return s.makeExternalAPICall(ctx, actualURL, apiConfig.TimeoutSeconds, apiConfig.Headers, identifier, lang, true)
}

// GetGenericServices returns generic services available at a service point
func (s *Service) GetGenericServices(ctx context.Context, language *string, servicePointId string) ([]dto.UserService, error) {
	// Default language to English if not provided
	lang := "en"
	if language != nil {
		lang = *language
	}

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

		externalServices, err := s.makeExternalAPICall(ctx, actualURL, apiConfig.TimeoutSeconds, apiConfig.Headers, "", lang, false)
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

	// Apply DeepL translation if configured for admin-created services
	if apiConfig != nil && apiConfig.UseDeepLTranslation != nil && *apiConfig.UseDeepLTranslation {
		log.Printf("DeepL translation is enabled for generic services")

		if s.translationService == nil {
			log.Printf("WARNING: DeepL translation is enabled but translation service is nil")
		} else {
			needsTranslation := (lang != "en")

			log.Printf("Generic services translation decision: needsTranslation=%v, targetLanguage=%s", needsTranslation, lang)

			if needsTranslation {
				log.Printf("Attempting to translate %d generic services to %s", len(services), lang)
				translatedServices, err := s.translateServices(services, lang)
				if err != nil {
					log.Printf("Failed to translate generic services: %v", err)
					// Return original services if translation fails
				} else {
					log.Printf("Successfully translated generic services")
					services = translatedServices
				}
			}
		}
	} else {
		log.Printf("DeepL translation not enabled for generic services")
	}

	log.Printf("Returning %d total generic services", len(services))
	return services, nil
}

// GetAppointmentServices returns appointment-specific services for a user
func (s *Service) GetAppointmentServices(ctx context.Context, identifier string, language *string) ([]dto.UserService, error) {
	// Default language to English if not provided
	lang := "en"
	if language != nil {
		lang = *language
	}

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

	return s.makeExternalAPICall(ctx, actualURL, apiConfig.TimeoutSeconds, apiConfig.Headers, identifier, lang, true)
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
func (s *Service) makeExternalAPICall(ctx context.Context, externalAPIURL string, timeoutSeconds int, headers map[string]string, identifier string, language string, isAppointmentServices bool) ([]dto.UserService, error) {
	// Get external API configuration to check multilingual settings
	apiConfig, err := s.configService.GetExternalAPIConfig(ctx)
	if err != nil {
		log.Printf("Failed to get external API config: %v", err)
		// Continue with basic call if config fails
	}

	// Determine HTTP method
	httpMethod := "GET"
	if apiConfig != nil {
		if isAppointmentServices && apiConfig.AppointmentServicesHttpMethod != nil {
			httpMethod = *apiConfig.AppointmentServicesHttpMethod
		} else if !isAppointmentServices && apiConfig.GenericServicesHttpMethod != nil {
			httpMethod = *apiConfig.GenericServicesHttpMethod
		}
	}

	// Create HTTP client with configured timeout
	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, httpMethod, externalAPIURL, nil)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to create request", 500, nil)
	}

	// Add custom headers if configured
	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// Add query parameters
	q := req.URL.Query()
	if identifier != "" {
		q.Add("identifier", identifier)
	}

	// Add language parameter based on language handling configuration
	var languageHandling *string
	var languageHeader *string

	if isAppointmentServices {
		if apiConfig != nil && apiConfig.AppointmentServicesLanguageHandling != nil {
			languageHandling = apiConfig.AppointmentServicesLanguageHandling
			languageHeader = apiConfig.AppointmentServicesLanguageHeader
		}
	} else {
		if apiConfig != nil && apiConfig.GenericServicesLanguageHandling != nil {
			languageHandling = apiConfig.GenericServicesLanguageHandling
			languageHeader = apiConfig.GenericServicesLanguageHeader
		}
	}

	if languageHandling != nil {
		switch *languageHandling {
		case "query_param":
			// Convert language code to uppercase for API
			langCode := strings.ToUpper(language)
			q.Add("lang", langCode)
			log.Printf("Added language parameter: lang=%s", langCode)
		case "header":
			// Add language to HTTP header
			headerName := "Accept-Language"
			if languageHeader != nil && *languageHeader != "" {
				headerName = *languageHeader
			}
			req.Header.Set(headerName, language)
			log.Printf("Added language header: %s=%s", headerName, language)
		case "none":
			// No language handling - will rely on DeepL translation
			log.Printf("No language handling configured - will use DeepL translation if enabled")
		}
	}

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

	// Apply DeepL translation if configured and needed
	if apiConfig != nil && apiConfig.UseDeepLTranslation != nil && *apiConfig.UseDeepLTranslation {
		log.Printf("DeepL translation is enabled in config")

		if s.translationService == nil {
			log.Printf("WARNING: DeepL translation is enabled but translation service is nil")
		} else {
			// Always translate if language is not English, regardless of language handling configuration
			needsTranslation := (language != "en")

			log.Printf("Translation decision: needsTranslation=%v, targetLanguage=%s", needsTranslation, language)

			if needsTranslation {
				log.Printf("Attempting to translate %d services to %s", len(services), language)
				translatedServices, err := s.translateServices(services, language)
				if err != nil {
					log.Printf("Failed to translate services: %v", err)
					// Return original services if translation fails
				} else {
					log.Printf("Successfully translated services %s", translatedServices)
					services = translatedServices
				}
			}
		}
	} else {
		log.Printf("DeepL translation not enabled or not configured properly")
	}

	return services, nil
}

// translateServices translates service names and descriptions using DeepL
func (s *Service) translateServices(services []dto.UserService, targetLanguage string) ([]dto.UserService, error) {
	if s.translationService == nil || !s.translationService.IsConfigured() {
		return services, fmt.Errorf("DeepL translation service not configured")
	}

	// Assume source language is English if not specified
	sourceLanguage := "en"

	// Skip translation if target language is English
	if targetLanguage == "en" {
		return services, nil
	}

	translatedServices := make([]dto.UserService, len(services))

	for i, service := range services {
		translatedService := service

		// Translate service name
		if service.ServiceName != "" {
			translatedName, err := s.translationService.Translate(service.ServiceName, sourceLanguage, targetLanguage)
			if err != nil {
				log.Printf("Failed to translate service name '%s': %v", service.ServiceName, err)
				// Keep original name if translation fails
			} else {
				translatedService.ServiceName = translatedName
			}
		}

		translatedServices[i] = translatedService
	}

	return translatedServices, nil
}
