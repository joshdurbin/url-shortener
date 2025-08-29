package shortener

import (
	"testing"
)

func TestBase62Encoding(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero", 0, "0"},
		{"one", 1, "1"},
		{"single digit", 9, "9"},
		{"double digit", 10, "A"},
		{"larger number", 61, "z"},
		{"larger number 2", 62, "10"},
		{"large number", 1000, "G8"},
		{"very large", 123456789, "8M0kX"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeBase62(tt.input)
			if result != tt.expected {
				t.Errorf("EncodeBase62(%d) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBase62Decoding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"zero", "0", 0},
		{"one", "1", 1},
		{"single digit", "9", 9},
		{"double digit", "A", 10},
		{"larger number", "z", 61},
		{"larger number 2", "10", 62},
		{"large number", "G8", 1000},
		{"very large", "8M0kX", 123456789},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodeBase62(tt.input)
			if result != tt.expected {
				t.Errorf("DecodeBase62(%s) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBase62RoundTrip(t *testing.T) {
	tests := []int64{0, 1, 10, 62, 100, 1000, 12345, 987654321}

	for _, num := range tests {
		encoded := EncodeBase62(num)
		decoded := DecodeBase62(encoded)
		if decoded != num {
			t.Errorf("Round trip failed: %d -> %s -> %d", num, encoded, decoded)
		}
	}
}

func TestPadBase62(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		minLength int
		expected  string
	}{
		{"no padding needed", "ABC", 3, "ABC"},
		{"no padding needed longer", "ABC", 2, "ABC"},
		{"padding needed", "A", 3, "00A"},
		{"padding needed longer", "AB", 5, "000AB"},
		{"zero padding", "", 3, "000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadBase62(tt.input, tt.minLength)
			if result != tt.expected {
				t.Errorf("PadBase62(%s, %d) = %s, want %s", tt.input, tt.minLength, result, tt.expected)
			}
		})
	}
}