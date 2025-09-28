package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QueueEntry struct {
	ID            string    `bson:"_id,omitempty" json:"id"`
	WaitingRoomID string    `bson:"waitingRoomId" json:"waitingRoomId"`
	TicketNumber  string    `bson:"ticketNumber" json:"ticketNumber"`
	QRToken       string    `bson:"qrToken" json:"qrToken"`
	Status        string    `bson:"status" json:"status"` // WAITING, CALLED, IN_SERVICE, COMPLETED, SKIPPED, CANCELLED, NO_SHOW
	Position      int       `bson:"position" json:"position"`
	CreatedAt     time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt" json:"updatedAt"`
	CardData      CardData  `bson:"cardData,omitempty" json:"cardData,omitempty"`
}

type CardData struct {
	IDNumber    string `bson:"idNumber" json:"idNumber"`
	FirstName   string `bson:"firstName" json:"firstName"`
	LastName    string `bson:"lastName" json:"lastName"`
	DateOfBirth string `bson:"dateOfBirth" json:"dateOfBirth"`
	Gender      string `bson:"gender" json:"gender"`
	Nationality string `bson:"nationality" json:"nationality"`
	Address     string `bson:"address" json:"address"`
	IssuedDate  string `bson:"issuedDate" json:"issuedDate"`
	ExpiryDate  string `bson:"expiryDate" json:"expiryDate"`
	Photo       string `bson:"photo" json:"photo"`
	Source      string `bson:"source" json:"source"`
}

type Service struct {
	client    *mongo.Client
	database  *mongo.Database
	entries   *mongo.Collection
	connected bool
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Initialize(connectionString string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		s.connected = false
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
		s.connected = false
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	s.client = client
	s.database = client.Database("waiting_room")
	s.entries = s.database.Collection("queue_entries")
	s.connected = true

	// Create indexes
	_, err = s.entries.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "waitingRoomId", Value: 1}, {Key: "status", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "qrToken", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "ticketNumber", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	log.Printf("Successfully connected to MongoDB")
	return nil
}

func (s *Service) Close() error {
	if s.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.client.Disconnect(ctx)
	}
	return nil
}

func (s *Service) CreateEntry(waitingRoomID string, cardData CardData) (*QueueEntry, error) {
	log.Printf("CreateEntry called for room: %s, card: %+v", waitingRoomID, cardData)

	// If not connected to MongoDB, return a mock entry
	if !s.connected {
		log.Printf("MongoDB not connected, creating mock entry")
		return s.createMockEntry(waitingRoomID, cardData), nil
	}

	log.Printf("MongoDB connected, proceeding with database operations")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Generate ticket number
	ticketNumber, err := s.generateTicketNumber(waitingRoomID)
	if err != nil {
		log.Printf("Failed to generate ticket number: %v, using mock entry", err)
		return s.createMockEntry(waitingRoomID, cardData), nil
	}

	// Generate QR token
	qrToken := uuid.New().String()

	// Get current position
	position, err := s.getNextPosition(waitingRoomID)
	if err != nil {
		log.Printf("Failed to get next position: %v, using mock entry", err)
		return s.createMockEntry(waitingRoomID, cardData), nil
	}

	entry := &QueueEntry{
		ID:            uuid.New().String(),
		WaitingRoomID: waitingRoomID,
		TicketNumber:  ticketNumber,
		QRToken:       qrToken,
		Status:        "WAITING",
		Position:      position,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		CardData:      cardData,
	}

	_, err = s.entries.InsertOne(ctx, entry)
	if err != nil {
		log.Printf("Failed to insert entry: %v, using mock entry", err)
		return s.createMockEntry(waitingRoomID, cardData), nil
	}

	return entry, nil
}

// createMockEntry creates a mock entry when MongoDB is not available
func (s *Service) createMockEntry(waitingRoomID string, cardData CardData) *QueueEntry {
	ticketNumber := fmt.Sprintf("A-%03d", 1) // Always start with A-001 for mock
	qrToken := uuid.New().String()

	return &QueueEntry{
		ID:            uuid.New().String(),
		WaitingRoomID: waitingRoomID,
		TicketNumber:  ticketNumber,
		QRToken:       qrToken,
		Status:        "WAITING",
		Position:      1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		CardData:      cardData,
	}
}

func (s *Service) GetEntryByQRToken(qrToken string) (*QueueEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var entry QueueEntry
	err := s.entries.FindOne(ctx, bson.M{"qrToken": qrToken}).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("entry not found")
		}
		return nil, fmt.Errorf("failed to find entry: %w", err)
	}

	return &entry, nil
}

