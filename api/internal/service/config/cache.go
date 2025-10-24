package config

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/arfis/waiting-room/internal/repository"
	"github.com/arfis/waiting-room/internal/types"
)

// ConfigCache manages in-memory configuration with automatic reloading
type ConfigCache struct {
	mu           sync.RWMutex
	config       *types.SystemConfiguration
	repo         repository.ConfigRepository
	lastReload   time.Time
	reloadTicker *time.Ticker
	stopChan     chan struct{}
}

// NewConfigCache creates a new configuration cache
func NewConfigCache(repo repository.ConfigRepository) *ConfigCache {
	cache := &ConfigCache{
		repo:     repo,
		stopChan: make(chan struct{}),
	}

	// Load initial configuration
	cache.ReloadConfig(context.Background())

	// Start periodic reload (every 30 seconds)
	cache.reloadTicker = time.NewTicker(30 * time.Second)
	go cache.startPeriodicReload()

	return cache
}

// GetSystemConfiguration returns the cached system configuration
func (c *ConfigCache) GetSystemConfiguration() *types.SystemConfiguration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// GetExternalAPIConfig returns the cached external API configuration
func (c *ConfigCache) GetExternalAPIConfig() *types.ExternalAPIConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.config != nil {
		return &c.config.ExternalAPI
	}
	return nil
}

// GetRoomsConfig returns the cached rooms configuration
func (c *ConfigCache) GetRoomsConfig() []types.RoomConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.config != nil {
		return c.config.Rooms
	}
	return nil
}

// ReloadConfig forces a reload of the configuration from the database
func (c *ConfigCache) ReloadConfig(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Println("Reloading configuration from database...")
	config, err := c.repo.GetSystemConfiguration(ctx)
	if err != nil {
		log.Printf("Failed to reload configuration: %v", err)
		return
	}

	if config != nil {
		c.config = config
		c.lastReload = time.Now()
		log.Printf("Configuration reloaded successfully - External API URL: %s", config.ExternalAPI.UserServicesURL)
	} else {
		log.Println("No configuration found in database")
	}
}

// startPeriodicReload starts a goroutine that periodically reloads the configuration
func (c *ConfigCache) startPeriodicReload() {
	for {
		select {
		case <-c.reloadTicker.C:
			c.ReloadConfig(context.Background())
		case <-c.stopChan:
			c.reloadTicker.Stop()
			return
		}
	}
}

// Stop stops the periodic reload
func (c *ConfigCache) Stop() {
	close(c.stopChan)
}

// UpdateConfiguration updates the configuration in the database and reloads the cache
func (c *ConfigCache) UpdateConfiguration(ctx context.Context, config *types.SystemConfiguration) error {
	// Update in database
	err := c.repo.SetSystemConfiguration(ctx, config)
	if err != nil {
		return err
	}

	// Reload cache immediately
	c.ReloadConfig(ctx)
	return nil
}

// UpdateExternalAPIConfiguration updates only the external API configuration
func (c *ConfigCache) UpdateExternalAPIConfiguration(ctx context.Context, apiConfig *types.ExternalAPIConfig) error {
	// Get current config
	currentConfig := c.GetSystemConfiguration()
	if currentConfig == nil {
		// Create new config if none exists
		currentConfig = &types.SystemConfiguration{
			ExternalAPI:   *apiConfig,
			Rooms:         []types.RoomConfig{},
			DefaultRoom:   "",
			WebSocketPath: "/ws/queue",
			AllowWildcard: false,
		}
	} else {
		// Update existing config
		currentConfig.ExternalAPI = *apiConfig
	}

	return c.UpdateConfiguration(ctx, currentConfig)
}

// UpdateRoomsConfiguration updates only the rooms configuration
func (c *ConfigCache) UpdateRoomsConfiguration(ctx context.Context, rooms []types.RoomConfig) error {
	// Get current config
	currentConfig := c.GetSystemConfiguration()
	if currentConfig == nil {
		// Create new config if none exists
		currentConfig = &types.SystemConfiguration{
			ExternalAPI:   types.ExternalAPIConfig{},
			Rooms:         rooms,
			DefaultRoom:   "",
			WebSocketPath: "/ws/queue",
			AllowWildcard: false,
		}
	} else {
		// Update existing config
		currentConfig.Rooms = rooms
	}

	return c.UpdateConfiguration(ctx, currentConfig)
}
