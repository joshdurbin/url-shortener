package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/joshdurbin/url-shortener/internal/cache/memory"
	"github.com/joshdurbin/url-shortener/internal/config"
	"github.com/joshdurbin/url-shortener/internal/repository/sqlite"
	"github.com/joshdurbin/url-shortener/internal/service"
	"github.com/joshdurbin/url-shortener/internal/shortener"
	"github.com/joshdurbin/url-shortener/internal/transport/client"
	httpTransport "github.com/joshdurbin/url-shortener/internal/transport/http"
)

var rootCmd = &cobra.Command{
	Use:   "url-shortener",
	Short: "A URL shortening service written in Go",
	Long:  "A high-performance URL shortening service with SQLite backend and configurable caching (memory or Redis)",
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the URL shortening server",
	RunE:  runServer,
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Client commands for interacting with the server",
}

var createCmd = &cobra.Command{
	Use:   "create [URL]",
	Short: "Create a short URL",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreateURL,
}

var getCmd = &cobra.Command{
	Use:   "get [SHORT_CODE]",
	Short: "Get information about a short URL",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetURL,
}

var deleteCmd = &cobra.Command{
	Use:   "delete [SHORT_CODE]",
	Short: "Delete a short URL",
	Args:  cobra.ExactArgs(1),
	RunE:  runDeleteURL,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all short URLs",
	RunE:  runListURLs,
}

func init() {
	// Server command flags
	serverCmd.Flags().StringP("port", "p", "8080", "Server port")
	serverCmd.Flags().String("server-url", "http://localhost:8080", "Server URL (for client communication)")
	serverCmd.Flags().String("db-path", "urls.db", "Database file path")
	serverCmd.Flags().Duration("sync-interval", 5*time.Second, "Cache sync interval")
	
	// Shortener configuration flags
	serverCmd.Flags().Int64("shortener-counter-step", 100, "Counter step size for counter-based generator")
	
	// Logging configuration flags
	serverCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging (HTTP requests/responses and error details)")
	
	// Client command flags
	clientCmd.PersistentFlags().StringP("server-url", "u", "http://localhost:8080", "Server URL")
	
	// Add subcommands
	clientCmd.AddCommand(createCmd, getCmd, deleteCmd, listCmd)
	rootCmd.AddCommand(serverCmd, clientCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	// Get configuration from CLI flags
	port, _ := cmd.Flags().GetString("port")
	serverURL, _ := cmd.Flags().GetString("server-url")
	dbPath, _ := cmd.Flags().GetString("db-path")
	syncInterval, _ := cmd.Flags().GetDuration("sync-interval")
	
	// Get shortener configuration
	shortenerCounterStep, _ := cmd.Flags().GetInt64("shortener-counter-step")
	
	// Get logging configuration
	verbose, _ := cmd.Flags().GetBool("verbose")
	
	shortenerConfig := shortener.Config{
		CounterStep: shortenerCounterStep,
	}
	
	// Create configuration
	cfg, err := config.New(port, serverURL, dbPath, syncInterval, verbose, shortenerConfig)
	if err != nil {
		return fmt.Errorf("failed to create configuration: %w", err)
	}

	log.Printf("Starting URL shortener server with config: port=%s", cfg.Server.Port)


	// Initialize database
	repo, err := sqlite.New(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing repository: %v", err)
		}
	}()

	// Initialize shortener generator
	generator, err := shortener.NewGenerator(cfg.Shortener, repo.GetQueries())
	if err != nil {
		return fmt.Errorf("failed to create shortener generator: %w", err)
	}
	log.Printf("Using %s shortener generator", generator.Type())

	// Initialize cache and service
	memoryCache := memory.New()
	urlShortener := service.NewURLShortener(repo, memoryCache, generator)
	log.Printf("Using in-memory cache")

	defer func() {
		if err := urlShortener.Close(); err != nil {
			log.Printf("Error closing shortener: %v", err)
		}
	}()

	// Initialize cache with existing data
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := urlShortener.InitializeCache(ctx); err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	// Start cache synchronization
	if err := urlShortener.StartCacheSync(ctx, cfg.Cache.SyncInterval); err != nil {
		return fmt.Errorf("failed to start cache sync: %w", err)
	}
	defer func() {
		if err := urlShortener.StopCacheSync(); err != nil {
			log.Printf("Error stopping cache sync: %v", err)
		}
	}()


	// Create and start HTTP server
	server := httpTransport.NewServer(urlShortener, cfg.Server.Port, cfg.Server.ServerURL, cfg.Logging.Verbose)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		
		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		
		// Shutdown server
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}

	log.Println("Server stopped")
	return nil
}

func runCreateURL(cmd *cobra.Command, args []string) error {
	serverURL, _ := cmd.Flags().GetString("server-url")
	c := client.NewClient(serverURL)
	commands := client.NewCommands(c)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return commands.Create(ctx, args[0])
}

func runGetURL(cmd *cobra.Command, args []string) error {
	serverURL, _ := cmd.Flags().GetString("server-url")
	c := client.NewClient(serverURL)
	commands := client.NewCommands(c)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return commands.Get(ctx, args[0])
}

func runDeleteURL(cmd *cobra.Command, args []string) error {
	serverURL, _ := cmd.Flags().GetString("server-url")
	c := client.NewClient(serverURL)
	commands := client.NewCommands(c)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return commands.Delete(ctx, args[0])
}

func runListURLs(cmd *cobra.Command, args []string) error {
	serverURL, _ := cmd.Flags().GetString("server-url")
	c := client.NewClient(serverURL)
	commands := client.NewCommands(c)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return commands.List(ctx)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}