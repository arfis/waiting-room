package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/ebfe/scard"
	"github.com/gorilla/websocket"
	"github.com/miekg/pkcs11"
)

type Payload struct {
	DeviceID   string    `json:"deviceId"`
	RoomID     string    `json:"roomId"`
	Token      string    `json:"token"`      // random per insertion
	Reader     string    `json:"reader"`     // reader name
	ATR        string    `json:"atr"`        // hex
	Protocol   string    `json:"protocol"`   // T=0/T=1/unknown
	OccurredAt string    `json:"occurredAt"` // RFC3339
	CardData   *CardData `json:"cardData,omitempty"`
	State      string    `json:"state"`   // "waiting", "reading", "success", "error"
	Message    string    `json:"message"` // Human readable message
}

type CardData struct {
	IDNumber    string `json:"id_number,omitempty"`     // cert serial / IC serial / UID
	FirstName   string `json:"first_name,omitempty"`    // from cert if present
	LastName    string `json:"last_name,omitempty"`     // from cert if present
	DateOfBirth string `json:"date_of_birth,omitempty"` // usually not available without auth
	Gender      string `json:"gender,omitempty"`
	Nationality string `json:"nationality,omitempty"`
	Address     string `json:"address,omitempty"`
	IssuedDate  string `json:"issued_date,omitempty"`
	ExpiryDate  string `json:"expiry_date,omitempty"`
	Photo       string `json:"photo,omitempty"`
	Source      string `json:"source,omitempty"` // "pkcs11-cert" | "cplc" | "uid"
}

