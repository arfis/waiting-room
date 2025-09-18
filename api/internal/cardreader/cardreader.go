package cardreader

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// CardData represents the data read from a national ID card
type CardData struct {
	IDNumber    string    `json:"id_number"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Nationality string    `json:"nationality"`
	Address     string    `json:"address"`
	IssuedDate  time.Time `json:"issued_date"`
	ExpiryDate  time.Time `json:"expiry_date"`
	Photo       string    `json:"photo,omitempty"` // Base64 encoded photo
	ReadTime    time.Time `json:"read_time"`
}

// CardReader interface for different types of card readers
type CardReader interface {
	Initialize() error
	ReadCard() (*CardData, error)
	Close() error
	IsConnected() bool
}

// MockCardReader simulates a card reader for development/testing
type MockCardReader struct {
	connected bool
}

func NewMockCardReader() *MockCardReader {
	return &MockCardReader{connected: true}
}

func (m *MockCardReader) Initialize() error {
	log.Println("Mock card reader initialized")
	return nil
}

func (m *MockCardReader) ReadCard() (*CardData, error) {
	if !m.connected {
		return nil, fmt.Errorf("card reader not connected")
	}

	// Simulate reading delay
	time.Sleep(2 * time.Second)

	// Return mock data
	return &CardData{
		IDNumber:    "1234567890123",
		FirstName:   "John",
		LastName:    "Doe",
		DateOfBirth: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		Gender:      "Male",
		Nationality: "Czech",
		Address:     "Prague, Czech Republic",
		IssuedDate:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		ExpiryDate:  time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
		ReadTime:    time.Now(),
	}, nil
}

func (m *MockCardReader) Close() error {
	m.connected = false
	log.Println("Mock card reader closed")
	return nil
}

func (m *MockCardReader) IsConnected() bool {
	return m.connected
}

// Service manages card reader operations
type Service struct {
	reader CardReader
}

func NewService() *Service {
	// For now, use mock reader. In production, you would initialize
	// the appropriate reader based on configuration
	reader := NewMockCardReader()
	return &Service{reader: reader}
}

func (s *Service) Initialize() error {
	return s.reader.Initialize()
}

func (s *Service) ReadCard() (*CardData, error) {
	return s.reader.ReadCard()
}

func (s *Service) IsConnected() bool {
	return s.reader.IsConnected()
}

func (s *Service) Close() error {
	return s.reader.Close()
}

// ToJSON converts CardData to JSON string
func (c *CardData) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
