package queue

import (
	"context"
	"testing"
	"time"

	"github.com/arfis/waiting-room/internal/config"
	"github.com/arfis/waiting-room/internal/priority"
	"github.com/arfis/waiting-room/internal/repository"
)

// Helper functions for creating pointers
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// TestCreateEntry_WithPriorityCalculation tests the integration between CreateEntry and priority calculator
func TestCreateEntry_WithPriorityCalculation(t *testing.T) {
	// Create a mock repository
	mockRepo := repository.NewMockQueueRepository()
	cfg := &config.Config{}

	// Create a mock priority repository (nil is acceptable for testing as we'll use GetDefaultConfig)
	// The priority repo is used to fetch config, but in tests it will default to GetDefaultConfig()
	var priorityRepo *priority.Repository // nil is acceptable for testing

	// Create waiting queue
	wq := NewWaitingQueue(mockRepo, cfg, nil, priorityRepo)

	ctx := context.Background()

	tests := []struct {
		name            string
		roomId          string
		cardData        CardData
		duration        int64
		serviceName     string
		symbols         []string
		appointmentTime *time.Time
		age             *int
		manualOverride  *float64
		expectedTier    int
		checkScore      func(float64) bool
	}{
		{
			name:   "STATIM patient should get tier 0",
			roomId: "triage-1",
			cardData: CardData{
				IDNumber:  "123456789",
				FirstName: "John",
				LastName:  "Doe",
			},
			duration:     300,
			serviceName:  "Emergency",
			symbols:      []string{"STATIM"},
			expectedTier: 0,
			checkScore: func(score float64) bool {
				// STATIM symbol = -1000
				return score <= -1000
			},
		},
		{
			name:   "VIP patient should get tier 1",
			roomId: "triage-1",
			cardData: CardData{
				IDNumber:  "987654321",
				FirstName: "Jane",
				LastName:  "Smith",
			},
			duration:     300,
			serviceName:  "Regular",
			symbols:      []string{"VIP"},
			expectedTier: 1,
			checkScore: func(score float64) bool {
				// VIP symbol = -500
				return score <= -500 && score > -1000
			},
		},
		{
			name:   "VIP elderly patient should have lower score",
			roomId: "triage-1",
			cardData: CardData{
				IDNumber:  "111222333",
				FirstName: "Senior",
				LastName:  "Citizen",
			},
			duration:     300,
			serviceName:  "Regular",
			symbols:      []string{"VIP"},
			age:          intPtr(75),
			expectedTier: 1,
			checkScore: func(score float64) bool {
				// VIP (-500) + age bonus (75-65)*-1 = -10
				// Total should be around -510
				return score < -500
			},
		},
		{
			name:   "Regular patient with young child",
			roomId: "triage-1",
			cardData: CardData{
				IDNumber:  "444555666",
				FirstName: "Young",
				LastName:  "Child",
			},
			duration:     300,
			serviceName:  "Pediatrics",
			symbols:      []string{},
			age:          intPtr(2),
			expectedTier: 2,
			checkScore: func(score float64) bool {
				// Age bonus (6-2)*-5 = -20
				return score <= -20 && score > -30
			},
		},
		{
			name:   "Regular patient with manual override",
			roomId: "triage-1",
			cardData: CardData{
				IDNumber:  "777888999",
				FirstName: "Manual",
				LastName:  "Override",
			},
			duration:       300,
			serviceName:    "Regular",
			symbols:        []string{},
			manualOverride: float64Ptr(-100),
			expectedTier:   2,
			checkScore: func(score float64) bool {
				// Manual override = -100
				return score <= -100 && score > -110
			},
		},
		{
			name:   "STATIM + VIP patient (STATIM takes precedence)",
			roomId: "triage-1",
			cardData: CardData{
				IDNumber:  "000111222",
				FirstName: "Emergency",
				LastName:  "VIP",
			},
			duration:     300,
			serviceName:  "Emergency",
			symbols:      []string{"STATIM", "VIP"},
			expectedTier: 0,
			checkScore: func(score float64) bool {
				// STATIM (-1000) + VIP (-500) = -1500
				return score <= -1500
			},
		},
		{
			name:   "Patient with appointment - late arrival",
			roomId: "triage-1",
			cardData: CardData{
				IDNumber:  "333444555",
				FirstName: "Late",
				LastName:  "Patient",
			},
			duration:        300,
			serviceName:     "Appointment",
			symbols:         []string{},
			appointmentTime: timePtr(time.Now().Add(-15 * time.Minute)), // 15 minutes late
			expectedTier:    2,
			checkScore: func(score float64) bool {
				// Late bonus: 15 * -3 = -45
				return score < 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create entry
			entry, err := wq.CreateEntry(ctx, tt.roomId, tt.cardData, tt.duration, tt.serviceName,
				tt.symbols, tt.appointmentTime, tt.age, tt.manualOverride)

			if err != nil {
				t.Fatalf("CreateEntry() error = %v", err)
			}

			// Check tier
			if entry.Tier != tt.expectedTier {
				t.Errorf("CreateEntry() tier = %v, want %v", entry.Tier, tt.expectedTier)
			}

			// Check fitness score
			if !tt.checkScore(entry.FitnessScore) {
				t.Errorf("CreateEntry() fitness score = %v, does not match expectations", entry.FitnessScore)
			}

			// Verify entry was created
			if entry.ID == "" {
				t.Error("CreateEntry() entry ID is empty")
			}

			if entry.TicketNumber == "" {
				t.Error("CreateEntry() ticket number is empty")
			}

			t.Logf("Created entry: ID=%s, Ticket=%s, Tier=%d, Score=%.2f",
				entry.ID, entry.TicketNumber, entry.Tier, entry.FitnessScore)
		})
	}
}

