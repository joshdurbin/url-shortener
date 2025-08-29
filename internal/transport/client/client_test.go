package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshdurbin/url-shortener/internal/domain"
)

func TestNewClient(t *testing.T) {
	serverURL := "http://localhost:8080"
	client := NewClient(serverURL)

	assert.NotNil(t, client)
	assert.Equal(t, serverURL, client.serverURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestClient_CreateURL(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		expectedResponse := domain.CreateURLResponse{
			ShortCode:   "abc123",
			ShortURL:    "http://localhost:8080/abc123",
			OriginalURL: "https://example.com",
			CreatedAt:   time.Now(),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/urls", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Verify request body
			var req domain.CreateURLRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			assert.Equal(t, "https://example.com", req.URL)

			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		response, err := client.CreateURL(ctx, "https://example.com")
		require.NoError(t, err)
		assert.Equal(t, expectedResponse.ShortCode, response.ShortCode)
		assert.Equal(t, expectedResponse.ShortURL, response.ShortURL)
		assert.Equal(t, expectedResponse.OriginalURL, response.OriginalURL)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		_, err := client.CreateURL(ctx, "invalid-url")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server returned status 400")
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		_, err := client.CreateURL(ctx, "https://example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.CreateURL(ctx, "https://example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestClient_GetURL(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		now := time.Now()
		expectedEntry := domain.URLEntry{
			ID:          1,
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
			CreatedAt:   now,
			LastUsedAt:  &now,
			UsageCount:  5,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/urls/abc123", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedEntry)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		entry, err := client.GetURL(ctx, "abc123")
		require.NoError(t, err)
		assert.Equal(t, expectedEntry.ShortCode, entry.ShortCode)
		assert.Equal(t, expectedEntry.OriginalURL, entry.OriginalURL)
		assert.Equal(t, expectedEntry.UsageCount, entry.UsageCount)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		_, err := client.GetURL(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		_, err := client.GetURL(ctx, "abc123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server returned status 500")
	})
}

func TestClient_DeleteURL(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "/api/urls/abc123", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		err := client.DeleteURL(ctx, "abc123")
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		err := client.DeleteURL(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("unexpected status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK) // Expecting NoContent
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		err := client.DeleteURL(ctx, "abc123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server returned status 200")
	})
}

func TestClient_ListURLs(t *testing.T) {
	t.Run("successful listing", func(t *testing.T) {
		now := time.Now()
		expectedEntries := []*domain.URLEntry{
			{
				ID:          1,
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				CreatedAt:   now,
				UsageCount:  5,
			},
			{
				ID:          2,
				ShortCode:   "def456",
				OriginalURL: "https://google.com",
				CreatedAt:   now.Add(-1 * time.Hour),
				UsageCount:  0,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/urls", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedEntries)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		entries, err := client.ListURLs(ctx)
		require.NoError(t, err)
		assert.Len(t, entries, 2)
		assert.Equal(t, expectedEntries[0].ShortCode, entries[0].ShortCode)
		assert.Equal(t, expectedEntries[1].ShortCode, entries[1].ShortCode)
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]*domain.URLEntry{})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		entries, err := client.ListURLs(ctx)
		require.NoError(t, err)
		assert.Len(t, entries, 0)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		_, err := client.ListURLs(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server returned status 500")
	})
}

func TestClient_NetworkErrors(t *testing.T) {
	client := NewClient("http://nonexistent-server:9999")
	ctx := context.Background()

	t.Run("create URL network error", func(t *testing.T) {
		_, err := client.CreateURL(ctx, "https://example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to make request")
	})

	t.Run("get URL network error", func(t *testing.T) {
		_, err := client.GetURL(ctx, "abc123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to make request")
	})

	t.Run("delete URL network error", func(t *testing.T) {
		err := client.DeleteURL(ctx, "abc123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to make request")
	})

	t.Run("list URLs network error", func(t *testing.T) {
		_, err := client.ListURLs(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to make request")
	})
}

func TestClient_InvalidRequests(t *testing.T) {
	// Test invalid request creation (this would typically not happen in practice)
	client := NewClient("://invalid-url")
	ctx := context.Background()

	t.Run("invalid URL in CreateURL", func(t *testing.T) {
		_, err := client.CreateURL(ctx, "https://example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create request")
	})

	t.Run("invalid URL in GetURL", func(t *testing.T) {
		_, err := client.GetURL(ctx, "abc123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create request")
	})

	t.Run("invalid URL in DeleteURL", func(t *testing.T) {
		err := client.DeleteURL(ctx, "abc123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create request")
	})

	t.Run("invalid URL in ListURLs", func(t *testing.T) {
		_, err := client.ListURLs(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create request")
	})
}

func TestClient_Timeout(t *testing.T) {
	// Create a server that sleeps longer than client timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with very short timeout
	client := NewClient(server.URL)
	client.httpClient.Timeout = 10 * time.Millisecond

	ctx := context.Background()
	_, err := client.CreateURL(ctx, "https://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestClient_MalformedJSON(t *testing.T) {
	// Test handling of various malformed JSON responses
	testCases := []struct {
		name         string
		responseBody string
		statusCode   int
	}{
		{"empty response", "", http.StatusOK},
		{"incomplete JSON", `{"short_code": "abc123"`, http.StatusOK},
		{"wrong JSON structure", `"this is just a string"`, http.StatusOK},
		{"null response", "null", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.responseBody))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			ctx := context.Background()

			_, err := client.CreateURL(ctx, "https://example.com")
			if tc.name == "null response" {
				// null JSON is valid JSON and decodes to zero values, should not error
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to decode response")
			}
		})
	}
}

func TestClient_LargeResponse(t *testing.T) {
	// Test handling of large response
	largeEntries := make([]*domain.URLEntry, 1000)
	for i := 0; i < 1000; i++ {
		largeEntries[i] = &domain.URLEntry{
			ID:          i,
			ShortCode:   strings.Repeat("a", 100), // Long short code
			OriginalURL: strings.Repeat("https://example.com/", 50), // Long URL
			CreatedAt:   time.Now(),
			UsageCount:  i,
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(largeEntries)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	entries, err := client.ListURLs(ctx)
	require.NoError(t, err)
	assert.Len(t, entries, 1000)
}