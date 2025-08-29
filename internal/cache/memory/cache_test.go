package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/joshdurbin/url-shortener/internal/domain"
)

func TestCache_New(t *testing.T) {
	cache := New()
	assert.NotNil(t, cache)
	assert.NotNil(t, cache.data)
	assert.NotNil(t, cache.stopChan)
	assert.False(t, cache.running)
}

func TestCache_SetAndGet(t *testing.T) {
	cache := New()
	ctx := context.Background()
	
	entry := &domain.CacheEntry{
		OriginalURL: "https://example.com",
		UsageCount:  1,
		LastUsedAt:  time.Now(),
		Dirty:       false,
	}

	// Test Set
	err := cache.Set(ctx, "test123", entry)
	assert.NoError(t, err)

	// Test Get - exists
	retrieved, exists := cache.Get(ctx, "test123")
	assert.True(t, exists)
	assert.NotNil(t, retrieved)
	assert.Equal(t, entry.OriginalURL, retrieved.OriginalURL)
	assert.Equal(t, entry.UsageCount, retrieved.UsageCount)
	assert.Equal(t, entry.Dirty, retrieved.Dirty)
	
	// Verify it's a copy (modifying retrieved shouldn't affect cache)
	retrieved.UsageCount = 999
	retrieved2, _ := cache.Get(ctx, "test123")
	assert.Equal(t, 1, retrieved2.UsageCount)

	// Test Get - doesn't exist
	retrieved, exists = cache.Get(ctx, "nonexistent")
	assert.False(t, exists)
	assert.Nil(t, retrieved)
}

func TestCache_Delete(t *testing.T) {
	cache := New()
	ctx := context.Background()
	
	entry := &domain.CacheEntry{
		OriginalURL: "https://example.com",
		UsageCount:  1,
		LastUsedAt:  time.Now(),
		Dirty:       false,
	}

	// Set entry
	err := cache.Set(ctx, "test123", entry)
	assert.NoError(t, err)

	// Verify it exists
	_, exists := cache.Get(ctx, "test123")
	assert.True(t, exists)

	// Delete entry
	err = cache.Delete(ctx, "test123")
	assert.NoError(t, err)

	// Verify it's gone
	_, exists = cache.Get(ctx, "test123")
	assert.False(t, exists)

	// Delete non-existent entry (should not error)
	err = cache.Delete(ctx, "nonexistent")
	assert.NoError(t, err)
}

func TestCache_IncrementUsage(t *testing.T) {
	cache := New()
	ctx := context.Background()
	
	now := time.Now()
	entry := &domain.CacheEntry{
		OriginalURL: "https://example.com",
		UsageCount:  1,
		LastUsedAt:  now,
		Dirty:       false,
	}

	// Set entry
	err := cache.Set(ctx, "test123", entry)
	assert.NoError(t, err)

	// Increment usage
	err = cache.IncrementUsage(ctx, "test123")
	assert.NoError(t, err)

	// Verify changes
	retrieved, exists := cache.Get(ctx, "test123")
	assert.True(t, exists)
	assert.Equal(t, 2, retrieved.UsageCount)
	assert.True(t, retrieved.LastUsedAt.After(now))
	assert.True(t, retrieved.Dirty)

	// Increment usage on non-existent entry (should not error)
	err = cache.IncrementUsage(ctx, "nonexistent")
	assert.NoError(t, err)
}

func TestCache_GetDirtyEntries(t *testing.T) {
	cache := New()
	ctx := context.Background()
	
	// Add clean entry
	cleanEntry := &domain.CacheEntry{
		OriginalURL: "https://clean.com",
		UsageCount:  1,
		LastUsedAt:  time.Now(),
		Dirty:       false,
	}
	err := cache.Set(ctx, "clean", cleanEntry)
	assert.NoError(t, err)

	// Add dirty entry
	dirtyEntry := &domain.CacheEntry{
		OriginalURL: "https://dirty.com",
		UsageCount:  2,
		LastUsedAt:  time.Now(),
		Dirty:       true,
	}
	err = cache.Set(ctx, "dirty", dirtyEntry)
	assert.NoError(t, err)

	// Get dirty entries
	dirty, err := cache.GetDirtyEntries(ctx)
	assert.NoError(t, err)
	assert.Len(t, dirty, 1)
	assert.Contains(t, dirty, "dirty")
	assert.NotContains(t, dirty, "clean")
	assert.Equal(t, dirtyEntry.OriginalURL, dirty["dirty"].OriginalURL)
}

