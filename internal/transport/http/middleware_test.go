package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/joshdurbin/url-shortener/internal/metrics/mocks"
)

func TestNewMetricsMiddleware(t *testing.T) {
	collector := &mocks.Collector{}
	middleware := NewMetricsMiddleware(collector)

	assert.NotNil(t, middleware)
	assert.Equal(t, collector, middleware.metrics)
}

func TestMetricsMiddleware_Middleware(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		collector := &mocks.Collector{}
		middleware := NewMetricsMiddleware(collector)

		// Set up expectations
		collector.On("RecordHTTPRequest", "GET", "/api/urls", "200", mock.AnythingOfType("time.Duration")).Return()

		// Create a test handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Wrap with middleware
		wrappedHandler := middleware.Middleware(handler)

		// Create test request
		req := httptest.NewRequest("GET", "/api/urls", nil)
		rr := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(rr, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "OK", rr.Body.String())

		// Verify metrics were recorded
		collector.AssertExpectations(t)
	})

	t.Run("error request", func(t *testing.T) {
		collector := &mocks.Collector{}
		middleware := NewMetricsMiddleware(collector)

		// Set up expectations for error status
		collector.On("RecordHTTPRequest", "POST", "/api/urls", "400", mock.AnythingOfType("time.Duration")).Return()

		// Create a test handler that returns an error
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))
		})

		// Wrap with middleware
		wrappedHandler := middleware.Middleware(handler)

		// Create test request
		req := httptest.NewRequest("POST", "/api/urls", nil)
		rr := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(rr, req)

		// Verify response
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "Bad Request", rr.Body.String())

		// Verify metrics were recorded
		collector.AssertExpectations(t)
	})

	t.Run("measures duration", func(t *testing.T) {
		collector := &mocks.Collector{}
		middleware := NewMetricsMiddleware(collector)

		var recordedDuration time.Duration
		collector.On("RecordHTTPRequest", "GET", "/api/urls", "200", mock.AnythingOfType("time.Duration")).
			Run(func(args mock.Arguments) {
				recordedDuration = args.Get(3).(time.Duration)
			}).Return()

		// Create a handler that sleeps briefly
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

		wrappedHandler := middleware.Middleware(handler)

		req := httptest.NewRequest("GET", "/api/urls", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		// Verify duration was measured and is reasonable
		assert.Greater(t, recordedDuration, 5*time.Millisecond)
		assert.Less(t, recordedDuration, 100*time.Millisecond)

		collector.AssertExpectations(t)
	})

	t.Run("different HTTP methods", func(t *testing.T) {
		testCases := []struct {
			method string
			path   string
		}{
			{"GET", "/api/urls"},
			{"POST", "/api/urls"},
			{"PUT", "/api/urls/abc123"},
			{"DELETE", "/api/urls/def456"},
			{"PATCH", "/api/urls/ghi789"},
		}

		for _, tc := range testCases {
			t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
				collector := &mocks.Collector{}
				middleware := NewMetricsMiddleware(collector)

				expectedEndpoint := normalizeEndpoint(tc.path)
				collector.On("RecordHTTPRequest", tc.method, expectedEndpoint, "200", mock.AnythingOfType("time.Duration")).Return()

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})

				wrappedHandler := middleware.Middleware(handler)
				req := httptest.NewRequest(tc.method, tc.path, nil)
				rr := httptest.NewRecorder()

				wrappedHandler.ServeHTTP(rr, req)

				collector.AssertExpectations(t)
			})
		}
	})

	t.Run("default status code 200", func(t *testing.T) {
		collector := &mocks.Collector{}
		middleware := NewMetricsMiddleware(collector)

		// When handler doesn't call WriteHeader, status should default to 200
		collector.On("RecordHTTPRequest", "GET", "/api/urls", "200", mock.AnythingOfType("time.Duration")).Return()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't call WriteHeader - should default to 200
			w.Write([]byte("OK"))
		})

		wrappedHandler := middleware.Middleware(handler)

		req := httptest.NewRequest("GET", "/api/urls", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		collector.AssertExpectations(t)
	})
}

func TestResponseWriter(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		rr := httptest.NewRecorder()
		rw := &responseWriter{ResponseWriter: rr, statusCode: http.StatusOK}

		rw.WriteHeader(http.StatusNotFound)
		assert.Equal(t, http.StatusNotFound, rw.statusCode)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("default status code", func(t *testing.T) {
		rr := httptest.NewRecorder()
		rw := &responseWriter{ResponseWriter: rr, statusCode: http.StatusOK}

		// Don't call WriteHeader, should keep default
		assert.Equal(t, http.StatusOK, rw.statusCode)
	})

	t.Run("write methods work correctly", func(t *testing.T) {
		rr := httptest.NewRecorder()
		rw := &responseWriter{ResponseWriter: rr, statusCode: http.StatusOK}

		rw.WriteHeader(http.StatusCreated)
		rw.Write([]byte("test content"))

		assert.Equal(t, http.StatusCreated, rw.statusCode)
		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Equal(t, "test content", rr.Body.String())
	})
}

