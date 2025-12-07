package dto

import (
	"fmt"
	"strings"
	"time"
)

// FlexibleTime is a custom type that can unmarshal various date-time formats
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements custom unmarshaling to handle multiple date-time formats
func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		return nil
	}

	// Try various formats
	formats := []string{
		time.RFC3339,                 // 2025-12-07T11:45:00Z
		"2006-01-02T15:04:05Z07:00",  // 2025-12-07T11:45:00+01:00
		"2006-01-02T15:04:05",        // 2025-12-07T11:45:00
		"2006-01-02T15:04",           // 2025-12-07T11:45
		"2006-01-02 15:04:05",        // 2025-12-07 11:45:00
		"2006-01-02 15:04",           // 2025-12-07 11:45
		"2006-01-02",                 // 2025-12-07
	}

	var parseErr error
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			ft.Time = t
			return nil
		}
		parseErr = err
	}

	return fmt.Errorf("unable to parse time %q: %w", s, parseErr)
}

// MarshalJSON implements custom marshaling to output RFC3339 format
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	if ft.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + ft.Time.Format(time.RFC3339) + `"`), nil
}
