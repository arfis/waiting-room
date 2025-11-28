package priority

import (
	"context"
	"log"

	"github.com/arfis/waiting-room/internal/priority"
	"github.com/arfis/waiting-room/internal/service"
)

// Service handles priority configuration management
type Service struct {
	priorityRepo *priority.Repository
}

// New creates a new priority configuration service
func New(priorityRepo *priority.Repository) *Service {
	return &Service{
		priorityRepo: priorityRepo,
	}
}

// GetPriorityConfig retrieves the priority configuration for a tenant/section
func (s *Service) GetPriorityConfig(ctx context.Context) (*priority.PriorityConfig, error) {
	// Extract tenant ID from context
	tenantIDHeader := service.GetTenantID(ctx)
	buildingID, sectionID := parseTenantID(tenantIDHeader)

	log.Printf("[PriorityService] Getting config for tenant: %s, section: %s", buildingID, sectionID)

	config, err := s.priorityRepo.GetConfig(ctx, buildingID, sectionID)
	if err != nil {
		log.Printf("[PriorityService] Error getting config: %v", err)
		return nil, err
	}

	return config, nil
}

// SavePriorityConfig saves the priority configuration for a tenant/section
func (s *Service) SavePriorityConfig(ctx context.Context, config *priority.PriorityConfig) error {
	// Extract tenant ID from context
	tenantIDHeader := service.GetTenantID(ctx)
	buildingID, sectionID := parseTenantID(tenantIDHeader)

	log.Printf("[PriorityService] Saving config for tenant: %s, section: %s", buildingID, sectionID)

	err := s.priorityRepo.SaveConfig(ctx, config, buildingID, sectionID)
	if err != nil {
		log.Printf("[PriorityService] Error saving config: %v", err)
		return err
	}

	log.Printf("[PriorityService] Config saved successfully")
	return nil
}

// GetDefaultConfig returns the default priority configuration
func (s *Service) GetDefaultConfig(ctx context.Context) (*priority.PriorityConfig, error) {
	log.Printf("[PriorityService] Getting default config")
	return priority.GetDefaultConfig(), nil
}

// parseTenantID parses the tenant ID header format "buildingId:sectionId"
func parseTenantID(tenantIDHeader string) (buildingID, sectionID string) {
	if tenantIDHeader == "" {
		return "", ""
	}

	// Split by colon
	parts := []rune(tenantIDHeader)
	colonIndex := -1
	for i, r := range parts {
		if r == ':' {
			colonIndex = i
			break
		}
	}

	if colonIndex == -1 {
		// No colon, treat entire string as building ID
		return tenantIDHeader, ""
	}

	buildingID = string(parts[:colonIndex])
	if colonIndex+1 < len(parts) {
		sectionID = string(parts[colonIndex+1:])
	}

	return buildingID, sectionID
}
