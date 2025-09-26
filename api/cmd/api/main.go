package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/cors"

	"github.com/arfis/waiting-room/internal/api"
	"github.com/arfis/waiting-room/internal/cardreader"
)

type server struct {
	cardService *cardreader.Service
}

func (s *server) PostWaitingRoomsRoomIdSwipe(w http.ResponseWriter, r *http.Request, roomId openapi_types.UUID) {
	// TODO: create entry, ticket, QR URL
	entryId := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	res := api.JoinResult{
		EntryId:      entryId,
		TicketNumber: "A-001",
		QrUrl:        "http://localhost:4201/q/demo",
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (s *server) GetQueueEntriesTokenQrToken(w http.ResponseWriter, r *http.Request, qrToken string) {
	// TODO: resolve token
	entryId := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	resp := api.PublicEntry{
		EntryId:      entryId,
		TicketNumber: "A-001",
		Status:       api.WAITING,
		Position:     5,
		EtaMinutes:   12,
		CanCancel:    false,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *server) PostWaitingRoomsRoomIdNext(w http.ResponseWriter, r *http.Request, roomId openapi_types.UUID) {
	// TODO: call next
	entryId := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	resp := api.QueueEntry{
		Id:            entryId,
		WaitingRoomId: roomId,
		TicketNumber:  "A-001",
		Status:        api.CALLED,
		Position:      0,
	}
	json.NewEncoder(w).Encode(resp)
}

// Card reader endpoints
func (s *server) GetCardReaderStatus(w http.ResponseWriter, r *http.Request) {
	// For now, always return false since we're using the standalone card-reader app
	// The real card reader status should come from the WebSocket connection
	status := map[string]interface{}{
		"connected": false,
		"status":    "standalone-card-reader-mode",
		"message":   "Using standalone card reader app via WebSocket",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *server) PostCardReaderRead(w http.ResponseWriter, r *http.Request) {
	if !s.cardService.IsConnected() {
		http.Error(w, "Card reader not connected", http.StatusServiceUnavailable)
		return
	}

	cardData, err := s.cardService.ReadCard()
	if err != nil {
		http.Error(w, "Failed to read card: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cardData)
}

func main() {
	r := chi.NewRouter()
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200", "http://localhost:4201", "http://localhost:4202", "http://localhost:4203"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler)

	// Initialize card reader service
	cardService := cardreader.NewService()
	if err := cardService.Initialize(); err != nil {
		log.Printf("Warning: Failed to initialize card reader: %v", err)
	}

	s := &server{cardService: cardService}
	r.Mount("/", api.HandlerFromMux(s, r))

	// Add health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Add card reader routes
	r.Get("/api/card-reader/status", s.GetCardReaderStatus)
	r.Post("/api/card-reader/read", s.PostCardReaderRead)

	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	log.Println("API listening on", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
