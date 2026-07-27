package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient communicates with the Gaggimate device over WebSocket.
type WebSocketClient struct {
	Host     string
	UseHTTPS bool
	Timeout  time.Duration
}

// NewWebSocketClient creates a new WebSocket client.
func NewWebSocketClient(host string, useHTTPS bool) *WebSocketClient {
	return &WebSocketClient{
		Host:     host,
		UseHTTPS: useHTTPS,
		Timeout:  15 * time.Second,
	}
}

func (c *WebSocketClient) wsURL() string {
	proto := "ws"
	if c.UseHTTPS {
		proto = "wss"
	}
	return fmt.Sprintf("%s://%s/ws", proto, c.Host)
}

func generateRequestID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	suffix := make([]byte, 9)
	for i := range suffix {
		suffix[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("cli-%d-%s", time.Now().UnixMilli(), string(suffix))
}

// wsResponse is a WebSocket response message.
type wsResponse struct {
	TP       string          `json:"tp"`
	RID      string          `json:"rid"`
	Error    string          `json:"error,omitempty"`
	Profiles json.RawMessage `json:"profiles,omitempty"`
	Profile  json.RawMessage `json:"profile,omitempty"`
	Notes    json.RawMessage `json:"notes,omitempty"`
	Msg      string          `json:"msg,omitempty"`
}

func (c *WebSocketClient) sendRequest(requestType string, payload interface{}) (*wsResponse, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: c.Timeout,
	}
	conn, _, err := dialer.Dial(c.wsURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("websocket dial: %w", err)
	}
	defer conn.Close()

	rid := generateRequestID()

	// Build request
	reqMap := map[string]interface{}{
		"tp":  requestType,
		"rid": rid,
	}
	// Merge payload
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		for k, v := range payloadMap {
			reqMap[k] = v
		}
	}

	if err := conn.WriteJSON(reqMap); err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	// Read responses until we get ours or timeout
	conn.SetReadDeadline(time.Now().Add(c.Timeout))
	responseType := "res:" + requestType[4:] // req:xxx -> res:xxx

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}

		var resp wsResponse
		if err := json.Unmarshal(msg, &resp); err != nil {
			continue // skip non-JSON messages
		}

		if resp.TP == responseType && resp.RID == rid {
			if resp.Error != "" {
				return nil, fmt.Errorf("API error: %s", resp.Error)
			}
			return &resp, nil
		}
		// Ignore other messages (evt:status, etc.)
	}
}

// Profile represents a Gaggimate brewing profile.
type Profile struct {
	ID          string          `json:"id,omitempty"`
	Label       string          `json:"label"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Temperature float64         `json:"temperature"`
	Favorite    bool            `json:"favorite"`
	Selected    bool            `json:"selected"`
	Utility     bool            `json:"utility"`
	Phases      json.RawMessage `json:"phases"`
}

// ListProfiles returns all profiles on the device.
func (c *WebSocketClient) ListProfiles() ([]Profile, error) {
	resp, err := c.sendRequest("req:profiles:list", nil)
	if err != nil {
		return nil, err
	}

	var profiles []Profile
	if err := json.Unmarshal(resp.Profiles, &profiles); err != nil {
		return nil, fmt.Errorf("parsing profiles: %w", err)
	}
	return profiles, nil
}

// GetProfile returns a specific profile by ID.
func (c *WebSocketClient) GetProfile(id string) (*Profile, error) {
	resp, err := c.sendRequest("req:profiles:load", map[string]interface{}{"id": id})
	if err != nil {
		return nil, err
	}

	var profile Profile
	if err := json.Unmarshal(resp.Profile, &profile); err != nil {
		return nil, fmt.Errorf("parsing profile: %w", err)
	}
	return &profile, nil
}

// DeleteProfile deletes a profile by ID.
func (c *WebSocketClient) DeleteProfile(id string) error {
	_, err := c.sendRequest("req:profiles:delete", map[string]interface{}{"id": id})
	return err
}

// SelectProfile selects a profile as the active brewing profile.
func (c *WebSocketClient) SelectProfile(id string) error {
	_, err := c.sendRequest("req:profiles:select", map[string]interface{}{"id": id})
	return err
}

// CreateProfile creates a new profile on the device. Returns the new profile ID.
func (c *WebSocketClient) CreateProfile(profile Profile) (string, error) {
	// Build the payload with all profile fields
	payload := map[string]interface{}{
		"label":       profile.Label,
		"type":        profile.Type,
		"description": profile.Description,
		"temperature": profile.Temperature,
		"phases":      profile.Phases,
	}

	resp, err := c.sendRequest("req:profiles:create", payload)
	if err != nil {
		return "", err
	}

	// The response should contain the new profile with its ID
	var created Profile
	if err := json.Unmarshal(resp.Profile, &created); err != nil {
		// If no profile returned, try to extract ID from response message
		if resp.Msg != "" {
			return resp.Msg, nil
		}
		return "", fmt.Errorf("parsing created profile: %w", err)
	}

	return created.ID, nil
}

// UpdateProfile updates an existing profile on the device.
func (c *WebSocketClient) UpdateProfile(id string, profile Profile) error {
	payload := map[string]interface{}{
		"id":          id,
		"label":       profile.Label,
		"type":        profile.Type,
		"description": profile.Description,
		"temperature": profile.Temperature,
		"phases":      profile.Phases,
	}

	_, err := c.sendRequest("req:profiles:update", payload)
	return err
}

// ShotNotes represents notes for a shot.
type ShotNotes struct {
	Rating       *int     `json:"rating,omitempty"`
	Notes        string   `json:"notes,omitempty"`
	BalanceTaste string   `json:"balanceTaste,omitempty"`
	GrindSetting string   `json:"grindSetting,omitempty"`
	DoseIn       *float64 `json:"doseIn,omitempty"`
	DoseOut      *float64 `json:"doseOut,omitempty"`
}

// GetShotNotes retrieves notes for a specific shot.
func (c *WebSocketClient) GetShotNotes(shotID string) (*ShotNotes, error) {
	normalizedID := normalizeID(shotID)
	resp, err := c.sendRequest("req:history:notes:get", map[string]interface{}{"id": normalizedID})
	if err != nil {
		return nil, err
	}

	var notes ShotNotes
	if err := json.Unmarshal(resp.Notes, &notes); err != nil {
		return nil, fmt.Errorf("parsing notes: %w", err)
	}
	return &notes, nil
}

// SaveShotNotes saves notes for a specific shot.
func (c *WebSocketClient) SaveShotNotes(shotID string, notes ShotNotes) error {
	normalizedID := normalizeID(shotID)

	notesData := map[string]interface{}{}
	if notes.Rating != nil {
		notesData["rating"] = *notes.Rating
	}
	if notes.Notes != "" {
		notesData["notes"] = notes.Notes
	}
	if notes.BalanceTaste != "" {
		notesData["balanceTaste"] = notes.BalanceTaste
	}
	if notes.GrindSetting != "" {
		notesData["grindSetting"] = notes.GrindSetting
	}
	if notes.DoseIn != nil {
		notesData["doseIn"] = *notes.DoseIn
	}
	if notes.DoseOut != nil {
		notesData["doseOut"] = *notes.DoseOut
	}

	_, err := c.sendRequest("req:history:notes:save", map[string]interface{}{
		"id":    normalizedID,
		"notes": notesData,
	})
	return err
}

func normalizeID(id string) string {
	n := 0
	for _, c := range id {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return fmt.Sprintf("%d", n)
}
