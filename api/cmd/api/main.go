package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/cors"

	"github.com/arfis/waiting-room/internal/api"
	"github.com/arfis/waiting-room/internal/cardreader"
	"github.com/arfis/waiting-room/internal/queue"
)

type server struct {
	cardService  *cardreader.Service
	queueService *queue.Service
}

func (s *server) PostWaitingRoomsRoomIdSwipe(w http.ResponseWriter, r *http.Request, roomId string) {
	log.Printf("Received swipe request for room: %s", roomId)

	var requestBody struct {
		IDCardRaw string `json:"idCardRaw"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received idCardRaw: %s", requestBody.IDCardRaw)

	// Parse the card data from the JSON string
	var parsedCardData struct {
		IDNumber  string `json:"id_number"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.Unmarshal([]byte(requestBody.IDCardRaw), &parsedCardData); err != nil {
		log.Printf("Failed to parse card data: %v", err)
		// Use fallback data
		parsedCardData = struct {
			IDNumber  string `json:"id_number"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}{
			IDNumber:  "123456789",
			FirstName: "John",
			LastName:  "Doe",
		}
	}

	cardData := queue.CardData{
		IDNumber:  parsedCardData.IDNumber,
		FirstName: parsedCardData.FirstName,
		LastName:  parsedCardData.LastName,
		Source:    "card-reader",
	}

	log.Printf("Creating queue entry with card data: %+v", cardData)

	// Create queue entry
	entry, err := s.queueService.CreateEntry(roomId, cardData)
	if err != nil {
		log.Printf("Failed to create queue entry: %v", err)
		http.Error(w, "Failed to create queue entry", http.StatusInternalServerError)
		return
	}

	log.Printf("Created queue entry: %+v", entry)

	// Generate QR URL
	qrUrl := fmt.Sprintf("http://localhost:4204/q/%s", entry.QRToken)

	res := api.JoinResult{
		EntryId:      uuid.MustParse(entry.ID),
		TicketNumber: entry.TicketNumber,
		QrUrl:        qrUrl,
	}

	log.Printf("Sending response: %+v", res)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (s *server) GetQueueEntriesTokenQrToken(w http.ResponseWriter, r *http.Request, qrToken string) {
	entry, err := s.queueService.GetEntryByQRToken(qrToken)
	if err != nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	// Calculate ETA (rough estimate: 5 minutes per person)
	etaMinutes := (entry.Position - 1) * 5
	if etaMinutes < 0 {
		etaMinutes = 0
	}

	resp := api.PublicEntry{
		EntryId:      uuid.MustParse(entry.ID),
		TicketNumber: entry.TicketNumber,
		Status:       api.QueueEntryStatus(entry.Status),
		Position:     entry.Position,
		EtaMinutes:   etaMinutes,
		CanCancel:    entry.Status == "WAITING",
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *server) PostWaitingRoomsRoomIdNext(w http.ResponseWriter, r *http.Request, roomId string) {
	entry, err := s.queueService.CallNext(roomId)
	if err != nil {
		http.Error(w, "No entries waiting", http.StatusNotFound)
		return
	}

	resp := api.QueueEntry{
		Id:            uuid.MustParse(entry.ID),
		WaitingRoomId: uuid.MustParse("00000000-0000-0000-0000-000000000001"), // Use a fixed UUID for triage-1
		TicketNumber:  entry.TicketNumber,
		Status:        api.CALLED,
		Position:      entry.Position,
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

func (s *server) GetWaitingRoomQueue(w http.ResponseWriter, r *http.Request) {
	roomId := chi.URLParam(r, "roomId")
	if roomId == "" {
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	entries, err := s.queueService.GetQueueEntries(roomId)
	if err != nil {
		log.Printf("Failed to get queue entries: %v", err)
		http.Error(w, "Failed to get queue entries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func main() {
	r := chi.NewRouter()
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200", "http://localhost:4201", "http://localhost:4204", "http://localhost:4203"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler)

	// Initialize card reader service
	cardService := cardreader.NewService()
	if err := cardService.Initialize(); err != nil {
		log.Printf("Warning: Failed to initialize card reader: %v", err)
	}

	// Initialize queue service
	queueService := queue.NewService()
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:admin@localhost:27017/waiting_room?authSource=admin"
	}
	if err := queueService.Initialize(mongoURI); err != nil {
		log.Printf("Warning: Failed to initialize queue service: %v", err)
		log.Printf("Queue service will use mock data until MongoDB is available")
	}

	s := &server{
		cardService:  cardService,
		queueService: queueService,
	}
	r.Mount("/", api.HandlerFromMux(s, r))

	// Add health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Add card reader routes
	r.Get("/api/card-reader/status", s.GetCardReaderStatus)
	r.Post("/api/card-reader/read", s.PostCardReaderRead)

	// Add queue routes
	r.Get("/api/waiting-rooms/{roomId}/queue", s.GetWaitingRoomQueue)

	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	log.Println("API listening on", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
