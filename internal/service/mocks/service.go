package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/joshdurbin/url-shortener/internal/domain"
)

// URLShortener is a mock implementation of service.URLShortener
type URLShortener struct {
	mock.Mock
}

// CreateShortURL creates a new short URL
func (m *URLShortener) CreateShortURL(ctx context.Context, originalURL string) (*domain.URLEntry, error) {
	args := m.Called(ctx, originalURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.URLEntry), args.Error(1)
}

// GetOriginalURL retrieves the original URL for a short code and increments usage
func (m *URLShortener) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	args := m.Called(ctx, shortCode)
	return args.String(0), args.Error(1)
}

// GetURLInfo retrieves detailed information about a short URL
func (m *URLShortener) GetURLInfo(ctx context.Context, shortCode string) (*domain.URLEntry, error) {
	args := m.Called(ctx, shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.URLEntry), args.Error(1)
}

// DeleteShortURL removes a short URL
func (m *URLShortener) DeleteShortURL(ctx context.Context, shortCode string) error {
	args := m.Called(ctx, shortCode)
	return args.Error(0)
}

// GetAllURLs retrieves all short URLs with current cache data
func (m *URLShortener) GetAllURLs(ctx context.Context) ([]*domain.URLEntry, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.URLEntry), args.Error(1)
}

// InitializeCache loads data from repository into cache
func (m *URLShortener) InitializeCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// StartCacheSync starts background cache synchronization
func (m *URLShortener) StartCacheSync(ctx context.Context, interval time.Duration) error {
	args := m.Called(ctx, interval)
	return args.Error(0)
}

// StopCacheSync stops background cache synchronization
func (m *URLShortener) StopCacheSync() error {
	args := m.Called()
	return args.Error(0)
}

// Close closes the service and its dependencies
func (m *URLShortener) Close() error {
	args := m.Called()
	return args.Error(0)
}