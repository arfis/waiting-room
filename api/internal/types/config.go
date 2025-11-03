package types

import (
	"fmt"
	"time"
)

// SystemConfiguration represents the complete system configuration stored in MongoDB
type SystemConfiguration struct {
	ID            string            `bson:"_id,omitempty" json:"id"`
	TenantID      string            `bson:"tenantId,omitempty" json:"tenantId,omitempty"` // Building/Hospital ID (e.g., "Nemocnica Spiska nova ves")
	SectionID     string            `bson:"sectionId,omitempty" json:"sectionId,omitempty"` // Section/Department within tenant (e.g., "Kardiologia pavilon B", "Dentist")
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
type CardReaderStatus struct {
	ID        string    `bson:"id" json:"id"`
	TenantID  string    `bson:"tenantId,omitempty" json:"tenantId,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Status    string    `bson:"status" json:"status"` // "online", "offline", "error"
	LastSeen  time.Time `bson:"lastSeen" json:"lastSeen"`
	IPAddress string    `bson:"ipAddress,omitempty" json:"ipAddress,omitempty"`
	Version   string    `bson:"version,omitempty" json:"version,omitempty"`
	LastError string    `bson:"lastError,omitempty" json:"lastError,omitempty"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// Tenant represents a tenant in the system
type Tenant struct {
	ID          string    `bson:"id" json:"id"`
	BuildingID  string    `bson:"buildingId" json:"buildingId"`
	SectionID   string    `bson:"sectionId" json:"sectionId"`
	Name        string    `bson:"name" json:"name"`
	Description string    `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
}

// GetTenantID returns the tenant ID (just the buildingId - the hospital/building)
// Note: This is now just the buildingId, not buildingId:sectionId
// The sectionId is a separate field that identifies departments within the tenant
func (t *Tenant) GetTenantID() string {
	return t.BuildingID
}

// GetFullTenantID returns the full identifier in format "buildingId:sectionId" for backwards compatibility
// This is used when sending the tenant ID in headers
func (t *Tenant) GetFullTenantID() string {
	return t.BuildingID + ":" + t.SectionID
}

// ParseTenantID parses a tenant ID string in the format "buildingId:sectionId"
// Returns buildingId (tenant), sectionId (section/department), and an error if the format is invalid
// For backwards compatibility, if no colon is present, treats the entire string as buildingId
func ParseTenantID(tenantID string) (buildingID, sectionID string, err error) {
	if tenantID == "" {
		return "", "", fmt.Errorf("invalid tenant ID format: tenant ID must not be empty")
	}
	
	// Check if it contains a colon (format: "buildingId:sectionId")
	for i, char := range tenantID {
		if char == ':' {
			buildingID = tenantID[:i]
			sectionID = tenantID[i+1:]
			if buildingID == "" {
				return "", "", fmt.Errorf("invalid tenant ID format: building ID must not be empty")
			}
			// sectionID can be empty (for tenant-level configs)
			return buildingID, sectionID, nil
		}
	}
	// If no colon, treat the entire string as buildingId (tenant only, no section)
	return tenantID, "", nil
}
