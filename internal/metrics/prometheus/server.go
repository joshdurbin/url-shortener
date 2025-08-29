package prometheus

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents a Prometheus metrics HTTP server
type Server struct {
	server   *http.Server
	port     string
	endpoint string
}

// NewServer creates a new metrics server
func NewServer(port, endpoint string) *Server {
	mux := http.NewServeMux()
	mux.Handle(endpoint, promhttp.Handler())
	
	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &Server{
		server:   server,
		port:     port,
		endpoint: endpoint,
	}
}

// Start starts the metrics server
func (s *Server) Start() error {
	log.Printf("Metrics server starting on port %s, endpoint %s", s.port, s.endpoint)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the metrics server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Metrics server shutting down...")
	return s.server.Shutdown(ctx)
}

// Port returns the server port
func (s *Server) Port() string {
	return s.port
}

// Endpoint returns the metrics endpoint path
func (s *Server) Endpoint() string {
	return s.endpoint
}

// URL returns the full metrics URL
func (s *Server) URL() string {
	return fmt.Sprintf("http://localhost:%s%s", s.port, s.endpoint)
}