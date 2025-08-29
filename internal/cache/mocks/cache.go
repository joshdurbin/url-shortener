package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/joshdurbin/url-shortener/internal/domain"
)

// Cache is a mock implementation of cache.Cache
type Cache struct {
	mock.Mock
}

// Get retrieves a cache entry by short code
func (m *Cache) Get(ctx context.Context, shortCode string) (*domain.CacheEntry, bool) {
	args := m.Called(ctx, shortCode)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*domain.CacheEntry), args.Bool(1)
}

// Set stores a cache entry
func (m *Cache) Set(ctx context.Context, shortCode string, entry *domain.CacheEntry) error {
	args := m.Called(ctx, shortCode, entry)
	return args.Error(0)
}

// Delete removes a cache entry
func (m *Cache) Delete(ctx context.Context, shortCode string) error {
	args := m.Called(ctx, shortCode)
	return args.Error(0)
}

// IncrementUsage increments the usage count for a short code
func (m *Cache) IncrementUsage(ctx context.Context, shortCode string) error {
	args := m.Called(ctx, shortCode)
	return args.Error(0)
}

// GetDirtyEntries returns all cache entries that need to be synced to the database
func (m *Cache) GetDirtyEntries(ctx context.Context) (map[string]*domain.CacheEntry, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*domain.CacheEntry), args.Error(1)
}

// MarkClean marks a cache entry as clean (synced to database)
func (m *Cache) MarkClean(ctx context.Context, shortCode string) error {
	args := m.Called(ctx, shortCode)
	return args.Error(0)
}

// LoadData loads data into the cache from a map
func (m *Cache) LoadData(ctx context.Context, data map[string]*domain.CacheEntry) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

// Close closes the cache connection (if applicable)
func (m *Cache) Close() error {
	args := m.Called()
	return args.Error(0)
}

// SyncableCache is a mock implementation of cache.SyncableCache
type SyncableCache struct {
	Cache
}

// StartBackgroundSync starts background synchronization with the given interval
func (m *SyncableCache) StartBackgroundSync(ctx context.Context, interval time.Duration, syncFunc func(map[string]*domain.CacheEntry) error) error {
	args := m.Called(ctx, interval, syncFunc)
	return args.Error(0)
}

// StopBackgroundSync stops background synchronization
func (m *SyncableCache) StopBackgroundSync() error {
	args := m.Called()
	return args.Error(0)
}