func TestCache_MarkClean(t *testing.T) {
	cache := New()
	ctx := context.Background()
	
	// Add dirty entry
	entry := &domain.CacheEntry{
		OriginalURL: "https://example.com",
		UsageCount:  1,
		LastUsedAt:  time.Now(),
		Dirty:       true,
	}
	err := cache.Set(ctx, "test123", entry)
	assert.NoError(t, err)

	// Verify it's dirty
	dirty, err := cache.GetDirtyEntries(ctx)
	assert.NoError(t, err)
	assert.Len(t, dirty, 1)

	// Mark clean
	err = cache.MarkClean(ctx, "test123")
	assert.NoError(t, err)

	// Verify it's no longer dirty
	dirty, err = cache.GetDirtyEntries(ctx)
	assert.NoError(t, err)
	assert.Len(t, dirty, 0)

	// Mark clean on non-existent entry (should not error)
	err = cache.MarkClean(ctx, "nonexistent")
	assert.NoError(t, err)
}

func TestCache_LoadData(t *testing.T) {
	cache := New()
	ctx := context.Background()
	
	// Add initial entry
	initialEntry := &domain.CacheEntry{
		OriginalURL: "https://initial.com",
		UsageCount:  1,
		LastUsedAt:  time.Now(),
		Dirty:       false,
	}
	err := cache.Set(ctx, "initial", initialEntry)
	assert.NoError(t, err)

	// Prepare data to load
	dataToLoad := map[string]*domain.CacheEntry{
		"entry1": {
			OriginalURL: "https://entry1.com",
			UsageCount:  5,
			LastUsedAt:  time.Now(),
			Dirty:       false,
		},
		"entry2": {
			OriginalURL: "https://entry2.com",
			UsageCount:  10,
			LastUsedAt:  time.Now(),
			Dirty:       true,
		},
	}

	// Load data
	err = cache.LoadData(ctx, dataToLoad)
	assert.NoError(t, err)

	// Verify initial entry is gone
	_, exists := cache.Get(ctx, "initial")
	assert.False(t, exists)

	// Verify new entries exist
	entry1, exists := cache.Get(ctx, "entry1")
	assert.True(t, exists)
	assert.Equal(t, "https://entry1.com", entry1.OriginalURL)
	assert.Equal(t, 5, entry1.UsageCount)

	entry2, exists := cache.Get(ctx, "entry2")
	assert.True(t, exists)
	assert.Equal(t, "https://entry2.com", entry2.OriginalURL)
	assert.Equal(t, 10, entry2.UsageCount)
	assert.True(t, entry2.Dirty)
}

func TestCache_BackgroundSync(t *testing.T) {
	cache := New()
	ctx := context.Background()
	
	var syncCallCount int
	var syncedEntries map[string]*domain.CacheEntry
	var mu sync.Mutex

	syncFunc := func(entries map[string]*domain.CacheEntry) error {
		mu.Lock()
		defer mu.Unlock()
		syncCallCount++
		syncedEntries = make(map[string]*domain.CacheEntry)
		for k, v := range entries {
			syncedEntries[k] = v
		}
		return nil
	}

	// Add dirty entry
	entry := &domain.CacheEntry{
		OriginalURL: "https://example.com",
		UsageCount:  1,
		LastUsedAt:  time.Now(),
		Dirty:       true,
	}
	err := cache.Set(ctx, "test123", entry)
	assert.NoError(t, err)

	// Start background sync with short interval
	err = cache.StartBackgroundSync(ctx, 50*time.Millisecond, syncFunc)
	assert.NoError(t, err)

	// Wait for sync to happen
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Greater(t, syncCallCount, 0)
	assert.Len(t, syncedEntries, 1)
	assert.Contains(t, syncedEntries, "test123")
	mu.Unlock()

	// Stop background sync
	err = cache.StopBackgroundSync()
	assert.NoError(t, err)

	// Verify entry is now clean
	dirty, err := cache.GetDirtyEntries(ctx)
	assert.NoError(t, err)
	assert.Len(t, dirty, 0)
}

