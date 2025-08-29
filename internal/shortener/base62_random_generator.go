package shortener

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"time"
)

// Base62RandomGenerator generates short codes using random base62 characters
type Base62RandomGenerator struct {
	length int
}

// NewBase62RandomGenerator creates a new random base62 generator
func NewBase62RandomGenerator(length int) *Base62RandomGenerator {
	return &Base62RandomGenerator{
		length: length,
	}
}

// GenerateShortCode generates a random base62 short code
func (g *Base62RandomGenerator) GenerateShortCode(ctx context.Context, originalURL string, timestamp time.Time) (string, error) {
	var result strings.Builder
	result.Grow(g.length)
	
	maxIndex := big.NewInt(int64(len(base62Chars)))
	
	for i := 0; i < g.length; i++ {
		randomIndex, err := rand.Int(rand.Reader, maxIndex)
		if err != nil {
			return "", err
		}
		result.WriteByte(base62Chars[randomIndex.Int64()])
	}
	
	return result.String(), nil
}

// Type returns the generator type
func (g *Base62RandomGenerator) Type() string {
	return TypeBase62Random
}

// Close performs cleanup (no-op for random generator)
func (g *Base62RandomGenerator) Close() error {
	return nil
}

// Ensure Base62RandomGenerator implements Generator interface
var _ Generator = (*Base62RandomGenerator)(nil)