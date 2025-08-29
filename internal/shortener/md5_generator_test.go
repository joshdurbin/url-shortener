package shortener

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMD5Generator(t *testing.T) {
	generator := NewMD5Generator()
	defer generator.Close()

	ctx := context.Background()
	testURL := "https://example.com/test"
	timestamp := time.Now()

	t.Run("generates consistent code for same input", func(t *testing.T) {
		code1, err := generator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		code2, err := generator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		if code1 != code2 {
			t.Errorf("Expected consistent codes, got %s and %s", code1, code2)
		}
	})

	t.Run("generates different codes for different URLs", func(t *testing.T) {
		url1 := "https://example.com/test1"
		url2 := "https://example.com/test2"

		code1, err := generator.GenerateShortCode(ctx, url1, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		code2, err := generator.GenerateShortCode(ctx, url2, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		if code1 == code2 {
			t.Errorf("Expected different codes for different URLs, got %s for both", code1)
		}
	})

	t.Run("generates different codes for different timestamps", func(t *testing.T) {
		timestamp1 := time.Unix(1000000000, 0)
		timestamp2 := time.Unix(2000000000, 0)

		code1, err := generator.GenerateShortCode(ctx, testURL, timestamp1)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		code2, err := generator.GenerateShortCode(ctx, testURL, timestamp2)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		if code1 == code2 {
			t.Errorf("Expected different codes for different timestamps, got %s for both", code1)
		}
	})

	t.Run("returns correct type", func(t *testing.T) {
		if generator.Type() != TypeMD5Hash {
			t.Errorf("Expected type %s, got %s", TypeMD5Hash, generator.Type())
		}
	})

	t.Run("generates non-empty codes", func(t *testing.T) {
		code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		if code == "" {
			t.Error("Expected non-empty code")
		}

		if len(code) == 0 {
			t.Error("Expected code with positive length")
		}
	})

	t.Run("generates URL-safe codes", func(t *testing.T) {
		code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		// Check that code contains only alphanumeric characters
		for _, r := range code {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
				t.Errorf("Code contains non-alphanumeric character: %c", r)
			}
		}
	})

	t.Run("code format matches expected pattern", func(t *testing.T) {
		// Test with known timestamp to verify format
		knownTime := time.Unix(1609459200, 0) // 2021-01-01 00:00:00 UTC
		code, err := generator.GenerateShortCode(ctx, testURL, knownTime)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		// Calculate expected timestamp part
		expectedTimestampPart := strconv.FormatInt(knownTime.Unix(), 36)
		if !strings.HasPrefix(code, expectedTimestampPart) {
			t.Errorf("Expected code to start with %s, got %s", expectedTimestampPart, code)
		}

		// Total length should be timestamp part + 4 hash chars
		expectedLength := len(expectedTimestampPart) + 4
		if len(code) != expectedLength {
			t.Errorf("Expected code length %d, got %d", expectedLength, len(code))
		}
	})
}