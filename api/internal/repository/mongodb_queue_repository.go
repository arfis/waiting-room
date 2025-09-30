package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/arfis/waiting-room/internal/types"
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
			Keys: bson.D{{Key: "waiting_room_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "qr_token", Value: 1}},
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
		// Get current count to generate ticket number
		count, err := r.collection.CountDocuments(ctx, bson.M{"waitingRoomId": entry.WaitingRoomID})
		if err != nil {
			log.Printf("MongoDB: Failed to count documents: %v", err)
			count = 0 // Fallback to 0 if count fails
		}
		entry.TicketNumber = fmt.Sprintf("A-%03d", count+1)
		log.Printf("MongoDB: Generated ticket number: %s", entry.TicketNumber)
	}

	if entry.QRToken == "" {
		// Generate a simple QR token (in production, use a proper UUID)
		entry.QRToken = fmt.Sprintf("qr-token-%s-%d", entry.WaitingRoomID, time.Now().Unix())
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

// GetQueueEntries retrieves all queue entries for a room
func (r *MongoDBQueueRepository) GetQueueEntries(ctx context.Context, roomId string) ([]*types.Entry, error) {
	filter := bson.M{"waitingRoomId": roomId}
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

// GetNextWaitingEntry gets the next waiting entry for a room
func (r *MongoDBQueueRepository) GetNextWaitingEntry(ctx context.Context, roomId string) (*types.Entry, error) {
	log.Printf("MongoDB: GetNextWaitingEntry for room %s", roomId)

	filter := bson.M{
		"waitingRoomId": roomId,
		"status":        "WAITING",
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "position", Value: 1}})

	log.Printf("MongoDB: Filter: %+v", filter)

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

// GetCurrentServedEntry gets the currently served entry for a room
func (r *MongoDBQueueRepository) GetCurrentServedEntry(ctx context.Context, roomId string) (*types.Entry, error) {
	filter := bson.M{
		"waitingRoomId": roomId,
		"status": bson.M{
			"$in": []string{"CALLED", "IN_SERVICE"},
		},
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

// RecalculatePositions recalculates positions for all waiting entries in a room
func (r *MongoDBQueueRepository) RecalculatePositions(ctx context.Context, roomId string) error {
	// Get all waiting entries sorted by creation time
	filter := bson.M{
		"waitingRoomId": roomId,
		"status":        "WAITING",
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
		if entry.Position != newPosition {
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

// Close closes the repository connection
func (r *MongoDBQueueRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.client.Disconnect(ctx)
}
