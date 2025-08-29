package shortener

import (
	"fmt"

	"github.com/joshdurbin/url-shortener/db/sqlc"
)

// NewGenerator creates a new counter-based generator
func NewGenerator(config Config, db *sqlc.Queries) (Generator, error) {
	if db == nil {
		return nil, fmt.Errorf("database queries required for counter-based generator")
	}
	
	counterProvider := NewCounterCache(db, config.CounterStep)
	return NewCounterGenerator(counterProvider), nil
}