package service

import (
	"context"
	"fmt"
	"time"
)

// TestGenerator is a simple generator for testing purposes
type TestGenerator struct {
	counter int
}

// NewTestGenerator creates a new test generator
func NewTestGenerator() *TestGenerator {
	return &TestGenerator{counter: 0}
}

// GenerateShortCode generates a simple test short code
func (g *TestGenerator) GenerateShortCode(ctx context.Context, originalURL string, timestamp time.Time) (string, error) {
	g.counter++
	return fmt.Sprintf("test%04d", g.counter), nil
}

// Type returns the generator type
func (g *TestGenerator) Type() string {
	return "test"
}

// Close performs cleanup
func (g *TestGenerator) Close() error {
	return nil
}