package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/joshdurbin/url-shortener/internal/cache/mocks"
	"github.com/joshdurbin/url-shortener/internal/domain"
	repoMocks "github.com/joshdurbin/url-shortener/internal/repository/mocks"
)

func TestURLShortener_CreateShortURL(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name        string
		originalURL string
		setupMocks  func(*repoMocks.URLRepository, *mocks.SyncableCache)
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful creation",
			originalURL: "https://example.com",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				repo.On("CreateURL", ctx, mock.AnythingOfType("string"), "https://example.com", mock.AnythingOfType("time.Time")).
					Return(&domain.URLEntry{
						ID:          1,
						ShortCode:   "abc123",
						OriginalURL: "https://example.com",
						CreatedAt:   time.Now(),
						UsageCount:  0,
					}, nil)
				
				cache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*domain.CacheEntry")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "invalid URL",
			originalURL: "not-a-url",
			setupMocks:  func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {},
			wantErr:     true,
			errContains: "invalid URL",
		},
		{
			name:        "repository error",
			originalURL: "https://example.com",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				repo.On("CreateURL", ctx, mock.AnythingOfType("string"), "https://example.com", mock.AnythingOfType("time.Time")).
					Return(nil, assert.AnError)
			},
			wantErr:     true,
			errContains: "failed to create URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repoMocks.URLRepository{}
			cache := &mocks.SyncableCache{}
			
			tt.setupMocks(repo, cache)
			
			shortener := NewURLShortener(repo, cache, NewTestGenerator())
			
			result, err := shortener.CreateShortURL(ctx, tt.originalURL)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.originalURL, result.OriginalURL)
				assert.NotEmpty(t, result.ShortCode)
			}
			
			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

func TestURLShortener_GetOriginalURL(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name       string
		shortCode  string
		setupMocks func(*repoMocks.URLRepository, *mocks.SyncableCache)
		wantURL    string
		wantErr    bool
	}{
		{
			name:      "found in cache",
			shortCode: "abc123",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				cache.On("Get", ctx, "abc123").
					Return(&domain.CacheEntry{
						OriginalURL: "https://example.com",
						UsageCount:  1,
						LastUsedAt:  time.Now(),
					}, true)
				
				cache.On("IncrementUsage", ctx, "abc123").
					Return(nil)
			},
			wantURL: "https://example.com",
			wantErr: false,
		},
		{
			name:      "not in cache, found in database",
			shortCode: "abc123",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				cache.On("Get", ctx, "abc123").
					Return(nil, false)
				
				repo.On("GetURL", ctx, "abc123").
					Return(&domain.URLEntry{
						ID:          1,
						ShortCode:   "abc123",
						OriginalURL: "https://example.com",
						CreatedAt:   time.Now(),
						UsageCount:  0,
					}, nil)
				
				cache.On("Set", ctx, "abc123", mock.AnythingOfType("*domain.CacheEntry")).
					Return(nil)
			},
			wantURL: "https://example.com",
			wantErr: false,
		},
		{
			name:      "not found anywhere",
			shortCode: "notfound",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				cache.On("Get", ctx, "notfound").
					Return(nil, false)
				
				repo.On("GetURL", ctx, "notfound").
					Return(nil, assert.AnError)
			},
			wantURL: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repoMocks.URLRepository{}
			cache := &mocks.SyncableCache{}
			
			tt.setupMocks(repo, cache)
			
			shortener := NewURLShortener(repo, cache, NewTestGenerator())
			
			result, err := shortener.GetOriginalURL(ctx, tt.shortCode)
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantURL, result)
			}
			
			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

func TestURLShortener_DeleteShortURL(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name       string
		shortCode  string
		setupMocks func(*repoMocks.URLRepository, *mocks.SyncableCache)
		wantErr    bool
		errContains string
	}{
		{
			name:      "successful deletion",
			shortCode: "abc123",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				repo.On("URLExists", ctx, "abc123").
					Return(true, nil)
				
				repo.On("DeleteURL", ctx, "abc123").
					Return(nil)
				
				cache.On("Delete", ctx, "abc123").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "short code not found",
			shortCode: "notfound",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				repo.On("URLExists", ctx, "notfound").
					Return(false, nil)
			},
			wantErr:     true,
			errContains: "short code not found",
		},
		{
			name:      "repository error on deletion",
			shortCode: "abc123",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				repo.On("URLExists", ctx, "abc123").
					Return(true, nil)
				
				repo.On("DeleteURL", ctx, "abc123").
					Return(assert.AnError)
			},
			wantErr:     true,
			errContains: "failed to delete URL from database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repoMocks.URLRepository{}
			cache := &mocks.SyncableCache{}
			
			tt.setupMocks(repo, cache)
			
			shortener := NewURLShortener(repo, cache, NewTestGenerator())
			
			err := shortener.DeleteShortURL(ctx, tt.shortCode)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
			
			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}


