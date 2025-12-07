package db

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TenantDatabaseManager manages separate MongoDB databases for each tenant
type TenantDatabaseManager struct {
	client      *mongo.Client
	masterDB    *mongo.Database
	tenantDBs   map[int64]*mongo.Database
	mu          sync.RWMutex
	dbNameFunc  func(int64) string // Function to generate database name from tenant ID
	nextTenantID int64              // Auto-increment counter for tenant IDs
	nextSectionID int64             // Auto-increment counter for section IDs
}

// NewTenantDatabaseManager creates a new tenant database manager
func NewTenantDatabaseManager(mongoURI, masterDBName string) (*TenantDatabaseManager, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	masterDB := client.Database(masterDBName)

	manager := &TenantDatabaseManager{
		client:    client,
		masterDB:  masterDB,
		tenantDBs: make(map[int64]*mongo.Database),
		dbNameFunc: func(tenantID int64) string {
			return fmt.Sprintf("tenant_%d", tenantID)
		},
		nextTenantID:  1000, // Start tenant IDs from 1000
		nextSectionID: 2000, // Start section IDs from 2000
	}

	// Initialize the counters from existing data
	if err := manager.initializeCounters(ctx); err != nil {
		log.Printf("Warning: Failed to initialize ID counters: %v", err)
	}

	log.Printf("Tenant database manager initialized with master database: %s", masterDBName)
	return manager, nil
}

// initializeCounters loads the max IDs from existing data to continue auto-increment
func (m *TenantDatabaseManager) initializeCounters(ctx context.Context) error {
	// Find max tenant ID
	tenantOpts := options.FindOne().SetSort(bson.D{{Key: "tenantId", Value: -1}})
	var tenantDoc struct {
		TenantID int64 `bson:"tenantId"`
	}
	err := m.masterDB.Collection("tenants").FindOne(ctx, bson.M{}, tenantOpts).Decode(&tenantDoc)
	if err == nil {
		m.nextTenantID = tenantDoc.TenantID + 1
		log.Printf("Initialized next tenant ID to %d", m.nextTenantID)
	} else if err != mongo.ErrNoDocuments {
		return fmt.Errorf("failed to find max tenant ID: %w", err)
	}

	// For section IDs, we need to check all tenant databases (not implemented yet)
	// For now, just start from 2000

	return nil
}

// GetMasterDatabase returns the master database (for tenant registry)
func (m *TenantDatabaseManager) GetMasterDatabase() *mongo.Database {
	return m.masterDB
}

// GetTenantDatabase returns the database for a specific tenant
// It caches database connections for performance
func (m *TenantDatabaseManager) GetTenantDatabase(tenantID int64) (*mongo.Database, error) {
	// Check cache first with read lock
	m.mu.RLock()
	db, exists := m.tenantDBs[tenantID]
	m.mu.RUnlock()

	if exists {
		return db, nil
	}

	// Not in cache, verify tenant exists and create connection
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if db, exists := m.tenantDBs[tenantID]; exists {
		return db, nil
	}

	// Verify tenant exists in master database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var tenant struct {
		TenantID     int64  `bson:"tenantId"`
		DatabaseName string `bson:"databaseName"`
		Status       string `bson:"status"`
	}

	err := m.masterDB.Collection("tenants").FindOne(ctx, bson.M{"tenantId": tenantID}).Decode(&tenant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("tenant %d not found", tenantID)
		}
		return nil, fmt.Errorf("failed to query tenant: %w", err)
	}

	if tenant.Status != "active" {
		return nil, fmt.Errorf("tenant %d is not active (status: %s)", tenantID, tenant.Status)
	}

	// Get database connection
	dbName := tenant.DatabaseName
	if dbName == "" {
		dbName = m.dbNameFunc(tenantID)
	}

	db = m.client.Database(dbName)
	m.tenantDBs[tenantID] = db

	log.Printf("Created database connection for tenant %d: %s", tenantID, dbName)
	return db, nil
}

