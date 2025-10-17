package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	CORS        CORSConfig        `yaml:"cors"`
	WebSocket   WebSocketConfig   `yaml:"websocket"`
	Rooms       RoomsConfig       `yaml:"rooms"`
	Logging     LoggingConfig     `yaml:"logging"`
	ExternalAPI ExternalAPIConfig `yaml:"external_api"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	MongoDB MongoDBConfig `yaml:"mongodb"`
}

// MongoDBConfig contains MongoDB-specific configuration
type MongoDBConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

// WebSocketConfig contains WebSocket configuration
type WebSocketConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// ServicePointConfig contains service point configuration
type ServicePointConfig struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	ManagerID   string `yaml:"manager_id,omitempty"`
	ManagerName string `yaml:"manager_name,omitempty"`
}

// RoomConfig contains configuration for a specific room
type RoomConfig struct {
	ID            string               `yaml:"id"`
	Name          string               `yaml:"name"`
	ServicePoints []ServicePointConfig `yaml:"service_points"`
}

// RoomsConfig contains room configuration
type RoomsConfig struct {
	DefaultRoom   string       `yaml:"default_room"`
	AllowWildcard bool         `yaml:"allow_wildcard"`
	Rooms         []RoomConfig `yaml:"rooms,omitempty"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// ExternalAPIConfig contains external API configuration
