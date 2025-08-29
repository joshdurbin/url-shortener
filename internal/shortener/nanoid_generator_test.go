package shortener

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNanoIDGenerator(t *testing.T) {
	length := 21
	generator := NewNanoIDGenerator(length)
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
		if generator.Type() != TypeNanoID {
			t.Errorf("Expected type %s, got %s", TypeNanoID, generator.Type())
		}
	})

	t.Run("generates codes using only NanoID characters", func(t *testing.T) {
		code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		for _, r := range code {
			if !containsChar(nanoIDChars, byte(r)) {
				t.Errorf("Code contains invalid character: %c", r)
			}
		}
	})

	t.Run("default length is 21", func(t *testing.T) {
		generator := NewNanoIDGenerator(0) // should default to 21
		defer generator.Close()

		code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		if len(code) != 21 {
			t.Errorf("Expected default code length 21, got %d", len(code))
		}
	})

	t.Run("handles different input parameters without affecting output format", func(t *testing.T) {
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
	})

	t.Run("URL-safe characters only", func(t *testing.T) {
		code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		// NanoID uses URL-safe characters: _-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
		for _, r := range code {
			char := byte(r)
			if !((char >= '0' && char <= '9') ||
				(char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				char == '_' || char == '-') {
				t.Errorf("Code contains non-URL-safe character: %c", r)
			}
		}
	})
}

func TestNanoIDGeneratorWithDifferentLengths(t *testing.T) {
	ctx := context.Background()
	testURL := "https://example.com/test"
	timestamp := time.Now()

	lengths := []int{1, 8, 16, 21, 32}

	for _, length := range lengths {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			generator := NewNanoIDGenerator(length)
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
func containsChar(s string, b byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return true
		}
	}
	return false
}