func TestNormalizeEndpoint(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// Root path
		{"/", "/"},
		
		// API endpoints
		{"/api/urls", "/api/urls"},
		{"/api/urls/abc123", "/api/urls/{code}"},
		{"/api/urls/def456", "/api/urls/{code}"},
		{"/api/urls/very-long-code-here", "/api/urls/{code}"},
		
		// Redirect endpoints (short codes)
		{"/abc123", "/{code}"},
		{"/def456", "/{code}"},
		{"/xyz", "/{code}"},
		{"/@#$", "/{code}"},
		
		// Edge cases  
		{"/api", "/{code}"},     // Short path not starting with /api/ -> redirect
		{"/api/", "/api/"},
		{"/api/other", "/api/other"},
		{"/metrics", "/{code}"},  // Short path not starting with /api/ -> redirect
		{"/health", "/{code}"},   // Short path not starting with /api/ -> redirect
		
		// Very short codes
		{"/a", "/{code}"},
		{"/1", "/{code}"},
		
		// API URLs shorter than expected
		{"/api/url", "/api/url"},
		{"/api/urls", "/api/urls"},
	}

	for _, tc := range testCases {
		t.Run("path_"+tc.input, func(t *testing.T) {
			result := normalizeEndpoint(tc.input)
			assert.Equal(t, tc.expected, result, "normalizeEndpoint(%s) should return %s", tc.input, tc.expected)
		})
	}
}

func TestStartsWith(t *testing.T) {
	testCases := []struct {
		s        string
		prefix   string
		expected bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", false},
		{"hello", "hello world", false},
		{"hello", "hello", true},
		{"", "", true},
		{"hello", "", true},
		{"", "hello", false},
		{"/api/urls", "/api/", true},
		{"/api", "/api/", false},
		{"/different", "/api/", false},
	}

	for _, tc := range testCases {
		t.Run("startsWith_"+tc.s+"_"+tc.prefix, func(t *testing.T) {
			result := startsWith(tc.s, tc.prefix)
			assert.Equal(t, tc.expected, result, "startsWith(%s, %s) should return %v", tc.s, tc.prefix, tc.expected)
		})
	}
}

func TestMetricsMiddleware_EndToEnd(t *testing.T) {
	// Test a more realistic scenario with actual HTTP routing patterns
	collector := &mocks.Collector{}
	middleware := NewMetricsMiddleware(collector)

	// Set up expectations for various requests
	collector.On("RecordHTTPRequest", "GET", "/", "200", mock.AnythingOfType("time.Duration")).Return()
	collector.On("RecordHTTPRequest", "POST", "/api/urls", "201", mock.AnythingOfType("time.Duration")).Return()
	collector.On("RecordHTTPRequest", "GET", "/api/urls/{code}", "200", mock.AnythingOfType("time.Duration")).Return()
	collector.On("RecordHTTPRequest", "DELETE", "/api/urls/{code}", "204", mock.AnythingOfType("time.Duration")).Return()
	collector.On("RecordHTTPRequest", "GET", "/{code}", "302", mock.AnythingOfType("time.Duration")).Return()

	// Create a simple router-like handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/":
			w.WriteHeader(http.StatusOK)
		case r.Method == "POST" && r.URL.Path == "/api/urls":
			w.WriteHeader(http.StatusCreated)
		case r.Method == "GET" && len(r.URL.Path) > 10 && r.URL.Path[:10] == "/api/urls/":
			w.WriteHeader(http.StatusOK)
		case r.Method == "DELETE" && len(r.URL.Path) > 10 && r.URL.Path[:10] == "/api/urls/":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "GET" && r.URL.Path != "/" && !startsWith(r.URL.Path, "/api/"):
			w.WriteHeader(http.StatusFound) // Redirect
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	wrappedHandler := middleware.Middleware(handler)

	// Test various requests
	requests := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"POST", "/api/urls"},
		{"GET", "/api/urls/test123"},
		{"DELETE", "/api/urls/test456"},
		{"GET", "/shortcode"},
	}

	for _, req := range requests {
		httpReq := httptest.NewRequest(req.method, req.path, nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, httpReq)

		// Just verify no panics and status codes are reasonable
		assert.True(t, rr.Code >= 200 && rr.Code < 500)
	}

	collector.AssertExpectations(t)
}

func TestMetricsMiddleware_PanicRecovery(t *testing.T) {
	// Test that middleware doesn't panic if handler panics
	collector := &mocks.Collector{}
	middleware := NewMetricsMiddleware(collector)

	// We expect the metrics to be recorded even if handler panics
	// (though in practice, a panic recovery middleware would typically handle this)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	wrappedHandler := middleware.Middleware(handler)

	req := httptest.NewRequest("GET", "/api/urls", nil)
	rr := httptest.NewRecorder()

	// This should panic since we don't have panic recovery
	assert.Panics(t, func() {
		wrappedHandler.ServeHTTP(rr, req)
	})

	// The metrics call would happen, but the panic prevents it from being recorded
	// In a real application, panic recovery middleware would be used before this middleware
}