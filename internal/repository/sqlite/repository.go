package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/joshdurbin/url-shortener/db/sqlc"
	"github.com/joshdurbin/url-shortener/internal/domain"
	"github.com/joshdurbin/url-shortener/internal/repository"
)

// Repository implements repository.URLRepository using SQLite
type Repository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

// New creates a new SQLite repository
func New(databasePath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys and WAL mode for better performance
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	repo := &Repository{
		db:      db,
		queries: sqlc.New(db),
	}

	if err := repo.runMigrations(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return repo, nil
}


// CreateURL creates a new short URL entry
func (r *Repository) CreateURL(ctx context.Context, shortCode, originalURL string, createdAt time.Time) (*domain.URLEntry, error) {
	url, err := r.queries.CreateURL(ctx, sqlc.CreateURLParams{
		ShortCode:   shortCode,
		OriginalUrl: originalURL,
		CreatedAt:   createdAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	return r.sqlcURLToDomain(url), nil
}

// GetURL retrieves a URL entry by its short code
func (r *Repository) GetURL(ctx context.Context, shortCode string) (*domain.URLEntry, error) {
	url, err := r.queries.GetURL(ctx, shortCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("short code not found")
		}
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return r.sqlcURLToDomain(url), nil
}

// GetAllURLs retrieves all URL entries ordered by creation date (desc)
func (r *Repository) GetAllURLs(ctx context.Context) ([]*domain.URLEntry, error) {
	urls, err := r.queries.GetAllURLs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all URLs: %w", err)
	}

	entries := make([]*domain.URLEntry, len(urls))
	for i, url := range urls {
		entries[i] = r.sqlcURLToDomain(url)
	}

	return entries, nil
}

// UpdateUsage updates the usage count and last used timestamp for a URL
func (r *Repository) UpdateUsage(ctx context.Context, shortCode string, usageCount int, lastUsedAt time.Time) error {
	err := r.queries.UpdateUsage(ctx, sqlc.UpdateUsageParams{
		UsageCount: sql.NullInt64{Int64: int64(usageCount), Valid: true},
		LastUsedAt: sql.NullTime{Time: lastUsedAt, Valid: true},
		ShortCode:  shortCode,
	})
	if err != nil {
		return fmt.Errorf("failed to update usage: %w", err)
	}
	return nil
}

// DeleteURL removes a URL entry by its short code
func (r *Repository) DeleteURL(ctx context.Context, shortCode string) error {
	err := r.queries.DeleteURL(ctx, shortCode)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}
	return nil
}

// URLExists checks if a short code exists
func (r *Repository) URLExists(ctx context.Context, shortCode string) (bool, error) {
	count, err := r.queries.URLExists(ctx, shortCode)
	if err != nil {
		return false, fmt.Errorf("failed to check URL existence: %w", err)
	}
	return count > 0, nil
}

// LoadCacheData loads all URL data for cache initialization
func (r *Repository) LoadCacheData(ctx context.Context) (map[string]*domain.CacheEntry, error) {
	urls, err := r.queries.GetAllURLs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load cache data: %w", err)
	}

	cache := make(map[string]*domain.CacheEntry)
	for _, url := range urls {
		cacheEntry := &domain.CacheEntry{
			OriginalURL: url.OriginalUrl,
			UsageCount:  int(url.UsageCount.Int64),
			Dirty:       false,
		}
		if url.LastUsedAt.Valid {
			cacheEntry.LastUsedAt = url.LastUsedAt.Time
		}
		cache[url.ShortCode] = cacheEntry
	}

	return cache, nil
}

// Close closes the repository connection
func (r *Repository) Close() error {
	return r.db.Close()
}

// sqlcURLToDomain converts a sqlc.Url to domain.URLEntry
func (r *Repository) sqlcURLToDomain(url sqlc.Url) *domain.URLEntry {
	entry := &domain.URLEntry{
		ID:          int(url.ID),
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalUrl,
		CreatedAt:   url.CreatedAt,
		UsageCount:  int(url.UsageCount.Int64),
	}

	if url.LastUsedAt.Valid {
		entry.LastUsedAt = &url.LastUsedAt.Time
	}

	return entry
}

// GetQueries returns the underlying sqlc queries for advanced operations
func (r *Repository) GetQueries() *sqlc.Queries {
	return r.queries
}

// Ensure Repository implements the interface
var _ repository.URLRepository = (*Repository)(nil)