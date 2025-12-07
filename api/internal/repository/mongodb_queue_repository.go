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

	"github.com/arfis/waiting-room/internal/db"
	"github.com/arfis/waiting-room/internal/middleware"
	"github.com/arfis/waiting-room/internal/types"
	"github.com/google/uuid"
)

// MongoDBQueueRepository implements QueueRepository using MongoDB with tenant isolation
type MongoDBQueueRepository struct {
	tenantManager *db.TenantDatabaseManager
}

// NewMongoDBQueueRepository creates a new MongoDB queue repository using tenant database manager
func NewMongoDBQueueRepository(uri, dbName string) (*MongoDBQueueRepository, error) {
	tenantManager, err := db.NewTenantDatabaseManager(uri, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant database manager: %w", err)
	}

	return &MongoDBQueueRepository{
		tenantManager: tenantManager,
	}, nil
}

// getTenantDatabase is a helper to get the tenant database from context
func (r *MongoDBQueueRepository) getTenantDatabase(ctx context.Context) (*mongo.Database, error) {
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok || tenantID == 0 {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	return r.tenantManager.GetTenantDatabase(tenantID)
}

// CreateEntry creates a new queue entry in the tenant's database
func (r *MongoDBQueueRepository) CreateEntry(ctx context.Context, entry *types.Entry) error {
	log.Printf("[QueueRepository] Creating entry for room %s", entry.WaitingRoomID)

	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

	// Get section ID from context if present
	sectionID := middleware.GetSectionIDOrZero(ctx)
	if sectionID != 0 {
		entry.SectionID = &sectionID
	}

	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	// Generate ticket number if not set
	if entry.TicketNumber == "" {
		// Build filter for counting entries: same room + same section
		countFilter := bson.M{"waitingRoomId": entry.WaitingRoomID}

		// Add section filter if section ID is set
		if entry.SectionID != nil && *entry.SectionID != 0 {
			countFilter["sectionId"] = *entry.SectionID
		} else {
			// Count only tenant-level entries (no section)
			countFilter["$or"] = []bson.M{
				{"sectionId": bson.M{"$exists": false}},
				{"sectionId": nil},
			}
		}

		// Get current count for this specific room + section to generate ticket number
		count, err := collection.CountDocuments(ctx, countFilter)
		if err != nil {
			log.Printf("[QueueRepository] Failed to count documents for room %s, sectionId %v: %v", entry.WaitingRoomID, entry.SectionID, err)
			count = 0 // Fallback to 0 if count fails
		}

		entry.TicketNumber = fmt.Sprintf("%s-%03d", strings.ToUpper(entry.WaitingRoomID), count+1)
		log.Printf("[QueueRepository] Generated ticket number: %s for room: %s, sectionId: %v (count: %d)", entry.TicketNumber, entry.WaitingRoomID, entry.SectionID, count)
	}

	if entry.QRToken == "" {
		entry.QRToken = uuid.NewString()
		log.Printf("[QueueRepository] Generated QR token: %s", entry.QRToken)
	}

	log.Printf("[QueueRepository] Inserting entry: %+v", entry)
	result, err := collection.InsertOne(ctx, entry)
	if err != nil {
		log.Printf("[QueueRepository] Insert failed: %v", err)
		return fmt.Errorf("failed to create queue entry: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		entry.ID = oid.Hex()
		log.Printf("[QueueRepository] Created entry with ID: %s", entry.ID)
	}

	return nil
}

// GetQueueEntries retrieves all queue entries for a room (filtered by section if provided in context)
func (r *MongoDBQueueRepository) GetQueueEntries(ctx context.Context, roomId string, states []string) ([]*types.Entry, error) {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

	// Get section ID from context (optional)
	sectionID := middleware.GetSectionIDOrZero(ctx)

	filter := bson.M{"waitingRoomId": roomId}

	// Add section filtering if section ID is provided
	if sectionID != 0 {
		filter["sectionId"] = sectionID
		log.Printf("[QueueRepository] Filtering by sectionId: %d", sectionID)
	} else {
		log.Printf("[QueueRepository] No sectionId provided (tenant-level query)")
	}

	if len(states) > 0 {
		filter["status"] = bson.M{"$in": states}
	}

	log.Printf("[QueueRepository] GetQueueEntries for room %s, sectionId: %d, filter: %+v", roomId, sectionID, filter)

	// Sort by priority: tier (lowest first), fitness score (lowest first), arrival time (earliest first), ticket number (alphabetically)
	opts := options.Find().SetSort(bson.D{
		{Key: "tier", Value: 1},
		{Key: "fitnessScore", Value: 1},
		{Key: "createdAt", Value: 1},
		{Key: "ticketNumber", Value: 1},
	})

	cursor, err := collection.Find(ctx, filter, opts)
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

// GetEntryByID retrieves a queue entry by ID from tenant database
func (r *MongoDBQueueRepository) GetEntryByID(ctx context.Context, id string) (*types.Entry, error) {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

	// Try to parse as ObjectID first, if that fails, use as string
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		// Use string ID (for UUIDs)
		filter = bson.M{"_id": id}
	}

	var entry types.Entry
	err = collection.FindOne(ctx, filter).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("queue entry not found")
		}
		return nil, fmt.Errorf("failed to find queue entry: %w", err)
	}

	return &entry, nil
}

