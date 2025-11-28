package websocket

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"github.com/arfis/waiting-room/internal/middleware"
	queueService "github.com/arfis/waiting-room/internal/service/queue"
)

// ClientInfo stores information about a WebSocket client
type ClientInfo struct {
	conn     *websocket.Conn
	tenantID string // tenantID from query parameter or header
}

// Hub manages WebSocket connections and broadcasts
type Hub struct {
	queueService *queueService.Service
	upgrader     websocket.Upgrader
	// clients structure: roomId -> tenantID -> []*ClientInfo
	// This allows us to efficiently find all clients for a specific tenant
	clients    map[string]map[string][]*ClientInfo
	clientsMux sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub(queueService *queueService.Service) *Hub {
	return &Hub{
		queueService: queueService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients: make(map[string]map[string][]*ClientInfo),
	}
}

// HandleConnection handles a WebSocket connection for queue updates
func (h *Hub) HandleConnection(w http.ResponseWriter, r *http.Request) {
	log.Printf("WebSocket handler called for room: %s", chi.URLParam(r, "roomId"))

	roomId := chi.URLParam(r, "roomId")
	if roomId == "" {
		log.Printf("Room ID is empty")
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	// Extract tenant ID from query parameter or header
	tenantID := extractTenantID(r)
	log.Printf("[WebSocket] Attempting to upgrade WebSocket connection for room: %s, tenantID: '%s'", roomId, tenantID)

	// Check if the response writer supports hijacking
	if _, ok := w.(http.Hijacker); !ok {
		log.Printf("ERROR: Response writer does not implement http.Hijacker")
		http.Error(w, "WebSocket not supported", http.StatusInternalServerError)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}
	defer conn.Close()

	// Normalize tenant ID: trim whitespace for consistency
	normalizedTenantID := strings.TrimSpace(tenantID)

	// Store client info with normalized tenantID
	clientInfo := &ClientInfo{
		conn:     conn,
		tenantID: normalizedTenantID,
	}

	// Use normalized tenant ID as key (use "default" for empty tenant ID)
	tenantKey := normalizedTenantID
	if tenantKey == "" {
		tenantKey = "default"
		log.Printf("[WebSocket] Empty tenantID provided, using 'default' as key")
	}

	log.Printf("[WebSocket] Normalized tenantID: '%s' -> '%s' (key: '%s')", tenantID, normalizedTenantID, tenantKey)

	// Add client to room, organized by tenantID
	h.addClient(roomId, tenantKey, clientInfo)

	log.Printf("[WebSocket] Client connected to room: %s, tenantID: '%s' (stored under key: '%s')", roomId, tenantID, tenantKey)

	// Send initial queue data to the newly connected client
	go h.sendInitialData(conn, roomId, normalizedTenantID, tenantKey)

	// Remove client when connection closes
	defer h.removeClient(roomId, tenantKey, conn)

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// sendInitialData sends initial queue data to a newly connected client
func (h *Hub) sendInitialData(conn *websocket.Conn, roomId, normalizedTenantID, tenantKey string) {
	// Small delay to ensure the client is fully connected
	time.Sleep(100 * time.Millisecond)
	log.Printf("[WebSocket] Sending initial queue data to newly connected client for tenantID: '%s' (key: '%s')", normalizedTenantID, tenantKey)

	// Create context with normalized tenantID for filtering
	ctx := context.Background()
	if normalizedTenantID != "" && normalizedTenantID != "default" {
		ctx = context.WithValue(ctx, middleware.TENANT, normalizedTenantID)
		log.Printf("[WebSocket] Using normalized tenantID '%s' for filtering initial data", normalizedTenantID)
	} else {
		log.Printf("[WebSocket] WARNING: normalized tenantID is empty or 'default', will get all entries (no tenant filter)")
	}

	// Get queue entries from service
	entries, err := h.queueService.GetQueueEntries(ctx, roomId, []string{"WAITING", "CALLED", "IN_SERVICE"})
	if err != nil {
		log.Printf("[WebSocket] Failed to get initial queue entries for tenantID '%s': %v", normalizedTenantID, err)
		return
	}

	log.Printf("[WebSocket] Retrieved %d initial entries for tenantID '%s' (key: '%s') in room %s", len(entries), normalizedTenantID, tenantKey, roomId)

	// Convert to WebSocket format
	wsEntries := convertEntriesToWebSocketFormat(entries)

	message := map[string]interface{}{
		"type":    "queue_update",
		"roomId":  roomId,
		"entries": wsEntries,
	}

	// Send to only this specific client
	h.clientsMux.RLock()
	var foundClient *ClientInfo
	if roomClients, exists := h.clients[roomId]; exists {
		if tenantClients, exists := roomClients[tenantKey]; exists {
			for _, client := range tenantClients {
				if client.conn == conn {
					foundClient = client
					break
				}
			}
		}
	}
	h.clientsMux.RUnlock()

	if foundClient != nil {
		if err := foundClient.conn.WriteJSON(message); err != nil {
			log.Printf("[WebSocket] Failed to send initial queue data: %v", err)
		} else {
			log.Printf("[WebSocket] Successfully sent initial queue data (%d entries) to client with tenantID: '%s'", len(wsEntries), normalizedTenantID)
		}
	} else {
		log.Printf("[WebSocket] Client not found in room %s for tenant '%s' for initial data send", roomId, tenantKey)
	}
}

// BroadcastQueueUpdate broadcasts queue update to clients for a specific tenant
func (h *Hub) BroadcastQueueUpdate(roomId string, targetTenantID string) {
	h.clientsMux.RLock()
	roomClients, roomExists := h.clients[roomId]
	h.clientsMux.RUnlock()

	if !roomExists || len(roomClients) == 0 {
		log.Printf("[WebSocket] No WebSocket clients connected to room: %s", roomId)
		return
	}

	// Normalize tenant ID
	normalizedTargetTenantID := strings.TrimSpace(targetTenantID)
	tenantKey := normalizedTargetTenantID
	if tenantKey == "" {
		tenantKey = "default"
		log.Printf("[WebSocket] WARNING: targetTenantID is empty, using 'default' as key")
	}

	log.Printf("[WebSocket] Broadcasting for tenantID: '%s' (normalized: '%s', key: '%s')", targetTenantID, normalizedTargetTenantID, tenantKey)

	// Get clients for this specific tenant
	h.clientsMux.RLock()
	tenantClients, tenantExists := h.clients[roomId][tenantKey]
	h.clientsMux.RUnlock()

	if !tenantExists || len(tenantClients) == 0 {
		log.Printf("[WebSocket] No clients found for tenantID '%s' (key: '%s') in room %s", targetTenantID, tenantKey, roomId)
		return
	}

	log.Printf("[WebSocket] Broadcasting queue update for room %s, tenantID: '%s' (key: '%s'), found %d clients",
		roomId, targetTenantID, tenantKey, len(tenantClients))

	// Create context with normalized tenantID
	ctx := context.Background()
	if normalizedTargetTenantID != "" && normalizedTargetTenantID != "default" {
		ctx = context.WithValue(ctx, middleware.TENANT, normalizedTargetTenantID)
		log.Printf("[WebSocket] Creating context with normalized tenantID: '%s' for %d clients", normalizedTargetTenantID, len(tenantClients))
	} else {
		log.Printf("[WebSocket] WARNING: Using default context (no tenantID) for %d clients - this will return all entries!", len(tenantClients))
	}

	// Get queue entries from service
	entries, err := h.queueService.GetQueueEntries(ctx, roomId, []string{"WAITING", "CALLED", "IN_SERVICE"})
	if err != nil {
		log.Printf("[WebSocket] Failed to get queue entries for broadcast (tenantID: '%s'): %v", targetTenantID, err)
		return
	}

	log.Printf("[WebSocket] Retrieved %d entries for tenantID '%s' in room %s", len(entries), targetTenantID, roomId)

	// Convert to WebSocket format
	wsEntries := convertEntriesToWebSocketFormat(entries)

	message := map[string]interface{}{
		"type":    "queue_update",
		"roomId":  roomId,
		"entries": wsEntries,
	}

	log.Printf("[WebSocket] Broadcasting queue update to %d clients with tenantID '%s' in room %s: %d entries", len(tenantClients), targetTenantID, roomId, len(wsEntries))

	// Send to clients in this tenant group
	sentCount := 0
	for _, clientInfo := range tenantClients {
		if err := clientInfo.conn.WriteJSON(message); err != nil {
			log.Printf("[WebSocket] Failed to send WebSocket message: %v", err)
			clientInfo.conn.Close()
		} else {
			sentCount++
			log.Printf("[WebSocket] Successfully sent message to client with tenantID: '%s'", clientInfo.tenantID)
		}
	}
	log.Printf("[WebSocket] Successfully sent queue update to %d/%d clients for tenantID '%s'", sentCount, len(tenantClients), targetTenantID)
}

// addClient adds a client to the hub
func (h *Hub) addClient(roomId, tenantKey string, clientInfo *ClientInfo) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	if h.clients[roomId] == nil {
		h.clients[roomId] = make(map[string][]*ClientInfo)
	}
	if h.clients[roomId][tenantKey] == nil {
		h.clients[roomId][tenantKey] = make([]*ClientInfo, 0)
	}
	h.clients[roomId][tenantKey] = append(h.clients[roomId][tenantKey], clientInfo)

	totalClients := 0
	for _, tenantClients := range h.clients[roomId] {
		totalClients += len(tenantClients)
	}

	log.Printf("[WebSocket] Total clients in room %s: %d (clients for tenant '%s': %d)", roomId, totalClients, tenantKey, len(h.clients[roomId][tenantKey]))
}

// removeClient removes a client from the hub
func (h *Hub) removeClient(roomId, tenantKey string, conn *websocket.Conn) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	if roomClients, exists := h.clients[roomId]; exists {
		if tenantClients, exists := roomClients[tenantKey]; exists {
			for i, client := range tenantClients {
				if client.conn == conn {
					// Remove this client from the slice
					h.clients[roomId][tenantKey] = append(tenantClients[:i], tenantClients[i+1:]...)
					break
				}
			}
			// If this was the last client for this tenant, remove the tenant entry
			if len(h.clients[roomId][tenantKey]) == 0 {
				delete(h.clients[roomId], tenantKey)
			}
		}
		// If this was the last tenant in the room, remove the room entry
		if len(h.clients[roomId]) == 0 {
			delete(h.clients, roomId)
		}
	}
	log.Printf("[WebSocket] Client disconnected from room: %s, tenantID key: '%s'", roomId, tenantKey)
}

// extractTenantID extracts the tenant ID from query parameters or headers
func extractTenantID(r *http.Request) string {
	log.Printf("[WebSocket] Raw request URL: %s", r.URL.String())
	log.Printf("[WebSocket] Raw request URL raw query: %s", r.URL.RawQuery)

	// Get all query parameters
	queryParams := r.URL.Query()
	log.Printf("[WebSocket] Query parameters map: %v", queryParams)

	// Try to get tenantId from query parameters with different casing
	tenantID := queryParams.Get("tenantId")
	if tenantID == "" {
		tenantID = queryParams.Get("tenantID")
	}
	if tenantID == "" {
		tenantID = queryParams.Get("tenant-id")
	}
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
		log.Printf("[WebSocket] Tenant ID not found in query params, checking header: '%s'", tenantID)
	} else {
		log.Printf("[WebSocket] Found tenant ID in query parameters: '%s'", tenantID)
	}

	log.Printf("[WebSocket] All query parameters: %v", queryParams)
	log.Printf("[WebSocket] Headers: X-Tenant-ID=%s", r.Header.Get("X-Tenant-ID"))

	return tenantID
}
