package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/joshdurbin/url-shortener/db/sqlc"
	"github.com/joshdurbin/url-shortener/internal/domain"
)

// URLRepository is a mock implementation of repository.URLRepository
type URLRepository struct {
	mock.Mock
}

// CreateURL creates a new short URL entry
func (m *URLRepository) CreateURL(ctx context.Context, shortCode, originalURL string, createdAt time.Time) (*domain.URLEntry, error) {
	args := m.Called(ctx, shortCode, originalURL, createdAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.URLEntry), args.Error(1)
}

// GetURL retrieves a URL entry by its short code
func (m *URLRepository) GetURL(ctx context.Context, shortCode string) (*domain.URLEntry, error) {
	args := m.Called(ctx, shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.URLEntry), args.Error(1)
}

// GetAllURLs retrieves all URL entries ordered by creation date (desc)
func (m *URLRepository) GetAllURLs(ctx context.Context) ([]*domain.URLEntry, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.URLEntry), args.Error(1)
}

// UpdateUsage updates the usage count and last used timestamp for a URL
func (m *URLRepository) UpdateUsage(ctx context.Context, shortCode string, usageCount int, lastUsedAt time.Time) error {
	args := m.Called(ctx, shortCode, usageCount, lastUsedAt)
	return args.Error(0)
}

// DeleteURL removes a URL entry by its short code
func (m *URLRepository) DeleteURL(ctx context.Context, shortCode string) error {
	args := m.Called(ctx, shortCode)
	return args.Error(0)
}

// URLExists checks if a short code exists
func (m *URLRepository) URLExists(ctx context.Context, shortCode string) (bool, error) {
	args := m.Called(ctx, shortCode)
	return args.Bool(0), args.Error(1)
}

// LoadCacheData loads all URL data for cache initialization
func (m *URLRepository) LoadCacheData(ctx context.Context) (map[string]*domain.CacheEntry, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*domain.CacheEntry), args.Error(1)
}

// GetQueries returns the underlying sqlc queries for advanced operations
func (m *URLRepository) GetQueries() *sqlc.Queries {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*sqlc.Queries)
}

// Close closes the repository connection
func (m *URLRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}