package shortener

import (
	"fmt"

	"github.com/joshdurbin/url-shortener/db/sqlc"
)

// NewGenerator creates a new generator based on configuration
func NewGenerator(config Config, db *sqlc.Queries) (Generator, error) {
	switch config.Type {
	case TypeMD5Hash:
		return NewMD5Generator(), nil
		
	case TypeBase62Counter:
		if db == nil {
			return nil, fmt.Errorf("database queries required for counter-based generator")
		}
		counterProvider := NewCounterCache(db, config.CounterStep)
		return NewBase62CounterGenerator(counterProvider, config.Length), nil
		
	case TypeBase62Random:
		if config.Length <= 0 {
			config.Length = 8 // Default length
		}
		return NewBase62RandomGenerator(config.Length), nil
		
	case TypeNanoID:
		if config.Length <= 0 {
			config.Length = 21 // Default NanoID length
		}
		return NewNanoIDGenerator(config.Length), nil
		
	default:
		return nil, fmt.Errorf("unknown generator type: %s", config.Type)
	}
}