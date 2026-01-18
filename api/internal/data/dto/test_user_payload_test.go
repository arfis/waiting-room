package dto

import (
	"encoding/json"
	"testing"
)

// Test with the exact payload the user is sending
func TestUserSwipePayload(t *testing.T) {
	// This is the exact payload the user reported
	jsonStr := `{"idCardRaw":"657798746","patientInformation":{"symbols":["STATIM"],"appointmentTime":"2025-12-04T18:54"}}`

	var swipeReq SwipeRequest
	err := json.Unmarshal([]byte(jsonStr), &swipeReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal user's SwipeRequest: %v", err)
	}

	// Verify it parsed correctly
	if swipeReq.IdCardRaw == nil || *swipeReq.IdCardRaw != "657798746" {
		t.Errorf("Expected idCardRaw '657798746', got %v", swipeReq.IdCardRaw)
	}

	if swipeReq.PatientInformation == nil {
		t.Fatal("Expected patientInformation to be non-nil")
	}

	patientInfo := swipeReq.PatientInformation
	if len(patientInfo.Symbols) != 1 || patientInfo.Symbols[0] != "STATIM" {
		t.Errorf("Expected symbols [STATIM], got %v", patientInfo.Symbols)
	}

	if patientInfo.AppointmentTime == nil {
		t.Fatal("Expected appointmentTime to be non-nil")
	}

	// Check date is December 4, 2025 at 18:54
	expectedYear, expectedMonth, expectedDay := 2025, 12, 4
	expectedHour, expectedMinute := 18, 54

	if patientInfo.AppointmentTime.Year() != expectedYear ||
		int(patientInfo.AppointmentTime.Month()) != expectedMonth ||
		patientInfo.AppointmentTime.Day() != expectedDay ||
		patientInfo.AppointmentTime.Hour() != expectedHour ||
		patientInfo.AppointmentTime.Minute() != expectedMinute {
		t.Errorf("Expected datetime 2025-12-04 18:54, got %v", patientInfo.AppointmentTime)
	}

	t.Logf("✓ Successfully parsed user's payload: %+v", swipeReq)
	t.Logf("✓ AppointmentTime: %v", patientInfo.AppointmentTime)
}
