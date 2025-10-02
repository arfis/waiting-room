package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	CORS      CORSConfig      `yaml:"cors"`
	WebSocket WebSocketConfig `yaml:"websocket"`
	Rooms     RoomsConfig     `yaml:"rooms"`
	Logging   LoggingConfig   `yaml:"logging"`
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

// RoomsConfig contains room configuration
type RoomsConfig struct {
	DefaultRoom   string `yaml:"default_room"`
	AllowWildcard bool   `yaml:"allow_wildcard"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
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
		config.Rooms.DefaultRoom = "triage-1"
	}

	// Default to allowing wildcard rooms if not specified
	if !config.Rooms.AllowWildcard {
		config.Rooms.AllowWildcard = true
	}

	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}

	if config.Logging.Format == "" {
		config.Logging.Format = "text"
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

	// Otherwise, only allow the default room (for strict mode)
	return roomID == c.Rooms.DefaultRoom
}
