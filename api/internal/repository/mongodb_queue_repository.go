package repository

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/arfis/waiting-room/internal/types"
	"github.com/google/uuid"
)

// MongoDBQueueRepository implements QueueRepository using MongoDB
type MongoDBQueueRepository struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

// NewMongoDBQueueRepository creates a new MongoDB queue repository
func NewMongoDBQueueRepository(uri, dbName string) (*MongoDBQueueRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)
	collection := database.Collection("queue_entries")

	// Create indexes (ignore errors for existing indexes)
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "waitingRoomId", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "qrToken", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "position", Value: 1}},
		},
	}

	// Try to create indexes, but don't fail if they already exist
	for _, index := range indexes {
		_, err := collection.Indexes().CreateOne(ctx, index)
		if err != nil {
			// Log but don't fail - index might already exist
			log.Printf("Index creation warning (may already exist): %v", err)
		}
	}

	// Clean up existing entries with null qrToken values
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cleanupCancel()

	_, err = collection.UpdateMany(cleanupCtx,
		bson.M{"qrToken": bson.M{"$in": []interface{}{nil, ""}}},
		bson.M{"$set": bson.M{"qrToken": ""}},
	)
	if err != nil {
		log.Printf("Warning: Failed to cleanup null qrToken values: %v", err)
	}

	return &MongoDBQueueRepository{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

// CreateEntry creates a new queue entry
func (r *MongoDBQueueRepository) CreateEntry(ctx context.Context, entry *types.Entry) error {
	log.Printf("MongoDB: Creating entry for room %s", entry.WaitingRoomID)

	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	// Generate ticket number and QR token if not set
	if entry.TicketNumber == "" {
		// Build filter for counting entries: same room + same tenant + same section
		// This ensures numbering is per tenant/section, not global
		countFilter := bson.M{"waitingRoomId": entry.WaitingRoomID}
		
		// Only add tenant filter if tenant ID is set
		if entry.TenantID != "" {
			countFilter["tenantId"] = entry.TenantID
		}
		
		// Only add section filter if section ID is set
		if entry.SectionID != "" {
			countFilter["sectionId"] = entry.SectionID
		}
		
		// Get current count for this specific room + tenant + section to generate ticket number
		count, err := r.collection.CountDocuments(ctx, countFilter)
		if err != nil {
			log.Printf("MongoDB: Failed to count documents for room %s, tenant %s, section %s: %v", entry.WaitingRoomID, entry.TenantID, entry.SectionID, err)
			count = 0 // Fallback to 0 if count fails
		}
		entry.TicketNumber = fmt.Sprintf("%s-%03d", strings.ToUpper(entry.WaitingRoomID), count+1)
		log.Printf("MongoDB: Generated ticket number: %s for room: %s, tenant: %s, section: %s (count: %d)", entry.TicketNumber, entry.WaitingRoomID, entry.TenantID, entry.SectionID, count)
	}

	if entry.QRToken == "" {
		// Generate a simple QR token (in production, use a proper UUID)
		entry.QRToken = uuid.NewString()
		log.Printf("MongoDB: Generated QR token: %s", entry.QRToken)
	}

	log.Printf("MongoDB: Inserting entry: %+v", entry)
	result, err := r.collection.InsertOne(ctx, entry)
	if err != nil {
		log.Printf("MongoDB: Insert failed: %v", err)
		return fmt.Errorf("failed to create queue entry: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		entry.ID = oid.Hex()
		log.Printf("MongoDB: Created entry with ID: %s", entry.ID)
	}

	return nil
}

// GetQueueEntries retrieves all queue entries for a room (filtered by tenant if provided)
func (r *MongoDBQueueRepository) GetQueueEntries(ctx context.Context, roomId string, states []string) ([]*types.Entry, error) {
	// Extract tenant ID from context (format: "buildingId:sectionId")
	tenantIDHeader := getTenantIDFromContext(ctx)
	buildingID, sectionID, _ := types.ParseTenantID(tenantIDHeader)
	
	filter := bson.M{"waitingRoomId": roomId}
	
	// Add tenant filtering if tenant ID is provided
	// If tenant ID is empty, we should NOT return all entries - we should return empty or only entries without tenant
	// But for now, if tenant ID is empty, we'll still filter by roomId only (for backward compatibility)
	// The caller should ensure tenant ID is always provided
	if buildingID != "" {
		filter["tenantId"] = buildingID
		log.Printf("[QueueRepository] Filtering by tenantId: '%s'", buildingID)
	} else {
		log.Printf("[QueueRepository] WARNING: buildingID is empty, filter will include entries from all tenants")
	}
	if sectionID != "" {
		filter["sectionId"] = sectionID
		log.Printf("[QueueRepository] Filtering by sectionId: '%s'", sectionID)
	} else if tenantIDHeader != "" {
		// If we have a tenant ID header but no section ID, that's OK (tenant-level only)
		log.Printf("[QueueRepository] No sectionId provided (tenant-level only)")
	}
	
	if len(states) > 0 {
		filter["status"] = bson.M{"$in": states}
	}
	
	log.Printf("[QueueRepository] GetQueueEntries for room %s, tenantIDHeader: '%s', buildingId: '%s', sectionId: '%s', filter: %+v", roomId, tenantIDHeader, buildingID, sectionID, filter)
	
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find queue entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*types.Entry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode queue entries: %w", err)
	}

	return entries, nil
}

// GetEntryByID retrieves a queue entry by ID
func (r *MongoDBQueueRepository) GetEntryByID(ctx context.Context, id string) (*types.Entry, error) {
	// Try to parse as ObjectID first, if that fails, use as string
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		// Use string ID (for UUIDs)
		filter = bson.M{"_id": id}
	}
	var entry types.Entry

	err := r.collection.FindOne(ctx, filter).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("queue entry not found")
		}
		return nil, fmt.Errorf("failed to find queue entry: %w", err)
	}

	return &entry, nil
}

