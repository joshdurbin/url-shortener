package shortener

import (
	"context"
	"time"
)

// Generator defines the interface for generating short codes
type Generator interface {
	// GenerateShortCode generates a short code for the given URL
	GenerateShortCode(ctx context.Context, originalURL string, timestamp time.Time) (string, error)
	
	// Type returns the type identifier of the generator
	Type() string
	
	// Close performs cleanup when the generator is no longer needed
	Close() error
}

// CounterProvider defines the interface for managing counters used by generators
type CounterProvider interface {
	// GetNextCounter returns the next counter value for a given key
	GetNextCounter(ctx context.Context, key string) (int64, error)
	
	// SetCounter sets the counter value for a given key
	SetCounter(ctx context.Context, key string, value int64) error
	
	// Close performs cleanup when the provider is no longer needed
	Close() error
}

// Config holds configuration for shortener generators
type Config struct {
	CounterStep int64 `json:"counter_step"` // Step size for counter-based generators
}

// GeneratorType constants
const (
	TypeCounter = "counter"
)

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		CounterStep: 1,
	}
}