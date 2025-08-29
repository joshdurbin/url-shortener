package shortener

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/joshdurbin/url-shortener/db/sqlc"
)

func setupTestDB(t *testing.T) *sqlc.Queries {
	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test.db")
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	
	// Create counters table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS counters (
			key TEXT PRIMARY KEY,
			value INTEGER NOT NULL DEFAULT 0,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create counters table: %v", err)
	}
	
	return sqlc.New(db)
}

func TestCounterCache(t *testing.T) {
	queries := setupTestDB(t)
	jumpAhead := int64(10)
	cache := NewCounterCache(queries, jumpAhead)
	defer cache.Close()

	ctx := context.Background()
	key := "test-counter"

	t.Run("first counter value is 1", func(t *testing.T) {
		value, err := cache.GetNextCounter(ctx, key)
		if err != nil {
			t.Fatalf("GetNextCounter failed: %v", err)
		}

		if value != 1 {
			t.Errorf("Expected first counter value to be 1, got %d", value)
		}
	})

	t.Run("consecutive calls increment counter", func(t *testing.T) {
		// Get several values
		for i := int64(2); i <= 5; i++ {
			value, err := cache.GetNextCounter(ctx, key)
			if err != nil {
				t.Fatalf("GetNextCounter failed: %v", err)
			}

			if value != i {
				t.Errorf("Expected counter value %d, got %d", i, value)
			}
		}
	})

	t.Run("handles different keys independently", func(t *testing.T) {
		key1 := "counter1"
		key2 := "counter2"

		value1a, err := cache.GetNextCounter(ctx, key1)
		if err != nil {
			t.Fatalf("GetNextCounter failed: %v", err)
		}

		value2a, err := cache.GetNextCounter(ctx, key2)
		if err != nil {
			t.Fatalf("GetNextCounter failed: %v", err)
		}

		value1b, err := cache.GetNextCounter(ctx, key1)
		if err != nil {
			t.Fatalf("GetNextCounter failed: %v", err)
		}

		if value1a != 1 || value2a != 1 || value1b != 2 {
			t.Errorf("Expected independent counters, got key1: %d, %d and key2: %d", value1a, value1b, value2a)
		}
	})

	t.Run("SetCounter works correctly", func(t *testing.T) {
		newKey := "set-counter-test"
		setValue := int64(100)

		err := cache.SetCounter(ctx, newKey, setValue)
		if err != nil {
			t.Fatalf("SetCounter failed: %v", err)
		}

		// Next value should be setValue + 1
		value, err := cache.GetNextCounter(ctx, newKey)
		if err != nil {
			t.Fatalf("GetNextCounter failed: %v", err)
		}

		expected := setValue + 1
		if value != expected {
			t.Errorf("Expected counter value %d after SetCounter, got %d", expected, value)
		}
	})
}

func TestCounterCacheConcurrency(t *testing.T) {
	queries := setupTestDB(t)
	jumpAhead := int64(5)
	cache := NewCounterCache(queries, jumpAhead)
	defer cache.Close()

	ctx := context.Background()
	key := "concurrent-test"
	numGoroutines := 10
	numIncrements := 10

	var wg sync.WaitGroup
	results := make([][]int64, numGoroutines)
	errors := make([]error, numGoroutines)

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			results[goroutineID] = make([]int64, numIncrements)
			for j := 0; j < numIncrements; j++ {
				value, err := cache.GetNextCounter(ctx, key)
				if err != nil {
					errors[goroutineID] = err
					return
				}
				results[goroutineID][j] = value
			}
		}(i)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Goroutine %d encountered error: %v", i, err)
		}
	}

	// Collect all values and check uniqueness
	allValues := make(map[int64]bool)
	for _, goroutineResults := range results {
		for _, value := range goroutineResults {
			if allValues[value] {
				t.Errorf("Duplicate value found: %d", value)
			}
			allValues[value] = true
		}
	}

	totalExpected := numGoroutines * numIncrements
	if len(allValues) != totalExpected {
		t.Errorf("Expected %d unique values, got %d", totalExpected, len(allValues))
	}
}

func TestCounterCacheSync(t *testing.T) {
	queries := setupTestDB(t)
	jumpAhead := int64(3)
	cache := NewCounterCache(queries, jumpAhead)
	
	ctx := context.Background()
	key := "sync-test"

	// Generate some values
	_, err := cache.GetNextCounter(ctx, key)
	if err != nil {
		t.Fatalf("GetNextCounter failed: %v", err)
	}

	// Sync cache to database
	err = cache.Sync(ctx)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Close the cache
	cache.Close()

	// Create new cache and verify it picks up from the right place
	cache2 := NewCounterCache(queries, jumpAhead)
	defer cache2.Close()

	// The new cache should allocate from the synced value
	value, err := cache2.GetNextCounter(ctx, key)
	if err != nil {
		t.Fatalf("GetNextCounter failed: %v", err)
	}

	// Should be greater than what we had before due to jump ahead
	if value <= 1 {
		t.Errorf("Expected counter to continue from synced value, got %d", value)
	}
}