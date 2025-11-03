package rest

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/dig"

	"github.com/arfis/waiting-room/internal/config"
	"github.com/arfis/waiting-room/internal/middleware"
	queueService "github.com/arfis/waiting-room/internal/queue"
	"github.com/arfis/waiting-room/internal/rest/register"
	kioskService "github.com/arfis/waiting-room/internal/service/kiosk"
	queueServiceGenerated "github.com/arfis/waiting-room/internal/service/queue"
)

// ClientInfo stores information about a WebSocket client
type ClientInfo struct {
	conn     *websocket.Conn
	tenantID string // tenantID from query parameter or header
}

// Server represents the HTTP server with WebSocket support
type Server struct {
	queueService *queueService.WaitingQueue
	upgrader     websocket.Upgrader
	// clients structure: roomId -> tenantID -> []*ClientInfo
	// This allows us to efficiently find all clients for a specific tenant
	clients    map[string]map[string][]*ClientInfo
	clientsMux sync.RWMutex
}

// NewServer creates and configures the HTTP server with all routes and middleware
func NewServer(diContainer *dig.Container, cfg *config.Config) *http.Server {
	// Create main router
	r := chi.NewRouter()

	// Add custom CORS middleware that doesn't interfere with WebSocket upgrades
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CORS for WebSocket routes
			if r.URL.Path == cfg.WebSocket.Path+"/"+cfg.GetDefaultRoom() || r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			// Apply CORS for other routes using configuration
			origin := r.Header.Get("Origin")
			allowedOrigins := cfg.GetAvailableCORSOrigins()

			// Check if the origin is in the allowed list
			if origin != "" && contains(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(allowedOrigins) > 0 {
				// If no origin or origin not in list, use the first allowed origin as fallback
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
			}

			w.Header().Set("Access-Control-Allow-Methods", cfg.GetCORSMethods())
			w.Header().Set("Access-Control-Allow-Headers", cfg.GetCORSHeaders())
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Temporarily remove other middleware to isolate WebSocket issue
	// diContainer.Invoke(func(loggingMiddleware *middleware.LoggingMiddleware) {
	// 	r.Use(loggingMiddleware.LoggingMiddleware)
	// })
	// r.Use(middleware.RequestIdMiddleware)

	// Create server instance for WebSocket handling
	var wsServer *Server
	diContainer.Invoke(func(queueService *queueService.WaitingQueue, kioskService *kioskService.Service, queueServiceGenerated *queueServiceGenerated.Service) {
		wsServer = &Server{
			queueService: queueService,
			upgrader: websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true // Allow all origins for development
				},
			},
			clients: make(map[string]map[string][]*ClientInfo),
		}

		// Set up broadcast function for services that need it
		kioskService.SetBroadcastFunc(wsServer.broadcastQueueUpdate)
		queueServiceGenerated.SetBroadcastFunc(wsServer.broadcastQueueUpdate)
		log.Println("Broadcast function set up for kiosk and queue services")
	})

	// todo: has to be later updated to use configuration.ServerContext
	r.Route("/api", func(router chi.Router) {
		register.Generated(router, diContainer)
	})
	// Register generated routes with /api prefix
	// r.Route("/api", func(apiRouter chi.Router) {
	// 	log.Println("Registering API routes...")
	// 	register.Generated(apiRouter, diContainer)

	// 	// Add admin routes as unprotected for development
	// 	diContainer.Invoke(func(adminHandler *admin.Handler) {
	// 		apiRouter.Get("/admin/configuration", adminHandler.GetSystemConfiguration)
	// 		apiRouter.Put("/admin/configuration", adminHandler.UpdateSystemConfiguration)
	// 		apiRouter.Get("/admin/configuration/external-api", adminHandler.GetExternalAPIConfiguration)
	// 		apiRouter.Put("/admin/configuration/external-api", adminHandler.UpdateExternalAPIConfiguration)
	// 		apiRouter.Get("/admin/configuration/rooms", adminHandler.GetRoomsConfiguration)
	// 		apiRouter.Put("/admin/configuration/rooms", adminHandler.UpdateRoomsConfiguration)
	// 		apiRouter.Get("/admin/card-readers", adminHandler.GetCardReaders)
	// 		apiRouter.Post("/admin/card-readers/{id}/restart", adminHandler.RestartCardReader)
	// 	})

	// 	// Add kiosk routes as unprotected for development
	// 	diContainer.Invoke(func(kioskHandler *kioskHandler.Handler) {
	// 		apiRouter.Get("/generic-services", kioskHandler.GetGenericServices)
	// 		apiRouter.Get("/appointment-services", kioskHandler.GetAppointmentServices)
	// 		apiRouter.Get("/user-services", kioskHandler.GetUserServices)
	// 		apiRouter.Get("/default-service-point", kioskHandler.GetDefaultServicePoint)
	// 		apiRouter.Post("/swipe", kioskHandler.SwipeCard)
	// 	})

	// 	log.Println("API routes registered successfully")
	// })

	// Add WebSocket routes AFTER middleware (like the original working version)
	if wsServer != nil && cfg.WebSocket.Enabled {
		r.Get(cfg.WebSocket.Path+"/{roomId}", wsServer.handleQueueWebSocket)
		r.Get("/health", wsServer.healthCheck)
		log.Printf("WebSocket routes registered at %s/{roomId}", cfg.WebSocket.Path)
	} else if !cfg.WebSocket.Enabled {
		log.Println("WebSocket disabled in configuration")
	} else {
		log.Println("ERROR: wsServer is nil, cannot register WebSocket routes")
	}

	// Create server with configuration
	return &http.Server{
		Addr:              cfg.GetAddress(),
		Handler:           r,
		ReadHeaderTimeout: 2 * time.Second,
	}
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// WebSocket handler for queue updates
func (s *Server) handleQueueWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("WebSocket handler called for room: %s", chi.URLParam(r, "roomId"))

	if s == nil {
		log.Printf("ERROR: Server instance is nil")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	roomId := chi.URLParam(r, "roomId")
	if roomId == "" {
		log.Printf("Room ID is empty")
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	// Extract tenant ID from query parameter or header
	// Query parameters are automatically URL-decoded by Go's URL.Query()
	// Log the raw URL for debugging
	log.Printf("[WebSocket] Raw request URL: %s", r.URL.String())
	log.Printf("[WebSocket] Raw request URL path: %s", r.URL.Path)
	log.Printf("[WebSocket] Raw request URL raw query: %s", r.URL.RawQuery)

	// Get all query parameters
	queryParams := r.URL.Query()
	log.Printf("[WebSocket] Query parameters map: %v", queryParams)
	log.Printf("[WebSocket] Query parameters count: %d", len(queryParams))

	// Try to get tenantId from query parameters
	tenantID := queryParams.Get("tenantId")
	if tenantID == "" {
		// Try with different casing
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

	// Log all query parameters for debugging
	log.Printf("[WebSocket] All query parameters: %v", queryParams)
	log.Printf("[WebSocket] Headers: X-Tenant-ID=%s", r.Header.Get("X-Tenant-ID"))
	log.Printf("[WebSocket] Attempting to upgrade WebSocket connection for room: %s, tenantID: '%s' (length: %d, raw bytes: %v)", roomId, tenantID, len(tenantID), []byte(tenantID))

	// Check if the response writer supports hijacking
	if _, ok := w.(http.Hijacker); !ok {
		log.Printf("ERROR: Response writer does not implement http.Hijacker")
		http.Error(w, "WebSocket not supported", http.StatusInternalServerError)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
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
	s.clientsMux.Lock()
	if s.clients[roomId] == nil {
		s.clients[roomId] = make(map[string][]*ClientInfo)
	}
	if s.clients[roomId][tenantKey] == nil {
		s.clients[roomId][tenantKey] = make([]*ClientInfo, 0)
	}
	s.clients[roomId][tenantKey] = append(s.clients[roomId][tenantKey], clientInfo)

	totalClients := 0
	for _, tenantClients := range s.clients[roomId] {
		totalClients += len(tenantClients)
	}
	s.clientsMux.Unlock()

	log.Printf("[WebSocket] Client connected to room: %s, tenantID: '%s' (stored under key: '%s')", roomId, tenantID, tenantKey)
	log.Printf("[WebSocket] Total clients in room %s: %d (clients for tenant '%s': %d)", roomId, totalClients, tenantKey, len(s.clients[roomId][tenantKey]))

	// Send initial queue data to the newly connected client ONLY
	// We need to send to just this one client, not broadcast to all clients with the same tenantID
	go func() {
		// Small delay to ensure the client is fully connected
		time.Sleep(100 * time.Millisecond)
		log.Printf("[WebSocket] Sending initial queue data to newly connected client for tenantID: '%s' (normalized: '%s', key: '%s')", tenantID, normalizedTenantID, tenantKey)

		// Create context with normalized tenantID for filtering
		ctx := context.Background()
		// Use the normalized tenantID from the connection (same format as stored)
		if normalizedTenantID != "" && normalizedTenantID != "default" {
			ctx = context.WithValue(ctx, middleware.TENANT, normalizedTenantID)
			log.Printf("[WebSocket] Using normalized tenantID '%s' for filtering initial data", normalizedTenantID)
		} else {
			log.Printf("[WebSocket] WARNING: normalized tenantID is empty or 'default', will get all entries (no tenant filter)")
			// Don't set tenant in context - this will get all entries
		}

		// Get queue entries filtered by tenant
		entries, err := s.queueService.GetQueueEntriesWithContext(ctx, roomId, []string{"WAITING", "CALLED", "IN_SERVICE"})
		if err != nil {
			log.Printf("[WebSocket] Failed to get initial queue entries for tenantID '%s': %v", tenantID, err)
			return
		}

		log.Printf("[WebSocket] Retrieved %d initial entries for tenantID '%s' (key: '%s') in room %s", len(entries), tenantID, tenantKey, roomId)

		// Log first few entry IDs and their tenant IDs for debugging
		if len(entries) > 0 {
			log.Printf("[WebSocket] First entry tenantID: '%s', sectionID: '%s'", entries[0].TenantID, entries[0].SectionID)
			if len(entries) > 5 {
				log.Printf("[WebSocket] ... and %d more entries", len(entries)-5)
			}
		}

		// Convert to WebSocket format
		var wsEntries []map[string]interface{}
		for _, entry := range entries {
			wsEntry := map[string]interface{}{
				"id":              entry.ID,
				"waitingRoomId":   entry.WaitingRoomID,
				"ticketNumber":    entry.TicketNumber,
				"qrToken":         entry.QRToken,
				"status":          entry.Status,
				"position":        entry.Position,
				"createdAt":       entry.CreatedAt,
				"updatedAt":       entry.UpdatedAt,
				"cardData":        entry.CardData,
				"servicePoint":    entry.ServicePoint,
				"serviceName":     entry.ServiceName,
				"serviceDuration": entry.ApproximateDurationMinutes,
			}
			wsEntries = append(wsEntries, wsEntry)
		}

		message := map[string]interface{}{
			"type":    "queue_update",
			"roomId":  roomId,
			"entries": wsEntries,
		}

		// Send to only this specific client
		s.clientsMux.RLock()
		// Find the client in the tenant's list
		var foundClient *ClientInfo
		if roomClients, exists := s.clients[roomId]; exists {
			if tenantClients, exists := roomClients[tenantKey]; exists {
				for _, client := range tenantClients {
					if client.conn == conn {
						foundClient = client
						break
					}
				}
			}
		}
		s.clientsMux.RUnlock()

		if foundClient != nil {
			if err := foundClient.conn.WriteJSON(message); err != nil {
				log.Printf("[WebSocket] Failed to send initial queue data: %v", err)
			} else {
				log.Printf("[WebSocket] Successfully sent initial queue data (%d entries) to client with tenantID: '%s' (normalized: '%s')", len(wsEntries), tenantID, normalizedTenantID)
			}
		} else {
			log.Printf("[WebSocket] Client not found in room %s for tenant '%s' for initial data send", roomId, tenantKey)
		}
	}()

	// Remove client when connection closes
	defer func() {
		s.clientsMux.Lock()
		// Find and remove this client from its tenant's list
		if roomClients, exists := s.clients[roomId]; exists {
			if tenantClients, exists := roomClients[tenantKey]; exists {
				for i, client := range tenantClients {
					if client.conn == conn {
						// Remove this client from the slice
						s.clients[roomId][tenantKey] = append(tenantClients[:i], tenantClients[i+1:]...)
						break
					}
				}
				// If this was the last client for this tenant, remove the tenant entry
				if len(s.clients[roomId][tenantKey]) == 0 {
					delete(s.clients[roomId], tenantKey)
				}
			}
			// If this was the last tenant in the room, remove the room entry
			if len(s.clients[roomId]) == 0 {
				delete(s.clients, roomId)
			}
		}
		s.clientsMux.Unlock()
		log.Printf("[WebSocket] Client disconnected from room: %s, tenantID: %s (key: '%s')", roomId, tenantID, tenantKey)
	}()

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

// Broadcast queue update to clients for a specific tenant (only refresh that tenant's data)
func (s *Server) broadcastQueueUpdate(roomId string, targetTenantID string) {
	s.clientsMux.RLock()
	roomClients, roomExists := s.clients[roomId]
	s.clientsMux.RUnlock()

	if !roomExists || len(roomClients) == 0 {
		log.Printf("[WebSocket] No WebSocket clients connected to room: %s", roomId)
		return
	}

	// Normalize tenant ID: trim whitespace and use "default" as key for empty tenant ID
	// This ensures consistent key matching with how clients are stored
	normalizedTargetTenantID := strings.TrimSpace(targetTenantID)
	tenantKey := normalizedTargetTenantID
	if tenantKey == "" {
		tenantKey = "default"
		log.Printf("[WebSocket] WARNING: targetTenantID is empty, using 'default' as key")
	}

	log.Printf("[WebSocket] Broadcasting for tenantID: '%s' (normalized: '%s', key: '%s')", targetTenantID, normalizedTargetTenantID, tenantKey)
	log.Printf("[WebSocket] Looking for clients with tenantKey: '%s'", tenantKey)

	// Get clients for this specific tenant
	s.clientsMux.RLock()
	tenantClients, tenantExists := s.clients[roomId][tenantKey]

	// Log all available tenant keys for debugging
	availableTenants := make([]string, 0, len(roomClients))
	for k := range roomClients {
		availableTenants = append(availableTenants, k)
	}
	s.clientsMux.RUnlock()

	if !tenantExists || len(tenantClients) == 0 {
		log.Printf("[WebSocket] No clients found for tenantID '%s' (key: '%s') in room %s", targetTenantID, tenantKey, roomId)
		log.Printf("[WebSocket] Available tenant keys in room %s: %v", roomId, availableTenants)
		log.Printf("[WebSocket] Available tenant keys count: %d", len(availableTenants))
		// Log each tenant key and its client count for debugging
		s.clientsMux.RLock()
		for k, clients := range roomClients {
			log.Printf("[WebSocket]   Tenant key '%s': %d clients", k, len(clients))
			for i, client := range clients {
				log.Printf("[WebSocket]     Client %d: tenantID='%s'", i, client.tenantID)
			}
		}
		s.clientsMux.RUnlock()
		return
	}

	log.Printf("[WebSocket] Broadcasting queue update for room %s, tenantID: '%s' (key: '%s'), found %d clients",
		roomId, targetTenantID, tenantKey, len(tenantClients))

	// Create context with normalized tenantID for this tenant group
	// Use normalizedTargetTenantID (not targetTenantID) to match what's stored
	ctx := context.Background()
	if normalizedTargetTenantID != "" && normalizedTargetTenantID != "default" {
		ctx = context.WithValue(ctx, middleware.TENANT, normalizedTargetTenantID)
		log.Printf("[WebSocket] Creating context with normalized tenantID: '%s' for %d clients", normalizedTargetTenantID, len(tenantClients))
	} else {
		log.Printf("[WebSocket] WARNING: Using default context (no tenantID) for %d clients - this will return all entries!", len(tenantClients))
		// Don't set tenant in context - this will get all entries (not what we want)
	}

	// Get queue entries filtered by tenant (repository will filter by tenantID from context)
	// Include all statuses that might be relevant for the queue display
	entries, err := s.queueService.GetQueueEntriesWithContext(ctx, roomId, []string{"WAITING", "CALLED", "IN_SERVICE"})
	if err != nil {
		log.Printf("[WebSocket] Failed to get queue entries for broadcast (tenantID: '%s'): %v", targetTenantID, err)
		return
	}

	log.Printf("[WebSocket] Retrieved %d entries for tenantID '%s' in room %s", len(entries), targetTenantID, roomId)

	// Convert to WebSocket format
	var wsEntries []map[string]interface{}
	for _, entry := range entries {
		wsEntry := map[string]interface{}{
			"id":              entry.ID,
			"waitingRoomId":   entry.WaitingRoomID,
			"ticketNumber":    entry.TicketNumber,
			"qrToken":         entry.QRToken,
			"status":          entry.Status,
			"position":        entry.Position,
			"createdAt":       entry.CreatedAt,
			"updatedAt":       entry.UpdatedAt,
			"cardData":        entry.CardData,
			"servicePoint":    entry.ServicePoint,
			"serviceName":     entry.ServiceName,
			"serviceDuration": entry.ApproximateDurationMinutes,
		}
		wsEntries = append(wsEntries, wsEntry)
	}

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
			// Note: Failed clients will be cleaned up on disconnect
		} else {
			sentCount++
			log.Printf("[WebSocket] Successfully sent message to client with tenantID: '%s'", clientInfo.tenantID)
		}
	}
	log.Printf("[WebSocket] Successfully sent queue update to %d/%d clients for tenantID '%s'", sentCount, len(tenantClients), targetTenantID)
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
