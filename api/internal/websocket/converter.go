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

		// Add timestamps (these will be added by the service in the future if needed)
		wsEntry["createdAt"] = time.Now().Format(time.RFC3339)
		wsEntry["updatedAt"] = time.Now().Format(time.RFC3339)

		wsEntries = append(wsEntries, wsEntry)
	}
	return wsEntries
}
