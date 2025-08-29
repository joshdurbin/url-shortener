package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/joshdurbin/url-shortener/internal/metrics"
)

// MetricsMiddleware creates HTTP middleware for collecting metrics
type MetricsMiddleware struct {
	metrics metrics.Collector
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(metricsCollector metrics.Collector) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: metricsCollector,
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Middleware returns the HTTP middleware function
func (m *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Extract endpoint from path
		endpoint := normalizeEndpoint(r.URL.Path)

		// Process the request
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start)
		status := strconv.Itoa(rw.statusCode)
		
		m.metrics.RecordHTTPRequest(r.Method, endpoint, status, duration)
	})
}

// normalizeEndpoint normalizes URL paths for consistent metrics
func normalizeEndpoint(path string) string {
	// Normalize specific paths for better grouping in metrics
	if path == "/" {
		return "/"
	}
	if path == "/api/urls" {
		return "/api/urls"
	}
	if len(path) > 10 && path[:10] == "/api/urls/" {
		return "/api/urls/{code}"
	}
	// For redirect endpoints, group them together
	if path != "/" && path != "/api/urls" && !startsWith(path, "/api/") {
		return "/{code}"
	}
	return path
}

// startsWith checks if string s starts with prefix
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}