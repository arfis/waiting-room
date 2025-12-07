package tenant

import (
	"context"
	"fmt"
	"strconv"

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
		TenantID:    tenantDTO.TenantId, // Will be generated if 0
		Name:        tenantDTO.Name,
		Description: getStringValue(tenantDTO.Description),
	}

	err := s.repo.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return s.convertToDTO(tenant), nil
}

// GetTenant retrieves a tenant by ID
func (s *Service) GetTenant(ctx context.Context, tenantID string) (*dto.Tenant, error) {
	// Parse tenant ID
	id, err := strconv.ParseInt(tenantID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID format: %w", err)
	}

	tenant, err := s.repo.GetTenant(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
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
		return nil, fmt.Errorf("failed to get all tenants: %w", err)
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
	if tenantDTO.TenantId == 0 {
		return nil, fmt.Errorf("tenant ID is required for update")
	}

	tenant := &types.Tenant{
		TenantID:    tenantDTO.TenantId,
		Name:        tenantDTO.Name,
		Description: getStringValue(tenantDTO.Description),
		Status:      getStringValue(tenantDTO.Status),
	}

	err := s.repo.UpdateTenant(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return s.convertToDTO(tenant), nil
}

// DeleteTenant deletes a tenant (marks as inactive)
func (s *Service) DeleteTenant(ctx context.Context, tenantID string) error {
	// Parse tenant ID
	id, err := strconv.ParseInt(tenantID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid tenant ID format: %w", err)
	}

	err = s.repo.DeleteTenant(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	return nil
}

// Section Management Methods

// CreateSection creates a new section within a tenant
func (s *Service) CreateSection(ctx context.Context, sectionDTO *dto.Section) (*dto.Section, error) {
	section := &types.Section{
		SectionID:   sectionDTO.SectionId, // Will be generated if 0
		Name:        sectionDTO.Name,
		Description: getStringValue(sectionDTO.Description),
	}

	err := s.repo.CreateSection(ctx, section)
	if err != nil {
		return nil, fmt.Errorf("failed to create section: %w", err)
	}

	return s.convertSectionToDTO(section, sectionDTO.TenantId), nil
}

// GetSection retrieves a section by ID
func (s *Service) GetSection(ctx context.Context, sectionID string) (*dto.Section, error) {
	// Parse section ID
	id, err := strconv.ParseInt(sectionID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid section ID format: %w", err)
	}

	section, err := s.repo.GetSection(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get section: %w", err)
	}
	if section == nil {
		return nil, nil
	}

	// Get tenant ID from context
	tenantID := ctx.Value("tenantId")
	var tenantIDInt int64
	if tid, ok := tenantID.(int64); ok {
		tenantIDInt = tid
	}

	return s.convertSectionToDTO(section, tenantIDInt), nil
}

// GetAllSections retrieves all sections for the tenant in context
func (s *Service) GetAllSections(ctx context.Context) ([]dto.Section, error) {
	sections, err := s.repo.GetAllSections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all sections: %w", err)
	}

	// Ensure we return an empty slice instead of nil
	if sections == nil {
		return []dto.Section{}, nil
	}

	// Get tenant ID from context
	tenantID := ctx.Value("tenantId")
	var tenantIDInt int64
	if tid, ok := tenantID.(int64); ok {
		tenantIDInt = tid
	}

	dtoSections := make([]dto.Section, 0, len(sections))
	for _, section := range sections {
		dtoSections = append(dtoSections, *s.convertSectionToDTO(&section, tenantIDInt))
	}

	return dtoSections, nil
}

// UpdateSection updates an existing section
func (s *Service) UpdateSection(ctx context.Context, sectionDTO *dto.Section) (*dto.Section, error) {
	if sectionDTO.SectionId == 0 {
		return nil, fmt.Errorf("section ID is required for update")
	}

	section := &types.Section{
		SectionID:   sectionDTO.SectionId,
		Name:        sectionDTO.Name,
		Description: getStringValue(sectionDTO.Description),
		Status:      getStringValue(sectionDTO.Status),
	}

	err := s.repo.UpdateSection(ctx, section)
	if err != nil {
		return nil, fmt.Errorf("failed to update section: %w", err)
	}

	return s.convertSectionToDTO(section, sectionDTO.TenantId), nil
}

// DeleteSection deletes a section (marks as inactive)
func (s *Service) DeleteSection(ctx context.Context, sectionID string) error {
	// Parse section ID
	id, err := strconv.ParseInt(sectionID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid section ID format: %w", err)
	}

	err = s.repo.DeleteSection(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}

	return nil
}

// Helper function to convert types.Tenant to dto.Tenant
func (s *Service) convertToDTO(tenant *types.Tenant) *dto.Tenant {
	dtoResult := &dto.Tenant{
		TenantId: tenant.TenantID,
		Name:     tenant.Name,
		Id:       &tenant.ID,
	}

	if tenant.Description != "" {
		dtoResult.Description = &tenant.Description
	}

	if tenant.DatabaseName != "" {
		dtoResult.DatabaseName = &tenant.DatabaseName
	}

	if tenant.Status != "" {
		dtoResult.Status = &tenant.Status
	}

	if !tenant.CreatedAt.IsZero() {
		dtoResult.CreatedAt = &tenant.CreatedAt
	}

	if !tenant.UpdatedAt.IsZero() {
		dtoResult.UpdatedAt = &tenant.UpdatedAt
	}

	return dtoResult
}

// Helper function to convert types.Section to dto.Section
func (s *Service) convertSectionToDTO(section *types.Section, tenantID int64) *dto.Section {
	dtoResult := &dto.Section{
		SectionId: section.SectionID,
		TenantId:  tenantID,
		Name:      section.Name,
		Id:        &section.ID,
	}

	if section.Description != "" {
		dtoResult.Description = &section.Description
	}

	if section.Status != "" {
		dtoResult.Status = &section.Status
	}

	if !section.CreatedAt.IsZero() {
		dtoResult.CreatedAt = &section.CreatedAt
	}

	if !section.UpdatedAt.IsZero() {
		dtoResult.UpdatedAt = &section.UpdatedAt
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
