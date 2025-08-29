package config

import (
	"fmt"
	"time"

	"github.com/joshdurbin/url-shortener/internal/shortener"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Cache     CacheConfig
	Metrics   MetricsConfig
	Shortener shortener.Config
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port      string
	ServerURL string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Path string
}

// CacheConfig holds cache-related configuration
type CacheConfig struct {
	SyncInterval time.Duration
}

// MetricsConfig holds metrics-related configuration
type MetricsConfig struct {
	Enabled  bool
	Port     string
	Endpoint string
}

// New creates a new config with the given parameters
func New(port, serverURL, dbPath string, syncInterval time.Duration, metricsEnabled bool, metricsPort, metricsEndpoint string, shortenerConfig shortener.Config) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:      port,
			ServerURL: serverURL,
		},
		Database: DatabaseConfig{
			Path: dbPath,
		},
		Cache: CacheConfig{
			SyncInterval: syncInterval,
		},
		Metrics: MetricsConfig{
			Enabled:  metricsEnabled,
			Port:     metricsPort,
			Endpoint: metricsEndpoint,
		},
		Shortener: shortenerConfig,
	}

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// validate validates the configuration values
func (c *Config) validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	if c.Server.ServerURL == "" {
		return fmt.Errorf("server URL cannot be empty")
	}

	if c.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	if c.Cache.SyncInterval <= 0 {
		return fmt.Errorf("cache sync interval must be positive, got: %v", c.Cache.SyncInterval)
	}

	if c.Metrics.Enabled {
		if c.Metrics.Port == "" {
			return fmt.Errorf("metrics port cannot be empty when metrics are enabled")
		}
		if c.Metrics.Endpoint == "" {
			return fmt.Errorf("metrics endpoint cannot be empty when metrics are enabled")
		}
		if c.Metrics.Port == c.Server.Port {
			return fmt.Errorf("metrics port cannot be the same as server port")
		}
	}

	return nil
}