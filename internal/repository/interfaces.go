package repository

import (
	"context"
	"time"

	"github.com/joshdurbin/url-shortener/db/sqlc"
	"github.com/joshdurbin/url-shortener/internal/domain"
)

// URLRepository defines the interface for URL data operations
type URLRepository interface {
	// CreateURL creates a new short URL entry
	CreateURL(ctx context.Context, shortCode, originalURL string, createdAt time.Time) (*domain.URLEntry, error)
	
	// GetURL retrieves a URL entry by its short code
	GetURL(ctx context.Context, shortCode string) (*domain.URLEntry, error)
	
	// GetAllURLs retrieves all URL entries ordered by creation date (desc)
	GetAllURLs(ctx context.Context) ([]*domain.URLEntry, error)
	
	// UpdateUsage updates the usage count and last used timestamp for a URL
	UpdateUsage(ctx context.Context, shortCode string, usageCount int, lastUsedAt time.Time) error
	
	// DeleteURL removes a URL entry by its short code
	DeleteURL(ctx context.Context, shortCode string) error
	
	// URLExists checks if a short code exists
	URLExists(ctx context.Context, shortCode string) (bool, error)
	
	// LoadCacheData loads all URL data for cache initialization
	LoadCacheData(ctx context.Context) (map[string]*domain.CacheEntry, error)
	
	// GetQueries returns the underlying sqlc queries for advanced operations
	GetQueries() *sqlc.Queries
	
	// Close closes the repository connection
	Close() error
}