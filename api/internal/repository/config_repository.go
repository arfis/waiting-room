package repository

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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
}

type MongoDBConfigRepository struct {
	collection           *mongo.Collection
	cardReaderCollection *mongo.Collection
}

func NewMongoDBConfigRepository(db *mongo.Database) *MongoDBConfigRepository {
	return &MongoDBConfigRepository{
		collection:           db.Collection("system_configuration"),
		cardReaderCollection: db.Collection("card_readers"),
	}
}

// System configuration management methods
func (r *MongoDBConfigRepository) GetSystemConfiguration(ctx context.Context) (*types.SystemConfiguration, error) {
	var config types.SystemConfiguration
	err := r.collection.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("No system configuration document found in MongoDB")
			return nil, nil
		}
		log.Printf("Error retrieving system configuration from MongoDB: %v", err)
		return nil, err
	}
	log.Printf("Retrieved system configuration from MongoDB - External API URL: %s", config.ExternalAPI.UserServicesURL)
	return &config, nil
}

func (r *MongoDBConfigRepository) SetSystemConfiguration(ctx context.Context, config *types.SystemConfiguration) error {
	now := time.Now()
	config.UpdatedAt = now

	// Use upsert to update or create
	opts := options.Update().SetUpsert(true)
	filter := bson.M{}
	update := bson.M{
		"$set": bson.M{
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
	now := time.Now()
	updates["updatedAt"] = now

	opts := options.Update().SetUpsert(true)
	filter := bson.M{}
	update := bson.M{
		"$set": updates,
		"$setOnInsert": bson.M{
			"createdAt": now,
		},
	}

	log.Printf("Updating MongoDB configuration with: %+v", updates)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Printf("Failed to update MongoDB configuration: %v", err)
		return err
	}

	log.Printf("MongoDB update result - Matched: %d, Modified: %d, Upserted: %d",
		result.MatchedCount, result.ModifiedCount, result.UpsertedCount)
	return nil
}

// Card reader management methods
func (r *MongoDBConfigRepository) GetCardReaderStatus(ctx context.Context, id string) (*types.CardReaderStatus, error) {
	var status types.CardReaderStatus
	err := r.cardReaderCollection.FindOne(ctx, bson.M{"id": id}).Decode(&status)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &status, nil
}

func (r *MongoDBConfigRepository) SetCardReaderStatus(ctx context.Context, status *types.CardReaderStatus) error {
	now := time.Now()
	status.UpdatedAt = now

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"id": status.ID}
	update := bson.M{
		"$set": bson.M{
			"id":        status.ID,
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
	cursor, err := r.cardReaderCollection.Find(ctx, bson.M{})
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
	_, err := r.cardReaderCollection.UpdateOne(
		ctx,
		bson.M{"id": id},
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
	_, err := r.cardReaderCollection.DeleteOne(ctx, bson.M{"id": id})
	return err
}
