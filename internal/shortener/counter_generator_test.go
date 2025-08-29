package shortener

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/joshdurbin/url-shortener/db/sqlc"
)

func setupCounterTestDB(t *testing.T) *sqlc.Queries {
	dbPath := filepath.Join(t.TempDir(), "counter_test.db")
	
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

func TestCounterGenerator_GenerateShortCode(t *testing.T) {
	queries := setupCounterTestDB(t)
	counterProvider := NewCounterCache(queries, 1)
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	ctx := context.Background()
	originalURL := "https://example.com"
	timestamp := time.Now()
	
	// Generate multiple short codes and verify they are unique and have correct format
	codes := make(map[string]bool)
	for i := 0; i < 15; i++ {
		code, err := generator.GenerateShortCode(ctx, originalURL, timestamp)
		if err != nil {
			t.Fatalf("GenerateShortCode failed: %v", err)
		}
		
		// Verify code length is exactly 7 characters (our target length)
		if len(code) != targetLength {
			t.Errorf("Expected code length %d, got %d for code %s", targetLength, len(code), code)
		}
		
		// Verify code contains only valid base62 characters
		for _, char := range code {
			if !strings.ContainsRune(base62Chars, char) {
				t.Errorf("Code %s contains invalid character %c", code, char)
			}
		}
		
		// Verify uniqueness
		if codes[code] {
			t.Errorf("Duplicate code generated: %s", code)
		}
		codes[code] = true
		
		t.Logf("Generated code %d: %s", i+1, code)
	}
}

func TestCounterGenerator_Type(t *testing.T) {
	queries := setupCounterTestDB(t)
	counterProvider := NewCounterCache(queries, 1)
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	expectedType := "counter"
	if generator.Type() != expectedType {
		t.Errorf("Expected type %s, got %s", expectedType, generator.Type())
	}
}

func TestCounterGenerator_Close(t *testing.T) {
	queries := setupCounterTestDB(t)
	counterProvider := NewCounterCache(queries, 1)
	
	generator := NewCounterGenerator(counterProvider)
	
	err := generator.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestCounterGenerator_ObfuscatedEncoding(t *testing.T) {
	// Test that the same counter value always produces the same obfuscated code
	queries := setupCounterTestDB(t)
	counterProvider := NewCounterCache(queries, 1)
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	// Test consistency for same ID
	testCases := []uint64{1, 2, 3, 4, 5, 100, 1000, 10000}
	
	for _, id := range testCases {
		code1 := generator.GenerateShortCodeForID(id)
		code2 := generator.GenerateShortCodeForID(id)
		
		if code1 != code2 {
			t.Errorf("Same ID %d produced different codes: %s vs %s", id, code1, code2)
		}
		
		// Verify code length
		if len(code1) != targetLength {
			t.Errorf("Expected code length %d, got %d for code %s", targetLength, len(code1), code1)
		}
		
		t.Logf("ID %d -> %s (length: %d)", id, code1, len(code1))
	}
}

func TestCounterGenerator_NonSequentialPattern(t *testing.T) {
	// Test that sequential counters don't produce sequential-looking codes
	queries := setupCounterTestDB(t)
	counterProvider := NewCounterCache(queries, 1)
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	codes := make([]string, 15)
	for i := uint64(1); i <= 15; i++ {
		codes[i-1] = generator.GenerateShortCodeForID(i)
	}
	
	// Verify that the codes don't follow an obvious pattern
	// (This is a heuristic test - we check that adjacent codes are different)
	for i := 0; i < len(codes)-1; i++ {
		if codes[i] == codes[i+1] {
			t.Errorf("Adjacent codes are identical: %s at positions %d and %d", codes[i], i, i+1)
		}
		
		// Check that codes don't differ by just incrementing last character
		if len(codes[i]) == len(codes[i+1]) && codes[i][:len(codes[i])-1] == codes[i+1][:len(codes[i+1])-1] {
			lastChar1 := codes[i][len(codes[i])-1]
			lastChar2 := codes[i+1][len(codes[i+1])-1]
			pos1 := strings.IndexByte(base62Chars, lastChar1)
			pos2 := strings.IndexByte(base62Chars, lastChar2)
			if pos1 >= 0 && pos2 >= 0 && pos2 == pos1+1 {
				t.Logf("Warning: Adjacent codes %s and %s look sequential", codes[i], codes[i+1])
			}
		}
		
		t.Logf("ID %2d -> %s", i+1, codes[i])
	}
}

func TestCounterGenerator_LengthDistribution(t *testing.T) {
	// Test length distribution - all codes should be exactly 7 characters
	queries := setupCounterTestDB(t)
	counterProvider := NewCounterCache(queries, 1)
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	lengthCount := make(map[int]int)
	const testCount = 100
	
	for i := uint64(1); i <= testCount; i++ {
		code := generator.GenerateShortCodeForID(i)
		lengthCount[len(code)]++
	}
	
	// All codes should be exactly targetLength characters
	if len(lengthCount) != 1 {
		t.Errorf("Expected all codes to have length %d, but got lengths: %v", targetLength, lengthCount)
	}
	
	if lengthCount[targetLength] != testCount {
		t.Errorf("Expected %d codes of length %d, got %d", testCount, targetLength, lengthCount[targetLength])
	}
	
	t.Logf("Length distribution: %v", lengthCount)
}

func TestCounterGenerator_Base62Conversion(t *testing.T) {
	queries := setupCounterTestDB(t)
	counterProvider := NewCounterCache(queries, 1)
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	testCases := []struct {
		name string
		num  uint64
	}{
		{"Zero", 0},
		{"One", 1},
		{"Base boundary", 61},
		{"Base boundary + 1", 62},
		{"Large number", 123456789},
		{"Max range", 3521614606207}, // 62^7-1
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test toBase62
			encoded := generator.toBase62(tc.num)
			
			// Test fromBase62
			decoded := generator.fromBase62(encoded)
			
			if decoded != tc.num {
				t.Errorf("Encoding/decoding failed: %d -> %s -> %d", tc.num, encoded, decoded)
			}
			
			// Verify encoded string contains only valid base62 characters
			for _, char := range encoded {
				if !strings.ContainsRune(base62Chars, char) {
					t.Errorf("Encoded string %s contains invalid character %c", encoded, char)
				}
			}
		})
	}
}

