package kiosk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/arfis/waiting-room/internal/config"
	"github.com/arfis/waiting-room/internal/data/dto"
	ngErrors "github.com/arfis/waiting-room/internal/errors"
	"github.com/arfis/waiting-room/internal/queue"
)

type Service struct {
	queueService  *queue.WaitingQueue
	broadcastFunc func(string) // Function to broadcast queue updates
	config        *config.Config
}

func New(queueService *queue.WaitingQueue, broadcastFunc func(string), config *config.Config) *Service {
	return &Service{
		queueService:  queueService,
		broadcastFunc: broadcastFunc,
		config:        config,
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
	// Get external API configuration
	externalAPIURL := s.config.GetExternalAPIUserServicesURL()
	timeoutSeconds := s.config.GetExternalAPITimeout()

	// Create HTTP client with configured timeout
	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", externalAPIURL, nil)
	if err != nil {
		return nil, ngErrors.New(ngErrors.InternalServerErrorCode, "failed to create request", 500, nil)
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
