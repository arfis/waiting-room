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
	"github.com/arfis/waiting-room/internal/service"
	configService "github.com/arfis/waiting-room/internal/service/config"
	"github.com/arfis/waiting-room/internal/service/translation"
	"github.com/arfis/waiting-room/internal/service/webhook"
)

type Service struct {
	queueService       *queue.WaitingQueue
	broadcastFunc      func(string, string) // Function to broadcast queue updates (roomId, tenantID)
	config             *config.Config
	configService      *configService.Service
	webhookService     *webhook.Service
	translationService *translation.DeepLTranslationService
}

func New(queueService *queue.WaitingQueue, broadcastFunc func(string, string), config *config.Config, configService *configService.Service, webhookService *webhook.Service, translationService *translation.DeepLTranslationService) *Service {
	return &Service{
		queueService:       queueService,
		broadcastFunc:      broadcastFunc,
		config:             config,
		configService:      configService,
		webhookService:     webhookService,
		translationService: translationService,
	}
}

func (s *Service) SetBroadcastFunc(f func(string, string)) {
	s.broadcastFunc = f
}

func (s *Service) SwipeCard(ctx context.Context, roomId string, req *dto.SwipeRequest) (*dto.JoinResult, error) {
	// Create CardData from the raw card data
	cardData := queue.CardData{
		IDNumber: *req.IdCardRaw,
		Source:   "card-reader",
	}

	// Use service duration from request, convert from minutes to seconds
	// Fallback to 5 minutes (300 seconds) if not provided
	approximateDurationSeconds := req.GetServiceDuration() * 60 // Convert minutes to seconds
	if approximateDurationSeconds == 0 {
		approximateDurationSeconds = 300 // Default fallback: 5 minutes = 300 seconds
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

	// Create queue entry using the existing queue service (pass context for tenant info)
	entry, err := s.queueService.CreateEntry(ctx, roomId, cardData, approximateDurationSeconds, serviceName)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to create queue entry", 500, nil)
	}

	// Generate QR URL
	qrUrl := "http://localhost:4204/q/" + entry.QRToken

	// Broadcast queue update - only to the tenant that changed
	// Extract tenant ID from context (format: "buildingId:sectionId")
	if s.broadcastFunc != nil {
		tenantID := service.GetTenantID(ctx)
		log.Printf("[KioskService] Broadcasting queue update for room %s, tenantID: '%s' (length: %d)", roomId, tenantID, len(tenantID))
		if tenantID == "" {
			log.Printf("[KioskService] WARNING: tenantID is empty, broadcasting to all clients")
		}
		s.broadcastFunc(roomId, tenantID)
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

	// Add service duration if provided (convert back to minutes for API response)
	if approximateDurationSeconds > 0 {
		durationMinutes := approximateDurationSeconds / 60
		result.ServiceDuration = &durationMinutes
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
		if externalAPIURL == "" {
			log.Printf("No external API URL configured for user services (fallback)")
			return []dto.UserService{}, nil // Return empty list if not configured
		}
		timeoutSeconds := s.config.GetExternalAPITimeout()
		log.Printf("Using fallback config - URL: %s, Timeout: %d, Error: %v", externalAPIURL, timeoutSeconds, err)
		return s.makeExternalAPICall(ctx, externalAPIURL, timeoutSeconds, nil, identifier, lang, true, "")
	}

	// Check if config is nil or appointment services URL is not configured
	if apiConfig == nil || apiConfig.AppointmentServicesURL == "" {
		log.Printf("Appointment services URL not configured for user services")
		return []dto.UserService{}, nil // Return empty list if not configured
	}

	// Replace ${identifier} placeholder with actual identifier value
	actualURL := s.replaceIdentifierInURL(apiConfig.AppointmentServicesURL, identifier)

	timeoutSeconds := apiConfig.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = 10 // Default timeout
	}

	return s.makeExternalAPICall(ctx, actualURL, timeoutSeconds, apiConfig.Headers, identifier, lang, true, "")
}

