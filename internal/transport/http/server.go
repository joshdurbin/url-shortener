package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joshdurbin/url-shortener/internal/service"
)

// Server represents the HTTP server
type Server struct {
	handler *Handler
	server  *http.Server
	port    string
}

// NewServer creates a new HTTP server
func NewServer(shortener service.URLShortener, port, serverURL string, verbose bool) *Server {
	handler := NewHandler(shortener, serverURL)
	
	mux := http.NewServeMux()
	
	// API endpoints
	mux.HandleFunc("/api/urls", handler.URLsHandler)
	mux.HandleFunc("/api/urls/", handler.URLsDetailHandler)
	
	// Redirect endpoint (catch-all)
	mux.HandleFunc("/", handler.Redirect)
	
	// Wrap with middlewares
	var finalHandler http.Handler = mux
	
	// Add logging middleware first (outermost)
	if verbose {
		loggingMiddleware := NewLoggingMiddleware(verbose)
		finalHandler = loggingMiddleware.Middleware(finalHandler)
	}
	
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      finalHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	return &Server{
		handler: handler,
		server:  server,
		port:    port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Server starting on port %s", s.port)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Server shutting down...")
	return s.server.Shutdown(ctx)
}

// Port returns the server port
func (s *Server) Port() string {
	return s.port
}

// Handler returns the server handler (useful for testing)
func (s *Server) Handler() *Handler {
	return s.handler
}