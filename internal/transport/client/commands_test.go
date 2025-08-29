package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshdurbin/url-shortener/internal/domain"
)

// captureOutput captures stdout for testing print statements
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	
	// Create a pipe to capture stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	
	// Save original stdout and restore it later
	origStdout := os.Stdout
	os.Stdout = w
	
	// Create a channel to read the output
	outputChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outputChan <- buf.String()
	}()
	
	// Execute the function
	fn()
	
	// Close writer and restore stdout
	w.Close()
	os.Stdout = origStdout
	
	// Read the captured output
	output := <-outputChan
	r.Close()
	
	return output
}

func TestNewCommands(t *testing.T) {
	client := NewClient("http://localhost:8080")
	commands := NewCommands(client)
	
	assert.NotNil(t, commands)
	assert.Equal(t, client, commands.client)
}

func TestCommands_Create(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		expectedResponse := domain.CreateURLResponse{
			ShortCode:   "abc123",
			ShortURL:    "http://localhost:8080/abc123",
			OriginalURL: "https://example.com",
			CreatedAt:   time.Now(),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		output := captureOutput(t, func() {
			err := commands.Create(ctx, "https://example.com")
			assert.NoError(t, err)
		})

		assert.Contains(t, output, "Short URL created:")
		assert.Contains(t, output, "abc123")
		assert.Contains(t, output, "http://localhost:8080/abc123")
		assert.Contains(t, output, "https://example.com")
		assert.Contains(t, output, "Created At:")
	})

	t.Run("creation error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		err := commands.Create(ctx, "invalid-url")
		assert.Error(t, err)
	})
}

func TestCommands_Get(t *testing.T) {
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedEntry)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		output := captureOutput(t, func() {
			err := commands.Get(ctx, "abc123")
			assert.NoError(t, err)
		})

		assert.Contains(t, output, "URL Information:")
		assert.Contains(t, output, "abc123")
		assert.Contains(t, output, "https://example.com")
		assert.Contains(t, output, "Usage Count: 5")
		assert.Contains(t, output, "Last Used At:")
		assert.NotContains(t, output, "Never")
	})

	t.Run("entry with no last used date", func(t *testing.T) {
		expectedEntry := domain.URLEntry{
			ID:          1,
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
			CreatedAt:   time.Now(),
			LastUsedAt:  nil, // Never used
			UsageCount:  0,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedEntry)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		output := captureOutput(t, func() {
			err := commands.Get(ctx, "abc123")
			assert.NoError(t, err)
		})

		assert.Contains(t, output, "Last Used At: Never")
		assert.Contains(t, output, "Usage Count: 0")
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		output := captureOutput(t, func() {
			err := commands.Get(ctx, "nonexistent")
			assert.NoError(t, err) // Should not return error, just print message
		})

		assert.Contains(t, output, "Short code 'nonexistent' not found")
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		err := commands.Get(ctx, "abc123")
		assert.Error(t, err)
	})
}

func TestCommands_Delete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		output := captureOutput(t, func() {
			err := commands.Delete(ctx, "abc123")
			assert.NoError(t, err)
		})

		assert.Contains(t, output, "Short URL 'abc123' deleted successfully")
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		output := captureOutput(t, func() {
			err := commands.Delete(ctx, "nonexistent")
			assert.NoError(t, err) // Should not return error, just print message
		})

		assert.Contains(t, output, "Short code 'nonexistent' not found")
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		err := commands.Delete(ctx, "abc123")
		assert.Error(t, err)
	})
}

