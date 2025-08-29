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
	Logging   LoggingConfig
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


// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Verbose bool
}

// New creates a new config with the given parameters
func New(port, serverURL, dbPath string, syncInterval time.Duration, verbose bool, shortenerConfig shortener.Config) (*Config, error) {
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
		Logging: LoggingConfig{
			Verbose: verbose,
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

	return nil
}