type ExternalAPIConfig struct {
	UserServicesURL string `yaml:"user_services_url"`
	Timeout         int    `yaml:"timeout_seconds"`
	RetryAttempts   int    `yaml:"retry_attempts"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	config := &Config{}

	// Load from YAML file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables
	overrideFromEnv(config)

	// Set defaults for any missing values
	setDefaults(config)

	return config, nil
}

// overrideFromEnv overrides configuration values with environment variables
func overrideFromEnv(config *Config) {
	if port := os.Getenv("WAITING_ROOM_PORT"); port != "" {
		config.Server.Port = port
	}

	if host := os.Getenv("WAITING_ROOM_HOST"); host != "" {
		config.Server.Host = host
	}

	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		config.Database.MongoDB.URI = uri
	}

	if database := os.Getenv("MONGODB_DATABASE"); database != "" {
		config.Database.MongoDB.Database = database
	}

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}

	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Logging.Format = format
	}

	if enabled := os.Getenv("WEBSOCKET_ENABLED"); enabled != "" {
		config.WebSocket.Enabled = enabled == "true"
	}

	if defaultRoom := os.Getenv("DEFAULT_ROOM"); defaultRoom != "" {
		config.Rooms.DefaultRoom = defaultRoom
	}

	if allowWildcard := os.Getenv("ALLOW_WILDCARD"); allowWildcard != "" {
		config.Rooms.AllowWildcard = strings.EqualFold(allowWildcard, "true")
	}

	if userServicesURL := os.Getenv("EXTERNAL_API_USER_SERVICES_URL"); userServicesURL != "" {
		config.ExternalAPI.UserServicesURL = userServicesURL
	}

	if timeout := os.Getenv("EXTERNAL_API_TIMEOUT_SECONDS"); timeout != "" {
		if timeoutInt, err := fmt.Sscanf(timeout, "%d", &config.ExternalAPI.Timeout); err == nil && timeoutInt > 0 {
			// timeout is already set by the scan
		}
	}

	if retryAttempts := os.Getenv("EXTERNAL_API_RETRY_ATTEMPTS"); retryAttempts != "" {
		if retryInt, err := fmt.Sscanf(retryAttempts, "%d", &config.ExternalAPI.RetryAttempts); err == nil && retryInt > 0 {
			// retryAttempts is already set by the scan
		}
	}
}

// setDefaults sets default values for missing configuration
func setDefaults(config *Config) {
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}

	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}

	if config.Database.MongoDB.URI == "" {
		config.Database.MongoDB.URI = "mongodb://admin:admin@localhost:27017/waiting_room?authSource=admin"
	}

	if config.Database.MongoDB.Database == "" {
		config.Database.MongoDB.Database = "waiting_room"
	}

	if len(config.CORS.AllowedOrigins) == 0 {
		config.CORS.AllowedOrigins = []string{
			"http://localhost:4200",
			"http://localhost:4201",
			"http://localhost:4203",
			"http://localhost:4204",
		}
	}

	if len(config.CORS.AllowedMethods) == 0 {
		config.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}

	if len(config.CORS.AllowedHeaders) == 0 {
		config.CORS.AllowedHeaders = []string{"*"}
	}

	if config.WebSocket.Path == "" {
		config.WebSocket.Path = "/ws/queue"
	}

	if config.Rooms.DefaultRoom == "" {
		if len(config.Rooms.Rooms) > 0 {
			config.Rooms.DefaultRoom = config.Rooms.Rooms[0].ID
		} else {
			config.Rooms.DefaultRoom = "triage-1"
		}
	}

	// Default to allowing wildcard rooms only when no rooms are explicitly configured.
	if len(config.Rooms.Rooms) == 0 && !config.Rooms.AllowWildcard {
		config.Rooms.AllowWildcard = true
	}

	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}

	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}

	if config.ExternalAPI.UserServicesURL == "" {
		config.ExternalAPI.UserServicesURL = "http://external-api.example.com/user-services"
	}

	if config.ExternalAPI.Timeout == 0 {
		config.ExternalAPI.Timeout = 10
	}

	if config.ExternalAPI.RetryAttempts == 0 {
		config.ExternalAPI.RetryAttempts = 3
	}
}

// GetAddress returns the server address in the format "host:port"
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// GetMongoURI returns the MongoDB URI
func (c *Config) GetMongoURI() string {
	return c.Database.MongoDB.URI
}

// GetMongoDatabase returns the MongoDB database name
func (c *Config) GetMongoDatabase() string {
	return c.Database.MongoDB.Database
}

// GetCORSOrigins returns CORS allowed origins as a comma-separated string
func (c *Config) GetCORSOrigins() string {
	return strings.Join(c.CORS.AllowedOrigins, ",")
}

// GetAvailableCORSOrigins returns the list of allowed CORS origins
func (c *Config) GetAvailableCORSOrigins() []string {
	return c.CORS.AllowedOrigins
}

// GetCORSMethods returns CORS allowed methods as a comma-separated string
func (c *Config) GetCORSMethods() string {
	return strings.Join(c.CORS.AllowedMethods, ",")
}

// GetCORSHeaders returns CORS allowed headers as a comma-separated string
func (c *Config) GetCORSHeaders() string {
	return strings.Join(c.CORS.AllowedHeaders, ",")
}

// GetDefaultRoom returns the default room ID
func (c *Config) GetDefaultRoom() string {
	return c.Rooms.DefaultRoom
}

// IsValidRoom checks if a room ID is valid
func (c *Config) IsValidRoom(roomID string) bool {
	if roomID == "" || len(roomID) == 0 {
		return false
	}

	// If wildcard is allowed, accept any non-empty string
	if c.Rooms.AllowWildcard {
		return true
	}

	if roomID == c.Rooms.DefaultRoom {
		return true
	}

	for _, room := range c.Rooms.Rooms {
		if room.ID == roomID {
			return true
		}
	}

	return false
}

// GetServicePointsForRoom returns the service points configured for a specific room
func (c *Config) GetServicePointsForRoom(roomID string) []ServicePointConfig {
	for _, room := range c.Rooms.Rooms {
		if room.ID == roomID {
			if len(room.ServicePoints) > 0 {
				return room.ServicePoints
			}
			break
		}
	}

	// Fallback to default room if explicit room not found
	for _, room := range c.Rooms.Rooms {
		if room.ID == c.Rooms.DefaultRoom {
			if len(room.ServicePoints) > 0 {
				return room.ServicePoints
			}
			break
		}
	}

	// If no specific room config found, return default service points
	return []ServicePointConfig{
		{ID: "window-1", Name: "Window 1", Description: "Main service window"},
		{ID: "window-2", Name: "Window 2", Description: "Secondary service window"},
	}
}

// GetDefaultServicePoint returns the first available service point for a room
func (c *Config) GetDefaultServicePoint(roomID string) string {
	servicePoints := c.GetServicePointsForRoom(roomID)
	if len(servicePoints) > 0 {
		return servicePoints[0].ID
	}
	return "window-1" // fallback
}

// GetExternalAPIUserServicesURL returns the external API URL for user services
func (c *Config) GetExternalAPIUserServicesURL() string {
	return c.ExternalAPI.UserServicesURL
}

// GetExternalAPITimeout returns the external API timeout in seconds
func (c *Config) GetExternalAPITimeout() int {
	return c.ExternalAPI.Timeout
}

// GetExternalAPIRetryAttempts returns the number of retry attempts for external API calls
func (c *Config) GetExternalAPIRetryAttempts() int {
	return c.ExternalAPI.RetryAttempts
}
