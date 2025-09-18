package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/cors"

	"github.com/arfis/waiting-room/internal/api"
	"github.com/arfis/waiting-room/internal/cardreader"
)

type server struct {
	cardService *cardreader.Service
}

func (s *server) PostWaitingRoomsRoomIdSwipe(w http.ResponseWriter, r *http.Request, roomId api.UUID, body api.PostWaitingRoomsRoomIdSwipeJSONRequestBody) {
	// TODO: create entry, ticket, QR URL
	res := api.JoinResult{
		EntryId:      "00000000-0000-0000-0000-000000000001",
		TicketNumber: "A-001",
		QrUrl:        "http://localhost:4201/q/demo",
	}
	w.WriteHeader(http.StatusCreated)
	api.EncodeJSONResponse(res, nil, w)
}

func (s *server) GetQueueEntriesTokenQrToken(w http.ResponseWriter, r *http.Request, qrToken string) {
	// TODO: resolve token
	resp := api.PublicEntry{
		EntryId:      "00000000-0000-0000-0000-000000000001",
		TicketNumber: "A-001",
		Status:       api.QueueEntryStatusWAITING,
		Position:     5,
		EtaMinutes:   12,
		CanCancel:    false,
	}
	api.EncodeJSONResponse(resp, nil, w)
}

func (s *server) PostWaitingRoomsRoomIdNext(w http.ResponseWriter, r *http.Request, roomId api.UUID) {
	// TODO: call next
	resp := api.QueueEntry{
		Id:            "00000000-0000-0000-0000-000000000001",
		WaitingRoomId: roomId,
		TicketNumber:  "A-001",
		Status:        api.QueueEntryStatusCALLED,
		Position:      0,
	}
	api.EncodeJSONResponse(resp, nil, w)
}

// Card reader endpoints
func (s *server) GetCardReaderStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"connected": s.cardService.IsConnected(),
		"status":    "ready",
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
	api.RegisterHandlers(r, s)

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