func TestCommands_List(t *testing.T) {
	t.Run("successful listing with entries", func(t *testing.T) {
		now := time.Now()
		entries := []*domain.URLEntry{
			{
				ID:          1,
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				CreatedAt:   now,
				LastUsedAt:  &now,
				UsageCount:  5,
			},
			{
				ID:          2,
				ShortCode:   "def456",
				OriginalURL: "https://google.com",
				CreatedAt:   now.Add(-1 * time.Hour),
				LastUsedAt:  nil, // Never used
				UsageCount:  0,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(entries)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		output := captureOutput(t, func() {
			err := commands.List(ctx)
			assert.NoError(t, err)
		})

		// Check table headers
		assert.Contains(t, output, "Short Code")
		assert.Contains(t, output, "Original URL")
		assert.Contains(t, output, "Created At")
		assert.Contains(t, output, "Last Used")
		assert.Contains(t, output, "Usage Count")

		// Check table separator
		assert.Contains(t, output, strings.Repeat("-", 120))

		// Check entry data
		assert.Contains(t, output, "abc123")
		assert.Contains(t, output, "https://example.com")
		assert.Contains(t, output, "def456")
		assert.Contains(t, output, "https://google.com")
		assert.Contains(t, output, "Never") // For entry with no LastUsedAt
	})

	t.Run("long URL truncation", func(t *testing.T) {
		longURL := "https://example.com/" + strings.Repeat("very-long-path/", 10)
		entries := []*domain.URLEntry{
			{
				ID:          1,
				ShortCode:   "abc123",
				OriginalURL: longURL,
				CreatedAt:   time.Now(),
				UsageCount:  0,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(entries)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		output := captureOutput(t, func() {
			err := commands.List(ctx)
			assert.NoError(t, err)
		})

		// Check that URL is truncated with "..."
		assert.Contains(t, output, "...")
		// Original long URL should not appear in full
		assert.NotContains(t, output, longURL)
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]*domain.URLEntry{})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		output := captureOutput(t, func() {
			err := commands.List(ctx)
			assert.NoError(t, err)
		})

		assert.Contains(t, output, "No URLs found")
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		err := commands.List(ctx)
		assert.Error(t, err)
	})
}

func TestCommands_OutputFormatting(t *testing.T) {
	t.Run("date formatting in list", func(t *testing.T) {
		// Test specific date formatting
		specificTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
		entries := []*domain.URLEntry{
			{
				ID:          1,
				ShortCode:   "test123",
				OriginalURL: "https://example.com",
				CreatedAt:   specificTime,
				LastUsedAt:  &specificTime,
				UsageCount:  1,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(entries)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		output := captureOutput(t, func() {
			err := commands.List(ctx)
			assert.NoError(t, err)
		})

		// Check for expected date format (2006-01-02 15:04:05)
		assert.Contains(t, output, "2023-12-25 15:30:45")
	})

	t.Run("RFC3339 formatting in get and create", func(t *testing.T) {
		specificTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
		
		// Test Create command output
		createResponse := domain.CreateURLResponse{
			ShortCode:   "test123",
			ShortURL:    "http://localhost:8080/test123",
			OriginalURL: "https://example.com",
			CreatedAt:   specificTime,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				json.NewEncoder(w).Encode(createResponse)
			} else {
				entry := domain.URLEntry{
					ID:          1,
					ShortCode:   "test123",
					OriginalURL: "https://example.com",
					CreatedAt:   specificTime,
					LastUsedAt:  &specificTime,
					UsageCount:  1,
				}
				json.NewEncoder(w).Encode(entry)
			}
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx := context.Background()

		// Test Create output
		createOutput := captureOutput(t, func() {
			err := commands.Create(ctx, "https://example.com")
			assert.NoError(t, err)
		})
		assert.Contains(t, createOutput, "2023-12-25T15:30:45Z")

		// Test Get output
		getOutput := captureOutput(t, func() {
			err := commands.Get(ctx, "test123")
			assert.NoError(t, err)
		})
		assert.Contains(t, getOutput, "2023-12-25T15:30:45Z")
	})
}

func TestCommands_ErrorHandling(t *testing.T) {
	t.Run("network timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond) // Longer than client timeout
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		client.httpClient.Timeout = 10 * time.Millisecond // Very short timeout
		commands := NewCommands(client)
		ctx := context.Background()

		err := commands.Create(ctx, "https://example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(50 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		commands := NewCommands(client)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := commands.Create(ctx, "https://example.com")
		assert.Error(t, err)
	})
}