func TestCounterGenerator_ObfuscationFunctions(t *testing.T) {
	queries := setupCounterTestDB(t)
	counterProvider := NewCounterCache(queries, 1)
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	// Test that obfuscation produces different results for sequential inputs
	values := []uint64{1, 2, 3, 4, 5}
	obfuscated := make([]uint64, len(values))
	
	for i, val := range values {
		obfuscated[i] = generator.obfuscateValue(val)
		t.Logf("obfuscateValue(%d) = %d", val, obfuscated[i])
	}
	
	// Verify all obfuscated values are different
	for i := 0; i < len(obfuscated); i++ {
		for j := i + 1; j < len(obfuscated); j++ {
			if obfuscated[i] == obfuscated[j] {
				t.Errorf("Obfuscation collision: obfuscateValue(%d) == obfuscateValue(%d) == %d", 
					values[i], values[j], obfuscated[i])
			}
		}
	}
	
	// Test that the same input always produces the same output
	for _, val := range values {
		result1 := generator.obfuscateValue(val)
		result2 := generator.obfuscateValue(val)
		if result1 != result2 {
			t.Errorf("obfuscateValue(%d) is not deterministic: %d vs %d", val, result1, result2)
		}
	}
}

func TestCounterGenerator_ConcurrentGeneration(t *testing.T) {
	queries := setupCounterTestDB(t)
	counterProvider := NewCounterCache(queries, 10) // Larger step for concurrency
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	ctx := context.Background()
	originalURL := "https://example.com"
	timestamp := time.Now()
	
	// Generate codes concurrently
	const numGoroutines = 5
	const codesPerGoroutine = 10
	
	codeChan := make(chan string, numGoroutines*codesPerGoroutine)
	errChan := make(chan error, numGoroutines*codesPerGoroutine)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < codesPerGoroutine; j++ {
				code, err := generator.GenerateShortCode(ctx, originalURL, timestamp)
				if err != nil {
					errChan <- err
					return
				}
				codeChan <- code
			}
		}()
	}
	
	// Collect all codes
	codes := make(map[string]bool)
	for i := 0; i < numGoroutines*codesPerGoroutine; i++ {
		select {
		case code := <-codeChan:
			if codes[code] {
				t.Errorf("Duplicate code generated concurrently: %s", code)
			}
			codes[code] = true
		case err := <-errChan:
			t.Errorf("Concurrent generation failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent code generation")
		}
	}
	
	// Verify we got the expected number of unique codes
	if len(codes) != numGoroutines*codesPerGoroutine {
		t.Errorf("Expected %d unique codes, got %d", numGoroutines*codesPerGoroutine, len(codes))
	}
}

func BenchmarkCounterGenerator_GenerateShortCode(b *testing.B) {
	queries := setupCounterTestDB(&testing.T{}) // This is a hack for benchmarks
	counterProvider := NewCounterCache(queries, 100)
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	ctx := context.Background()
	originalURL := "https://example.com"
	timestamp := time.Now()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := generator.GenerateShortCode(ctx, originalURL, timestamp)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCounterGenerator_ObfuscateValue(b *testing.B) {
	queries := setupCounterTestDB(&testing.T{})
	counterProvider := NewCounterCache(queries, 1)
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.obfuscateValue(uint64(i))
	}
}

func BenchmarkCounterGenerator_ToBase62(b *testing.B) {
	queries := setupCounterTestDB(&testing.T{})
	counterProvider := NewCounterCache(queries, 1)
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	numbers := []uint64{0, 1, 61, 62, 3844, 123456789, 3521614606207}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		num := numbers[i%len(numbers)]
		_ = generator.toBase62(num)
	}
}

func BenchmarkCounterGenerator_FromBase62(b *testing.B) {
	queries := setupCounterTestDB(&testing.T{})
	counterProvider := NewCounterCache(queries, 1)
	defer counterProvider.Close()
	
	generator := NewCounterGenerator(counterProvider)
	defer generator.Close()
	
	codes := []string{
		"0",
		"1",
		"z",
		"10", 
		"100",
		"8M0kX",
		"zzzzzz",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code := codes[i%len(codes)]
		_ = generator.fromBase62(code)
	}
}