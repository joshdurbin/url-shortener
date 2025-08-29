package service

import (
	"context"
	"time"

	"github.com/joshdurbin/url-shortener/internal/domain"
)

// URLShortener defines the interface for URL shortening operations
type URLShortener interface {
	// CreateShortURL creates a new short URL
	CreateShortURL(ctx context.Context, originalURL string) (*domain.URLEntry, error)
	
	// GetOriginalURL retrieves the original URL for a short code and increments usage
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
	
	// GetURLInfo retrieves detailed information about a short URL
	GetURLInfo(ctx context.Context, shortCode string) (*domain.URLEntry, error)
	
	// DeleteShortURL removes a short URL
	DeleteShortURL(ctx context.Context, shortCode string) error
	
	// GetAllURLs retrieves all short URLs with current cache data
	GetAllURLs(ctx context.Context) ([]*domain.URLEntry, error)
	
	// InitializeCache loads data from repository into cache
	InitializeCache(ctx context.Context) error
	
	// StartCacheSync starts background cache synchronization
	StartCacheSync(ctx context.Context, interval time.Duration) error
	
	// StopCacheSync stops background cache synchronization
	StopCacheSync() error
	
	// Close closes the service and its dependencies
	Close() error
}