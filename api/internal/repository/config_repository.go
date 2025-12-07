package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/arfis/waiting-room/internal/db"
	"github.com/arfis/waiting-room/internal/middleware"
	"github.com/arfis/waiting-room/internal/types"
)

type ConfigRepository interface {
	// System configuration management (operates on tenant-specific database)
	GetSystemConfiguration(ctx context.Context) (*types.SystemConfiguration, error)
	SetSystemConfiguration(ctx context.Context, config *types.SystemConfiguration) error
	UpdateSystemConfiguration(ctx context.Context, updates map[string]interface{}) error

	// Card reader management (operates on tenant-specific database)
	GetCardReaderStatus(ctx context.Context, id string) (*types.CardReaderStatus, error)
	SetCardReaderStatus(ctx context.Context, status *types.CardReaderStatus) error
	GetAllCardReaders(ctx context.Context) ([]types.CardReaderStatus, error)
	UpdateCardReaderLastSeen(ctx context.Context, id string) error
	DeleteCardReader(ctx context.Context, id string) error

	// Tenant management (operates on master database)
	CreateTenant(ctx context.Context, tenant *types.Tenant) error
	GetTenant(ctx context.Context, tenantID int64) (*types.Tenant, error)
	GetAllTenants(ctx context.Context) ([]types.Tenant, error)
	UpdateTenant(ctx context.Context, tenant *types.Tenant) error
	DeleteTenant(ctx context.Context, tenantID int64) error

	// Section management (operates on tenant-specific database)
	CreateSection(ctx context.Context, section *types.Section) error
	GetSection(ctx context.Context, sectionID int64) (*types.Section, error)
	GetAllSections(ctx context.Context) ([]types.Section, error)
	UpdateSection(ctx context.Context, section *types.Section) error
	DeleteSection(ctx context.Context, sectionID int64) error
}

type MongoDBConfigRepository struct {
	tenantManager *db.TenantDatabaseManager
}

func NewMongoDBConfigRepository(tenantManager *db.TenantDatabaseManager) *MongoDBConfigRepository {
	return &MongoDBConfigRepository{
		tenantManager: tenantManager,
	}
}

// getTenantDatabase is a helper to get the tenant database from context
func (r *MongoDBConfigRepository) getTenantDatabase(ctx context.Context) (*mongo.Database, error) {
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok || tenantID == 0 {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	return r.tenantManager.GetTenantDatabase(tenantID)
}

// System configuration management methods

// GetSystemConfiguration gets the system configuration for the tenant (and optionally section) from context
func (r *MongoDBConfigRepository) GetSystemConfiguration(ctx context.Context) (*types.SystemConfiguration, error) {
	// Get tenant database
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	// Extract section ID from context (optional)
	sectionID := middleware.GetSectionIDOrZero(ctx)

	// Build filter
	var filter bson.M
	if sectionID != 0 {
		// Section-specific config
		filter = bson.M{"sectionId": sectionID}
		log.Printf("[ConfigRepository] Querying section-specific config: sectionId=%d", sectionID)
	} else {
		// Tenant-level config (sectionId is null or doesn't exist)
		filter = bson.M{
			"$or": []bson.M{
				{"sectionId": bson.M{"$exists": false}},
				{"sectionId": nil},
			},
		}
		log.Printf("[ConfigRepository] Querying tenant-level config")
	}

	var config types.SystemConfiguration
	err = tenantDB.Collection("system_configuration").FindOne(ctx, filter).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("[ConfigRepository] No configuration found (sectionId=%d)", sectionID)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve configuration: %w", err)
	}

	log.Printf("[ConfigRepository] Retrieved configuration: ID=%s, sectionId=%v", config.ID, config.SectionID)
	return &config, nil
}