// TestCreateMultipleEntries_PriorityOrdering tests that multiple entries are ordered correctly by priority
func TestCreateMultipleEntries_PriorityOrdering(t *testing.T) {
	// Create a mock repository
	mockRepo := repository.NewMockQueueRepository()
	cfg := &config.Config{}
	var priorityRepo *priority.Repository // nil is acceptable for testing
	wq := NewWaitingQueue(mockRepo, cfg, nil, priorityRepo)

	ctx := context.Background()
	roomId := "triage-1"

	// Create multiple patients in different priority tiers
	patients := []struct {
		name       string
		cardData   CardData
		symbols    []string
		age        *int
		expectTier int
	}{
		{
			name: "Regular adult",
			cardData: CardData{
				IDNumber:  "100",
				FirstName: "Regular",
				LastName:  "Adult",
			},
			symbols:    []string{},
			age:        intPtr(30),
			expectTier: 2,
		},
		{
			name: "STATIM patient",
			cardData: CardData{
				IDNumber:  "200",
				FirstName: "Emergency",
				LastName:  "STATIM",
			},
			symbols:    []string{"STATIM"},
			expectTier: 0,
		},
		{
			name: "VIP elderly",
			cardData: CardData{
				IDNumber:  "300",
				FirstName: "VIP",
				LastName:  "Senior",
			},
			symbols:    []string{"VIP"},
			age:        intPtr(80),
			expectTier: 1,
		},
		{
			name: "Regular child",
			cardData: CardData{
				IDNumber:  "400",
				FirstName: "Young",
				LastName:  "Child",
			},
			symbols:    []string{},
			age:        intPtr(3),
			expectTier: 2,
		},
	}

	// Create all entries
	var entries []*Entry
	for _, p := range patients {
		entry, err := wq.CreateEntry(ctx, roomId, p.cardData, 300, "Service",
			p.symbols, nil, p.age, nil)
		if err != nil {
			t.Fatalf("Failed to create entry for %s: %v", p.name, err)
		}

		if entry.Tier != p.expectTier {
			t.Errorf("%s: expected tier %d, got %d", p.name, p.expectTier, entry.Tier)
		}

		entries = append(entries, entry)
		t.Logf("Created %s: Tier=%d, Score=%.2f", p.name, entry.Tier, entry.FitnessScore)
	}

	// Get all entries from repository
	allEntries, err := mockRepo.GetQueueEntries(ctx, roomId, []string{"WAITING"})
	if err != nil {
		t.Fatalf("Failed to get queue entries: %v", err)
	}

	if len(allEntries) != len(patients) {
		t.Errorf("Expected %d entries, got %d", len(patients), len(allEntries))
	}

	// Verify ordering: STATIM (tier 0) should be first, VIP (tier 1) second, then regular (tier 2)
	if len(allEntries) >= 2 {
		// First should be STATIM (tier 0)
		if allEntries[0].Tier != 0 {
			t.Errorf("First entry should be tier 0 (STATIM), got tier %d", allEntries[0].Tier)
		}

		// Second should be VIP (tier 1)
		if allEntries[1].Tier != 1 {
			t.Errorf("Second entry should be tier 1 (VIP), got tier %d", allEntries[1].Tier)
		}
	}

	// Log final ordering
	t.Log("Final queue order:")
	for i, entry := range allEntries {
		t.Logf("  %d. Ticket=%s, Tier=%d, Score=%.2f, Position=%d",
			i+1, entry.TicketNumber, entry.Tier, entry.FitnessScore, entry.Position)
	}
}

