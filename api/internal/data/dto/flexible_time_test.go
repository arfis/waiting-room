package dto

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFlexibleTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Full RFC3339 format",
			input:   `"2025-04-09T19:18:00Z"`,
			wantErr: false,
		},
		{
			name:    "Without seconds and timezone (problematic format)",
			input:   `"2025-04-09T19:18"`,
			wantErr: false,
		},
		{
			name:    "Without timezone",
			input:   `"2025-04-09T19:18:00"`,
			wantErr: false,
		},
		{
			name:    "With timezone offset",
			input:   `"2025-04-09T19:18:00+02:00"`,
			wantErr: false,
		},
		{
			name:    "Date only",
			input:   `"2025-04-09"`,
			wantErr: false,
		},
		{
			name:    "Null value",
			input:   `null`,
			wantErr: false,
		},
		{
			name:    "Empty string",
			input:   `""`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ft FlexibleTime
			err := json.Unmarshal([]byte(tt.input), &ft)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlexibleTime.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && err == nil && tt.input != `null` && tt.input != `""` {
				t.Logf("Successfully parsed: %s -> %v", tt.input, ft.Time)
			}
		})
	}
}

func TestPatientInformation_UnmarshalJSON(t *testing.T) {
	// Test the full PatientInformation struct with flexible datetime
	jsonStr := `{
		"age": 3,
		"appointmentTime": "2025-04-09T19:18",
		"symbols": ["VIP"],
		"manualOverride": 1.5
	}`

	var patientInfo PatientInformation
	err := json.Unmarshal([]byte(jsonStr), &patientInfo)
	if err != nil {
		t.Fatalf("Failed to unmarshal PatientInformation: %v", err)
	}

	if patientInfo.Age == nil || *patientInfo.Age != 3 {
		t.Errorf("Expected age 3, got %v", patientInfo.Age)
	}

	if patientInfo.AppointmentTime == nil {
		t.Fatal("Expected appointmentTime to be non-nil")
	}

	expectedYear, expectedMonth, expectedDay := 2025, time.April, 9
	if patientInfo.AppointmentTime.Year() != expectedYear ||
		patientInfo.AppointmentTime.Month() != expectedMonth ||
		patientInfo.AppointmentTime.Day() != expectedDay {
		t.Errorf("Expected date 2025-04-09, got %v", patientInfo.AppointmentTime.Time)
	}

	if len(patientInfo.Symbols) != 1 || patientInfo.Symbols[0] != "VIP" {
		t.Errorf("Expected symbols [VIP], got %v", patientInfo.Symbols)
	}

	t.Logf("Successfully parsed PatientInformation: %+v", patientInfo)
}

func TestSwipeRequest_UnmarshalJSON(t *testing.T) {
	// Test the full SwipeRequest with embedded PatientInformation
	jsonStr := `{
		"idCardRaw": "314494288",
		"patientInformation": {
			"symbols": ["VIP"],
			"appointmentTime": "2025-04-09T19:18",
			"age": 3
		}
	}`

	var swipeReq SwipeRequest
	err := json.Unmarshal([]byte(jsonStr), &swipeReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal SwipeRequest: %v", err)
	}

	if swipeReq.IdCardRaw == nil || *swipeReq.IdCardRaw != "314494288" {
		t.Errorf("Expected idCardRaw '314494288', got %v", swipeReq.IdCardRaw)
	}

	if swipeReq.PatientInformation == nil {
		t.Fatal("Expected patientInformation to be non-nil")
	}

	patientInfo := swipeReq.PatientInformation
	if patientInfo.Age == nil || *patientInfo.Age != 3 {
		t.Errorf("Expected age 3, got %v", patientInfo.Age)
	}

	if patientInfo.AppointmentTime == nil {
		t.Fatal("Expected appointmentTime to be non-nil")
	}

	t.Logf("Successfully parsed SwipeRequest: %+v", swipeReq)
	t.Logf("AppointmentTime: %v", patientInfo.AppointmentTime.Time)
}
