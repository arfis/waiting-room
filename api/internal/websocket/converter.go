package websocket

import (
	"time"

	"github.com/arfis/waiting-room/internal/data/dto"
)

// convertEntriesToWebSocketFormat converts DTO entries to WebSocket message format
func convertEntriesToWebSocketFormat(entries []dto.QueueEntry) []map[string]interface{} {
	var wsEntries []map[string]interface{}
	for _, entry := range entries {
		wsEntry := map[string]interface{}{
			"id":            entry.ID,
			"waitingRoomId": entry.WaitingRoomID,
			"ticketNumber":  entry.TicketNumber,
			"status":        entry.Status,
			"position":      entry.Position,
		}

		// Add optional fields
		if entry.ServicePoint != nil {
			wsEntry["servicePoint"] = *entry.ServicePoint
		}
		if entry.ServiceName != nil {
			wsEntry["serviceName"] = *entry.ServiceName
		}
		if entry.ServiceDuration != nil {
			wsEntry["serviceDuration"] = *entry.ServiceDuration
		}
		if entry.Age != nil {
			wsEntry["age"] = *entry.Age
		}
		if len(entry.Symbols) > 0 {
			wsEntry["symbols"] = entry.Symbols
		}
		if entry.AppointmentTime != nil {
			wsEntry["appointmentTime"] = entry.AppointmentTime.Time.Format(time.RFC3339)
		}

		// Add timestamps from the entry
		if !entry.CreatedAt.IsZero() {
			wsEntry["createdAt"] = entry.CreatedAt.Format(time.RFC3339)
			wsEntry["updatedAt"] = entry.CreatedAt.Format(time.RFC3339) // Use createdAt as updatedAt for now
		}

		wsEntries = append(wsEntries, wsEntry)
	}
	return wsEntries
}
