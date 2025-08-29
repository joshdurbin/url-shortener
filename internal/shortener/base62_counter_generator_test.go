package shortener

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/joshdurbin/url-shortener/db/sqlc"
)

func TestBase62CounterGenerator(t *testing.T) {
	queries := setupTestDBForGenerator(t)
	jumpAhead := int64(10)
	counterProvider := NewCounterCache(queries, jumpAhead)
	defer counterProvider.Close()

	length := 6
	generator := NewBase62CounterGenerator(counterProvider, length)
	defer generator.Close()

	ctx := context.Background()
	testURL := "https://example.com/test"
	timestamp := time.Now()

	t.Run("generates sequential codes", func(t *testing.T) {
		codes := make([]string, 5)
		
		for i := 0; i < 5; i++ {
			code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
			if err != nil {
				t.Fatalf("GenerateShortCode failed: %v", err)
			}
			codes[i] = code
		}

		// Decode and verify they are sequential
		for i := 1; i < len(codes); i++ {
			prev := DecodeBase62(codes[i-1])
			curr := DecodeBase62(codes[i])
			
			if curr != prev+1 {
				t.Errorf("Expected sequential codes, got %d then %d", prev, curr)
			}
		}
	})

	t.Run("returns correct type", func(t *testing.T) {
		if generator.Type() != TypeBase62Counter {
			t.Errorf("Expected type %s, got %s", TypeBase62Counter, generator.Type())
		}
	})

	t.Run("generates codes of correct minimum length", func(t *testing.T) {
		code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		if len(code) < length {
			t.Errorf("Expected code length >= %d, got %d", length, len(code))
		}
	})

	t.Run("pads short codes to minimum length", func(t *testing.T) {
		// Create generator with larger minimum length
		bigLength := 10
		bigLengthGenerator := NewBase62CounterGenerator(counterProvider, bigLength)
		defer bigLengthGenerator.Close()

		code, err := bigLengthGenerator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		if len(code) != bigLength {
			t.Errorf("Expected code length %d (with padding), got %d", bigLength, len(code))
		}

		// Should start with zeros for small numbers
		if code[0] != '0' {
			t.Logf("Code doesn't start with padding zero (counter may be large): %s", code)
		}
	})

	t.Run("generates codes using only base62 characters", func(t *testing.T) {
		code, err := generator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		for _, r := range code {
			if !containsChar(base62Chars, byte(r)) {
				t.Errorf("Code contains invalid character: %c", r)
			}
		}
	})

	t.Run("input parameters don't affect sequence", func(t *testing.T) {
		// Different URLs and timestamps should still produce sequential codes
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

		// Should be sequential regardless of input
		val1 := DecodeBase62(code1)
		val2 := DecodeBase62(code2)
		
		if val2 != val1+1 {
			t.Errorf("Expected sequential codes regardless of input, got %d then %d", val1, val2)
		}
	})

	t.Run("zero length allows natural code length", func(t *testing.T) {
		noLengthGenerator := NewBase62CounterGenerator(counterProvider, 0)
		defer noLengthGenerator.Close()

		code, err := noLengthGenerator.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}

		// Should have natural length (no padding)
		if len(code) == 0 {
			t.Error("Expected non-empty code")
		}

		// Code should not start with '0' (no padding)
		if code[0] == '0' && len(code) > 1 {
			t.Errorf("Unexpected padding in natural length code: %s", code)
		}
	})
}

func TestBase62CounterGeneratorPersistence(t *testing.T) {
	queries := setupTestDBForGenerator(t)
	jumpAhead := int64(5)
	
	// Create first generator and use it
	counterProvider1 := NewCounterCache(queries, jumpAhead)
	generator1 := NewBase62CounterGenerator(counterProvider1, 6)
	
	ctx := context.Background()
	testURL := "https://example.com/test"
	timestamp := time.Now()

	// Generate some codes
	codes1 := make([]string, 3)
	for i := 0; i < 3; i++ {
		code, err := generator1.GenerateShortCode(ctx, testURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}
		codes1[i] = code
	}

	// Close first generator
	generator1.Close()
	counterProvider1.Close()

	// Create second generator (should continue sequence)
	counterProvider2 := NewCounterCache(queries, jumpAhead)
	generator2 := NewBase62CounterGenerator(counterProvider2, 6)
	defer generator2.Close()
	defer counterProvider2.Close()

	// Generate more codes
	code4, err := generator2.GenerateShortCode(ctx, testURL, timestamp)
	if err != nil {
		t.Fatalf("GenerateShortCode failed: %v", err)
	}

	// Should continue from where the first generator left off
	lastVal := DecodeBase62(codes1[2])
	nextVal := DecodeBase62(code4)

	// Due to jump-ahead, the next value might be higher than lastVal + 1
	// but should be greater than lastVal
	if nextVal <= lastVal {
		t.Errorf("Expected continuation of sequence, got %d after %d", nextVal, lastVal)
	}
}

func setupTestDBForGenerator(t *testing.T) *sqlc.Queries {
	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "generator_test.db")
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	
	// Create counters table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS counters (
			key TEXT PRIMARY KEY,
			value INTEGER NOT NULL DEFAULT 0,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create counters table: %v", err)
	}
	
	return sqlc.New(db)
}