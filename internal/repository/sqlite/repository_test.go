package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_New(t *testing.T) {
	dbPath := createTempDB(t)
	defer os.Remove(dbPath)

	repo, err := New(dbPath)
	require.NoError(t, err)
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.db)
	assert.NotNil(t, repo.queries)

	// Verify database connection is working
	err = repo.db.Ping()
	assert.NoError(t, err)

	// Close repository
	err = repo.Close()
	assert.NoError(t, err)
}

func TestRepository_New_InvalidPath(t *testing.T) {
	// Test with invalid database path
	repo, err := New("/invalid/path/to/database.db")
	assert.Error(t, err)
	assert.Nil(t, repo)
}

func TestRepository_CreateURL(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()
	shortCode := "test123"
	originalURL := "https://example.com"
	createdAt := time.Now().UTC()

	// Create URL
	entry, err := repo.CreateURL(ctx, shortCode, originalURL, createdAt)
	require.NoError(t, err)
	assert.NotNil(t, entry)
	assert.NotZero(t, entry.ID)
	assert.Equal(t, shortCode, entry.ShortCode)
	assert.Equal(t, originalURL, entry.OriginalURL)
	assert.WithinDuration(t, createdAt, entry.CreatedAt, time.Second)
	assert.Nil(t, entry.LastUsedAt)
	assert.Equal(t, 0, entry.UsageCount)
}

func TestRepository_CreateURL_Duplicate(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()
	shortCode := "test123"
	originalURL := "https://example.com"
	createdAt := time.Now().UTC()

	// Create first URL
	_, err := repo.CreateURL(ctx, shortCode, originalURL, createdAt)
	require.NoError(t, err)

	// Try to create duplicate
	_, err = repo.CreateURL(ctx, shortCode, "https://different.com", createdAt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create URL")
}

func TestRepository_GetURL(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()
	shortCode := "test123"
	originalURL := "https://example.com"
	createdAt := time.Now().UTC()

	// Create URL first
	created, err := repo.CreateURL(ctx, shortCode, originalURL, createdAt)
	require.NoError(t, err)

	// Get URL
	retrieved, err := repo.GetURL(ctx, shortCode)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.ShortCode, retrieved.ShortCode)
	assert.Equal(t, created.OriginalURL, retrieved.OriginalURL)
	assert.WithinDuration(t, created.CreatedAt, retrieved.CreatedAt, time.Second)
	assert.Equal(t, created.UsageCount, retrieved.UsageCount)
}

func TestRepository_GetURL_NotFound(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()

	// Try to get non-existent URL
	_, err := repo.GetURL(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "short code not found")
}

func TestRepository_GetAllURLs(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()

	// Initially should be empty
	urls, err := repo.GetAllURLs(ctx)
	require.NoError(t, err)
	assert.Len(t, urls, 0)

	// Create multiple URLs with different timestamps
	now := time.Now().UTC()
	urls1, err := repo.CreateURL(ctx, "test1", "https://example1.com", now.Add(-2*time.Hour))
	require.NoError(t, err)

	urls2, err := repo.CreateURL(ctx, "test2", "https://example2.com", now.Add(-1*time.Hour))
	require.NoError(t, err)

	urls3, err := repo.CreateURL(ctx, "test3", "https://example3.com", now)
	require.NoError(t, err)

	// Get all URLs
	allURLs, err := repo.GetAllURLs(ctx)
	require.NoError(t, err)
	assert.Len(t, allURLs, 3)

	// Should be ordered by creation date (desc), so newest first
	assert.Equal(t, urls3.ID, allURLs[0].ID)
	assert.Equal(t, urls2.ID, allURLs[1].ID)
	assert.Equal(t, urls1.ID, allURLs[2].ID)
}

func TestRepository_UpdateUsage(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()
	shortCode := "test123"
	originalURL := "https://example.com"
	createdAt := time.Now().UTC()

	// Create URL first
	_, err := repo.CreateURL(ctx, shortCode, originalURL, createdAt)
	require.NoError(t, err)

	// Update usage
	lastUsedAt := time.Now().UTC()
	err = repo.UpdateUsage(ctx, shortCode, 5, lastUsedAt)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetURL(ctx, shortCode)
	require.NoError(t, err)
	assert.Equal(t, 5, retrieved.UsageCount)
	assert.NotNil(t, retrieved.LastUsedAt)
	assert.WithinDuration(t, lastUsedAt, *retrieved.LastUsedAt, time.Second)
}

func TestRepository_UpdateUsage_NonExistent(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()

	// Try to update non-existent URL (should not error, just no rows affected)
	err := repo.UpdateUsage(ctx, "nonexistent", 5, time.Now())
	assert.NoError(t, err)
}

func TestRepository_DeleteURL(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()
	shortCode := "test123"
	originalURL := "https://example.com"
	createdAt := time.Now().UTC()

	// Create URL first
	_, err := repo.CreateURL(ctx, shortCode, originalURL, createdAt)
	require.NoError(t, err)

	// Verify it exists
	_, err = repo.GetURL(ctx, shortCode)
	require.NoError(t, err)

	// Delete URL
	err = repo.DeleteURL(ctx, shortCode)
	require.NoError(t, err)

	// Verify it's gone
	_, err = repo.GetURL(ctx, shortCode)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "short code not found")
}

