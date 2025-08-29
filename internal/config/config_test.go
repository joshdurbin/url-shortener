package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshdurbin/url-shortener/internal/shortener"
)

func TestConfig_New_Valid(t *testing.T) {
	cfg, err := New(
		"8080",
		"http://localhost:8080",
		"/tmp/test.db",
		5*time.Second,
		true, shortener.DefaultConfig(),
	)

	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify server config
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "http://localhost:8080", cfg.Server.ServerURL)

	// Verify database config
	assert.Equal(t, "/tmp/test.db", cfg.Database.Path)

	// Verify cache config
	assert.Equal(t, 5*time.Second, cfg.Cache.SyncInterval)

	// Verify logging config
	assert.True(t, cfg.Logging.Verbose)
}

func TestConfig_Validate_EmptyServerPort(t *testing.T) {
	_, err := New(
		"",                      // empty port
		"http://localhost:8080",
		"/tmp/test.db",
		5*time.Second,
		true, shortener.DefaultConfig(),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server port cannot be empty")
}

func TestConfig_Validate_EmptyServerURL(t *testing.T) {
	_, err := New(
		"8080",
		"", // empty server URL
		"/tmp/test.db",
		5*time.Second,
		true, shortener.DefaultConfig(),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server URL cannot be empty")
}

func TestConfig_Validate_EmptyDatabasePath(t *testing.T) {
	_, err := New(
		"8080",
		"http://localhost:8080",
		"", // empty database path
		5*time.Second,
		true, shortener.DefaultConfig(),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database path cannot be empty")
}

func TestConfig_Validate_InvalidSyncInterval(t *testing.T) {
	testCases := []struct {
		name         string
		syncInterval time.Duration
	}{
		{"zero interval", 0},
		{"negative interval", -5 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(
				"8080",
				"http://localhost:8080",
				"/tmp/test.db",
				tc.syncInterval,
				true, shortener.DefaultConfig(),
			)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cache sync interval must be positive")
		})
	}
}

func TestConfig_Validate_DirectCall(t *testing.T) {
	// Test validate method directly
	cfg := &Config{
		Server: ServerConfig{
			Port:      "8080",
			ServerURL: "http://localhost:8080",
		},
		Database: DatabaseConfig{
			Path: "/tmp/test.db",
		},
		Cache: CacheConfig{
			SyncInterval: 5 * time.Second,
		},
		Logging: LoggingConfig{
			Verbose: false,
		},
	}

	err := cfg.validate()
	assert.NoError(t, err)
}

func TestConfig_EdgeCases(t *testing.T) {
	t.Run("minimal valid sync interval", func(t *testing.T) {
		cfg, err := New(
			"8080",
			"http://localhost:8080",
			"/tmp/test.db",
			1*time.Nanosecond, // minimal positive duration
			true, shortener.DefaultConfig(),
		)
		require.NoError(t, err)
		assert.Equal(t, 1*time.Nanosecond, cfg.Cache.SyncInterval)
	})

	t.Run("large sync interval", func(t *testing.T) {
		cfg, err := New(
			"8080",
			"http://localhost:8080",
			"/tmp/test.db",
			24*time.Hour, // large duration
			false,
			shortener.DefaultConfig(),
		)
		require.NoError(t, err)
		assert.Equal(t, 24*time.Hour, cfg.Cache.SyncInterval)
		assert.False(t, cfg.Logging.Verbose)
	})
}

func TestConfig_RealWorldScenarios(t *testing.T) {
	t.Run("development config", func(t *testing.T) {
		cfg, err := New(
			"8080",
			"http://localhost:8080",
			"./dev.db",
			1*time.Second,
			true, shortener.DefaultConfig(),
		)
		require.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("production config", func(t *testing.T) {
		cfg, err := New(
			"80",
			"https://myapp.com",
			"/var/lib/myapp/urls.db",
			30*time.Second,
			false, shortener.DefaultConfig(),
		)
		require.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("testing config", func(t *testing.T) {
		cfg, err := New(
			"0", // Let OS assign port
			"http://localhost",
			":memory:",
			100*time.Millisecond,
			false, // No verbose logging in tests
			shortener.DefaultConfig(),
		)
		require.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("docker config", func(t *testing.T) {
		cfg, err := New(
			"8080",
			"http://0.0.0.0:8080",
			"/data/urls.db",
			5*time.Second,
			true, shortener.DefaultConfig(),
		)
		require.NoError(t, err)
		assert.NotNil(t, cfg)
	})
}