// GetGenericServices returns generic services available
func (s *Service) GetGenericServices(ctx context.Context, language *string) ([]dto.UserService, error) {
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

	// If no config found, return empty list
	if apiConfig == nil {
		log.Printf("No external API config found for generic services")
		return []dto.UserService{}, nil
	}

	var services []dto.UserService
	var adminCreatedServices []dto.UserService

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
				adminCreatedServices = append(adminCreatedServices, userService)
			}
		}
		log.Printf("Added %d enabled admin-created services", len(adminCreatedServices))
	}

	// If external URL is configured, also fetch from external API
	if apiConfig.GenericServicesURL != "" {
		log.Printf("Fetching generic services from external URL: %s", apiConfig.GenericServicesURL)

		var actualURL string
		var postBody string

		if apiConfig.GenericServicesHttpMethod != nil && *apiConfig.GenericServicesHttpMethod == "POST" {
			// For POST requests, use the URL as-is and prepare the body
			actualURL = apiConfig.GenericServicesURL
			if apiConfig.GenericServicesPostBody != "" {
				// Replace placeholders in POST body (no servicePointId needed)
				postBody = strings.ReplaceAll(apiConfig.GenericServicesPostBody, "${language}", lang)
			}
		} else {
			// For GET requests, use the URL as-is (no servicePointId replacement needed)
			actualURL = apiConfig.GenericServicesURL
		}

		externalServices, err := s.makeExternalAPICall(ctx, actualURL, apiConfig.TimeoutSeconds, apiConfig.Headers, "", lang, false, postBody)
		if err != nil {
			log.Printf("Failed to fetch external generic services: %v", err)
		} else {
			// Append external services to admin-created services
			log.Printf("Fetched %d external generic services, appending to %d admin-created services", len(externalServices), len(services))
			services = append(services, externalServices...)
			log.Printf("Total services before translation: %d (admin-created + external)", len(services))
		}
	}

	// If no admin-created services and no external URL, return empty list
	if len(apiConfig.GenericServices) == 0 && apiConfig.GenericServicesURL == "" {
		log.Printf("No generic services configured (neither admin-created nor external URL)")
		return []dto.UserService{}, nil
	}

	// Apply DeepL translation if configured for all generic services (both admin-created and external)
	if apiConfig != nil && apiConfig.UseDeepLTranslation != nil && *apiConfig.UseDeepLTranslation {
		log.Printf("DeepL translation is enabled for generic services (admin-created + external)")

		if s.translationService == nil {
			log.Printf("WARNING: DeepL translation is enabled but translation service is nil")
		} else {
			// Always attempt translation if we have external services (they might be in Slovak)
			// or if target language is not English
			needsTranslation := true // Always try to translate generic services

			log.Printf("Generic services translation decision: needsTranslation=%v, targetLanguage=%s", needsTranslation, lang)

			if needsTranslation {
				if len(services) == 0 {
					log.Printf("No services to translate (services array is empty)")
				} else {
					// Separate admin-created and external services to translate them with different source languages
					// We track admin-created services separately when building the list
					externalServices := make([]dto.UserService, 0)

					// Find external services by comparing with admin-created list
					for _, service := range services {
						isAdminCreated := false
						for _, adminService := range adminCreatedServices {
							if service.Id == adminService.Id && service.ServiceName == adminService.ServiceName {
								isAdminCreated = true
								break
							}
						}
						if !isAdminCreated {
							externalServices = append(externalServices, service)
						}
					}

					log.Printf("Services breakdown: %d admin-created (assumed EN), %d external (source varies)", len(adminCreatedServices), len(externalServices))

					// Determine source language for external services based on language handling configuration
					externalSourceLanguage := "en" // Default
					if apiConfig.GenericServicesLanguageHandling != nil {
						if *apiConfig.GenericServicesLanguageHandling == "none" {
							// When language handling is "none", external API returns in its default language (SK)
							externalSourceLanguage = "sk"
							log.Printf("Language handling is 'none' - external services are in Slovak (sk)")
						} else {
							// If language handling is query_param or header, API might return in requested language
							// In this case, check if the requested language matches target (no translation needed)
							externalSourceLanguage = "en" // Assuming API can return in requested language or default EN
							log.Printf("Language handling is '%s' - external services may already be in target language", *apiConfig.GenericServicesLanguageHandling)
						}
					} else {
						// No language handling config - assume external services are in their default language (SK)
						externalSourceLanguage = "sk"
						log.Printf("No language handling config - assuming external services are in Slovak (sk)")
					}

					// Use tracked admin-created services
					adminServices := adminCreatedServices

					var allTranslatedServices []dto.UserService

					// Translate admin-created services from English (only if target is not English)
					if len(adminServices) > 0 {
						if lang != "en" {
							log.Printf("Translating %d admin-created services from en to %s", len(adminServices), lang)
							translatedAdmin, err := s.translateServices(adminServices, "en", lang)
							if err != nil {
								log.Printf("Failed to translate admin-created services: %v (keeping original)", err)
								allTranslatedServices = append(allTranslatedServices, adminServices...)
							} else {
								allTranslatedServices = append(allTranslatedServices, translatedAdmin...)
							}
						} else {
							log.Printf("Target language is English - keeping %d admin-created services as-is", len(adminServices))
							allTranslatedServices = append(allTranslatedServices, adminServices...)
						}
					}

					// Translate external services from their source language
					if len(externalServices) > 0 {
						// Only translate if source and target are different
						if externalSourceLanguage != lang {
							log.Printf("Translating %d external services from %s to %s", len(externalServices), externalSourceLanguage, lang)
							translatedExternal, err := s.translateServices(externalServices, externalSourceLanguage, lang)
							if err != nil {
								log.Printf("Failed to translate external services: %v (keeping original)", err)
								allTranslatedServices = append(allTranslatedServices, externalServices...)
							} else {
								allTranslatedServices = append(allTranslatedServices, translatedExternal...)
							}
						} else {
							log.Printf("External services source language (%s) matches target (%s) - keeping as-is", externalSourceLanguage, lang)
							allTranslatedServices = append(allTranslatedServices, externalServices...)
						}
					}

					log.Printf("Translation complete: %d total services processed (admin-created + external)", len(allTranslatedServices))
					services = allTranslatedServices
				}
			} else {
				log.Printf("Translation not needed - language is English (en)")
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

	// Check if config is nil or appointment services URL is not configured
	if apiConfig == nil || apiConfig.AppointmentServicesURL == "" {
		log.Printf("Appointment services URL not configured")
		return []dto.UserService{}, nil // Return empty list if not configured
	}

	// Replace ${identifier} placeholder with actual identifier value
	actualURL := s.replaceIdentifierInURL(apiConfig.AppointmentServicesURL, identifier)

	timeoutSeconds := apiConfig.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = 10 // Default timeout
	}

	return s.makeExternalAPICall(ctx, actualURL, timeoutSeconds, apiConfig.Headers, identifier, lang, true, "")
}

// replaceIdentifierInURL replaces ${identifier} placeholder with the actual identifier value
func (s *Service) replaceIdentifierInURL(urlTemplate, identifier string) string {
	return strings.ReplaceAll(urlTemplate, "${identifier}", identifier)
}

// replaceServicePointIdInURL replaces ${servicePointId} placeholder with the actual service point ID
func (s *Service) replaceServicePointIdInURL(urlTemplate, servicePointId string) string {
	return strings.ReplaceAll(urlTemplate, "${servicePointId}", servicePointId)
}

// GetDefaultServicePoint returns the default service point for a room
func (s *Service) GetDefaultServicePoint(ctx context.Context, roomId string) (*string, error) {
	// Get the default service point from config
	servicePointId := s.config.GetDefaultServicePoint(roomId)
	if servicePointId == "" {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "no service points configured for room", 500, nil)
	}
	return &servicePointId, nil
}

