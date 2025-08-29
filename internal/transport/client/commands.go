package client

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Commands provides command-line operations for the client
type Commands struct {
	client *Client
}

// NewCommands creates a new Commands instance
func NewCommands(client *Client) *Commands {
	return &Commands{
		client: client,
	}
}

// Create creates a short URL and displays the result
func (c *Commands) Create(ctx context.Context, originalURL string) error {
	result, err := c.client.CreateURL(ctx, originalURL)
	if err != nil {
		return err
	}

	fmt.Printf("Short URL created:\n")
	fmt.Printf("Short Code: %s\n", result.ShortCode)
	fmt.Printf("Short URL: %s\n", result.ShortURL)
	fmt.Printf("Original URL: %s\n", result.OriginalURL)
	fmt.Printf("Created At: %s\n", result.CreatedAt.Format(time.RFC3339))

	return nil
}

// Get retrieves and displays information about a short URL
func (c *Commands) Get(ctx context.Context, shortCode string) error {
	entry, err := c.client.GetURL(ctx, shortCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fmt.Printf("Short code '%s' not found\n", shortCode)
			return nil
		}
		return err
	}

	fmt.Printf("URL Information:\n")
	fmt.Printf("Short Code: %s\n", entry.ShortCode)
	fmt.Printf("Original URL: %s\n", entry.OriginalURL)
	fmt.Printf("Created At: %s\n", entry.CreatedAt.Format(time.RFC3339))
	if entry.LastUsedAt != nil {
		fmt.Printf("Last Used At: %s\n", entry.LastUsedAt.Format(time.RFC3339))
	} else {
		fmt.Printf("Last Used At: Never\n")
	}
	fmt.Printf("Usage Count: %d\n", entry.UsageCount)

	return nil
}

// Delete removes a short URL
func (c *Commands) Delete(ctx context.Context, shortCode string) error {
	err := c.client.DeleteURL(ctx, shortCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fmt.Printf("Short code '%s' not found\n", shortCode)
			return nil
		}
		return err
	}

	fmt.Printf("Short URL '%s' deleted successfully\n", shortCode)
	return nil
}

// List displays all short URLs in a table format
func (c *Commands) List(ctx context.Context) error {
	entries, err := c.client.ListURLs(ctx)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No URLs found")
		return nil
	}

	fmt.Printf("%-15s %-50s %-20s %-20s %s\n", "Short Code", "Original URL", "Created At", "Last Used", "Usage Count")
	fmt.Println(strings.Repeat("-", 120))

	for _, entry := range entries {
		lastUsed := "Never"
		if entry.LastUsedAt != nil {
			lastUsed = entry.LastUsedAt.Format("2006-01-02 15:04:05")
		}

		originalURL := entry.OriginalURL
		if len(originalURL) > 50 {
			originalURL = originalURL[:47] + "..."
		}

		fmt.Printf("%-15s %-50s %-20s %-20s %d\n",
			entry.ShortCode,
			originalURL,
			entry.CreatedAt.Format("2006-01-02 15:04:05"),
			lastUsed,
			entry.UsageCount,
		)
	}

	return nil
}