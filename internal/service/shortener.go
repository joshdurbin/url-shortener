package service

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/joshdurbin/url-shortener/internal/cache"
	"github.com/joshdurbin/url-shortener/internal/domain"
	"github.com/joshdurbin/url-shortener/internal/repository"
	"github.com/joshdurbin/url-shortener/internal/shortener"
)

// urlShortener implements URLShortener interface
type urlShortener struct {
	repo      repository.URLRepository
	cache     cache.SyncableCache
	generator shortener.Generator
}

// NewURLShortener creates a new URL shortener service
func NewURLShortener(repo repository.URLRepository, cache cache.SyncableCache, generator shortener.Generator) URLShortener {
	return &urlShortener{
		repo:      repo,
		cache:     cache,
		generator: generator,
	}
}

// StartCacheSync starts the background cache synchronization
func (s *urlShortener) StartCacheSync(ctx context.Context, interval time.Duration) error {
	syncFunc := func(dirtyEntries map[string]*domain.CacheEntry) error {
		for shortCode, entry := range dirtyEntries {
			if err := s.repo.UpdateUsage(ctx, shortCode, entry.UsageCount, entry.LastUsedAt); err != nil {
				return fmt.Errorf("failed to sync entry %s: %w", shortCode, err)
			}
		}
		return nil
	}
	
	return s.cache.StartBackgroundSync(ctx, interval, syncFunc)
}

// StopCacheSync stops the background cache synchronization
func (s *urlShortener) StopCacheSync() error {
	return s.cache.StopBackgroundSync()
}

// InitializeCache loads data from the repository into the cache
func (s *urlShortener) InitializeCache(ctx context.Context) error {
	data, err := s.repo.LoadCacheData(ctx)
	if err != nil {
		return fmt.Errorf("failed to load cache data: %w", err)
	}
	
	return s.cache.LoadData(ctx, data)
}


// CreateShortURL creates a new short URL
func (s *urlShortener) CreateShortURL(ctx context.Context, originalURL string) (*domain.URLEntry, error) {
	// Validate URL
	parsedURL, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	
	// Only allow HTTP and HTTPS schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL: only HTTP and HTTPS are supported")
	}

	createdAt := time.Now()
	shortCode, err := s.generator.GenerateShortCode(ctx, originalURL, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate short code: %w", err)
	}

	// Insert into database
	entry, err := s.repo.CreateURL(ctx, shortCode, originalURL, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	// Add to cache
	cacheEntry := &domain.CacheEntry{
		OriginalURL: originalURL,
		UsageCount:  0,
		LastUsedAt:  createdAt,
		Dirty:       false,
	}
	if err := s.cache.Set(ctx, shortCode, cacheEntry); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to cache new entry %s: %v\n", shortCode, err)
	}

	return entry, nil
}

// GetOriginalURL retrieves the original URL for a short code and increments usage
func (s *urlShortener) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	// Try cache first
	if entry, exists := s.cache.Get(ctx, shortCode); exists {
		if err := s.cache.IncrementUsage(ctx, shortCode); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to increment usage in cache for %s: %v\n", shortCode, err)
		}
		
		return entry.OriginalURL, nil
	}

	// Fall back to database
	entry, err := s.repo.GetURL(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("short code not found")
	}

	// Add to cache and increment usage
	cacheEntry := &domain.CacheEntry{
		OriginalURL: entry.OriginalURL,
		UsageCount:  entry.UsageCount + 1,
		LastUsedAt:  time.Now(),
		Dirty:       true,
	}
	if err := s.cache.Set(ctx, shortCode, cacheEntry); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to cache entry %s: %v\n", shortCode, err)
	}

	return entry.OriginalURL, nil
}

// GetURLInfo retrieves detailed information about a short URL
func (s *urlShortener) GetURLInfo(ctx context.Context, shortCode string) (*domain.URLEntry, error) {
	entry, err := s.repo.GetURL(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("short code not found")
	}

	// Update with cache data if available
	if cacheEntry, exists := s.cache.Get(ctx, shortCode); exists {
		entry.UsageCount = cacheEntry.UsageCount
		entry.LastUsedAt = &cacheEntry.LastUsedAt
	}

	return entry, nil
}

// DeleteShortURL removes a short URL
func (s *urlShortener) DeleteShortURL(ctx context.Context, shortCode string) error {
	// Check if URL exists
	exists, err := s.repo.URLExists(ctx, shortCode)
	if err != nil {
		return fmt.Errorf("failed to check URL existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("short code not found")
	}

	// Delete from database
	if err := s.repo.DeleteURL(ctx, shortCode); err != nil {
		return fmt.Errorf("failed to delete URL from database: %w", err)
	}

	// Delete from cache
	if err := s.cache.Delete(ctx, shortCode); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to delete from cache %s: %v\n", shortCode, err)
	}

	return nil
}

// GetAllURLs retrieves all short URLs with current cache data
func (s *urlShortener) GetAllURLs(ctx context.Context) ([]*domain.URLEntry, error) {
	entries, err := s.repo.GetAllURLs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get URLs from database: %w", err)
	}

	// Update with cache data
	for _, entry := range entries {
		if cacheEntry, exists := s.cache.Get(ctx, entry.ShortCode); exists {
			entry.UsageCount = cacheEntry.UsageCount
			entry.LastUsedAt = &cacheEntry.LastUsedAt
		}
	}

	return entries, nil
}

// Close closes the service and its dependencies
func (s *urlShortener) Close() error {
	if err := s.generator.Close(); err != nil {
		return fmt.Errorf("failed to close generator: %w", err)
	}
	if err := s.cache.Close(); err != nil {
		return fmt.Errorf("failed to close cache: %w", err)
	}
	if err := s.repo.Close(); err != nil {
		return fmt.Errorf("failed to close repository: %w", err)
	}
	return nil
}

// Ensure urlShortener implements URLShortener interface
var _ URLShortener = (*urlShortener)(nil)