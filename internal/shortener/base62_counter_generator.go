package shortener

import (
	"context"
	"time"
)

// Base62CounterGenerator generates short codes using base62 encoding with a counter
type Base62CounterGenerator struct {
	counterProvider CounterProvider
	length          int
	counterKey      string
}

// NewBase62CounterGenerator creates a new base62 counter-based generator
func NewBase62CounterGenerator(counterProvider CounterProvider, length int) *Base62CounterGenerator {
	return &Base62CounterGenerator{
		counterProvider: counterProvider,
		length:          length,
		counterKey:      "base62_counter",
	}
}

// GenerateShortCode generates a short code using base62 counter encoding
func (g *Base62CounterGenerator) GenerateShortCode(ctx context.Context, originalURL string, timestamp time.Time) (string, error) {
	counter, err := g.counterProvider.GetNextCounter(ctx, g.counterKey)
	if err != nil {
		return "", err
	}
	
	encoded := EncodeBase62(counter)
	
	// Pad to minimum length if specified
	if g.length > 0 {
		encoded = PadBase62(encoded, g.length)
	}
	
	return encoded, nil
}

// Type returns the generator type
func (g *Base62CounterGenerator) Type() string {
	return TypeBase62Counter
}

// Close performs cleanup
func (g *Base62CounterGenerator) Close() error {
	if g.counterProvider != nil {
		return g.counterProvider.Close()
	}
	return nil
}

// Ensure Base62CounterGenerator implements Generator interface
var _ Generator = (*Base62CounterGenerator)(nil)