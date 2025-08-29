package memory

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/joshdurbin/url-shortener/internal/cache"
	"github.com/joshdurbin/url-shortener/internal/domain"
)

// Cache implements cache.SyncableCache using in-memory storage
type Cache struct {
	data     map[string]*domain.CacheEntry
	mutex    sync.RWMutex
	stopChan chan struct{}
	running  bool
}

// New creates a new in-memory cache
func New() *Cache {
	return &Cache{
		data:     make(map[string]*domain.CacheEntry),
		stopChan: make(chan struct{}),
	}
}

// Get retrieves a cache entry by short code
func (c *Cache) Get(ctx context.Context, shortCode string) (*domain.CacheEntry, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	entry, exists := c.data[shortCode]
	if !exists {
		return nil, false
	}
	
	// Return a copy to prevent external modification
	return &domain.CacheEntry{
		OriginalURL: entry.OriginalURL,
		UsageCount:  entry.UsageCount,
		LastUsedAt:  entry.LastUsedAt,
		Dirty:       entry.Dirty,
	}, true
}

// Set stores a cache entry
func (c *Cache) Set(ctx context.Context, shortCode string, entry *domain.CacheEntry) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Store a copy to prevent external modification
	c.data[shortCode] = &domain.CacheEntry{
		OriginalURL: entry.OriginalURL,
		UsageCount:  entry.UsageCount,
		LastUsedAt:  entry.LastUsedAt,
		Dirty:       entry.Dirty,
	}
	
	return nil
}

// Delete removes a cache entry
func (c *Cache) Delete(ctx context.Context, shortCode string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	delete(c.data, shortCode)
	return nil
}

// IncrementUsage increments the usage count for a short code
func (c *Cache) IncrementUsage(ctx context.Context, shortCode string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if entry, exists := c.data[shortCode]; exists {
		entry.UsageCount++
		entry.LastUsedAt = time.Now()
		entry.Dirty = true
	}
	
	return nil
}

// GetDirtyEntries returns all cache entries that need to be synced to the database
func (c *Cache) GetDirtyEntries(ctx context.Context) (map[string]*domain.CacheEntry, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	dirty := make(map[string]*domain.CacheEntry)
	for shortCode, entry := range c.data {
		if entry.Dirty {
			// Return a copy
			dirty[shortCode] = &domain.CacheEntry{
				OriginalURL: entry.OriginalURL,
				UsageCount:  entry.UsageCount,
				LastUsedAt:  entry.LastUsedAt,
				Dirty:       entry.Dirty,
			}
		}
	}
	
	return dirty, nil
}

// MarkClean marks a cache entry as clean (synced to database)
func (c *Cache) MarkClean(ctx context.Context, shortCode string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if entry, exists := c.data[shortCode]; exists {
		entry.Dirty = false
	}
	
	return nil
}

// LoadData loads data into the cache from a map
func (c *Cache) LoadData(ctx context.Context, data map[string]*domain.CacheEntry) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Clear existing data and load new data
	c.data = make(map[string]*domain.CacheEntry)
	for shortCode, entry := range data {
		// Store a copy
		c.data[shortCode] = &domain.CacheEntry{
			OriginalURL: entry.OriginalURL,
			UsageCount:  entry.UsageCount,
			LastUsedAt:  entry.LastUsedAt,
			Dirty:       entry.Dirty,
		}
	}
	
	return nil
}

// StartBackgroundSync starts background synchronization with the given interval
func (c *Cache) StartBackgroundSync(ctx context.Context, interval time.Duration, syncFunc func(map[string]*domain.CacheEntry) error) error {
	c.mutex.Lock()
	if c.running {
		c.mutex.Unlock()
		return nil // Already running
	}
	c.running = true
	c.mutex.Unlock()

	go c.backgroundSync(ctx, interval, syncFunc)
	return nil
}

// StopBackgroundSync stops background synchronization
func (c *Cache) StopBackgroundSync() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if !c.running {
		return nil
	}
	
	c.running = false
	close(c.stopChan)
	
	// Create new channel for potential restart
	c.stopChan = make(chan struct{})
	return nil
}

// backgroundSync runs the background synchronization loop
func (c *Cache) backgroundSync(ctx context.Context, interval time.Duration, syncFunc func(map[string]*domain.CacheEntry) error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Get a copy of stopChan to avoid race condition
	c.mutex.RLock()
	stopChan := c.stopChan
	c.mutex.RUnlock()

	for {
		select {
		case <-ticker.C:
			c.syncToDatabase(ctx, syncFunc)
		case <-stopChan:
			// Final sync before stopping
			c.syncToDatabase(ctx, syncFunc)
			return
		case <-ctx.Done():
			return
		}
	}
}

// syncToDatabase syncs dirty entries to the database
func (c *Cache) syncToDatabase(ctx context.Context, syncFunc func(map[string]*domain.CacheEntry) error) {
	dirtyEntries, err := c.GetDirtyEntries(ctx)
	if err != nil {
		log.Printf("Error getting dirty entries: %v", err)
		return
	}
	
	if len(dirtyEntries) == 0 {
		return
	}
	
	if err := syncFunc(dirtyEntries); err != nil {
		log.Printf("Error syncing cache entries to database: %v", err)
		return
	}
	
	// Mark entries as clean
	for shortCode := range dirtyEntries {
		if err := c.MarkClean(ctx, shortCode); err != nil {
			log.Printf("Error marking entry %s as clean: %v", shortCode, err)
		}
	}
}

// Close closes the cache (stops background sync)
func (c *Cache) Close() error {
	return c.StopBackgroundSync()
}

// Ensure Cache implements the interfaces
var _ cache.Cache = (*Cache)(nil)
var _ cache.SyncableCache = (*Cache)(nil)