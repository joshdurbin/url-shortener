package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/joshdurbin/url-shortener/internal/domain"
)

// Client represents an HTTP client for the URL shortener API
type Client struct {
	serverURL  string
	httpClient *http.Client
}

// NewClient creates a new URL shortener client
func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateURL creates a short URL
func (c *Client) CreateURL(ctx context.Context, originalURL string) (*domain.CreateURLResponse, error) {
	reqBody := domain.CreateURLRequest{URL: originalURL}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL+"/api/urls", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result domain.CreateURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetURL retrieves information about a short URL
func (c *Client) GetURL(ctx context.Context, shortCode string) (*domain.URLEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.serverURL+"/api/urls/"+shortCode, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("short code '%s' not found", shortCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var entry domain.URLEntry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &entry, nil
}

// DeleteURL deletes a short URL
func (c *Client) DeleteURL(ctx context.Context, shortCode string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.serverURL+"/api/urls/"+shortCode, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("short code '%s' not found", shortCode)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// ListURLs retrieves all short URLs
func (c *Client) ListURLs(ctx context.Context) ([]*domain.URLEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.serverURL+"/api/urls", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var entries []*domain.URLEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return entries, nil
}