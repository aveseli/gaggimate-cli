// Package api provides HTTP and WebSocket clients for the Gaggimate device.
package api

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aveseli/gaggimate-cli/internal/parser"
)

// HTTPClient communicates with the Gaggimate device over HTTP.
type HTTPClient struct {
	Host       string
	UseHTTPS   bool
	Timeout    time.Duration
	MaxRetries int
}

// NewHTTPClient creates a new HTTP client with default settings.
func NewHTTPClient(host string, useHTTPS bool) *HTTPClient {
	return &HTTPClient{
		Host:       host,
		UseHTTPS:   useHTTPS,
		Timeout:    15 * time.Second,
		MaxRetries: 3,
	}
}

func (c *HTTPClient) baseURL() string {
	proto := "http"
	if c.UseHTTPS {
		proto = "https"
	}
	return fmt.Sprintf("%s://%s/api/history", proto, c.Host)
}

// FetchShotIndex fetches and parses the shot history index.
func (c *HTTPClient) FetchShotIndex() ([]parser.ShotListItem, error) {
	url := c.baseURL() + "/index.bin"

	resp, err := c.doGet(url)
	if err != nil {
		return nil, fmt.Errorf("fetching shot index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []parser.ShotListItem{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading index response: %w", err)
	}

	if parser.IsHTMLResponse(data) {
		return nil, fmt.Errorf("device returned HTML instead of binary data for index.bin")
	}

	indexData, err := parser.ParseBinaryIndex(data)
	if err != nil {
		return nil, fmt.Errorf("parsing index: %w", err)
	}

	return parser.IndexToShotList(indexData), nil
}

// FetchShot fetches and parses a specific shot by ID.
func (c *HTTPClient) FetchShot(shotID string) (*parser.ShotData, error) {
	paddedID := shotID
	for len(paddedID) < 6 {
		paddedID = "0" + paddedID
	}
	url := fmt.Sprintf("%s/%s.slog", c.baseURL(), paddedID)

	var lastErr error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
		}

		data, err := c.fetchBytes(url)
		if err != nil {
			lastErr = err
			continue
		}

		shot, err := parser.ParseBinaryShot(data, paddedID)
		if err != nil {
			// HTML responses are retryable
			if parser.IsHTMLResponse(data) {
				lastErr = err
				continue
			}
			return nil, fmt.Errorf("parsing shot %s: %w", paddedID, err)
		}
		return shot, nil
	}

	return nil, fmt.Errorf("fetching shot %s after %d retries: %w", paddedID, c.MaxRetries, lastErr)
}

func (c *HTTPClient) doGet(url string) (*http.Response, error) {
	client := &http.Client{Timeout: c.Timeout}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")
	return client.Do(req)
}

func (c *HTTPClient) fetchBytes(url string) ([]byte, error) {
	resp, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("shot not found (404)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}