func TestURLShortener_GetURLInfo(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name       string
		shortCode  string
		setupMocks func(*repoMocks.URLRepository, *mocks.SyncableCache)
		wantErr    bool
	}{
		{
			name:      "successful retrieval",
			shortCode: "abc123",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				repo.On("GetURL", ctx, "abc123").
					Return(&domain.URLEntry{
						ID:          1,
						ShortCode:   "abc123",
						OriginalURL: "https://example.com",
						CreatedAt:   time.Now(),
						UsageCount:  5,
					}, nil)
				// Cache miss - no cache entry found
				cache.On("Get", ctx, "abc123").
					Return((*domain.CacheEntry)(nil), false)
			},
			wantErr: false,
		},
		{
			name:      "not found",
			shortCode: "notfound",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				repo.On("GetURL", ctx, "notfound").
					Return(nil, assert.AnError)
				// Cache is not called when repo returns error
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repoMocks.URLRepository{}
			cache := &mocks.SyncableCache{}
			
			tt.setupMocks(repo, cache)
			
			shortener := NewURLShortener(repo, cache, NewTestGenerator())
			
			result, err := shortener.GetURLInfo(ctx, tt.shortCode)
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.shortCode, result.ShortCode)
			}
			
			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

func TestURLShortener_GetAllURLs(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name       string
		setupMocks func(*repoMocks.URLRepository, *mocks.SyncableCache)
		wantCount  int
		wantErr    bool
	}{
		{
			name: "successful retrieval with URLs",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				repo.On("GetAllURLs", ctx).
					Return([]*domain.URLEntry{
						{ID: 1, ShortCode: "abc123", OriginalURL: "https://example.com"},
						{ID: 2, ShortCode: "def456", OriginalURL: "https://google.com"},
					}, nil)
				// Cache miss for both entries
				cache.On("Get", ctx, "abc123").
					Return((*domain.CacheEntry)(nil), false)
				cache.On("Get", ctx, "def456").
					Return((*domain.CacheEntry)(nil), false)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "empty list",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				repo.On("GetAllURLs", ctx).
					Return([]*domain.URLEntry{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "repository error",
			setupMocks: func(repo *repoMocks.URLRepository, cache *mocks.SyncableCache) {
				repo.On("GetAllURLs", ctx).
					Return(nil, assert.AnError)
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repoMocks.URLRepository{}
			cache := &mocks.SyncableCache{}
			
			tt.setupMocks(repo, cache)
			
			shortener := NewURLShortener(repo, cache, NewTestGenerator())
			
			result, err := shortener.GetAllURLs(ctx)
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.wantCount)
			}
			
			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

func TestURLShortener_CacheOperations(t *testing.T) {
	ctx := context.Background()
	
	t.Run("InitializeCache", func(t *testing.T) {
		repo := &repoMocks.URLRepository{}
		cache := &mocks.SyncableCache{}
			
		cacheData := map[string]*domain.CacheEntry{
			"abc123": {OriginalURL: "https://example.com", UsageCount: 1},
			"def456": {OriginalURL: "https://google.com", UsageCount: 2},
		}
		
		repo.On("LoadCacheData", ctx).Return(cacheData, nil)
		cache.On("LoadData", ctx, cacheData).Return(nil)
		
		shortener := NewURLShortener(repo, cache, NewTestGenerator())
		err := shortener.InitializeCache(ctx)
		
		require.NoError(t, err)
		repo.AssertExpectations(t)
		cache.AssertExpectations(t)
	})

	t.Run("InitializeCache error", func(t *testing.T) {
		repo := &repoMocks.URLRepository{}
		cache := &mocks.SyncableCache{}
			
		repo.On("LoadCacheData", ctx).Return(nil, assert.AnError)
		
		shortener := NewURLShortener(repo, cache, NewTestGenerator())
		err := shortener.InitializeCache(ctx)
		
		require.Error(t, err)
		repo.AssertExpectations(t)
		cache.AssertExpectations(t)
	})

	t.Run("StartCacheSync", func(t *testing.T) {
		repo := &repoMocks.URLRepository{}
		cache := &mocks.SyncableCache{}
			
		syncInterval := 100 * time.Millisecond
		cache.On("StartBackgroundSync", ctx, syncInterval, mock.AnythingOfType("func(map[string]*domain.CacheEntry) error")).Return(nil)
		
		shortener := NewURLShortener(repo, cache, NewTestGenerator())
		err := shortener.StartCacheSync(ctx, syncInterval)
		
		require.NoError(t, err)
		repo.AssertExpectations(t)
		cache.AssertExpectations(t)
	})

	t.Run("StopCacheSync", func(t *testing.T) {
		repo := &repoMocks.URLRepository{}
		cache := &mocks.SyncableCache{}
			
		cache.On("StopBackgroundSync").Return(nil)
		
		shortener := NewURLShortener(repo, cache, NewTestGenerator())
		err := shortener.StopCacheSync()
		
		require.NoError(t, err)
		repo.AssertExpectations(t)
		cache.AssertExpectations(t)
	})

	t.Run("Close", func(t *testing.T) {
		repo := &repoMocks.URLRepository{}
		cache := &mocks.SyncableCache{}
			
		cache.On("Close").Return(nil)
		repo.On("Close").Return(nil)
		
		shortener := NewURLShortener(repo, cache, NewTestGenerator())
		err := shortener.Close()
		
		require.NoError(t, err)
		repo.AssertExpectations(t)
		cache.AssertExpectations(t)
	})
}

func TestURLShortener_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	
	t.Run("cache error in CreateShortURL", func(t *testing.T) {
		repo := &repoMocks.URLRepository{}
		cache := &mocks.SyncableCache{}
			
		repo.On("CreateURL", ctx, mock.AnythingOfType("string"), "https://example.com", mock.AnythingOfType("time.Time")).
			Return(&domain.URLEntry{
				ID:          1,
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				CreatedAt:   time.Now(),
				UsageCount:  0,
			}, nil)
		
		cache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*domain.CacheEntry")).
			Return(assert.AnError)
		
		shortener := NewURLShortener(repo, cache, NewTestGenerator())
		
		// Should still succeed even if cache fails
		result, err := shortener.CreateShortURL(ctx, "https://example.com")
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		repo.AssertExpectations(t)
		cache.AssertExpectations(t)
	})

	t.Run("cache error in GetOriginalURL fallback", func(t *testing.T) {
		repo := &repoMocks.URLRepository{}
		cache := &mocks.SyncableCache{}
			
		cache.On("Get", ctx, "abc123").Return(nil, false)
		
		repo.On("GetURL", ctx, "abc123").
			Return(&domain.URLEntry{
				ID:          1,
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				CreatedAt:   time.Now(),
				UsageCount:  0,
			}, nil)
		
		cache.On("Set", ctx, "abc123", mock.AnythingOfType("*domain.CacheEntry")).
			Return(assert.AnError) // Cache set fails
		
		shortener := NewURLShortener(repo, cache, NewTestGenerator())
		
		// Should still work even if cache set fails
		result, err := shortener.GetOriginalURL(ctx, "abc123")
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", result)
		
		repo.AssertExpectations(t)
		cache.AssertExpectations(t)
	})
}

func TestURLShortener_URLValidation(t *testing.T) {
	repo := &repoMocks.URLRepository{}
	cache := &mocks.SyncableCache{}
	shortener := NewURLShortener(repo, cache, NewTestGenerator())
	ctx := context.Background()
	
	invalidURLs := []string{
		"",
		"not-a-url",
		"ftp://example.com", // Only http/https supported
		"javascript:alert(1)",
		"data:text/plain,hello",
		"file:///etc/passwd",
	}
	
	for _, url := range invalidURLs {
		t.Run("invalid_url_"+url, func(t *testing.T) {
			_, err := shortener.CreateShortURL(ctx, url)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid URL")
		})
	}
	
	validURLs := []string{
		"http://example.com",
		"https://example.com",
		"https://example.com:8080",
		"https://example.com/path?query=1",
		"https://subdomain.example.com",
	}
	
	for _, url := range validURLs {
		t.Run("valid_url_"+url, func(t *testing.T) {
			repo.On("CreateURL", ctx, mock.AnythingOfType("string"), url, mock.AnythingOfType("time.Time")).
				Return(&domain.URLEntry{
					ID:          1,
					ShortCode:   "abc123",
					OriginalURL: url,
					CreatedAt:   time.Now(),
					UsageCount:  0,
				}, nil)
			
			cache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*domain.CacheEntry")).
				Return(nil)
			
			result, err := shortener.CreateShortURL(ctx, url)
			assert.NoError(t, err)
			assert.Equal(t, url, result.OriginalURL)
		})
	}
	
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}