// makeExternalAPICall makes the actual HTTP call to the external API
func (s *Service) makeExternalAPICall(ctx context.Context, externalAPIURL string, timeoutSeconds int, headers map[string]string, identifier string, language string, isAppointmentServices bool, postBody string) ([]dto.UserService, error) {
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

	// Ensure timeout is valid (default to 10 seconds if 0)
	if timeoutSeconds == 0 {
		timeoutSeconds = 10
	}

	// Create HTTP client with configured timeout
	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	// Create request with body if POST method
	var bodyReader io.Reader
	if httpMethod == "POST" && postBody != "" {
		bodyReader = strings.NewReader(postBody)
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, externalAPIURL, bodyReader)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to create request", 500, nil)
	}

	// Set content type for POST requests
	if httpMethod == "POST" && postBody != "" {
		req.Header.Set("Content-Type", "application/json")
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
		log.Printf("Failed to call external API %s: %v", externalAPIURL, err)
		// Return empty list instead of error to allow proceeding without services
		return []dto.UserService{}, nil
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("External API %s returned status %d", externalAPIURL, resp.StatusCode)
		// Return empty list instead of error to allow proceeding without services
		return []dto.UserService{}, nil
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body from external API %s: %v", externalAPIURL, err)
		// Return empty list instead of error to allow proceeding without services
		return []dto.UserService{}, nil
	}

	log.Printf("External API response body: %s", string(body))

	// Try to parse as external API format first (with code, id as int64, name)
	type ExternalService struct {
		Code     string `json:"code"`
		Duration int64  `json:"duration"`
		Id       int64  `json:"id"`
		Name     string `json:"name"`
	}

	var externalServices []ExternalService
	var services []dto.UserService

	// Try parsing as external API format first
	if err := json.Unmarshal(body, &externalServices); err == nil {
		if len(externalServices) > 0 {
			// Transform external format to UserService format
			services = make([]dto.UserService, len(externalServices))
			for i, ext := range externalServices {
				services[i] = dto.UserService{
					Id:          fmt.Sprintf("%d", ext.Id), // Convert int64 id to string
					ServiceName: ext.Name,
					Duration:    ext.Duration,
				}
			}
			log.Printf("Successfully parsed %d services from external API format", len(services))
		} else {
			log.Printf("External API format parsed successfully but returned empty array")
			services = []dto.UserService{}
		}
	} else {
		// Fallback: try parsing as direct UserService format
		externalErr := err
		if parseErr := json.Unmarshal(body, &services); parseErr != nil {
			log.Printf("Failed to parse response in both formats. External format error: %v, Direct format error: %v", externalErr, parseErr)
			log.Printf("Response body was: %s", string(body))
			// Return empty list instead of error to allow proceeding without services
			return []dto.UserService{}, nil
		}
		log.Printf("Successfully parsed %d services from direct UserService format", len(services))
	}

	// Apply DeepL translation if configured and needed
	// Note: For generic services, translation will be handled in GetGenericServices after merging with admin-created services
	// For appointment services, translation happens here since they are returned directly
	if apiConfig != nil && apiConfig.UseDeepLTranslation != nil && *apiConfig.UseDeepLTranslation {
		// Only translate here for appointment services (isAppointmentServices = true)
		// Generic services will be translated later in GetGenericServices after merging
		if isAppointmentServices {
			log.Printf("DeepL translation is enabled in config for appointment services")

			if s.translationService == nil {
				log.Printf("WARNING: DeepL translation is enabled but translation service is nil")
			} else {
				// Always attempt translation for appointment services
				needsTranslation := true

				log.Printf("Translation decision for appointment services: needsTranslation=%v, targetLanguage=%s", needsTranslation, language)

				if needsTranslation {
					// Appointment services are typically in English, so use "en" as source
					sourceLanguage := "en"
					log.Printf("Attempting to translate %d appointment services from %s to %s", len(services), sourceLanguage, language)
					translatedServices, err := s.translateServices(services, sourceLanguage, language)
					if err != nil {
						log.Printf("Failed to translate appointment services: %v", err)
						// Return original services if translation fails
					} else {
						log.Printf("Successfully translated appointment services")
						services = translatedServices
					}
				}
			}
		} else {
			log.Printf("Skipping translation in makeExternalAPICall for generic services - will be handled in GetGenericServices after merging")
		}
	} else {
		log.Printf("DeepL translation not enabled or not configured properly")
	}

	return services, nil
}

