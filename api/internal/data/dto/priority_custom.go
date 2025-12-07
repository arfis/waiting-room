package dto

import (
	"encoding/json"
)

// UnmarshalJSON implements custom JSON unmarshaling for PatientInformation
// to handle flexible date-time formats for appointmentTime
func (p *PatientInformation) UnmarshalJSON(data []byte) error {
	// Create a temporary struct with FlexibleTime for unmarshaling
	type Alias struct {
		Age             *int64        `json:"age,omitempty"`
		AppointmentTime *FlexibleTime `json:"appointmentTime,omitempty"`
		ManualOverride  *float64      `json:"manualOverride,omitempty"`
		Symbols         []string      `json:"symbols,omitempty"`
	}

	var temp Alias
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy values from temp to p
	p.Age = temp.Age
	p.ManualOverride = temp.ManualOverride
	p.Symbols = temp.Symbols

	// Convert FlexibleTime to *time.Time
	if temp.AppointmentTime != nil {
		t := temp.AppointmentTime.Time
		p.AppointmentTime = &t
	}

	return nil
}