func main() {
	roomID := envOr("ROOM_ID", "triage-1")
	deviceID := envOr("DEVICE_ID", "reader-01")
	wantReader := strings.TrimSpace(os.Getenv("READER_NAME"))
	wsURL := envOr("WS_URL", "ws://localhost:4201/ws/card-reader")

	// PC/SC context
	ctx, err := scard.EstablishContext()
	must(err, "establish PC/SC context")
	defer ctx.Release()

	readers, err := ctx.ListReaders()
	must(err, "list readers")
	if len(readers) == 0 {
		log.Fatal("no smart card readers found")
	}

	var reader string
	if wantReader != "" {
		found := false
		for _, r := range readers {
			if r == wantReader {
				reader = r
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("reader %q not found. available: %v", wantReader, readers)
		}
	} else {
		reader = readers[0]
	}
	log.Printf("Using reader: %s", reader)
	log.Printf("WebSocket URL: %s", wsURL)
	log.Println("Waiting for card...")

	// Send initial waiting state
	sendStateUpdate(wsURL, deviceID, roomID, reader, "waiting", "Please insert your ID card")

	// Event-driven monitor (no polling races)
	monitor(context.Background(), *ctx, reader, func(atr []byte, proto string) {
		token := randToken(16)

		// Send reading state
		sendStateUpdate(wsURL, deviceID, roomID, reader, "reading", "Reading card data...")

		pl := Payload{
			DeviceID:   deviceID,
			RoomID:     roomID,
			Token:      token,
			Reader:     reader,
			ATR:        strings.ToUpper(hex.EncodeToString(atr)),
			Protocol:   proto,
			OccurredAt: time.Now().Format(time.RFC3339),
		}

		// Read while the card is present
		cardData := readCardData(*ctx, reader, proto, atr)
		if cardData != nil {
			pl.CardData = cardData
			pl.State = "success"
			pl.Message = "Card read successfully"
		} else {
			pl.State = "error"
			pl.Message = "Failed to read card data"
		}

		// Send final result to WebSocket
		sendToWebSocket(wsURL, pl)

		// Also print to console for debugging
		b, _ := json.MarshalIndent(pl, "", "  ")
		fmt.Println(string(b))

		// Send success state
		if cardData != nil {
			sendStateUpdate(wsURL, deviceID, roomID, reader, "success", "Card read successfully - please remove card")
		} else {
			sendStateUpdate(wsURL, deviceID, roomID, reader, "error", "Failed to read card - please try again")
		}
	}, func() {
		// Card removed callback
		sendStateUpdate(wsURL, deviceID, roomID, reader, "removed", "Card removed - ready for next card")
	})
}

// -----------------------------
// Event-driven monitor
// -----------------------------

// monitor waits for insert/remove using PC/SC GetStatusChange.
// On insert -> onInsert(); then blocks until removal; then goes back to waiting.
// REPLACE your monitor(...) with this version.
// (Fix: copy updated ReaderState back from the slice after GetStatusChange.)
func monitor(_ context.Context, c scard.Context, reader string, onInsert func(atr []byte, proto string), onRemove func()) {
	state := scard.ReaderState{
		Reader:       reader,
		CurrentState: scard.StateUnaware,
	}
	var seenATR string

	for {
		// Arm for next change based on the last event state
		state.CurrentState = state.EventState

		// Block (up to 2s) for a state change
		states := []scard.ReaderState{state}
		if err := c.GetStatusChange(states, 2000); err != nil && !errors.Is(err, scard.ErrTimeout) {
			log.Printf("GetStatusChange error: %v", err)
			// If service is not available, try to reconnect
			if strings.Contains(err.Error(), "Service not available") {
				log.Println("PC/SC service not available, attempting to reconnect...")
				time.Sleep(5 * time.Second)
				// Try to establish a new context
				newCtx, err := scard.EstablishContext()
				if err != nil {
					log.Printf("Failed to re-establish context: %v", err)
					time.Sleep(time.Second)
					continue
				}
				c = *newCtx
				// Reset state
				state = scard.ReaderState{
					Reader:       reader,
					CurrentState: scard.StateUnaware,
				}
				seenATR = ""
			}
			time.Sleep(time.Second)
			continue
		}
		// 🔧 IMPORTANT: read back the updated state
		state = states[0]

		// Only act when something actually changed
		if state.EventState&scard.StateChanged == 0 {
			continue
		}

		// Card inserted?
		if state.EventState&scard.StatePresent != 0 {
			card, err := c.Connect(reader, scard.ShareShared, scard.ProtocolAny)
			if err != nil {
				continue
			}
			status, err := card.Status()
			card.Disconnect(scard.LeaveCard)
			if err != nil {
				continue
			}
			atrHex := strings.ToUpper(hex.EncodeToString(status.Atr))
			proto := protocolName(status.ActiveProtocol)

			if atrHex != seenATR {
				seenATR = atrHex
				log.Printf("Card inserted (ATR=%s, proto=%s)", atrHex, proto)
				onInsert(status.Atr, proto)
				log.Println("Reading done - waiting for removal")
			}
			continue
		}

		// Card removed?
		if state.EventState&scard.StateEmpty != 0 && seenATR != "" {
			seenATR = ""
			log.Println("Card removed")
			if onRemove != nil {
				onRemove()
			}
			log.Println("Waiting for card...")
		}
	}
}

func waitUntilRemoved(c scard.Context, reader string) {
	for {
		// Try a quick connect: if it fails, the card is gone
		if _, err := c.Connect(reader, scard.ShareShared, scard.ProtocolAny); err != nil {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// -----------------------------
// Utilities
// -----------------------------

func protocolName(p scard.Protocol) string {
	switch p {
	case scard.ProtocolT0:
		return "T=0"
	case scard.ProtocolT1:
		return "T=1"
	default:
		return "unknown"
	}
}

func randToken(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

func envOr(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func must(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

// -----------------------------
// Data reading implementations
// -----------------------------

type CertInfo struct {
	Label   string
	Subject pkix.Name
	RawDER  []byte
}

// readCardData tries (1) PKCS#11 certs, then (2) CPLC serial, then (3) UID.
func readCardData(ctx scard.Context, reader string, protocol string, atr []byte) *CardData {
	log.Printf("Reading card data (protocol: %s)", protocol)

	// 1) Public certificates via PKCS#11 (no PIN needed to read certs)
	if info, ok := readCertSubject(); ok {
		first, last := extractName(info.Subject)
		return &CardData{
			IDNumber:  strings.TrimSpace(info.Subject.SerialNumber), // may or may not be personal
			FirstName: first,
			LastName:  last,
			Source:    "pkcs11-cert",
		}
	}

	// 2) CPLC (chip serial) via APDU GET DATA 9F7F (GlobalPlatform)
	if cplcHex, icSerial, err := readCPLC(ctx, reader); err == nil {
		_ = cplcHex
		return &CardData{
			IDNumber: strings.ToUpper(icSerial),
			Source:   "cplc",
		}
	} else {
		if swmsg := explainSW(err); swmsg != "" {
			log.Printf("CPLC read failed: %s", swmsg)
		} else {
			log.Printf("CPLC read failed: %v", err)
		}
	}

	// 3) Contactless UID (vendor command; may not be supported)
	if uid, err := readUID(ctx, reader); err == nil && uid != "" {
		return &CardData{
			IDNumber: strings.ToUpper(uid),
			Source:   "uid",
		}
	} else if err != nil {
		if swmsg := explainSW(err); swmsg != "" {
			log.Printf("UID read failed: %s", swmsg)
		} else {
			log.Printf("UID read failed: %v", err)
		}
	}
	// 4) ATR hash
	return &CardData{
		IDNumber: atrID(atr),
		Source:   "atr-hash",
	}
}

// ----- PKCS#11 cert reading -----

// func readCertSubject() (CertInfo, bool) {
// 	module := strings.TrimSpace(os.Getenv("PKCS11_MODULE"))
// 	if module == "" {
// 		module = defaultPKCS11ModulePath()
// 	}
// 	if module == "" {
// 		log.Println("PKCS11_MODULE not set and no default module path for this OS; skipping PKCS#11 cert read")
// 		return CertInfo{}, false
// 	}

// 	certs, err := readPublicCertsPKCS11(module)
// 	if err != nil || len(certs) == 0 {
// 		if err != nil {
// 			log.Printf("PKCS#11 cert read failed (%s): %v", module, err)
// 		} else {
// 			log.Printf("PKCS#11 found no certificates (%s)", module)
// 		}
// 		return CertInfo{}, false
// 	}
// 	return certs[0], true // pick first; filter by label if desired
// }

// func defaultPKCS11ModulePath() string {
// 	switch runtime.GOOS {
// 	case "linux":
// 		return "/usr/lib/x86_64-linux-gnu/libeopproxy11.so" // CZ eObčanka proxy (adjust if needed)
// 	case "darwin":
// 		return "/usr/local/lib/eOPCZE/libeopproxy11.dylib" // adjust if needed
// 	case "windows":
// 		// Set PKCS11_MODULE to the middleware DLL path (varies by install)
// 		return ""
// 	default:
// 		return ""
// 	}
// }

func readPublicCertsPKCS11(modulePath string) ([]CertInfo, error) {
	p := pkcs11.New(modulePath)
	if p == nil {
		return nil, fmt.Errorf("pkcs11.New returned nil for %q (CGO disabled or bad module?)", modulePath)
	}
	if err := p.Initialize(); err != nil {
		return nil, fmt.Errorf("pkcs11 initialize %q failed: %w", modulePath, err)
	}
	defer func() {
		p.Destroy()
		p.Finalize()
	}()

	slots, err := p.GetSlotList(true)
	if err != nil {
		return nil, fmt.Errorf("GetSlotList: %w", err)
	}
	if len(slots) == 0 {
		return nil, errors.New("no PKCS#11 slots with token")
	}

	for _, slot := range slots {
		sess, err := p.OpenSession(slot, pkcs11.CKF_SERIAL_SESSION)
		if err != nil {
			continue
		}
		defer p.CloseSession(sess)

		if err := p.FindObjectsInit(sess, []*pkcs11.Attribute{
			pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_CERTIFICATE),
		}); err != nil {
			continue
		}
		objs, _, err := p.FindObjects(sess, 50)
		p.FindObjectsFinal(sess)
		if err != nil {
			continue
		}

		var out []CertInfo
		for _, o := range objs {
			attrs, err := p.GetAttributeValue(sess, o, []*pkcs11.Attribute{
				pkcs11.NewAttribute(pkcs11.CKA_VALUE, nil),
				pkcs11.NewAttribute(pkcs11.CKA_LABEL, nil),
			})
			if err != nil {
				continue
			}
			var der, label []byte
			for _, a := range attrs {
				switch a.Type {
				case pkcs11.CKA_VALUE:
					der = a.Value
				case pkcs11.CKA_LABEL:
					label = a.Value
				}
			}
			if len(der) == 0 {
				continue
			}
			cert, err := x509.ParseCertificate(der)
			if err != nil {
				continue
			}
			out = append(out, CertInfo{
				Label:   string(label),
				Subject: cert.Subject, // pkix.Name
				RawDER:  der,
			})
		}
		if len(out) > 0 {
			return out, nil
		}
	}
	return nil, errors.New("no certificates found via PKCS#11")
}

// ----- Name extraction helpers -----

var (
	oidGivenName = asn1.ObjectIdentifier{2, 5, 4, 42}
	oidSurname   = asn1.ObjectIdentifier{2, 5, 4, 4}
)

func extractName(n pkix.Name) (given, surname string) {
	for _, atv := range append(n.Names, n.ExtraNames...) {
		switch {
		case atv.Type.Equal(oidGivenName):
			if s, ok := atv.Value.(string); ok && given == "" {
				given = s
			}
		case atv.Type.Equal(oidSurname):
			if s, ok := atv.Value.(string); ok && surname == "" {
				surname = s
			}
		}
	}
	// Fallback: sometimes CommonName contains "Given Surname"
	if given == "" && surname == "" && n.CommonName != "" {
		parts := strings.Fields(n.CommonName)
		if len(parts) > 0 {
			given = parts[0]
		}
		if len(parts) > 1 {
			surname = strings.Join(parts[1:], " ")
		}
	}
	return
}

// ----- APDU fallbacks -----

// readCPLC tries GlobalPlatform GET DATA (CPLC) 9F7F and extracts an IC serial
func readCPLC(ctx scard.Context, reader string) (cplcHex string, icSerial string, err error) {
	card, err := ctx.Connect(reader, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		return "", "", err
	}
	defer card.Disconnect(scard.LeaveCard)

	apdu := []byte{0x80, 0xCA, 0x9F, 0x7F, 0x00}
	resp, err := card.Transmit(apdu)
	if err != nil || len(resp) < 2 {
		return "", "", fmt.Errorf("CPLC transmit failed")
	}

	sw1, sw2 := resp[len(resp)-2], resp[len(resp)-1]
	if sw1 != 0x90 || sw2 != 0x00 {
		return "", "", fmt.Errorf("GET DATA 9F7F failed SW=%02X%02X", sw1, sw2)
	}
	data := resp[:len(resp)-2]

	// Optional TLV 9F7F <len> <value>
	if len(data) > 3 && data[0] == 0x9F && data[1] == 0x7F {
		l := int(data[2])
		if 3+l <= len(data) {
			data = data[3 : 3+l]
		}
	}
	cplcHex = strings.ToUpper(hex.EncodeToString(data))

	// Common mapping: bytes 13..16 as IC serial (best-effort)
	if len(data) >= 16 {
		icSerial = strings.ToUpper(hex.EncodeToString(data[12:16]))
	}
	return cplcHex, icSerial, nil
}

// readUID tries PC/SC vendor GET DATA for contactless UID (not universal)
func readUID(ctx scard.Context, reader string) (string, error) {
	card, err := ctx.Connect(reader, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		return "", err
	}
	defer card.Disconnect(scard.LeaveCard)

	apdu := []byte{0xFF, 0xCA, 0x00, 0x00, 0x00}
	resp, err := card.Transmit(apdu)
	if err != nil || len(resp) < 2 {
		return "", fmt.Errorf("UID transmit failed")
	}
	sw1, sw2 := resp[len(resp)-2], resp[len(resp)-1]
	if sw1 != 0x90 || sw2 != 0x00 {
		return "", fmt.Errorf("UID get failed SW=%02X%02X", sw1, sw2)
	}
	data := resp[:len(resp)-2]
	return strings.ToUpper(hex.EncodeToString(data)), nil
}

// explainSW adds a human hint for common SW codes found above.
func explainSW(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "SW=6E00"):
		return "GET DATA (CPLC): CLA not supported on this card (normal for many eIDs)."
	case strings.Contains(msg, "SW=6D00"):
		return "GET DATA: INS not supported on this card."
	case strings.Contains(msg, "SW=6881"):
		return "UID command: function/logical channel not supported by this reader/card."
	default:
		return ""
	}
}

// ADD THIS HELPER SOMEWHERE NEAR YOUR UTILS
func firstExistingPath(paths ...string) string {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// OPTIONAL: SMALL LOGGING TWEAK — REPLACE readCertSubject()’s module resolution with this snippet
func readCertSubject() (CertInfo, bool) {
	cands := pkcs11ModuleCandidates()
	if len(cands) == 0 {
		log.Println("PKCS#11: no candidate module paths; set PKCS11_MODULE or install OpenSC.")
		return CertInfo{}, false
	}
	for _, mod := range cands {
		if mod == "" {
			continue
		}
		if _, err := os.Stat(mod); err != nil {
			continue
		}
		log.Printf("Trying PKCS#11 module: %s", mod)
		certs, err := readPublicCertsPKCS11(mod)
		if err != nil {
			log.Printf("PKCS#11 attempt failed (%s): %v", mod, err)
			continue
		}
		if len(certs) == 0 {
			log.Printf("PKCS#11 found no certificates (%s)", mod)
			continue
		}
		// success
		return certs[0], true
	}
	log.Println("PKCS#11: no usable module initialized; all candidates failed.")
	return CertInfo{}, false
}

func atrID(atr []byte) string {
	sum := sha256.Sum256(atr)
	// short stable ID: first 8 bytes (16 hex chars)
	return "ATR-" + strings.ToUpper(hex.EncodeToString(sum[:8]))
}

// ADD: candidate list helper (near your utils)
func pkcs11ModuleCandidates() []string {
	// Respect explicit override first
	if m := strings.TrimSpace(os.Getenv("PKCS11_MODULE")); m != "" {
		return []string{m}
	}
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Library/OpenSC/lib/opensc-pkcs11.so", // ← cask (correct)
			"/Library/OpenSC/lib/opensc-pkcs11.dylib",
			"/opt/homebrew/lib/opensc-pkcs11.dylib",
			"/usr/local/lib/opensc-pkcs11.dylib",
			"/usr/local/lib/eOPCZE/libeopproxy11.dylib",
		}
	case "linux":
		return []string{
			"/usr/lib/x86_64-linux-gnu/opensc-pkcs11.so",
			"/usr/lib64/opensc-pkcs11.so",
			"/usr/lib/opensc-pkcs11.so",
			"/usr/local/lib/opensc-pkcs11.so",
			"/usr/lib/x86_64-linux-gnu/libeopproxy11.so",
		}
	default:
		return nil
	}
}

// -----------------------------
// WebSocket Communication
// -----------------------------

func sendStateUpdate(wsURL string, deviceID, roomID, reader, state, message string) {
	payload := Payload{
		DeviceID:   deviceID,
		RoomID:     roomID,
		Token:      randToken(8), // Shorter token for state updates
		Reader:     reader,
		ATR:        "",
		Protocol:   "",
		OccurredAt: time.Now().Format(time.RFC3339),
		State:      state,
		Message:    message,
	}

	sendToWebSocket(wsURL, payload)
}

func sendToWebSocket(wsURL string, payload Payload) {
	// Parse WebSocket URL
	u, err := url.Parse(wsURL)
	if err != nil {
		log.Printf("Invalid WebSocket URL: %v", err)
		return
	}

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("Failed to connect to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Send payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, payloadBytes)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}

	if payload.State != "" {
		log.Printf("State update sent: %s - %s", payload.State, payload.Message)
	} else {
		log.Printf("Card data sent to WebSocket successfully")
	}
}