// CreateTenantDatabase creates a new database for a tenant
func (m *TenantDatabaseManager) CreateTenantDatabase(ctx context.Context, tenantID int64) error {
	dbName := m.dbNameFunc(tenantID)
	db := m.client.Database(dbName)

	// Create collections with indexes
	collections := []string{
		"system_configuration",
		"card_readers",
		"waiting_queue",
		"sections",
		"priority_config",
	}

	for _, collName := range collections {
		// Create collection if it doesn't exist
		err := db.CreateCollection(ctx, collName)
		if err != nil {
			// Ignore "already exists" errors
			if !mongo.IsDuplicateKeyError(err) && err.Error() != "Collection already exists" {
				log.Printf("Warning: Failed to create collection %s: %v", collName, err)
			}
		}
	}

	// Create indexes
	if err := m.createIndexes(ctx, db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Cache the database connection
	m.mu.Lock()
	m.tenantDBs[tenantID] = db
	m.mu.Unlock()

	log.Printf("Created tenant database: %s", dbName)
	return nil
}

// createIndexes creates necessary indexes for tenant database collections
func (m *TenantDatabaseManager) createIndexes(ctx context.Context, db *mongo.Database) error {
	// Index on sections collection
	sectionsIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "sectionId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := db.Collection("sections").Indexes().CreateMany(ctx, sectionsIndexes)
	if err != nil {
		return fmt.Errorf("failed to create sections indexes: %w", err)
	}

	// Index on waiting_queue collection
	queueIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "sectionId", Value: 1}, {Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "ticketNumber", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "qrToken", Value: 1}},
		},
	}
	_, err = db.Collection("waiting_queue").Indexes().CreateMany(ctx, queueIndexes)
	if err != nil {
		return fmt.Errorf("failed to create queue indexes: %w", err)
	}

	// Index on card_readers collection
	readerIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "id", Value: 1}, {Key: "sectionId", Value: 1}},
		},
	}
	_, err = db.Collection("card_readers").Indexes().CreateMany(ctx, readerIndexes)
	if err != nil {
		return fmt.Errorf("failed to create card_readers indexes: %w", err)
	}

	// Index on system_configuration collection
	configIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "sectionId", Value: 1}},
		},
	}
	_, err = db.Collection("system_configuration").Indexes().CreateMany(ctx, configIndexes)
	if err != nil {
		return fmt.Errorf("failed to create system_configuration indexes: %w", err)
	}

	return nil
}

// DeleteTenantDatabase marks a tenant as inactive (does not actually drop the database)
// This is a safer approach than dropping the database immediately
func (m *TenantDatabaseManager) DeleteTenantDatabase(ctx context.Context, tenantID int64) error {
	// Mark tenant as inactive in master database
	result, err := m.masterDB.Collection("tenants").UpdateOne(
		ctx,
		bson.M{"tenantId": tenantID},
		bson.M{"$set": bson.M{
			"status":    "inactive",
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to mark tenant as inactive: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("tenant %d not found", tenantID)
	}

	// Remove from cache
	m.mu.Lock()
	delete(m.tenantDBs, tenantID)
	m.mu.Unlock()

	log.Printf("Marked tenant %d as inactive", tenantID)
	return nil
}

// DropTenantDatabase permanently drops the tenant database
// WARNING: This is destructive and cannot be undone!
func (m *TenantDatabaseManager) DropTenantDatabase(ctx context.Context, tenantID int64) error {
	var tenant struct {
		DatabaseName string `bson:"databaseName"`
	}

	err := m.masterDB.Collection("tenants").FindOne(ctx, bson.M{"tenantId": tenantID}).Decode(&tenant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("tenant %d not found", tenantID)
		}
		return fmt.Errorf("failed to query tenant: %w", err)
	}

	dbName := tenant.DatabaseName
	if dbName == "" {
		dbName = m.dbNameFunc(tenantID)
	}

	// Drop the database
	if err := m.client.Database(dbName).Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}

	// Remove from cache
	m.mu.Lock()
	delete(m.tenantDBs, tenantID)
	m.mu.Unlock()

	log.Printf("Permanently dropped tenant database: %s", dbName)
	return nil
}

// GenerateNextTenantID generates the next auto-increment tenant ID
func (m *TenantDatabaseManager) GenerateNextTenantID() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := m.nextTenantID
	m.nextTenantID++
	return id
}

// GenerateNextSectionID generates the next auto-increment section ID
func (m *TenantDatabaseManager) GenerateNextSectionID() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := m.nextSectionID
	m.nextSectionID++
	return id
}

// Close closes all database connections
func (m *TenantDatabaseManager) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// ListActiveTenants returns all active tenant IDs
func (m *TenantDatabaseManager) ListActiveTenants(ctx context.Context) ([]int64, error) {
	cursor, err := m.masterDB.Collection("tenants").Find(ctx, bson.M{"status": "active"})
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer cursor.Close(ctx)

	var tenantIDs []int64
	for cursor.Next(ctx) {
		var tenant struct {
			TenantID int64 `bson:"tenantId"`
		}
		if err := cursor.Decode(&tenant); err != nil {
			log.Printf("Warning: Failed to decode tenant: %v", err)
			continue
		}
		tenantIDs = append(tenantIDs, tenant.TenantID)
	}

	return tenantIDs, cursor.Err()
}
