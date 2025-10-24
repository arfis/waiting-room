package types

import "time"

// SystemConfiguration represents the complete system configuration stored in MongoDB
type SystemConfiguration struct {
	ID            string            `bson:"_id,omitempty" json:"id"`
	ExternalAPI   ExternalAPIConfig `bson:"externalAPI" json:"externalAPI"`
	Rooms         []RoomConfig      `bson:"rooms" json:"rooms"`
	DefaultRoom   string            `bson:"defaultRoom" json:"defaultRoom"`
	WebSocketPath string            `bson:"webSocketPath" json:"webSocketPath"`
	AllowWildcard bool              `bson:"allowWildcard" json:"allowWildcard"`
	CreatedAt     time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time         `bson:"updatedAt" json:"updatedAt"`
}

// ExternalAPIConfig represents external API configuration
type ExternalAPIConfig struct {
	UserServicesURL string            `bson:"userServicesUrl" json:"userServicesUrl"`
	TimeoutSeconds  int               `bson:"timeoutSeconds" json:"timeoutSeconds"`
	RetryAttempts   int               `bson:"retryAttempts" json:"retryAttempts"`
	Headers         map[string]string `bson:"headers,omitempty" json:"headers,omitempty"`
}

// RoomConfig represents room configuration
type RoomConfig struct {
	ID            string               `bson:"id" json:"id"`
	Name          string               `bson:"name" json:"name"`
	Description   string               `bson:"description,omitempty" json:"description,omitempty"`
	ServicePoints []ServicePointConfig `bson:"servicePoints" json:"servicePoints"`
	IsDefault     bool                 `bson:"isDefault" json:"isDefault"`
}

// ServicePointConfig represents service point configuration
type ServicePointConfig struct {
	ID          string `bson:"id" json:"id"`
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	ManagerID   string `bson:"managerId,omitempty" json:"managerId,omitempty"`
	ManagerName string `bson:"managerName,omitempty" json:"managerName,omitempty"`
}

// CardReaderStatus represents the status of a card reader
type CardReaderStatus struct {
	ID        string    `bson:"id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	Status    string    `bson:"status" json:"status"` // "online", "offline", "error"
	LastSeen  time.Time `bson:"lastSeen" json:"lastSeen"`
	IPAddress string    `bson:"ipAddress,omitempty" json:"ipAddress,omitempty"`
	Version   string    `bson:"version,omitempty" json:"version,omitempty"`
	LastError string    `bson:"lastError,omitempty" json:"lastError,omitempty"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}
