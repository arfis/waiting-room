package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/arfis/waiting-room/internal/middleware"
	"github.com/arfis/waiting-room/internal/types"
)

type ConfigRepository interface {
	// System configuration management
	GetSystemConfiguration(ctx context.Context) (*types.SystemConfiguration, error)
	SetSystemConfiguration(ctx context.Context, config *types.SystemConfiguration) error
	UpdateSystemConfiguration(ctx context.Context, updates map[string]interface{}) error

	// Card reader management
	GetCardReaderStatus(ctx context.Context, id string) (*types.CardReaderStatus, error)
	SetCardReaderStatus(ctx context.Context, status *types.CardReaderStatus) error
	GetAllCardReaders(ctx context.Context) ([]types.CardReaderStatus, error)
	UpdateCardReaderLastSeen(ctx context.Context, id string) error
	DeleteCardReader(ctx context.Context, id string) error

	// Tenant management
	CreateTenant(ctx context.Context, tenant *types.Tenant) error
	GetTenant(ctx context.Context, tenantID string) (*types.Tenant, error)
	GetAllTenants(ctx context.Context) ([]types.Tenant, error)
	UpdateTenant(ctx context.Context, tenant *types.Tenant) error
	DeleteTenant(ctx context.Context, tenantID string) error
}

type MongoDBConfigRepository struct {
	collection           *mongo.Collection
	cardReaderCollection *mongo.Collection
	tenantCollection     *mongo.Collection
}

func NewMongoDBConfigRepository(db *mongo.Database) *MongoDBConfigRepository {
	return &MongoDBConfigRepository{
		collection:           db.Collection("system_configuration"),
		cardReaderCollection: db.Collection("card_readers"),
		tenantCollection:     db.Collection("tenants"),
	}
}

// System configuration management methods
func (r *MongoDBConfigRepository) GetSystemConfiguration(ctx context.Context) (*types.SystemConfiguration, error) {
	// Extract tenant ID from context (format: "buildingId:sectionId" or just "buildingId")
	tenantIDHeader := getTenantIDFromContext(ctx)

	// Parse tenant ID to extract buildingId and sectionId
	buildingID, sectionID, _ := types.ParseTenantID(tenantIDHeader)

	var filter bson.M
	if tenantIDHeader != "" {
		// When tenant ID is provided, filter by both tenantId (building) and sectionId (section)
		// This allows multiple sections within the same tenant to have different configurations
		filter = bson.M{}
		if buildingID != "" {
			filter["tenantId"] = buildingID
		}
		if sectionID != "" {
			filter["sectionId"] = sectionID
		}
		log.Printf("[ConfigRepository] Querying with tenant filter: buildingId=%s, sectionId=%s, filter=%+v", buildingID, sectionID, filter)
	} else {
		// When no tenant ID is provided, only return documents without tenantId (legacy/system configs)
		// Use $or to match documents where tenantId doesn't exist OR is null/empty
		filter = bson.M{
			"$or": []bson.M{
				{"tenantId": bson.M{"$exists": false}},
				{"tenantId": ""},
				{"tenantId": nil},
			},
		}
		log.Printf("[ConfigRepository] Querying without tenant (legacy/system config), filter=%+v", filter)
	}

	var config types.SystemConfiguration
	err := r.collection.FindOne(ctx, filter).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("[ConfigRepository] No system configuration document found in MongoDB for buildingId: %s, sectionId: %s", buildingID, sectionID)
			return nil, nil
		}
		log.Printf("[ConfigRepository] Error retrieving system configuration from MongoDB: %v", err)
		return nil, err
	}
	log.Printf("[ConfigRepository] Retrieved system configuration from MongoDB - buildingId: %s, sectionId: %s, config ID: %s", buildingID, sectionID, config.ID)
	return &config, nil
}

// Helper function to extract tenant ID from context
func getTenantIDFromContext(ctx context.Context) string {
	// Use middleware.TENANT constant to match what the middleware sets
	tenantID := ctx.Value(middleware.TENANT)
	if tenantID == nil {
		return ""
	}
	return tenantID.(string)
}

