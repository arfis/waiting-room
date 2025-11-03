package tenant

import (
	"context"
	"fmt"

	"github.com/arfis/waiting-room/internal/data/dto"
	"github.com/arfis/waiting-room/internal/repository"
	"github.com/arfis/waiting-room/internal/types"
)

type Service struct {
	repo repository.ConfigRepository
}

func NewService(repo repository.ConfigRepository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateTenant creates a new tenant
func (s *Service) CreateTenant(ctx context.Context, tenantDTO *dto.Tenant) (*dto.Tenant, error) {
	tenant := &types.Tenant{
		BuildingID:  tenantDTO.BuildingId,
		SectionID:   tenantDTO.SectionId,
		Name:        tenantDTO.Name,
		Description: getStringValue(tenantDTO.Description),
	}

	// Validate tenant ID format
	if tenant.BuildingID == "" || tenant.SectionID == "" {
		return nil, fmt.Errorf("building ID and section ID are required")
	}

	tenantID := tenant.GetTenantID()
	if _, _, err := types.ParseTenantID(tenantID); err != nil {
		return nil, err
	}

	err := s.repo.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}

	return s.convertToDTO(tenant), nil
}

// GetTenant retrieves a tenant by ID
func (s *Service) GetTenant(ctx context.Context, tenantID string) (*dto.Tenant, error) {
	tenant, err := s.repo.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, nil
	}
	return s.convertToDTO(tenant), nil
}

// GetAllTenants retrieves all tenants
func (s *Service) GetAllTenants(ctx context.Context) ([]dto.Tenant, error) {
	tenants, err := s.repo.GetAllTenants(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure we return an empty slice instead of nil
	if tenants == nil {
		return []dto.Tenant{}, nil
	}

	dtoTenants := make([]dto.Tenant, 0, len(tenants))
	for _, tenant := range tenants {
		dtoTenants = append(dtoTenants, *s.convertToDTO(&tenant))
	}
	return dtoTenants, nil
}

// UpdateTenant updates an existing tenant
func (s *Service) UpdateTenant(ctx context.Context, tenantDTO *dto.Tenant) (*dto.Tenant, error) {
	tenant := &types.Tenant{
		BuildingID:  tenantDTO.BuildingId,
		SectionID:   tenantDTO.SectionId,
		Name:        tenantDTO.Name,
		Description: getStringValue(tenantDTO.Description),
	}

	// Validate tenant ID format
	if tenant.BuildingID == "" || tenant.SectionID == "" {
		return nil, fmt.Errorf("building ID and section ID are required")
	}

	tenantID := tenant.GetTenantID()
	if _, _, err := types.ParseTenantID(tenantID); err != nil {
		return nil, err
	}

	err := s.repo.UpdateTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}

	return s.convertToDTO(tenant), nil
}

// DeleteTenant deletes a tenant and its configuration
func (s *Service) DeleteTenant(ctx context.Context, tenantID string) error {
	return s.repo.DeleteTenant(ctx, tenantID)
}

// Helper function to convert types.Tenant to dto.Tenant
func (s *Service) convertToDTO(tenant *types.Tenant) *dto.Tenant {
	dtoResult := &dto.Tenant{
		BuildingId: tenant.BuildingID,
		SectionId:  tenant.SectionID,
		Name:       tenant.Name,
		Id:         &tenant.ID,
	}

	if tenant.Description != "" {
		dtoResult.Description = &tenant.Description
	}

	if !tenant.CreatedAt.IsZero() {
		dtoResult.CreatedAt = &tenant.CreatedAt
	}

	if !tenant.UpdatedAt.IsZero() {
		dtoResult.UpdatedAt = &tenant.UpdatedAt
	}

	return dtoResult
}

// Helper function to get string value from pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