// translateServices translates service names and descriptions using DeepL
func (s *Service) translateServices(services []dto.UserService, sourceLanguage, targetLanguage string) ([]dto.UserService, error) {
	log.Printf("translateServices called with %d services, sourceLanguage=%s, targetLanguage=%s", len(services), sourceLanguage, targetLanguage)

	if s.translationService == nil {
		log.Printf("ERROR: translationService is nil")
		return services, fmt.Errorf("DeepL translation service is nil")
	}

	if !s.translationService.IsConfigured() {
		log.Printf("ERROR: translationService is not configured")
		return services, fmt.Errorf("DeepL translation service not configured")
	}

	// Skip translation if source and target languages are the same
	if sourceLanguage == targetLanguage {
		log.Printf("Skipping translation - source and target languages are the same (%s)", targetLanguage)
		return services, nil
	}

	log.Printf("Starting translation of %d services from %s to %s", len(services), sourceLanguage, targetLanguage)
	translatedServices := make([]dto.UserService, len(services))
	successCount := 0
	failCount := 0

	for i, service := range services {
		translatedService := service

		// Translate service name
		if service.ServiceName != "" {
			log.Printf("Translating service %d: '%s' from %s to %s", i, service.ServiceName, sourceLanguage, targetLanguage)
			translatedName, err := s.translationService.Translate(service.ServiceName, sourceLanguage, targetLanguage)
			if err != nil {
				log.Printf("Failed to translate service name '%s': %v (keeping original)", service.ServiceName, err)
				failCount++
				// Keep original name if translation fails
			} else {
				log.Printf("Successfully translated '%s' -> '%s'", service.ServiceName, translatedName)
				translatedService.ServiceName = translatedName
				successCount++
			}
		} else {
			log.Printf("Service %d has empty ServiceName, skipping translation", i)
		}

		translatedServices[i] = translatedService
	}

	log.Printf("Translation complete: %d succeeded, %d failed out of %d total", successCount, failCount, len(services))
	return translatedServices, nil
}
