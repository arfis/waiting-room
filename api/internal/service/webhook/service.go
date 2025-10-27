package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/arfis/waiting-room/internal/service/config"
)

type Service struct {
	configService *config.Service
	httpClient    *http.Client
}

type WebhookPayload struct {
	Event          string                 `json:"event"`
	TicketID       string                 `json:"ticketId"`
	ServiceID      string                 `json:"serviceId,omitempty"`
	State          string                 `json:"state"`
	Timestamp      time.Time              `json:"timestamp"`
	RoomID         string                 `json:"roomId"`
	ServicePointID string                 `json:"servicePointId,omitempty"`
	UserID         string                 `json:"userId,omitempty"`
	AdditionalData map[string]interface{} `json:"additionalData,omitempty"`
}

func NewService(configService *config.Service) *Service {
	return &Service{
		configService: configService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // Default timeout, will be overridden by config
		},
	}
}

// SendWebhook sends a webhook notification for ticket state changes
func (s *Service) SendWebhook(ctx context.Context, payload WebhookPayload) error {
	// Get webhook configuration
	webhookConfig, err := s.getWebhookConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get webhook config: %w", err)
	}

	// If no webhook URL is configured, skip
	if webhookConfig.WebhookURL == "" {
		return nil
	}

	// Create HTTP request
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookConfig.WebhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "WaitingRoom-Webhook/1.0")

	// Add custom headers from configuration
	for key, value := range webhookConfig.Headers {
		req.Header.Set(key, value)
	}

	// Set timeout from configuration
	timeout := time.Duration(webhookConfig.WebhookTimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second // Default timeout
	}

	// Create client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Send webhook with retry logic
	return s.sendWithRetry(ctx, client, req, webhookConfig.WebhookRetryAttempts)
}

// sendWithRetry sends the webhook with retry logic
func (s *Service) sendWithRetry(ctx context.Context, client *http.Client, req *http.Request, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("webhook request failed (attempt %d/%d): %w", attempt+1, maxRetries+1, err)
			if attempt < maxRetries {
				// Wait before retry (exponential backoff)
				waitTime := time.Duration(attempt+1) * time.Second
				time.Sleep(waitTime)
				continue
			}
			return lastErr
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil // Success
		}

		// Read response body for error details
		body, _ := io.ReadAll(resp.Body)
		lastErr = fmt.Errorf("webhook returned status %d (attempt %d/%d): %s",
			resp.StatusCode, attempt+1, maxRetries+1, string(body))

		if attempt < maxRetries {
			// Wait before retry
			waitTime := time.Duration(attempt+1) * time.Second
			time.Sleep(waitTime)
		}
	}

	return lastErr
}

// getWebhookConfig retrieves webhook configuration
func (s *Service) getWebhookConfig(ctx context.Context) (*WebhookConfig, error) {
	config, err := s.configService.GetExternalAPIConfiguration(ctx)
	if err != nil {
		return nil, err
	}

	return &WebhookConfig{
		WebhookURL:            config.WebhookURL,
		WebhookTimeoutSeconds: config.WebhookTimeoutSeconds,
		WebhookRetryAttempts:  config.WebhookRetryAttempts,
		Headers:               config.Headers,
	}, nil
}

type WebhookConfig struct {
	WebhookURL            string
	WebhookTimeoutSeconds int
	WebhookRetryAttempts  int
	Headers               map[string]string
}

// Helper methods for different webhook events

// SendServiceSelectedWebhook sends webhook when a service is selected
func (s *Service) SendServiceSelectedWebhook(ctx context.Context, ticketID, serviceID, roomID, servicePointID, userID string) error {
	payload := WebhookPayload{
		Event:          "service_selected",
		TicketID:       ticketID,
		ServiceID:      serviceID,
		State:          "service_selected",
		Timestamp:      time.Now(),
		RoomID:         roomID,
		ServicePointID: servicePointID,
		UserID:         userID,
	}
	return s.SendWebhook(ctx, payload)
}

// SendTicketCalledWebhook sends webhook when a ticket is called
func (s *Service) SendTicketCalledWebhook(ctx context.Context, ticketID, roomID, servicePointID, userID string) error {
	payload := WebhookPayload{
		Event:          "ticket_called",
		TicketID:       ticketID,
		State:          "called",
		Timestamp:      time.Now(),
		RoomID:         roomID,
		ServicePointID: servicePointID,
		UserID:         userID,
	}
	return s.SendWebhook(ctx, payload)
}

// SendTicketCompletedWebhook sends webhook when a ticket is completed
func (s *Service) SendTicketCompletedWebhook(ctx context.Context, ticketID, roomID, servicePointID, userID string) error {
	payload := WebhookPayload{
		Event:          "ticket_completed",
		TicketID:       ticketID,
		State:          "completed",
		Timestamp:      time.Now(),
		RoomID:         roomID,
		ServicePointID: servicePointID,
		UserID:         userID,
	}
	return s.SendWebhook(ctx, payload)
}

// SendTicketCancelledWebhook sends webhook when a ticket is cancelled
func (s *Service) SendTicketCancelledWebhook(ctx context.Context, ticketID, roomID, servicePointID, userID string) error {
	payload := WebhookPayload{
		Event:          "ticket_cancelled",
		TicketID:       ticketID,
		State:          "cancelled",
		Timestamp:      time.Now(),
		RoomID:         roomID,
		ServicePointID: servicePointID,
		UserID:         userID,
	}
	return s.SendWebhook(ctx, payload)
}

// SendGenericStateChangeWebhook sends webhook for any state change
func (s *Service) SendGenericStateChangeWebhook(ctx context.Context, ticketID, state, roomID, servicePointID, userID string, additionalData map[string]interface{}) error {
	payload := WebhookPayload{
		Event:          "ticket_state_changed",
		TicketID:       ticketID,
		State:          state,
		Timestamp:      time.Now(),
		RoomID:         roomID,
		ServicePointID: servicePointID,
		UserID:         userID,
		AdditionalData: additionalData,
	}
	return s.SendWebhook(ctx, payload)
}
