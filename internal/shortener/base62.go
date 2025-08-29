package shortener

import "strings"

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// EncodeBase62 encodes a number to base62 string
func EncodeBase62(num int64) string {
	if num == 0 {
		return "0"
	}
	
	var result strings.Builder
	base := int64(len(base62Chars))
	
	for num > 0 {
		remainder := num % base
		result.WriteByte(base62Chars[remainder])
		num = num / base
	}
	
	// Reverse the string
	encoded := result.String()
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	
	return string(runes)
}

// DecodeBase62 decodes a base62 string to number
func DecodeBase62(encoded string) int64 {
	var result int64
	base := int64(len(base62Chars))
	
	for _, char := range encoded {
		result = result*base + int64(strings.IndexRune(base62Chars, char))
	}
	
	return result
}

// PadBase62 pads a base62 string to a minimum length
func PadBase62(encoded string, minLength int) string {
	if len(encoded) >= minLength {
		return encoded
	}
	
	padding := minLength - len(encoded)
	return strings.Repeat("0", padding) + encoded
}