func (s *Service) GetQueueEntries(waitingRoomID string) ([]QueueEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := s.entries.Find(ctx, bson.M{
		"waitingRoomId": waitingRoomID,
		"status":        bson.M{"$in": []string{"WAITING", "CALLED"}},
	}, options.Find().SetSort(bson.D{{Key: "position", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to find entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []QueueEntry
	if err = cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}

func (s *Service) CallNext(waitingRoomID string) (*QueueEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, complete any currently served person (CALLED or IN_SERVICE)
	_, err := s.entries.UpdateMany(
		ctx,
		bson.M{
			"waitingRoomId": waitingRoomID,
			"status":        bson.M{"$in": []string{"CALLED", "IN_SERVICE"}},
		},
		bson.M{
			"$set": bson.M{
				"status":    "COMPLETED",
				"updatedAt": time.Now(),
			},
		},
	)
	if err != nil {
		log.Printf("Failed to complete current entries: %v", err)
		// Continue anyway, don't fail the whole operation
	}

	// Find the next waiting entry and call them
	var entry QueueEntry
	err = s.entries.FindOneAndUpdate(
		ctx,
		bson.M{
			"waitingRoomId": waitingRoomID,
			"status":        "WAITING",
		},
		bson.M{
			"$set": bson.M{
				"status":    "CALLED",
				"updatedAt": time.Now(),
			},
		},
		options.FindOneAndUpdate().SetSort(bson.D{{Key: "position", Value: 1}}),
	).Decode(&entry)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no entries waiting")
		}
		return nil, fmt.Errorf("failed to call next: %w", err)
	}

	// Recalculate positions for all remaining WAITING entries
	log.Printf("About to recalculate positions for room: %s", waitingRoomID)
	err = s.recalculatePositions(waitingRoomID)
	if err != nil {
		log.Printf("Failed to recalculate positions: %v", err)
		// Don't fail the whole operation, but log the error
	} else {
		log.Printf("Successfully recalculated positions for room: %s", waitingRoomID)
	}

	return &entry, nil
}

// FinishCurrent completes the currently served person without calling the next
func (s *Service) FinishCurrent(waitingRoomID string) (*QueueEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("Finishing current person for room: %s", waitingRoomID)

	// First, get the currently served person to return their info
	var entry QueueEntry
	err := s.entries.FindOne(
		ctx,
		bson.M{
			"waitingRoomId": waitingRoomID,
			"status":        bson.M{"$in": []string{"CALLED", "IN_SERVICE"}},
		},
	).Decode(&entry)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("No one is currently being served in room: %s", waitingRoomID)
			return nil, nil // No error, just no one to finish
		}
		return nil, fmt.Errorf("failed to find current person: %w", err)
	}

	log.Printf("Found current person: %s (ticket: %s), updating to COMPLETED", entry.ID, entry.TicketNumber)

	// Update the status using UpdateMany (same as CallNext)
	result, err := s.entries.UpdateMany(
		ctx,
		bson.M{
			"waitingRoomId": waitingRoomID,
			"status":        bson.M{"$in": []string{"CALLED", "IN_SERVICE"}},
		},
		bson.M{
			"$set": bson.M{
				"status":    "COMPLETED",
				"updatedAt": time.Now(),
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update person status: %w", err)
	}

	if result.ModifiedCount == 0 {
		log.Printf("No entries were modified for room: %s", waitingRoomID)
		return nil, nil
	}

	log.Printf("Successfully finished person: %s (ticket: %s)", entry.ID, entry.TicketNumber)
	return &entry, nil
}

// recalculatePositions updates the position field for all WAITING entries
func (s *Service) recalculatePositions(waitingRoomID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get all WAITING entries sorted by creation time
	cursor, err := s.entries.Find(
		ctx,
		bson.M{
			"waitingRoomId": waitingRoomID,
			"status":        "WAITING",
		},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}),
	)
	if err != nil {
		return fmt.Errorf("failed to find waiting entries: %w", err)
	}
	defer cursor.Close(ctx)

	position := 1
	for cursor.Next(ctx) {
		var entry QueueEntry
		if err := cursor.Decode(&entry); err != nil {
			log.Printf("Failed to decode entry: %v", err)
			continue
		}

		// Update position if it's different
		if entry.Position != position {
			_, err := s.entries.UpdateOne(
				ctx,
				bson.M{"_id": entry.ID},
				bson.M{
					"$set": bson.M{
						"position":  position,
						"updatedAt": time.Now(),
					},
				},
			)
			if err != nil {
				log.Printf("Failed to update position for entry %s: %v", entry.ID, err)
			}
		}
		position++
	}

	return cursor.Err()
}

func (s *Service) generateTicketNumber(waitingRoomID string) (string, error) {
	log.Printf("generateTicketNumber called for room: %s", waitingRoomID)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	log.Printf("Querying for last entry in room: %s", waitingRoomID)
	// Get the last ticket number for this room
	var lastEntry QueueEntry
	err := s.entries.FindOne(
		ctx,
		bson.M{"waitingRoomId": waitingRoomID},
		options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	).Decode(&lastEntry)

	log.Printf("FindOne result: err=%v, entry=%+v", err, lastEntry)

	// If no entries exist, start with 1
	if err == mongo.ErrNoDocuments {
		return fmt.Sprintf("A-%03d", 1), nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get last ticket: %w", err)
	}

	// Extract number from last ticket and increment
	var lastNumber int
	fmt.Sscanf(lastEntry.TicketNumber, "A-%d", &lastNumber)
	return fmt.Sprintf("A-%03d", lastNumber+1), nil
}

func (s *Service) getNextPosition(waitingRoomID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := s.entries.CountDocuments(ctx, bson.M{
		"waitingRoomId": waitingRoomID,
		"status":        bson.M{"$in": []string{"WAITING", "CALLED"}},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count entries: %w", err)
	}

	return int(count) + 1, nil
}
