package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshdurbin/url-shortener/internal/cache/memory"
	"github.com/joshdurbin/url-shortener/internal/repository/sqlite"
	"github.com/joshdurbin/url-shortener/internal/service"
	"github.com/joshdurbin/url-shortener/internal/shortener"
)

func TestIntegration_FullWorkflow(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test_urls_%d.db", time.Now().UnixNano())
	defer os.Remove(dbPath)

	// Set up components
	repo, err := sqlite.New(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	cache := memory.New()
	defer cache.Close()
	
	// Create counter generator
	config := shortener.DefaultConfig()
	generator, err := shortener.NewGenerator(config, repo.GetQueries())
	require.NoError(t, err)
	defer generator.Close()

	urlShortener := service.NewURLShortener(repo, cache, generator)
	defer urlShortener.Close()

	// Initialize cache
	ctx := context.Background()
	require.NoError(t, urlShortener.InitializeCache(ctx))

	// Start cache sync
	require.NoError(t, urlShortener.StartCacheSync(ctx, 100*time.Millisecond))
	defer urlShortener.StopCacheSync()

	// Test: Create a short URL directly via service
	originalURL := "https://example.com/very/long/path/to/resource"
	
	result, err := urlShortener.CreateShortURL(ctx, originalURL)
	require.NoError(t, err)
	assert.NotEmpty(t, result.ShortCode)
	assert.Equal(t, originalURL, result.OriginalURL)

	shortCode := result.ShortCode

	// Test: Get URL info
	urlInfo, err := urlShortener.GetURLInfo(ctx, shortCode)
	require.NoError(t, err)
	assert.Equal(t, shortCode, urlInfo.ShortCode)
	assert.Equal(t, originalURL, urlInfo.OriginalURL)
	assert.Equal(t, 0, urlInfo.UsageCount)

	// Test: Get original URL (simulates redirect)
	retrievedURL, err := urlShortener.GetOriginalURL(ctx, shortCode)
	require.NoError(t, err)
	assert.Equal(t, originalURL, retrievedURL)

	// Verify usage was incremented
	urlInfo, err = urlShortener.GetURLInfo(ctx, shortCode)
	require.NoError(t, err)
	assert.Equal(t, 1, urlInfo.UsageCount)
	assert.NotNil(t, urlInfo.LastUsedAt)

	// Test: List URLs
	urls, err := urlShortener.GetAllURLs(ctx)
	require.NoError(t, err)
	assert.Len(t, urls, 1)
	assert.Equal(t, shortCode, urls[0].ShortCode)
	assert.Equal(t, originalURL, urls[0].OriginalURL)

	// Test: Create another URL
	secondURL := "https://google.com"
	result2, err := urlShortener.CreateShortURL(ctx, secondURL)
	require.NoError(t, err)
	assert.NotEqual(t, shortCode, result2.ShortCode)

	// Verify we have 2 URLs now
	urls, err = urlShortener.GetAllURLs(ctx)
	require.NoError(t, err)
	assert.Len(t, urls, 2)

	// Test: Delete URL
	err = urlShortener.DeleteShortURL(ctx, shortCode)
	require.NoError(t, err)

	// Verify URL is deleted
	_, err = urlShortener.GetURLInfo(ctx, shortCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Verify only 1 URL remains
	urls, err = urlShortener.GetAllURLs(ctx)
	require.NoError(t, err)
	assert.Len(t, urls, 1)
	assert.Equal(t, result2.ShortCode, urls[0].ShortCode)

	// Test cache sync by waiting and checking database directly
	time.Sleep(200 * time.Millisecond) // Wait for sync

	// Get URL info for the remaining URL to increment usage
	_, err = urlShortener.GetOriginalURL(ctx, result2.ShortCode)
	require.NoError(t, err)

	// Wait for sync
	time.Sleep(200 * time.Millisecond)

	// Verify in database
	dbEntry, err := repo.GetURL(ctx, result2.ShortCode)
	require.NoError(t, err)
	assert.Equal(t, 1, dbEntry.UsageCount)
}

func TestIntegration_ErrorCases(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test_urls_error_%d.db", time.Now().UnixNano())
	defer os.Remove(dbPath)

	// Set up components
	repo, err := sqlite.New(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	cache := memory.New()
	defer cache.Close()
	
	// Create counter generator
	config := shortener.DefaultConfig()
	generator, err := shortener.NewGenerator(config, repo.GetQueries())
	require.NoError(t, err)
	defer generator.Close()

	urlShortener := service.NewURLShortener(repo, cache, generator)
	defer urlShortener.Close()

	ctx := context.Background()
	require.NoError(t, urlShortener.InitializeCache(ctx))

	// Test: Invalid URL
	_, err = urlShortener.CreateShortURL(ctx, "not-a-url")
	require.Error(t, err)

	// Test: Get non-existent URL
	_, err = urlShortener.GetOriginalURL(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test: Delete non-existent URL
	err = urlShortener.DeleteShortURL(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestIntegration_ConcurrentAccess(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test_urls_concurrent_%d.db", time.Now().UnixNano())
	defer os.Remove(dbPath)

	// Set up components
	repo, err := sqlite.New(dbPath)
	require.NoError(t, err)
	defer repo.Close()

	cache := memory.New()
	defer cache.Close()
	
	// Create counter generator
	config := shortener.DefaultConfig()
	generator, err := shortener.NewGenerator(config, repo.GetQueries())
	require.NoError(t, err)
	defer generator.Close()

	urlShortener := service.NewURLShortener(repo, cache, generator)
	defer urlShortener.Close()

	ctx := context.Background()
	require.NoError(t, urlShortener.InitializeCache(ctx))
	require.NoError(t, urlShortener.StartCacheSync(ctx, 50*time.Millisecond))
	defer urlShortener.StopCacheSync()

	// Create a URL to test concurrent access
	originalURL := "https://example.com/concurrent"
	entry, err := urlShortener.CreateShortURL(ctx, originalURL)
	require.NoError(t, err)

	shortCode := entry.ShortCode

	// Concurrently access the URL to increment usage
	concurrency := 10
	done := make(chan struct{}, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			
			// Each goroutine accesses the URL 5 times
			for j := 0; j < 5; j++ {
				url, err := urlShortener.GetOriginalURL(ctx, shortCode)
				assert.NoError(t, err)
				assert.Equal(t, originalURL, url)
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// Wait for final cache sync
	time.Sleep(100 * time.Millisecond)

	// Verify final usage count
	info, err := urlShortener.GetURLInfo(ctx, shortCode)
	require.NoError(t, err)
	assert.Equal(t, concurrency*5, info.UsageCount)
}

