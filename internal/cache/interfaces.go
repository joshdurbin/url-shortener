package cache

import (
	"context"
	"time"

	"github.com/joshdurbin/url-shortener/internal/domain"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a cache entry by short code
	Get(ctx context.Context, shortCode string) (*domain.CacheEntry, bool)
	
	// Set stores a cache entry
	Set(ctx context.Context, shortCode string, entry *domain.CacheEntry) error
	
	// Delete removes a cache entry
	Delete(ctx context.Context, shortCode string) error
	
	// IncrementUsage increments the usage count for a short code
	IncrementUsage(ctx context.Context, shortCode string) error
	
	// GetDirtyEntries returns all cache entries that need to be synced to the database
	GetDirtyEntries(ctx context.Context) (map[string]*domain.CacheEntry, error)
	
	// MarkClean marks a cache entry as clean (synced to database)
	MarkClean(ctx context.Context, shortCode string) error
	
	// LoadData loads data into the cache from a map
	LoadData(ctx context.Context, data map[string]*domain.CacheEntry) error
	
	// Close closes the cache connection (if applicable)
	Close() error
}

// SyncableCache extends Cache with sync capabilities
type SyncableCache interface {
	Cache
	
	// StartBackgroundSync starts background synchronization with the given interval
	StartBackgroundSync(ctx context.Context, interval time.Duration, syncFunc func(map[string]*domain.CacheEntry) error) error
	
	// StopBackgroundSync stops background synchronization
	StopBackgroundSync() error
}