package shortener

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"time"
)

// MD5Generator generates short codes using MD5 hash + timestamp (existing implementation)
type MD5Generator struct {
}

// NewMD5Generator creates a new MD5-based generator
func NewMD5Generator() *MD5Generator {
	return &MD5Generator{}
}

// GenerateShortCode generates a short code using MD5 hash + timestamp
func (g *MD5Generator) GenerateShortCode(ctx context.Context, originalURL string, timestamp time.Time) (string, error) {
	// Create a hash of URL + timestamp for uniqueness
	hasher := md5.New()
	hasher.Write([]byte(originalURL + timestamp.String()))
	hash := hex.EncodeToString(hasher.Sum(nil))

	// Use base36 encoding for the timestamp + first few chars of hash
	timestampStr := strconv.FormatInt(timestamp.Unix(), 36)
	hashPart := hash[:4]

	return timestampStr + hashPart, nil
}

// Type returns the generator type
func (g *MD5Generator) Type() string {
	return TypeMD5Hash
}

// Close performs cleanup (no-op for MD5Generator)
func (g *MD5Generator) Close() error {
	return nil
}

// Ensure MD5Generator implements Generator interface
var _ Generator = (*MD5Generator)(nil)