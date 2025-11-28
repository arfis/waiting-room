package rest

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/dig"

	"github.com/arfis/waiting-room/internal/config"
	"github.com/arfis/waiting-room/internal/middleware"
	"github.com/arfis/waiting-room/internal/rest/register"
	"github.com/arfis/waiting-room/internal/websocket"
	kioskService "github.com/arfis/waiting-room/internal/service/kiosk"
	queueServiceGenerated "github.com/arfis/waiting-room/internal/service/queue"
)

// NewServer creates and configures the HTTP server with all routes and middleware
func NewServer(diContainer *dig.Container, cfg *config.Config) *http.Server {
	// Create main router
	r := chi.NewRouter()

	// Add custom CORS middleware that doesn't interfere with WebSocket upgrades
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CORS for WebSocket routes
			if strings.HasPrefix(r.URL.Path, cfg.WebSocket.Path) || r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			// Apply CORS for other routes using configuration
			// Get Origin header (Go's Header.Get is case-insensitive, but we'll try anyway)
			origin := r.Header.Get("Origin")

			// Log request details (less verbose)
			log.Printf("[CORS] %s %s - Origin: '%s'", r.Method, r.URL.Path, origin)

			allowedOrigins := cfg.GetAvailableCORSOrigins()

			// Normalize origin (trim whitespace, remove trailing slash)
			normalizedOrigin := strings.TrimSpace(origin)
			if normalizedOrigin != "" && strings.HasSuffix(normalizedOrigin, "/") {
				normalizedOrigin = normalizedOrigin[:len(normalizedOrigin)-1]
			}

			// Normalize allowed origins (trim whitespace, remove trailing slashes)
			normalizedAllowedOrigins := make([]string, len(allowedOrigins))
			for i, allowed := range allowedOrigins {
				normalized := strings.TrimSpace(allowed)
				if strings.HasSuffix(normalized, "/") {
					normalized = normalized[:len(normalized)-1]
				}
				normalizedAllowedOrigins[i] = normalized
			}

			log.Printf("[CORS] Request origin: '%s' (normalized: '%s')", origin, normalizedOrigin)
			log.Printf("[CORS] Allowed origins: %v (normalized: %v)", allowedOrigins, normalizedAllowedOrigins)

			// Check if origin is localhost (for development - allow any localhost port)
			isLocalhost := normalizedOrigin != "" && (strings.HasPrefix(normalizedOrigin, "http://localhost:") || strings.HasPrefix(normalizedOrigin, "http://127.0.0.1:"))

			// Determine which origin to allow
			var allowedOrigin string
			if normalizedOrigin != "" && contains(normalizedAllowedOrigins, normalizedOrigin) {
				// Origin is in the allowed list
				allowedOrigin = normalizedOrigin
				log.Printf("[CORS] Allowed origin (from config): %s", normalizedOrigin)
			} else if isLocalhost {
				// For development: allow any localhost origin even if not explicitly in config
				allowedOrigin = normalizedOrigin
				log.Printf("[CORS] Allowed localhost origin (development): %s", normalizedOrigin)
			} else if len(normalizedAllowedOrigins) > 0 && normalizedOrigin != "" {
				// Origin provided but not in allowed list and not localhost - reject
				log.Printf("[CORS] Rejected origin: '%s' (normalized: '%s') not in allowed list: %v", origin, normalizedOrigin, normalizedAllowedOrigins)
				// Set CORS headers even on rejection so browser can see the error
				w.Header().Set("Access-Control-Allow-Origin", normalizedOrigin) // Echo back the origin for debugging
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", cfg.GetCORSMethods())
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Tenant-ID, Authorization, Accept, Origin, X-Requested-With")
				w.WriteHeader(http.StatusForbidden)
				return
			} else if len(normalizedAllowedOrigins) > 0 {
				// No origin header (e.g., same-origin request or curl) - allow all for development
				// For browser requests, we should check Referer header if available
				referer := r.Header.Get("Referer")
				if referer != "" {
					// Extract origin from Referer URL (e.g., "http://localhost:4201/some/path" -> "http://localhost:4201")
					refererOrigin := strings.TrimSpace(referer)
					// Parse URL to extract just the origin (scheme + host + port)
					if strings.HasPrefix(refererOrigin, "http://") || strings.HasPrefix(refererOrigin, "https://") {
						// Find the third slash (after http://host:port/)
						parts := strings.Split(refererOrigin, "/")
						if len(parts) >= 3 {
							refererOrigin = parts[0] + "//" + parts[2]
						}
						// Remove trailing slash if present
						if strings.HasSuffix(refererOrigin, "/") {
							refererOrigin = refererOrigin[:len(refererOrigin)-1]
						}
						// Check if referer is localhost
						if strings.HasPrefix(refererOrigin, "http://localhost:") || strings.HasPrefix(refererOrigin, "http://127.0.0.1:") {
							allowedOrigin = refererOrigin
							log.Printf("[CORS] No origin header, using Referer origin: %s (extracted from: %s)", allowedOrigin, referer)
						} else {
							allowedOrigin = normalizedAllowedOrigins[0]
							log.Printf("[CORS] No origin header, Referer not localhost, allowing first config origin: %s", allowedOrigin)
						}
					} else {
						allowedOrigin = normalizedAllowedOrigins[0]
						log.Printf("[CORS] No origin header, Referer format invalid, allowing first config origin: %s", allowedOrigin)
					}
				} else {
					allowedOrigin = normalizedAllowedOrigins[0]
					log.Printf("[CORS] No origin or Referer header, allowing first config origin: %s", allowedOrigin)
				}
			} else {
				// No allowed origins configured - allow all (for development only)
				allowedOrigin = "*"
				log.Printf("[CORS] No allowed origins configured, allowing all origins (*)")
			}

			// Always set CORS headers for the determined origin
			if allowedOrigin != "" {
				if allowedOrigin == "*" {
					// Can't use "*" with credentials
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				log.Printf("[CORS] Set Access-Control-Allow-Origin: %s", allowedOrigin)
			}

			// Set CORS headers (must be set for all requests, including preflight)
			w.Header().Set("Access-Control-Allow-Methods", cfg.GetCORSMethods())

			// Handle allowed headers - if "*" is in the list, use explicit headers
			corsHeaders := cfg.GetCORSHeaders()
			allowedHeadersList := cfg.CORS.AllowedHeaders

			// Check if "*" is in the allowed headers list
			if len(allowedHeadersList) > 0 && contains(allowedHeadersList, "*") {
				// Use common headers explicitly since browsers don't accept "*" with credentials
				// Include all headers that kiosk and other apps might use
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Tenant-ID, Authorization, Accept, Origin, X-Requested-With, Cache-Control, Pragma, Expires")
			} else {
				w.Header().Set("Access-Control-Allow-Headers", corsHeaders)
			}

			w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

			// Handle preflight OPTIONS requests
			// IMPORTANT: Must handle OPTIONS AFTER setting all CORS headers
			if r.Method == "OPTIONS" {
				log.Printf("[CORS] Handling preflight OPTIONS request for origin: %s", normalizedOrigin)
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Apply tenant middleware to extract tenant ID from headers/query params
	diContainer.Invoke(func(tenantMiddleware *middleware.TenantMiddleware) {
		r.Use(tenantMiddleware.Middleware())
		log.Println("Tenant middleware registered")
	})

	// Temporarily remove other middleware to isolate WebSocket issue
	// diContainer.Invoke(func(loggingMiddleware *middleware.LoggingMiddleware) {
	// 	r.Use(loggingMiddleware.LoggingMiddleware)
	// })
	// r.Use(middleware.RequestIdMiddleware)

	// Create WebSocket hub for handling WebSocket connections
	var wsHub *websocket.Hub
	diContainer.Invoke(func(kioskService *kioskService.Service, queueServiceGenerated *queueServiceGenerated.Service) {
		wsHub = websocket.NewHub(queueServiceGenerated)

		// Set up broadcast function for services that need it
		kioskService.SetBroadcastFunc(wsHub.BroadcastQueueUpdate)
		queueServiceGenerated.SetBroadcastFunc(wsHub.BroadcastQueueUpdate)
		log.Println("Broadcast function set up for kiosk and queue services")
	})

	// todo: has to be later updated to use configuration.ServerContext
	// Register API routes - CORS middleware is already applied above
	r.Route("/api", func(router chi.Router) {
		register.Generated(router, diContainer)
	})

	// Add WebSocket routes AFTER middleware (like the original working version)
	if wsHub != nil && cfg.WebSocket.Enabled {
		r.Get(cfg.WebSocket.Path+"/{roomId}", wsHub.HandleConnection)
		r.Get("/health", healthCheck)
		log.Printf("WebSocket routes registered at %s/{roomId}", cfg.WebSocket.Path)
	} else if !cfg.WebSocket.Enabled {
		log.Println("WebSocket disabled in configuration")
	} else {
		log.Println("ERROR: wsHub is nil, cannot register WebSocket routes")
	}

	// Create server with configuration
	return &http.Server{
		Addr:              cfg.GetAddress(),
		Handler:           r,
		ReadHeaderTimeout: 2 * time.Second,
	}
}

// healthCheck is a simple health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
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