// SetSystemConfiguration saves the system configuration to the tenant database
func (r *MongoDBConfigRepository) SetSystemConfiguration(ctx context.Context, config *types.SystemConfiguration) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	sectionID := middleware.GetSectionIDOrZero(ctx)

	// Set section ID on config
	if sectionID != 0 {
		config.SectionID = &sectionID
	} else {
		config.SectionID = nil
	}

	now := time.Now()
	config.UpdatedAt = now

	// Build filter
	filter := bson.M{}
	if config.SectionID != nil {
		filter["sectionId"] = *config.SectionID
	} else {
		filter["$or"] = []bson.M{
			{"sectionId": bson.M{"$exists": false}},
			{"sectionId": nil},
		}
	}

	update := bson.M{
		"$set": bson.M{
			"sectionId":     config.SectionID,
			"externalAPI":   config.ExternalAPI,
			"rooms":         config.Rooms,
			"defaultRoom":   config.DefaultRoom,
			"webSocketPath": config.WebSocketPath,
			"allowWildcard": config.AllowWildcard,
			"updatedAt":     config.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"createdAt": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err = tenantDB.Collection("system_configuration").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	log.Printf("[ConfigRepository] Saved configuration (sectionId=%v)", config.SectionID)
	return nil
}

// UpdateSystemConfiguration updates the system configuration
func (r *MongoDBConfigRepository) UpdateSystemConfiguration(ctx context.Context, updates map[string]interface{}) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	sectionID := middleware.GetSectionIDOrZero(ctx)

	// Add section ID to updates
	if sectionID != 0 {
		updates["sectionId"] = sectionID
	}

	now := time.Now()
	updates["updatedAt"] = now

	// Build filter
	filter := bson.M{}
	if sectionID != 0 {
		filter["sectionId"] = sectionID
	} else {
		filter["$or"] = []bson.M{
			{"sectionId": bson.M{"$exists": false}},
			{"sectionId": nil},
		}
	}

	update := bson.M{
		"$set": updates,
		"$setOnInsert": bson.M{
			"createdAt": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := tenantDB.Collection("system_configuration").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	log.Printf("[ConfigRepository] Updated configuration (matched=%d, modified=%d, sectionId=%d)",
		result.MatchedCount, result.ModifiedCount, sectionID)
	return nil
}

// Card reader management methods

// GetCardReaderStatus retrieves a card reader by ID from tenant database
func (r *MongoDBConfigRepository) GetCardReaderStatus(ctx context.Context, id string) (*types.CardReaderStatus, error) {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	sectionID := middleware.GetSectionIDOrZero(ctx)

	filter := bson.M{"id": id}
	if sectionID != 0 {
		filter["sectionId"] = sectionID
	}

	var status types.CardReaderStatus
	err = tenantDB.Collection("card_readers").FindOne(ctx, filter).Decode(&status)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve card reader: %w", err)
	}

	return &status, nil
}

// SetCardReaderStatus saves or updates a card reader in tenant database
func (r *MongoDBConfigRepository) SetCardReaderStatus(ctx context.Context, status *types.CardReaderStatus) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	sectionID := middleware.GetSectionIDOrZero(ctx)

	// Set section ID on status
	if sectionID != 0 {
		status.SectionID = &sectionID
	}

	now := time.Now()
	status.UpdatedAt = now

	filter := bson.M{"id": status.ID}
	if sectionID != 0 {
		filter["sectionId"] = sectionID
	}

	update := bson.M{
		"$set": bson.M{
			"id":        status.ID,
			"sectionId": status.SectionID,
			"name":      status.Name,
			"status":    status.Status,
			"lastSeen":  status.LastSeen,
			"ipAddress": status.IPAddress,
			"version":   status.Version,
			"lastError": status.LastError,
			"updatedAt": status.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"createdAt": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err = tenantDB.Collection("card_readers").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save card reader: %w", err)
	}

	log.Printf("[ConfigRepository] Saved card reader: id=%s, sectionId=%v", status.ID, status.SectionID)
	return nil
}

// GetAllCardReaders retrieves all card readers for the tenant/section from context
func (r *MongoDBConfigRepository) GetAllCardReaders(ctx context.Context) ([]types.CardReaderStatus, error) {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	sectionID := middleware.GetSectionIDOrZero(ctx)

	filter := bson.M{}
	if sectionID != 0 {
		filter["sectionId"] = sectionID
	}

	cursor, err := tenantDB.Collection("card_readers").Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find card readers: %w", err)
	}
	defer cursor.Close(ctx)

	var readers []types.CardReaderStatus
	if err = cursor.All(ctx, &readers); err != nil {
		return nil, fmt.Errorf("failed to decode card readers: %w", err)
	}

	return readers, nil
}