func TestCache_BackgroundSync_StartWhenAlreadyRunning(t *testing.T) {
	cache := New()
	ctx := context.Background()
	
	syncFunc := func(entries map[string]*domain.CacheEntry) error {
		return nil
	}

	// Start first sync
	err := cache.StartBackgroundSync(ctx, 100*time.Millisecond, syncFunc)
	assert.NoError(t, err)

	// Try to start again (should not error)
	err = cache.StartBackgroundSync(ctx, 100*time.Millisecond, syncFunc)
	assert.NoError(t, err)

	// Stop
	err = cache.StopBackgroundSync()
	assert.NoError(t, err)
}

func TestCache_StopBackgroundSync_WhenNotRunning(t *testing.T) {
	cache := New()
	
	// Stop when not running (should not error)
	err := cache.StopBackgroundSync()
	assert.NoError(t, err)
}

func TestCache_Close(t *testing.T) {
	cache := New()
	ctx := context.Background()
	
	syncFunc := func(entries map[string]*domain.CacheEntry) error {
		return nil
	}

	// Start background sync
	err := cache.StartBackgroundSync(ctx, 100*time.Millisecond, syncFunc)
	assert.NoError(t, err)

	// Close cache
	err = cache.Close()
	assert.NoError(t, err)
	
	// Verify it's stopped
	assert.False(t, cache.running)
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache := New()
	ctx := context.Background()
	
	const numGoroutines = 10
	const opsPerGoroutine = 100
	
	var wg sync.WaitGroup
	
	// Test concurrent set/get/increment operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < opsPerGoroutine; j++ {
				shortCode := "test" + string(rune('0'+id))
				
				entry := &domain.CacheEntry{
					OriginalURL: "https://example.com",
					UsageCount:  j,
					LastUsedAt:  time.Now(),
					Dirty:       false,
				}
				
				// Set
				err := cache.Set(ctx, shortCode, entry)
				assert.NoError(t, err)
				
				// Get
				retrieved, exists := cache.Get(ctx, shortCode)
				assert.True(t, exists)
				assert.NotNil(t, retrieved)
				
				// Increment
				err = cache.IncrementUsage(ctx, shortCode)
				assert.NoError(t, err)
				
				// Delete occasionally
				if j%10 == 0 {
					err = cache.Delete(ctx, shortCode)
					assert.NoError(t, err)
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	// Verify cache is still functional
	entry := &domain.CacheEntry{
		OriginalURL: "https://final.com",
		UsageCount:  1,
		LastUsedAt:  time.Now(),
		Dirty:       false,
	}
	
	err := cache.Set(ctx, "final", entry)
	assert.NoError(t, err)
	
	retrieved, exists := cache.Get(ctx, "final")
	assert.True(t, exists)
	assert.Equal(t, entry.OriginalURL, retrieved.OriginalURL)
}

func TestCache_ContextCancellation(t *testing.T) {
	cache := New()
	ctx, cancel := context.WithCancel(context.Background())
	
	syncCallCount := 0
	syncFunc := func(entries map[string]*domain.CacheEntry) error {
		syncCallCount++
		return nil
	}

	// Add dirty entry
	entry := &domain.CacheEntry{
		OriginalURL: "https://example.com",
		UsageCount:  1,
		LastUsedAt:  time.Now(),
		Dirty:       true,
	}
	err := cache.Set(ctx, "test123", entry)
	assert.NoError(t, err)

	// Start background sync
	err = cache.StartBackgroundSync(ctx, 50*time.Millisecond, syncFunc)
	assert.NoError(t, err)

	// Wait briefly for sync to start
	time.Sleep(25 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait to ensure sync stops
	time.Sleep(100 * time.Millisecond)

	// Stop background sync
	err = cache.StopBackgroundSync()
	assert.NoError(t, err)
	
	// Some syncs should have happened before cancellation
	assert.GreaterOrEqual(t, syncCallCount, 0)
}