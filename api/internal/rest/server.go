package rest

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/dig"

	"github.com/arfis/waiting-room/internal/config"
	queueService "github.com/arfis/waiting-room/internal/queue"
	"github.com/arfis/waiting-room/internal/rest/register"
	kioskService "github.com/arfis/waiting-room/internal/service/kiosk"
	queueServiceGenerated "github.com/arfis/waiting-room/internal/service/queue"
)

// Server represents the HTTP server with WebSocket support
type Server struct {
	queueService *queueService.WaitingQueue
	upgrader     websocket.Upgrader
	clients      map[string]map[*websocket.Conn]bool
	clientsMux   sync.RWMutex
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
			clients: make(map[string]map[*websocket.Conn]bool),
		}

		// Set up broadcast function for services that need it
		kioskService.SetBroadcastFunc(wsServer.broadcastQueueUpdate)
		queueServiceGenerated.SetBroadcastFunc(wsServer.broadcastQueueUpdate)
		log.Println("Broadcast function set up for kiosk and queue services")
	})

	// Register generated routes with /api prefix
	r.Route("/api", func(apiRouter chi.Router) {
		log.Println("Registering API routes...")
		register.Generated(apiRouter, diContainer)
		log.Println("API routes registered successfully")
	})

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

	log.Printf("Attempting to upgrade WebSocket connection for room: %s", roomId)

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

	// Add client to room
	s.clientsMux.Lock()
	if s.clients[roomId] == nil {
		s.clients[roomId] = make(map[*websocket.Conn]bool)
	}
	s.clients[roomId][conn] = true
	s.clientsMux.Unlock()

	log.Printf("WebSocket client connected to room: %s", roomId)

	// Remove client when connection closes
	defer func() {
		s.clientsMux.Lock()
		delete(s.clients[roomId], conn)
		if len(s.clients[roomId]) == 0 {
			delete(s.clients, roomId)
		}
		s.clientsMux.Unlock()
		log.Printf("WebSocket client disconnected from room: %s", roomId)
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

// Broadcast queue update to all connected clients
func (s *Server) broadcastQueueUpdate(roomId string) {
	s.clientsMux.RLock()
	clients, exists := s.clients[roomId]
	s.clientsMux.RUnlock()

	if !exists || len(clients) == 0 {
		log.Printf("No WebSocket clients connected to room: %s", roomId)
		return
	}

	// Get current queue entries
	// for now we only broadcast waiting entries
	entries, err := s.queueService.GetQueueEntries(roomId, []string{"WAITING", "CALLED"})
	if err != nil {
		log.Printf("Failed to get queue entries for broadcast: %v", err)
		return
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

	log.Printf("Broadcasting queue update to %d clients in room %s: %d entries", len(clients), roomId, len(wsEntries))

	// Send to all clients
	s.clientsMux.RLock()
	for client := range clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("Failed to send WebSocket message: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
	s.clientsMux.RUnlock()
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
