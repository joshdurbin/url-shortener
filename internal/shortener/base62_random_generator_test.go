package shortener

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestBase62RandomGenerator(t *testing.T) {
	length := 8
	generator := NewBase62RandomGenerator(length)
	defer generator.Close()

	ctx := context.Background()
	testURL := "https://example.com/test"
	timestamp := time.Now()

	t.Run("generates codes of correct length", func(t *testing.T) {
		code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		if len(code) != length {
			t.Errorf("Expected code length %d, got %d", length, len(code))
		}
	})

	t.Run("generates different codes on consecutive calls", func(t *testing.T) {
		codes := make(map[string]bool)
		const numCodes = 100

		for i := 0; i < numCodes; i++ {
			code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
			if err != nil {
				t.Fatalf("GenerateShortCode failed: %v", err)
			}

			if codes[code] {
				t.Errorf("Generated duplicate code: %s", code)
			}
			codes[code] = true
		}

		if len(codes) < numCodes {
			t.Errorf("Expected %d unique codes, got %d", numCodes, len(codes))
		}
	})

	t.Run("returns correct type", func(t *testing.T) {
		if generator.Type() != TypeBase62Random {
			t.Errorf("Expected type %s, got %s", TypeBase62Random, generator.Type())
		}
	})

	t.Run("generates codes using only base62 characters", func(t *testing.T) {
		code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		for _, r := range code {
			if !contains(base62Chars, byte(r)) {
				t.Errorf("Code contains invalid character: %c", r)
			}
		}
	})

	t.Run("handles different input parameters without affecting output", func(t *testing.T) {
		// Random generator should not be affected by URL or timestamp
		url1 := "https://example.com/test1"
		url2 := "https://different-example.com/test2"
		timestamp1 := time.Unix(1000000000, 0)
		timestamp2 := time.Unix(2000000000, 0)

		code1, err := generator.GenerateShortCode(ctx, url1, timestamp1)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		code2, err := generator.GenerateShortCode(ctx, url2, timestamp2)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		// Both should be valid length and character set
		if len(code1) != length || len(code2) != length {
			t.Errorf("Expected both codes to be length %d, got %d and %d", length, len(code1), len(code2))
		}

		// They should (very likely) be different due to randomness
		if code1 == code2 {
			t.Logf("Note: Generated identical codes %s (very unlikely but possible with randomness)", code1)
		}
	})
}

func TestBase62RandomGeneratorWithDifferentLengths(t *testing.T) {
	ctx := context.Background()
	testURL := "https://example.com/test"
	timestamp := time.Now()

	lengths := []int{1, 4, 8, 12, 16, 21}

	for _, length := range lengths {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			generator := NewBase62RandomGenerator(length)
			defer generator.Close()

			code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
			if err != nil {
				t.Fatalf("GenerateShortCode failed: %v", err)
			}

			if len(code) != length {
				t.Errorf("Expected code length %d, got %d", length, len(code))
			}
		})
	}
}

// Helper function to check if byte is in string
func contains(s string, b byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return true
		}
	}
	return false
}