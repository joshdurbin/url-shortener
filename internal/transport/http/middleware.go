package http

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware creates HTTP middleware for logging requests and responses
type LoggingMiddleware struct {
	verbose bool
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(verbose bool) *LoggingMiddleware {
	return &LoggingMiddleware{
		verbose: verbose,
	}
}

// loggingResponseWriter wraps http.ResponseWriter to capture response details
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.body != nil {
		lrw.body.Write(b)
	}
	return lrw.ResponseWriter.Write(b)
}

// Middleware returns the HTTP logging middleware function
func (l *LoggingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.verbose {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		// Log request
		log.Printf("[HTTP REQUEST] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		
		// Log request body for POST/PUT requests
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					log.Printf("[HTTP REQUEST] Error reading request body: %v", err)
				} else {
					// Create a new reader for the handler
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					if len(bodyBytes) > 0 {
						log.Printf("[HTTP REQUEST] Body: %s", string(bodyBytes))
					}
				}
			}
		}

		// Wrap the response writer to capture response details
		var responseBody *bytes.Buffer
		if l.verbose {
			responseBody = &bytes.Buffer{}
		}
		
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:          responseBody,
		}

		// Process the request
		next.ServeHTTP(lrw, r)

		// Log response
		duration := time.Since(start)
		log.Printf("[HTTP RESPONSE] %s %s -> %d in %v", r.Method, r.URL.Path, lrw.statusCode, duration)
		
		if responseBody != nil && responseBody.Len() > 0 && lrw.statusCode >= 400 {
			log.Printf("[HTTP RESPONSE] Error body: %s", responseBody.String())
		}
	})
}