// GetEntryByQRToken retrieves a queue entry by QR token
func (r *MongoDBQueueRepository) GetEntryByQRToken(ctx context.Context, qrToken string) (*types.Entry, error) {
	filter := bson.M{"qrToken": qrToken}
	var entry types.Entry

	err := r.collection.FindOne(ctx, filter).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("queue entry not found")
		}
		return nil, fmt.Errorf("failed to find queue entry: %w", err)
	}

	return &entry, nil
}

// UpdateEntryStatus updates the status of a queue entry
func (r *MongoDBQueueRepository) UpdateEntryStatus(ctx context.Context, id string, status string) error {
	// Try to parse as ObjectID first, if that fails, use as string
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		// Use string ID (for UUIDs)
		filter = bson.M{"_id": id}
	}
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update entry status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("queue entry not found")
	}

	return nil
}

// UpdateEntryPosition updates the position of a queue entry
func (r *MongoDBQueueRepository) UpdateEntryPosition(ctx context.Context, id string, position int) error {
	// Try to parse as ObjectID first, if that fails, use as string
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		// Use string ID (for UUIDs)
		filter = bson.M{"_id": id}
	}
	update := bson.M{
		"$set": bson.M{
			"position":  position,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update entry position: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("queue entry not found")
	}

	return nil
}

// UpdateEntryServicePoint updates the service point of a queue entry
func (r *MongoDBQueueRepository) UpdateEntryServicePoint(ctx context.Context, id string, servicePoint string) error {
	// Try to parse as ObjectID first, if that fails, use as string
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		// Use string ID (for UUIDs)
		filter = bson.M{"_id": id}
	}
	update := bson.M{
		"$set": bson.M{
			"servicePoint": servicePoint,
			"updatedAt":    time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update entry service point: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("queue entry not found")
	}

	return nil
}

// GetNextWaitingEntry gets the next waiting entry for a room (filtered by tenant if provided)
func (r *MongoDBQueueRepository) GetNextWaitingEntry(ctx context.Context, roomId string) (*types.Entry, error) {
	// Extract tenant ID from context (format: "buildingId:sectionId")
	tenantIDHeader := getTenantIDFromContext(ctx)
	buildingID, sectionID, _ := types.ParseTenantID(tenantIDHeader)
	
	log.Printf("[QueueRepository] GetNextWaitingEntry for room %s, buildingId: %s, sectionId: %s", roomId, buildingID, sectionID)

	filter := bson.M{
		"waitingRoomId": roomId,
		"status":        "WAITING",
	}
	
	// Add tenant filtering if tenant ID is provided
	if buildingID != "" {
		filter["tenantId"] = buildingID
	}
	if sectionID != "" {
		filter["sectionId"] = sectionID
	}
	
	opts := options.FindOne().SetSort(bson.D{{Key: "position", Value: 1}})

	log.Printf("[QueueRepository] GetNextWaitingEntry filter: %+v", filter)

	// First, let's count how many documents match this filter
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("MongoDB: Error counting documents: %v", err)
	} else {
		log.Printf("MongoDB: Found %d documents matching filter", count)
	}

	var entry types.Entry
	err = r.collection.FindOne(ctx, filter, opts).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("MongoDB: No waiting entries found (FindOne returned no documents)")
			return nil, nil // No waiting entries
		}
		log.Printf("MongoDB: Error finding next waiting entry: %v", err)
		return nil, fmt.Errorf("failed to find next waiting entry: %w", err)
	}

	log.Printf("MongoDB: Successfully found and decoded entry: %+v", entry)
	return &entry, nil
}