// TestCreateEntry_WithWaitingTime tests that waiting time affects fitness score
func TestCreateEntry_WithWaitingTime(t *testing.T) {
	mockRepo := repository.NewMockQueueRepository()
	cfg := &config.Config{}
	var priorityRepo *priority.Repository // nil is acceptable for testing
	wq := NewWaitingQueue(mockRepo, cfg, nil, priorityRepo)

	ctx := context.Background()
	roomId := "triage-1"

	// Create first entry
	cardData1 := CardData{IDNumber: "111", FirstName: "First", LastName: "Patient"}
	entry1, err := wq.CreateEntry(ctx, roomId, cardData1, 300, "Service", []string{}, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create first entry: %v", err)
	}

	// Wait a bit to simulate passage of time
	time.Sleep(100 * time.Millisecond)

	// Create second entry (same priority tier, but entered later)
	cardData2 := CardData{IDNumber: "222", FirstName: "Second", LastName: "Patient"}
	entry2, err := wq.CreateEntry(ctx, roomId, cardData2, 300, "Service", []string{}, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create second entry: %v", err)
	}

	// First entry should have slightly lower score due to waiting time
	// (though the difference might be negligible in this test)
	if entry1.Tier != entry2.Tier {
		t.Errorf("Both entries should be in same tier, got %d and %d", entry1.Tier, entry2.Tier)
	}

	t.Logf("Entry1: Score=%.2f, Entry2: Score=%.2f", entry1.FitnessScore, entry2.FitnessScore)

	// Get all entries to verify order
	allEntries, err := mockRepo.GetQueueEntries(ctx, roomId, []string{"WAITING"})
	if err != nil {
		t.Fatalf("Failed to get queue entries: %v", err)
	}

	if len(allEntries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(allEntries))
	}

	// First created entry should still be first (FIFO within same tier and similar score)
	if allEntries[0].CardData.IDNumber != "111" {
		t.Errorf("Expected first entry to be '111', got '%s'", allEntries[0].CardData.IDNumber)
	}
}

// TestCreateEntry_PersistsMetadata tests that priority metadata is correctly persisted
func TestCreateEntry_PersistsMetadata(t *testing.T) {
	mockRepo := repository.NewMockQueueRepository()
	cfg := &config.Config{}
	var priorityRepo *priority.Repository // nil is acceptable for testing
	wq := NewWaitingQueue(mockRepo, cfg, nil, priorityRepo)

	ctx := context.Background()
	roomId := "triage-1"

	appointmentTime := time.Now().Add(1 * time.Hour)
	age := 45
	manualOverride := -25.5
	symbols := []string{"VIP", "IMMOBILE"}

	cardData := CardData{
		IDNumber:  "METADATA_TEST",
		FirstName: "Meta",
		LastName:  "Data",
	}

	entry, err := wq.CreateEntry(ctx, roomId, cardData, 600, "Test Service",
		symbols, &appointmentTime, &age, &manualOverride)
	if err != nil {
		t.Fatalf("CreateEntry() error = %v", err)
	}

	// Verify all metadata is persisted
	if len(entry.Symbols) != 2 || entry.Symbols[0] != "VIP" || entry.Symbols[1] != "IMMOBILE" {
		t.Errorf("Symbols not persisted correctly: %v", entry.Symbols)
	}

	if entry.AppointmentTime == nil || !entry.AppointmentTime.Equal(appointmentTime) {
		t.Errorf("AppointmentTime not persisted correctly: %v", entry.AppointmentTime)
	}

	if entry.Age == nil || *entry.Age != age {
		t.Errorf("Age not persisted correctly: %v", entry.Age)
	}

	if entry.ManualOverride == nil || *entry.ManualOverride != manualOverride {
		t.Errorf("ManualOverride not persisted correctly: %v", entry.ManualOverride)
	}

	// Verify calculated fields
	if entry.Tier != 1 { // VIP tier
		t.Errorf("Expected tier 1 (VIP), got %d", entry.Tier)
	}

	if entry.FitnessScore == 0 {
		t.Error("FitnessScore should not be zero")
	}

	t.Logf("Entry created with all metadata: Tier=%d, Score=%.2f", entry.Tier, entry.FitnessScore)
}