func TestRepository_DeleteURL_NonExistent(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()

	// Try to delete non-existent URL (should not error)
	err := repo.DeleteURL(ctx, "nonexistent")
	assert.NoError(t, err)
}

func TestRepository_URLExists(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()
	shortCode := "test123"
	originalURL := "https://example.com"
	createdAt := time.Now().UTC()

	// Initially should not exist
	exists, err := repo.URLExists(ctx, shortCode)
	require.NoError(t, err)
	assert.False(t, exists)

	// Create URL
	_, err = repo.CreateURL(ctx, shortCode, originalURL, createdAt)
	require.NoError(t, err)

	// Now should exist
	exists, err = repo.URLExists(ctx, shortCode)
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete URL
	err = repo.DeleteURL(ctx, shortCode)
	require.NoError(t, err)

	// Should not exist again
	exists, err = repo.URLExists(ctx, shortCode)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRepository_LoadCacheData(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()

	// Initially should be empty
	data, err := repo.LoadCacheData(ctx)
	require.NoError(t, err)
	assert.Len(t, data, 0)

	// Create URLs with usage data
	now := time.Now().UTC()
	
	// URL with no usage
	_, err = repo.CreateURL(ctx, "test1", "https://example1.com", now)
	require.NoError(t, err)

	// URL with usage
	_, err = repo.CreateURL(ctx, "test2", "https://example2.com", now)
	require.NoError(t, err)
	err = repo.UpdateUsage(ctx, "test2", 5, now.Add(time.Hour))
	require.NoError(t, err)

	// Load cache data
	data, err = repo.LoadCacheData(ctx)
	require.NoError(t, err)
	assert.Len(t, data, 2)

	// Verify first URL
	entry1, exists := data["test1"]
	assert.True(t, exists)
	assert.Equal(t, "https://example1.com", entry1.OriginalURL)
	assert.Equal(t, 0, entry1.UsageCount)
	assert.False(t, entry1.Dirty)

	// Verify second URL
	entry2, exists := data["test2"]
	assert.True(t, exists)
	assert.Equal(t, "https://example2.com", entry2.OriginalURL)
	assert.Equal(t, 5, entry2.UsageCount)
	assert.WithinDuration(t, now.Add(time.Hour), entry2.LastUsedAt, time.Second)
	assert.False(t, entry2.Dirty)
}

func TestRepository_ConcurrentOperations(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()
	
	// Test concurrent creates
	numGoroutines := 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			shortCode := "test" + string(rune('0'+id))
			originalURL := "https://example" + string(rune('0'+id)) + ".com"
			createdAt := time.Now().UTC()
			
			_, err := repo.CreateURL(ctx, shortCode, originalURL, createdAt)
			done <- err
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		assert.NoError(t, err)
	}

	// Verify all URLs were created
	allURLs, err := repo.GetAllURLs(ctx)
	require.NoError(t, err)
	assert.Len(t, allURLs, numGoroutines)
}

func TestRepository_DatabaseConstraints(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	ctx := context.Background()

	t.Run("empty short code", func(t *testing.T) {
		// SQLite NOT NULL allows empty strings, only prevents NULL values
		_, err := repo.CreateURL(ctx, "", "https://example.com", time.Now())
		assert.NoError(t, err)
	})

	t.Run("empty original URL", func(t *testing.T) {
		// SQLite NOT NULL allows empty strings, only prevents NULL values
		_, err := repo.CreateURL(ctx, "test123", "", time.Now())
		assert.NoError(t, err)
	})
}

func TestRepository_Close(t *testing.T) {
	dbPath := createTempDB(t)
	defer os.Remove(dbPath)

	repo, err := New(dbPath)
	require.NoError(t, err)

	// Close repository
	err = repo.Close()
	assert.NoError(t, err)

	// Try to use after close (should fail)
	ctx := context.Background()
	_, err = repo.GetAllURLs(ctx)
	assert.Error(t, err)
}

func TestRepository_ContextCancellation(t *testing.T) {
	repo := setupTestRepo(t)
	defer teardownTestRepo(t, repo)

	// Create context that gets cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Operations should respect context cancellation
	_, err := repo.CreateURL(ctx, "test123", "https://example.com", time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// Helper functions

func createTempDB(t *testing.T) string {
	t.Helper()
	file, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	file.Close()
	return file.Name()
}

func setupTestRepo(t *testing.T) *Repository {
	t.Helper()
	dbPath := createTempDB(t)
	t.Cleanup(func() {
		os.Remove(dbPath)
	})

	repo, err := New(dbPath)
	require.NoError(t, err)
	
	return repo
}

func teardownTestRepo(t *testing.T, repo *Repository) {
	t.Helper()
	if repo != nil {
		repo.Close()
	}
}