// GetEntryByQRToken retrieves a queue entry by QR token from tenant database
func (r *MongoDBQueueRepository) GetEntryByQRToken(ctx context.Context, qrToken string) (*types.Entry, error) {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

	filter := bson.M{"qrToken": qrToken}
	var entry types.Entry

	err = collection.FindOne(ctx, filter).Decode(&entry)
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
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

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

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update entry status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("queue entry not found")
	}

	return nil
}

// UpdateEntryStatusAndSymbols updates the status and symbols of a queue entry
func (r *MongoDBQueueRepository) UpdateEntryStatusAndSymbols(ctx context.Context, id string, status string, symbols []string) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

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
			"symbols":   symbols,
			"updatedAt": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update entry status and symbols: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("queue entry not found")
	}

	return nil
}

// UpdateEntryPosition updates the position of a queue entry
func (r *MongoDBQueueRepository) UpdateEntryPosition(ctx context.Context, id string, position int) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

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

	result, err := collection.UpdateOne(ctx, filter, update)
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
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

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

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update entry service point: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("queue entry not found")
	}

	return nil
}

// GetNextWaitingEntry gets the next waiting entry for a room (filtered by section if provided in context)
func (r *MongoDBQueueRepository) GetNextWaitingEntry(ctx context.Context, roomId string) (*types.Entry, error) {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

	// Get section ID from context (optional)
	sectionID := middleware.GetSectionIDOrZero(ctx)

	log.Printf("[QueueRepository] GetNextWaitingEntry for room %s, sectionId: %d", roomId, sectionID)

	filter := bson.M{
		"waitingRoomId": roomId,
		"status":        "WAITING",
	}

	// Add section filtering if section ID is provided
	if sectionID != 0 {
		filter["sectionId"] = sectionID
	}

	// Sort by priority: tier (lowest first), fitness score (lowest first), arrival time (earliest first), ticket number (alphabetically)
	opts := options.FindOne().SetSort(bson.D{
		{Key: "tier", Value: 1},
		{Key: "fitnessScore", Value: 1},
		{Key: "createdAt", Value: 1},
		{Key: "ticketNumber", Value: 1},
	})

	log.Printf("[QueueRepository] GetNextWaitingEntry filter: %+v", filter)

	var entry types.Entry
	err = collection.FindOne(ctx, filter, opts).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("[QueueRepository] No waiting entries found")
			return nil, nil // No waiting entries
		}
		log.Printf("[QueueRepository] Error finding next waiting entry: %v", err)
		return nil, fmt.Errorf("failed to find next waiting entry: %w", err)
	}

	log.Printf("[QueueRepository] Successfully found entry: %+v", entry)
	return &entry, nil
}

// GetCurrentServedEntry gets the currently served entry for a room (filtered by section if provided in context)
func (r *MongoDBQueueRepository) GetCurrentServedEntry(ctx context.Context, roomId string) (*types.Entry, error) {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

	// Get section ID from context (optional)
	sectionID := middleware.GetSectionIDOrZero(ctx)

	filter := bson.M{
		"waitingRoomId": roomId,
		"status": bson.M{
			"$in": []string{"CALLED", "IN_SERVICE"},
		},
	}

	// Add section filtering if section ID is provided
	if sectionID != 0 {
		filter["sectionId"] = sectionID
	}

	var entry types.Entry
	err = collection.FindOne(ctx, filter).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No one currently being served
		}
		return nil, fmt.Errorf("failed to find current served entry: %w", err)
	}

	return &entry, nil
}

