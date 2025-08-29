package shortener

import (
	"context"
	"math/bits"
	"time"
)

const (
	// Base62 characters: 0-9, a-z, A-Z (case sensitive)
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	targetLength = 7 // Target length for short codes
)

// CounterGenerator generates obfuscated short codes using a monotonic counter with bit manipulation
type CounterGenerator struct {
	counterProvider CounterProvider
	counterKey      string
	multiplier      uint64 // Large prime multiplier for obfuscation
	salt           uint64 // Salt value to add entropy
}

// NewCounterGenerator creates a new counter-based generator with obfuscation
func NewCounterGenerator(counterProvider CounterProvider) *CounterGenerator {
	return &CounterGenerator{
		counterProvider: counterProvider,
		counterKey:      "url_counter",
		multiplier:      0x5DEECE66D,         // Large odd multiplier (used in LCGs)
		salt:           0x9E3779B97F4A7C15,   // Large prime-like constant
	}
}

// GenerateShortCode generates an obfuscated short code from a monotonic counter
func (g *CounterGenerator) GenerateShortCode(ctx context.Context, originalURL string, timestamp time.Time) (string, error) {
	counter, err := g.counterProvider.GetNextCounter(ctx, g.counterKey)
	if err != nil {
		return "", err
	}
	
	return g.encodeCounter(uint64(counter)), nil
}

// encodeCounter transforms the counter value and converts it to a short code
func (g *CounterGenerator) encodeCounter(counter uint64) string {
	// Apply multiple transformations to completely obscure the original counter
	transformed := g.obfuscateValue(counter)
	
	// Ensure the result falls within our desired length range
	// For 7 characters: 62^6 to 62^7-1 (to ensure 7 chars)
	minVal := uint64(56800235584)     // 62^6 (ensures at least 7 characters)
	maxVal := uint64(3521614606207)   // 62^7-1 (ensures exactly 7 characters)
	
	// Map the transformed value to our target range
	rangeSize := maxVal - minVal + 1
	finalValue := (transformed % rangeSize) + minVal
	
	return g.toBase62(finalValue)
}

// obfuscateValue applies multiple transformations to hide the original value
func (g *CounterGenerator) obfuscateValue(value uint64) uint64 {
	// Step 1: XOR with salt
	result := value ^ g.salt
	
	// Step 2: Multiply by large odd number (this scrambles bits significantly)
	result *= g.multiplier
	
	// Step 3: Bit rotation to further scramble
	result = bits.RotateLeft64(result, 21)
	
	// Step 4: XOR with rotated version of itself
	result ^= bits.RotateLeft64(result, 32)
	
	// Step 5: Apply bit reversal on lower 32 bits for extra scrambling
	lower := uint32(result & 0xFFFFFFFF)
	upper := uint32(result >> 32)
	result = (uint64(bits.Reverse32(lower)) << 32) | uint64(upper)
	
	return result
}

// toBase62 converts a number to base62 representation
func (g *CounterGenerator) toBase62(num uint64) string {
	if num == 0 {
		return "0"
	}
	
	result := ""
	for num > 0 {
		result = string(base62Chars[num%62]) + result
		num /= 62
	}
	
	return result
}

// fromBase62 converts a base62 string back to a number
func (g *CounterGenerator) fromBase62(str string) uint64 {
	result := uint64(0)
	for _, char := range str {
		var value uint64
		if char >= '0' && char <= '9' {
			value = uint64(char - '0')
		} else if char >= 'a' && char <= 'z' {
			value = uint64(char - 'a' + 10)
		} else if char >= 'A' && char <= 'Z' {
			value = uint64(char - 'A' + 36)
		}
		result = result*62 + value
	}
	return result
}

// Type returns the generator type
func (g *CounterGenerator) Type() string {
	return "counter"
}

// Close performs cleanup
func (g *CounterGenerator) Close() error {
	if g.counterProvider != nil {
		return g.counterProvider.Close()
	}
	return nil
}

// GenerateShortCodeForID generates a short code for a specific ID/counter value (for testing)
func (g *CounterGenerator) GenerateShortCodeForID(id uint64) string {
	return g.encodeCounter(id)
}

// Ensure CounterGenerator implements Generator interface
var _ Generator = (*CounterGenerator)(nil)