func (r *MongoDBConfigRepository) SetSystemConfiguration(ctx context.Context, config *types.SystemConfiguration) error {
	// Extract tenant ID from context (format: "buildingId:sectionId")
	tenantIDHeader := getTenantIDFromContext(ctx)

	// Parse tenant ID to extract buildingId and sectionId
	buildingID, sectionID, _ := types.ParseTenantID(tenantIDHeader)

	// Set both tenantId (building) and sectionId (section) on the config
	if buildingID != "" {
		config.TenantID = buildingID
	}
	if sectionID != "" {
		config.SectionID = sectionID
	}

	now := time.Now()
	config.UpdatedAt = now

	// Use upsert to update or create
	opts := options.Update().SetUpsert(true)
	filter := bson.M{}
	if buildingID != "" {
		filter["tenantId"] = buildingID
	}
	if sectionID != "" {
		filter["sectionId"] = sectionID
	}

	update := bson.M{
		"$set": bson.M{
			"tenantId":      config.TenantID,
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

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoDBConfigRepository) UpdateSystemConfiguration(ctx context.Context, updates map[string]interface{}) error {
	// Extract tenant ID from context (format: "buildingId:sectionId")
	tenantIDHeader := getTenantIDFromContext(ctx)

	// Parse tenant ID to extract buildingId and sectionId
	buildingID, sectionID, _ := types.ParseTenantID(tenantIDHeader)

	// Set both tenantId (building) and sectionId (section) in updates if provided
	if buildingID != "" {
		updates["tenantId"] = buildingID
	}
	if sectionID != "" {
		updates["sectionId"] = sectionID
	}

	now := time.Now()
	updates["updatedAt"] = now

	opts := options.Update().SetUpsert(true)
	filter := bson.M{}
	if buildingID != "" {
		filter["tenantId"] = buildingID
	}
	if sectionID != "" {
		filter["sectionId"] = sectionID
	}

	update := bson.M{
		"$set": updates,
		"$setOnInsert": bson.M{
			"createdAt": now,
		},
	}

	log.Printf("Updating MongoDB configuration with: %+v for buildingId: %s, sectionId: %s", updates, buildingID, sectionID)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Printf("Failed to update MongoDB configuration: %v", err)
		return err
	}

	log.Printf("MongoDB update result - Matched: %d, Modified: %d, Upserted: %d for buildingId: %s, sectionId: %s",
		result.MatchedCount, result.ModifiedCount, result.UpsertedCount, buildingID, sectionID)
	return nil
}

// Card reader management methods
func (r *MongoDBConfigRepository) GetCardReaderStatus(ctx context.Context, id string) (*types.CardReaderStatus, error) {
	// Extract tenant ID from context
	tenantID := getTenantIDFromContext(ctx)

	filter := bson.M{"id": id}
	if tenantID != "" {
		filter["tenantId"] = tenantID
	}

	var status types.CardReaderStatus
	err := r.cardReaderCollection.FindOne(ctx, filter).Decode(&status)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &status, nil
}

func (r *MongoDBConfigRepository) SetCardReaderStatus(ctx context.Context, status *types.CardReaderStatus) error {
	// Extract tenant ID from context and set it on the status
	tenantID := getTenantIDFromContext(ctx)
	if tenantID != "" {
		status.TenantID = tenantID
	}

	now := time.Now()
	status.UpdatedAt = now

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"id": status.ID}
	if tenantID != "" {
		filter["tenantId"] = tenantID
	}

	update := bson.M{
		"$set": bson.M{
			"id":        status.ID,
			"tenantId":  status.TenantID,
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

	_, err := r.cardReaderCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoDBConfigRepository) GetAllCardReaders(ctx context.Context) ([]types.CardReaderStatus, error) {
	// Extract tenant ID from context
	tenantID := getTenantIDFromContext(ctx)

	filter := bson.M{}
	if tenantID != "" {
		filter["tenantId"] = tenantID
	}

	cursor, err := r.cardReaderCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var readers []types.CardReaderStatus
	if err = cursor.All(ctx, &readers); err != nil {
		return nil, err
	}
	return readers, nil
}

func (r *MongoDBConfigRepository) UpdateCardReaderLastSeen(ctx context.Context, id string) error {
	// Extract tenant ID from context
	tenantID := getTenantIDFromContext(ctx)

	filter := bson.M{"id": id}
	if tenantID != "" {
		filter["tenantId"] = tenantID
	}

	_, err := r.cardReaderCollection.UpdateOne(
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
	return err
}

func (r *MongoDBConfigRepository) DeleteCardReader(ctx context.Context, id string) error {
	// Extract tenant ID from context
	tenantID := getTenantIDFromContext(ctx)

	filter := bson.M{"id": id}
	if tenantID != "" {
		filter["tenantId"] = tenantID
	}

	_, err := r.cardReaderCollection.DeleteOne(ctx, filter)
	return err
}

// Tenant management methods
func (r *MongoDBConfigRepository) CreateTenant(ctx context.Context, tenant *types.Tenant) error {
	now := time.Now()
	tenantID := tenant.GetTenantID()
	tenant.ID = tenantID
	tenant.CreatedAt = now
	tenant.UpdatedAt = now

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"id": tenantID}
	update := bson.M{
		"$set": bson.M{
			"id":          tenant.ID,
			"buildingId":  tenant.BuildingID,
			"sectionId":   tenant.SectionID,
			"name":        tenant.Name,
			"description": tenant.Description,
			"updatedAt":   tenant.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"createdAt": tenant.CreatedAt,
		},
	}

	_, err := r.tenantCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Printf("Failed to create tenant in MongoDB: %v", err)
		return err
	}
	log.Printf("Created tenant with ID: %s", tenantID)
	return nil
}

func (r *MongoDBConfigRepository) GetTenant(ctx context.Context, tenantID string) (*types.Tenant, error) {
	var tenant types.Tenant
	err := r.tenantCollection.FindOne(ctx, bson.M{"id": tenantID}).Decode(&tenant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		log.Printf("Error retrieving tenant from MongoDB: %v", err)
		return nil, err
	}
	return &tenant, nil
}

func (r *MongoDBConfigRepository) GetAllTenants(ctx context.Context) ([]types.Tenant, error) {
	cursor, err := r.tenantCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tenants []types.Tenant
	if err = cursor.All(ctx, &tenants); err != nil {
		log.Printf("Error retrieving all tenants from MongoDB: %v", err)
		return nil, err
	}

	// Ensure we return an empty slice instead of nil
	if tenants == nil {
		return []types.Tenant{}, nil
	}

	return tenants, nil
}

func (r *MongoDBConfigRepository) UpdateTenant(ctx context.Context, tenant *types.Tenant) error {
	now := time.Now()
	tenantID := tenant.GetTenantID()
	tenant.ID = tenantID
	tenant.UpdatedAt = now

	filter := bson.M{"id": tenantID}
	update := bson.M{
		"$set": bson.M{
			"buildingId":  tenant.BuildingID,
			"sectionId":   tenant.SectionID,
			"name":        tenant.Name,
			"description": tenant.Description,
			"updatedAt":   tenant.UpdatedAt,
		},
	}

	result, err := r.tenantCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Failed to update tenant in MongoDB: %v", err)
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("tenant with ID %s not found", tenantID)
	}
	log.Printf("Updated tenant with ID: %s", tenantID)
	return nil
}

func (r *MongoDBConfigRepository) DeleteTenant(ctx context.Context, tenantID string) error {
	result, err := r.tenantCollection.DeleteOne(ctx, bson.M{"id": tenantID})
	if err != nil {
		log.Printf("Failed to delete tenant from MongoDB: %v", err)
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("tenant with ID %s not found", tenantID)
	}
	log.Printf("Deleted tenant with ID: %s", tenantID)
	return nil
}