// UpdateCardReaderLastSeen updates the last seen timestamp for a card reader
func (r *MongoDBConfigRepository) UpdateCardReaderLastSeen(ctx context.Context, id string) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	sectionID := middleware.GetSectionIDOrZero(ctx)

	filter := bson.M{"id": id}
	if sectionID != 0 {
		filter["sectionId"] = sectionID
	}

	_, err = tenantDB.Collection("card_readers").UpdateOne(
		ctx,
		filter,
		bson.M{
			"$set": bson.M{
				"lastSeen":  time.Now(),
				"status":    "online",
				"updatedAt": time.Now(),
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to update card reader last seen: %w", err)
	}

	return nil
}

// DeleteCardReader removes a card reader from tenant database
func (r *MongoDBConfigRepository) DeleteCardReader(ctx context.Context, id string) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	sectionID := middleware.GetSectionIDOrZero(ctx)

	filter := bson.M{"id": id}
	if sectionID != 0 {
		filter["sectionId"] = sectionID
	}

	_, err = tenantDB.Collection("card_readers").DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete card reader: %w", err)
	}

	log.Printf("[ConfigRepository] Deleted card reader: id=%s", id)
	return nil
}

// Tenant management methods (operate on master database)

// CreateTenant creates a new tenant in the master database and creates its dedicated database
func (r *MongoDBConfigRepository) CreateTenant(ctx context.Context, tenant *types.Tenant) error {
	masterDB := r.tenantManager.GetMasterDatabase()

	// Generate tenant ID if not set
	if tenant.TenantID == 0 {
		tenant.TenantID = r.tenantManager.GenerateNextTenantID()
	}

	// Generate database name
	tenant.DatabaseName = fmt.Sprintf("tenant_%d", tenant.TenantID)
	tenant.Status = "active"

	now := time.Now()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now

	// Insert into master database
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"tenantId": tenant.TenantID}
	update := bson.M{
		"$set": bson.M{
			"tenantId":     tenant.TenantID,
			"name":         tenant.Name,
			"description":  tenant.Description,
			"databaseName": tenant.DatabaseName,
			"status":       tenant.Status,
			"updatedAt":    tenant.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"createdAt": tenant.CreatedAt,
		},
	}

	_, err := masterDB.Collection("tenants").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create tenant database
	if err := r.tenantManager.CreateTenantDatabase(ctx, tenant.TenantID); err != nil {
		return fmt.Errorf("failed to create tenant database: %w", err)
	}

	log.Printf("[ConfigRepository] Created tenant: tenantId=%d, name=%s", tenant.TenantID, tenant.Name)
	return nil
}

// GetTenant retrieves a tenant by ID from master database
func (r *MongoDBConfigRepository) GetTenant(ctx context.Context, tenantID int64) (*types.Tenant, error) {
	masterDB := r.tenantManager.GetMasterDatabase()

	var tenant types.Tenant
	err := masterDB.Collection("tenants").FindOne(ctx, bson.M{"tenantId": tenantID}).Decode(&tenant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve tenant: %w", err)
	}

	return &tenant, nil
}

// GetAllTenants retrieves all tenants from master database
func (r *MongoDBConfigRepository) GetAllTenants(ctx context.Context) ([]types.Tenant, error) {
	masterDB := r.tenantManager.GetMasterDatabase()

	cursor, err := masterDB.Collection("tenants").Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find tenants: %w", err)
	}
	defer cursor.Close(ctx)

	var tenants []types.Tenant
	if err = cursor.All(ctx, &tenants); err != nil {
		return nil, fmt.Errorf("failed to decode tenants: %w", err)
	}

	// Ensure we return an empty slice instead of nil
	if tenants == nil {
		return []types.Tenant{}, nil
	}

	return tenants, nil
}

