package checkin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type Request struct {
	DeviceID string `json:"deviceId"`
	Token    string `json:"token"`  // random per insertion (or stable hash later)
	RoomID   string `json:"roomId"` // e.g. "triage-1"
}

type Response struct {
	TicketNumber int    `json:"ticketNumber"`
	Message      string `json:"message"`
}

// NewHTTPClient returns a tuned client you can reuse.
// Reuse this across calls (don’t make a new client each time).
func NewHTTPClient() *http.Client {
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
	}
	return &http.Client{
		Timeout:   5 * time.Second, // end-to-end request timeout
		Transport: tr,
	}
}

// CheckIn posts {deviceId, token, roomId} to POST {baseURL}/api/checkin
// and returns the assigned ticket number.
func CheckIn(ctx context.Context, client *http.Client, baseURL string, req Request) (int, error) {
	if client == nil {
		client = NewHTTPClient()
	}
	payload, _ := json.Marshal(req)

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/checkin", bytes.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "waiting-room-reader/1.0")

	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("checkin request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB cap

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("checkin HTTP %d: %s", resp.StatusCode, string(body))
	}

	var out Response
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, fmt.Errorf("decode checkin response: %w", err)
	}
	if out.TicketNumber == 0 {
		return 0, errors.New("missing ticketNumber in response")
	}
	return out.TicketNumber, nil
}
