package types

import (
	"fmt"
	"time"
)

// SystemConfiguration represents the complete system configuration stored in MongoDB
// NOTE: Now stored in tenant-specific databases, so no tenantId field needed
type SystemConfiguration struct {
	ID            string            `bson:"_id,omitempty" json:"id"`
	SectionID     *int64            `bson:"sectionId,omitempty" json:"sectionId,omitempty"` // Section/Department within tenant (numeric ID). Null = tenant-level config
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
	AppointmentServicesURL        string            `bson:"appointmentServicesUrl,omitempty" json:"appointmentServicesUrl,omitempty"`
	AppointmentServicesHttpMethod *string           `bson:"appointmentServicesHttpMethod,omitempty" json:"appointmentServicesHttpMethod,omitempty"`
	GenericServicesURL            string            `bson:"genericServicesUrl,omitempty" json:"genericServicesUrl,omitempty"`
	GenericServicesHttpMethod     *string           `bson:"genericServicesHttpMethod,omitempty" json:"genericServicesHttpMethod,omitempty"`
	GenericServicesPostBody       string            `bson:"genericServicesPostBody,omitempty" json:"genericServicesPostBody,omitempty"`
	GenericServices               []GenericService  `bson:"genericServices,omitempty" json:"genericServices,omitempty"`
	WebhookURL                    string            `bson:"webhookUrl,omitempty" json:"webhookUrl,omitempty"`
	WebhookHttpMethod             *string           `bson:"webhookHttpMethod,omitempty" json:"webhookHttpMethod,omitempty"`
	WebhookTimeoutSeconds         int               `bson:"webhookTimeoutSeconds,omitempty" json:"webhookTimeoutSeconds,omitempty"`
	WebhookRetryAttempts          int               `bson:"webhookRetryAttempts,omitempty" json:"webhookRetryAttempts,omitempty"`
	TimeoutSeconds                int               `bson:"timeoutSeconds" json:"timeoutSeconds"`
	RetryAttempts                 int               `bson:"retryAttempts" json:"retryAttempts"`
	Headers                       map[string]string `bson:"headers,omitempty" json:"headers,omitempty"`
	// Multilingual configuration
	MultilingualSupport *bool    `bson:"multilingualSupport,omitempty" json:"multilingualSupport,omitempty"`
	SupportedLanguages  []string `bson:"supportedLanguages,omitempty" json:"supportedLanguages,omitempty"`
	UseDeepLTranslation *bool    `bson:"useDeepLTranslation,omitempty" json:"useDeepLTranslation,omitempty"`
	// Appointment services language handling
	AppointmentServicesLanguageHandling *string `bson:"appointmentServicesLanguageHandling,omitempty" json:"appointmentServicesLanguageHandling,omitempty"`
	AppointmentServicesLanguageHeader   *string `bson:"appointmentServicesLanguageHeader,omitempty" json:"appointmentServicesLanguageHeader,omitempty"`
	// Generic services language handling
	GenericServicesLanguageHandling *string `bson:"genericServicesLanguageHandling,omitempty" json:"genericServicesLanguageHandling,omitempty"`
	GenericServicesLanguageHeader   *string `bson:"genericServicesLanguageHeader,omitempty" json:"genericServicesLanguageHeader,omitempty"`
}

// GenericService represents a generic service that can be created by admin
type GenericService struct {
	ID          string `bson:"id" json:"id"`
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	Duration    int    `bson:"duration,omitempty" json:"duration,omitempty"` // Duration in minutes
	Enabled     bool   `bson:"enabled" json:"enabled"`
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
// NOTE: Now stored in tenant-specific databases, so no tenantId field needed
type CardReaderStatus struct {
	ID        string    `bson:"id" json:"id"`
	SectionID *int64    `bson:"sectionId,omitempty" json:"sectionId,omitempty"` // Optional: which section this reader belongs to
	Name      string    `bson:"name" json:"name"`
	Status    string    `bson:"status" json:"status"` // "online", "offline", "error"
	LastSeen  time.Time `bson:"lastSeen" json:"lastSeen"`
	IPAddress string    `bson:"ipAddress,omitempty" json:"ipAddress,omitempty"`
	Version   string    `bson:"version,omitempty" json:"version,omitempty"`
	LastError string    `bson:"lastError,omitempty" json:"lastError,omitempty"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// Tenant represents a tenant in the system (stored in master database)
type Tenant struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	TenantID     int64     `bson:"tenantId" json:"tenantId"`           // Numeric tenant ID (auto-increment)
	Name         string    `bson:"name" json:"name"`                   // Tenant name (e.g., "Hospital A")
	Description  string    `bson:"description,omitempty" json:"description,omitempty"`
	DatabaseName string    `bson:"databaseName" json:"databaseName"`   // Name of tenant's dedicated database
	Status       string    `bson:"status" json:"status"`               // "active" or "inactive"
	CreatedAt    time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time `bson:"updatedAt" json:"updatedAt"`
}

// Section represents a sub-tenant (e.g., department) within a tenant's database
type Section struct {
	ID          string    `bson:"_id,omitempty" json:"id"`
	SectionID   int64     `bson:"sectionId" json:"sectionId"`         // Numeric section ID (unique within tenant)
	Name        string    `bson:"name" json:"name"`                   // Section name (e.g., "Cardiology")
	Description string    `bson:"description,omitempty" json:"description,omitempty"`
	Status      string    `bson:"status" json:"status"`               // "active" or "inactive"
	CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
}

// ParseTenantAndSectionID parses a composite ID string in format "tenantId:sectionId"
// Returns tenantID (numeric), sectionID (numeric, 0 if not present), and error if invalid
func ParseTenantAndSectionID(compositeID string) (tenantID int64, sectionID int64, err error) {
	if compositeID == "" {
		return 0, 0, fmt.Errorf("composite ID must not be empty")
	}

	var tenantIDStr, sectionIDStr string

	// Find the colon separator
	colonIdx := -1
	for i, char := range compositeID {
		if char == ':' {
			colonIdx = i
			break
		}
	}

	if colonIdx == -1 {
		// No colon, just tenant ID
		tenantIDStr = compositeID
	} else {
		tenantIDStr = compositeID[:colonIdx]
		sectionIDStr = compositeID[colonIdx+1:]
	}

	// Parse tenant ID
	var parsedTenantID int64
	_, err = fmt.Sscanf(tenantIDStr, "%d", &parsedTenantID)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid tenant ID format: %w", err)
	}

	// Parse section ID if present
	var parsedSectionID int64
	if sectionIDStr != "" {
		_, err = fmt.Sscanf(sectionIDStr, "%d", &parsedSectionID)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid section ID format: %w", err)
		}
	}

	return parsedTenantID, parsedSectionID, nil
}