// UpdateTenant updates a tenant in the master database
func (r *MongoDBConfigRepository) UpdateTenant(ctx context.Context, tenant *types.Tenant) error {
	masterDB := r.tenantManager.GetMasterDatabase()

	now := time.Now()
	tenant.UpdatedAt = now

	filter := bson.M{"tenantId": tenant.TenantID}
	update := bson.M{
		"$set": bson.M{
			"name":        tenant.Name,
			"description": tenant.Description,
			"status":      tenant.Status,
			"updatedAt":   tenant.UpdatedAt,
		},
	}

	result, err := masterDB.Collection("tenants").UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("tenant with ID %d not found", tenant.TenantID)
	}

	log.Printf("[ConfigRepository] Updated tenant: tenantId=%d", tenant.TenantID)
	return nil
}

// DeleteTenant marks a tenant as inactive (soft delete)
func (r *MongoDBConfigRepository) DeleteTenant(ctx context.Context, tenantID int64) error {
	// Use tenant manager to mark as inactive
	if err := r.tenantManager.DeleteTenantDatabase(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	log.Printf("[ConfigRepository] Deleted tenant: tenantId=%d", tenantID)
	return nil
}

// Section management methods (operate on tenant-specific database)

// CreateSection creates a new section in the tenant database
func (r *MongoDBConfigRepository) CreateSection(ctx context.Context, section *types.Section) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	// Generate section ID if not set
	if section.SectionID == 0 {
		section.SectionID = r.tenantManager.GenerateNextSectionID()
	}

	section.Status = "active"

	now := time.Now()
	section.CreatedAt = now
	section.UpdatedAt = now

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"sectionId": section.SectionID}
	update := bson.M{
		"$set": bson.M{
			"sectionId":   section.SectionID,
			"name":        section.Name,
			"description": section.Description,
			"status":      section.Status,
			"updatedAt":   section.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"createdAt": section.CreatedAt,
		},
	}

	_, err = tenantDB.Collection("sections").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to create section: %w", err)
	}

	log.Printf("[ConfigRepository] Created section: sectionId=%d, name=%s", section.SectionID, section.Name)
	return nil
}

// GetSection retrieves a section by ID from tenant database
func (r *MongoDBConfigRepository) GetSection(ctx context.Context, sectionID int64) (*types.Section, error) {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	var section types.Section
	err = tenantDB.Collection("sections").FindOne(ctx, bson.M{"sectionId": sectionID}).Decode(&section)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve section: %w", err)
	}

	return &section, nil
}

// GetAllSections retrieves all sections from tenant database
func (r *MongoDBConfigRepository) GetAllSections(ctx context.Context) ([]types.Section, error) {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	cursor, err := tenantDB.Collection("sections").Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find sections: %w", err)
	}
	defer cursor.Close(ctx)

	var sections []types.Section
	if err = cursor.All(ctx, &sections); err != nil {
		return nil, fmt.Errorf("failed to decode sections: %w", err)
	}

	// Ensure we return an empty slice instead of nil
	if sections == nil {
		return []types.Section{}, nil
	}

	return sections, nil
}

// UpdateSection updates a section in the tenant database
func (r *MongoDBConfigRepository) UpdateSection(ctx context.Context, section *types.Section) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	now := time.Now()
	section.UpdatedAt = now

	filter := bson.M{"sectionId": section.SectionID}
	update := bson.M{
		"$set": bson.M{
			"name":        section.Name,
			"description": section.Description,
			"status":      section.Status,
			"updatedAt":   section.UpdatedAt,
		},
	}

	result, err := tenantDB.Collection("sections").UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update section: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("section with ID %d not found", section.SectionID)
	}

	log.Printf("[ConfigRepository] Updated section: sectionId=%d", section.SectionID)
	return nil
}

// DeleteSection marks a section as inactive (soft delete)
func (r *MongoDBConfigRepository) DeleteSection(ctx context.Context, sectionID int64) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	filter := bson.M{"sectionId": sectionID}
	update := bson.M{
		"$set": bson.M{
			"status":    "inactive",
			"updatedAt": time.Now(),
		},
	}

	result, err := tenantDB.Collection("sections").UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("section with ID %d not found", sectionID)
	}

	log.Printf("[ConfigRepository] Deleted section: sectionId=%d", sectionID)
	return nil
}
