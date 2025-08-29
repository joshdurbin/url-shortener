package shortener

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"time"
)

// NanoID alphabet (URL-safe characters)
const nanoIDChars = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// NanoIDGenerator generates short codes using NanoID algorithm
type NanoIDGenerator struct {
	length int
}

// NewNanoIDGenerator creates a new NanoID generator
func NewNanoIDGenerator(length int) *NanoIDGenerator {
	if length <= 0 {
		length = 21 // Default NanoID length
	}
	return &NanoIDGenerator{
		length: length,
	}
}

// GenerateShortCode generates a NanoID-style short code
func (g *NanoIDGenerator) GenerateShortCode(ctx context.Context, originalURL string, timestamp time.Time) (string, error) {
	var result strings.Builder
	result.Grow(g.length)
	
	alphabetLength := int64(len(nanoIDChars))
	maxIndex := big.NewInt(alphabetLength)
	
	for i := 0; i < g.length; i++ {
		randomIndex, err := rand.Int(rand.Reader, maxIndex)
		if err != nil {
			return "", err
		}
		result.WriteByte(nanoIDChars[randomIndex.Int64()])
	}
	
	return result.String(), nil
}

// Type returns the generator type
func (g *NanoIDGenerator) Type() string {
	return TypeNanoID
}

// Close performs cleanup (no-op for NanoID generator)
func (g *NanoIDGenerator) Close() error {
	return nil
}

// Ensure NanoIDGenerator implements Generator interface
var _ Generator = (*NanoIDGenerator)(nil)