// RecalculatePositions recalculates positions for all waiting entries in a room (filtered by section if provided in context)
func (r *MongoDBQueueRepository) RecalculatePositions(ctx context.Context, roomId string) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

	// Get section ID from context (optional)
	sectionID := middleware.GetSectionIDOrZero(ctx)

	// Get all waiting entries sorted by priority
	filter := bson.M{
		"waitingRoomId": roomId,
		"status":        "WAITING",
	}

	// Add section filtering if section ID is provided
	if sectionID != 0 {
		filter["sectionId"] = sectionID
	}

	// Sort by: tier (lowest first), fitness score (lowest first), arrival time (earliest first), ticket number (alphabetically)
	opts := options.Find().SetSort(bson.D{
		{Key: "tier", Value: 1},
		{Key: "fitnessScore", Value: 1},
		{Key: "createdAt", Value: 1},
		{Key: "ticketNumber", Value: 1},
	})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return fmt.Errorf("failed to find waiting entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []types.Entry
	if err := cursor.All(ctx, &entries); err != nil {
		return fmt.Errorf("failed to decode waiting entries: %w", err)
	}

	// Update positions based on the priority-sorted order
	for i, entry := range entries {
		newPosition := i + 1
		if entry.Position != int64(newPosition) {
			if err := r.UpdateEntryPosition(ctx, entry.ID, newPosition); err != nil {
				return fmt.Errorf("failed to update position for entry %s: %w", entry.ID, err)
			}
		}
	}

	log.Printf("[QueueRepository] Recalculated positions for %d entries in room %s (sectionId: %d)",
		len(entries), roomId, sectionID)
	return nil
}

// DeleteEntry deletes a queue entry from tenant database
func (r *MongoDBQueueRepository) DeleteEntry(ctx context.Context, id string) error {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

	// Try to parse as ObjectID first, if that fails, use as string
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objectID}
	} else {
		// Use string ID (for UUIDs)
		filter = bson.M{"_id": id}
	}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete queue entry: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("queue entry not found")
	}

	return nil
}

// GetNextWaitingEntryForServicePoint gets the next waiting entry for a specific service point (filtered by section if provided in context)
func (r *MongoDBQueueRepository) GetNextWaitingEntryForServicePoint(ctx context.Context, roomId, servicePointId string) (*types.Entry, error) {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

	// Get section ID from context (optional)
	sectionID := middleware.GetSectionIDOrZero(ctx)

	filter := bson.M{
		"waitingRoomId": roomId,
		"servicePoint":  servicePointId,
		"status":        "WAITING",
	}

	// Add section filtering if section ID is provided
	if sectionID != 0 {
		filter["sectionId"] = sectionID
	}

	// Sort by priority: tier (lowest first), fitness score (lowest first), arrival time (earliest first), ticket number (alphabetically)
	opts := options.FindOne().SetSort(bson.D{
		{Key: "tier", Value: 1},
		{Key: "fitnessScore", Value: 1},
		{Key: "createdAt", Value: 1},
		{Key: "ticketNumber", Value: 1},
	})

	var entry types.Entry
	err = collection.FindOne(ctx, filter, opts).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get next waiting entry for service point: %w", err)
	}

	return &entry, nil
}

// GetCurrentServedEntryForServicePoint gets the currently served entry for a specific service point (filtered by section if provided in context)
func (r *MongoDBQueueRepository) GetCurrentServedEntryForServicePoint(ctx context.Context, roomId, servicePointId string) (*types.Entry, error) {
	tenantDB, err := r.getTenantDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	collection := tenantDB.Collection("waiting_queue")

	// Get section ID from context (optional)
	sectionID := middleware.GetSectionIDOrZero(ctx)

	filter := bson.M{
		"waitingRoomId": roomId,
		"servicePoint":  servicePointId,
		"status":        bson.M{"$in": []string{"CALLED", "IN_ROOM", "IN_SERVICE"}},
	}

	// Add section filtering if section ID is provided
	if sectionID != 0 {
		filter["sectionId"] = sectionID
	}

	opts := options.FindOne().SetSort(bson.M{"updatedAt": -1})

	var entry types.Entry
	err = collection.FindOne(ctx, filter, opts).Decode(&entry)
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

	return r.tenantManager.Close(ctx)
}