// GetCurrentServedEntry gets the currently served entry for a room (filtered by tenant if provided)
func (r *MongoDBQueueRepository) GetCurrentServedEntry(ctx context.Context, roomId string) (*types.Entry, error) {
	// Extract tenant ID from context (format: "buildingId:sectionId")
	tenantIDHeader := getTenantIDFromContext(ctx)
	buildingID, sectionID, _ := types.ParseTenantID(tenantIDHeader)
	
	filter := bson.M{
		"waitingRoomId": roomId,
		"status": bson.M{
			"$in": []string{"CALLED", "IN_SERVICE"},
		},
	}
	
	// Add tenant filtering if tenant ID is provided
	if buildingID != "" {
		filter["tenantId"] = buildingID
	}
	if sectionID != "" {
		filter["sectionId"] = sectionID
	}

	var entry types.Entry
	err := r.collection.FindOne(ctx, filter).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No one currently being served
		}
		return nil, fmt.Errorf("failed to find current served entry: %w", err)
	}

	return &entry, nil
}

// RecalculatePositions recalculates positions for all waiting entries in a room (filtered by tenant if provided)
func (r *MongoDBQueueRepository) RecalculatePositions(ctx context.Context, roomId string) error {
	// Extract tenant ID from context (format: "buildingId:sectionId")
	tenantIDHeader := getTenantIDFromContext(ctx)
	buildingID, sectionID, _ := types.ParseTenantID(tenantIDHeader)
	
	// Get all waiting entries sorted by creation time
	filter := bson.M{
		"waitingRoomId": roomId,
		"status":        "WAITING",
	}
	
	// Add tenant filtering if tenant ID is provided
	if buildingID != "" {
		filter["tenantId"] = buildingID
	}
	if sectionID != "" {
		filter["sectionId"] = sectionID
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return fmt.Errorf("failed to find waiting entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []types.Entry
	if err := cursor.All(ctx, &entries); err != nil {
		return fmt.Errorf("failed to decode waiting entries: %w", err)
	}

	// Update positions
	for i, entry := range entries {
		newPosition := i + 1
		if entry.Position != int64(newPosition) {
			if err := r.UpdateEntryPosition(ctx, entry.ID, newPosition); err != nil {
				return fmt.Errorf("failed to update position for entry %s: %w", entry.ID, err)
			}
		}
	}

	return nil
}

// DeleteEntry deletes a queue entry
func (r *MongoDBQueueRepository) DeleteEntry(ctx context.Context, id string) error {
	// Try to parse as ObjectID first, if that fails, use as string
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		// Use string ID (for UUIDs)
		filter = bson.M{"_id": id}
	}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete queue entry: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("queue entry not found")
	}

	return nil
}

// GetNextWaitingEntryForServicePoint gets the next waiting entry for a specific service point (filtered by tenant if provided)
func (r *MongoDBQueueRepository) GetNextWaitingEntryForServicePoint(ctx context.Context, roomId, servicePointId string) (*types.Entry, error) {
	// Extract tenant ID from context (format: "buildingId:sectionId")
	tenantIDHeader := getTenantIDFromContext(ctx)
	buildingID, sectionID, _ := types.ParseTenantID(tenantIDHeader)
	
	collection := r.database.Collection("queue_entries")

	filter := bson.M{
		"waitingRoomId": roomId,
		"servicePoint":  servicePointId,
		"status":        "WAITING",
	}
	
	// Add tenant filtering if tenant ID is provided
	if buildingID != "" {
		filter["tenantId"] = buildingID
	}
	if sectionID != "" {
		filter["sectionId"] = sectionID
	}

	opts := options.FindOne().SetSort(bson.M{"position": 1})

	var entry types.Entry
	err := collection.FindOne(ctx, filter, opts).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get next waiting entry for service point: %w", err)
	}

	return &entry, nil
}

// GetCurrentServedEntryForServicePoint gets the currently served entry for a specific service point (filtered by tenant if provided)
func (r *MongoDBQueueRepository) GetCurrentServedEntryForServicePoint(ctx context.Context, roomId, servicePointId string) (*types.Entry, error) {
	// Extract tenant ID from context (format: "buildingId:sectionId")
	tenantIDHeader := getTenantIDFromContext(ctx)
	buildingID, sectionID, _ := types.ParseTenantID(tenantIDHeader)
	
	collection := r.database.Collection("queue_entries")

	filter := bson.M{
		"waitingRoomId": roomId,
		"servicePoint":  servicePointId,
		"status":        bson.M{"$in": []string{"CALLED", "IN_ROOM", "IN_SERVICE"}},
	}
	
	// Add tenant filtering if tenant ID is provided
	if buildingID != "" {
		filter["tenantId"] = buildingID
	}
	if sectionID != "" {
		filter["sectionId"] = sectionID
	}

	opts := options.FindOne().SetSort(bson.M{"updatedAt": -1})

	var entry types.Entry
	err := collection.FindOne(ctx, filter, opts).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get current served entry for service point: %w", err)
	}

	return &entry, nil
}

// Close closes the repository connection
func (r *MongoDBQueueRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.client.Disconnect(ctx)
}
