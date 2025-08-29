package domain

import (
	"time"
)

// URLEntry represents a shortened URL with its metadata
type URLEntry struct {
	ID          int        `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	UsageCount  int        `json:"usage_count"`
}

// CacheEntry represents an entry in the cache
type CacheEntry struct {
	OriginalURL string    `json:"original_url"`
	UsageCount  int       `json:"usage_count"`
	LastUsedAt  time.Time `json:"last_used_at"`
	Dirty       bool      `json:"dirty"` // Indicates if the entry needs to be synced to DB
}

// CreateURLRequest represents the request to create a short URL
type CreateURLRequest struct {
	URL string `json:"url"`
}

// CreateURLResponse represents the response when creating a short URL
type CreateURLResponse struct {
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}