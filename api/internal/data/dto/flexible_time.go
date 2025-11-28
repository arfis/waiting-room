package dto

import (
	"fmt"
	"strings"
	"time"
)

// FlexibleTime is a custom time type that can unmarshal from multiple datetime formats
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling to support multiple datetime formats
func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	if s == "null" || s == "" {
		return nil
	}

	// List of datetime formats to try (in order of preference)
	formats := []string{
		time.RFC3339,                // "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano,            // "2006-01-02T15:04:05.999999999Z07:00"
		"2006-01-02T15:04:05",       // Without timezone
		"2006-01-02T15:04",          // Without seconds and timezone (the problematic format)
		"2006-01-02T15:04Z07:00",    // With timezone but no seconds
		"2006-01-02T15:04Z",         // With Z timezone but no seconds
		time.DateTime,               // "2006-01-02 15:04:05"
		time.DateOnly,               // "2006-01-02"
	}

	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			ft.Time = t
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("unable to parse datetime '%s': %w", s, lastErr)
}

// MarshalJSON implements custom JSON marshaling to always output RFC3339 format
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	if ft.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ft.Format(time.RFC3339))), nil
}

// ToTimePtr converts FlexibleTime to *time.Time for compatibility
func (ft *FlexibleTime) ToTimePtr() *time.Time {
	if ft == nil || ft.IsZero() {
		return nil
	}
	return &ft.Time
}
