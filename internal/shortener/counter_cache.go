package shortener

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/joshdurbin/url-shortener/db/sqlc"
)

// CounterCache provides an in-memory counter cache with async writeback to database
type CounterCache struct {
	mu           sync.RWMutex
	db           *sqlc.Queries
	counters     map[string]*cacheEntry
	jumpAhead    int64
	writebackCh  chan writebackRequest
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

type cacheEntry struct {
	current   int64
	allocated int64
	dirty     bool
}

type writebackRequest struct {
	key   string
	value int64
	resp  chan error
}

// NewCounterCache creates a new counter cache
func NewCounterCache(db *sqlc.Queries, jumpAhead int64) *CounterCache {
	cache := &CounterCache{
		db:          db,
		counters:    make(map[string]*cacheEntry),
		jumpAhead:   jumpAhead,
		writebackCh: make(chan writebackRequest, 100),
		stopCh:      make(chan struct{}),
	}
	
	// Start writeback goroutine
	cache.wg.Add(1)
	go cache.writebackWorker()
	
	return cache
}

// GetNextCounter returns the next counter value, allocating more from DB if needed
func (c *CounterCache) GetNextCounter(ctx context.Context, key string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	entry, exists := c.counters[key]
	if !exists {
		// Initialize from database
		dbValue, err := c.db.GetCounter(ctx, key)
		if err != nil && err != sql.ErrNoRows {
			return 0, fmt.Errorf("failed to get counter from DB: %w", err)
		}
		
		if err == sql.ErrNoRows {
			dbValue = 0
		}
		
		entry = &cacheEntry{
			current:   dbValue,
			allocated: dbValue + c.jumpAhead,
			dirty:     true,
		}
		c.counters[key] = entry
		
		// Async writeback of allocated value
		go c.asyncWriteback(key, entry.allocated)
	}
	
	// Check if we need to allocate more
	if entry.current >= entry.allocated {
		newAllocated := entry.allocated + c.jumpAhead
		oldAllocated := entry.allocated
		entry.allocated = newAllocated
		entry.dirty = true
		
		// Async writeback of new allocated value
		go c.asyncWriteback(key, newAllocated)
		
		// Update current to continue from where we left off
		entry.current = oldAllocated
	}
	
	entry.current++
	return entry.current, nil
}

// SetCounter sets a counter value
func (c *CounterCache) SetCounter(ctx context.Context, key string, value int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	entry := &cacheEntry{
		current:   value,
		allocated: value + c.jumpAhead,
		dirty:     true,
	}
	c.counters[key] = entry
	
	// Async writeback
	go c.asyncWriteback(key, entry.allocated)
	
	return nil
}

// asyncWriteback performs async writeback without blocking
func (c *CounterCache) asyncWriteback(key string, value int64) {
	select {
	case c.writebackCh <- writebackRequest{key: key, value: value}:
	case <-c.stopCh:
		return
	default:
		// Channel full, skip this writeback (will be handled by sync on close)
		return
	}
}

// writebackWorker handles async writebacks
func (c *CounterCache) writebackWorker() {
	defer c.wg.Done()
	
	for {
		select {
		case req := <-c.writebackCh:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := c.db.SetCounter(ctx, sqlc.SetCounterParams{
				Key:   req.key,
				Value: req.value,
			})
			cancel()
			
			if req.resp != nil {
				req.resp <- err
			}
			
		case <-c.stopCh:
			// Process remaining requests
			for {
				select {
				case req := <-c.writebackCh:
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					err := c.db.SetCounter(ctx, sqlc.SetCounterParams{
						Key:   req.key,
						Value: req.value,
					})
					cancel()
					
					if req.resp != nil {
						req.resp <- err
					}
				default:
					return
				}
			}
		}
	}
}

// Sync synchronously writes all dirty entries to database
func (c *CounterCache) Sync(ctx context.Context) error {
	c.mu.RLock()
	dirtyEntries := make(map[string]int64)
	for key, entry := range c.counters {
		if entry.dirty {
			dirtyEntries[key] = entry.allocated
		}
	}
	c.mu.RUnlock()
	
	for key, value := range dirtyEntries {
		if err := c.db.SetCounter(ctx, sqlc.SetCounterParams{
			Key:   key,
			Value: value,
		}); err != nil {
			return fmt.Errorf("failed to sync counter %s: %w", key, err)
		}
		
		// Mark as clean
		c.mu.Lock()
		if entry, exists := c.counters[key]; exists {
			entry.dirty = false
		}
		c.mu.Unlock()
	}
	
	return nil
}

// Close closes the counter cache and syncs all dirty entries
func (c *CounterCache) Close() error {
	select {
	case <-c.stopCh:
		// Already closed
		return nil
	default:
		close(c.stopCh)
	}
	
	c.wg.Wait()
	
	// Sync all dirty entries
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	return c.Sync(ctx)
}

// Ensure CounterCache implements CounterProvider
var _ CounterProvider = (*CounterCache)(nil)