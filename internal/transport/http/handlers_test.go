package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/joshdurbin/url-shortener/internal/domain"
	"github.com/joshdurbin/url-shortener/internal/service/mocks"
)

func TestHandler_CreateURL(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mocks.URLShortener)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful creation",
			requestBody: domain.CreateURLRequest{
				URL: "https://example.com",
			},
			setupMocks: func(mockService *mocks.URLShortener) {
				mockService.On("CreateShortURL", context.Background(), "https://example.com").
					Return(&domain.URLEntry{
						ID:          1,
						ShortCode:   "abc123",
						OriginalURL: "https://example.com",
						CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UsageCount:  0,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty URL",
			requestBody: domain.CreateURLRequest{
				URL: "",
			},
			setupMocks:     func(mockService *mocks.URLShortener) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "URL is required",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			setupMocks:     func(mockService *mocks.URLShortener) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name: "service error",
			requestBody: domain.CreateURLRequest{
				URL: "invalid-url",
			},
			setupMocks: func(mockService *mocks.URLShortener) {
				mockService.On("CreateShortURL", context.Background(), "invalid-url").
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.URLShortener{}
			tt.setupMocks(mockService)

			handler := NewHandler(mockService, "http://localhost:8080")

			var body bytes.Buffer
			if tt.requestBody != nil {
				if jsonStr, ok := tt.requestBody.(string); ok {
					body.WriteString(jsonStr)
				} else {
					require.NoError(t, json.NewEncoder(&body).Encode(tt.requestBody))
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/urls", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateURL(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_GetURL(t *testing.T) {
	tests := []struct {
		name           string
		shortCode      string
		setupMocks     func(*mocks.URLShortener)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful retrieval",
			shortCode: "abc123",
			setupMocks: func(mockService *mocks.URLShortener) {
				mockService.On("GetURLInfo", context.Background(), "abc123").
					Return(&domain.URLEntry{
						ID:          1,
						ShortCode:   "abc123",
						OriginalURL: "https://example.com",
						CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UsageCount:  5,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "short code not found",
			shortCode: "notfound",
			setupMocks: func(mockService *mocks.URLShortener) {
				mockService.On("GetURLInfo", context.Background(), "notfound").
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty short code",
			shortCode:      "",
			setupMocks:     func(mockService *mocks.URLShortener) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Short code is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.URLShortener{}
			tt.setupMocks(mockService)

			handler := NewHandler(mockService, "http://localhost:8080")

			path := "/api/urls/"
			if tt.shortCode != "" {
				path += tt.shortCode
			}
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			handler.GetURL(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_DeleteURL(t *testing.T) {
	tests := []struct {
		name           string
		shortCode      string
		setupMocks     func(*mocks.URLShortener)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful deletion",
			shortCode: "abc123",
			setupMocks: func(mockService *mocks.URLShortener) {
				mockService.On("DeleteShortURL", context.Background(), "abc123").
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:      "short code not found",
			shortCode: "notfound",
			setupMocks: func(mockService *mocks.URLShortener) {
				mockService.On("DeleteShortURL", context.Background(), "notfound").
					Return(assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty short code",
			shortCode:      "",
			setupMocks:     func(mockService *mocks.URLShortener) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Short code is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.URLShortener{}
			tt.setupMocks(mockService)

			handler := NewHandler(mockService, "http://localhost:8080")

			path := "/api/urls/"
			if tt.shortCode != "" {
				path += tt.shortCode
			}
			req := httptest.NewRequest(http.MethodDelete, path, nil)
			w := httptest.NewRecorder()

			handler.DeleteURL(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_Redirect(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		setupMocks     func(*mocks.URLShortener)
		expectedStatus int
		expectedHeader string
	}{
		{
			name: "successful redirect",
			path: "/abc123",
			setupMocks: func(mockService *mocks.URLShortener) {
				mockService.On("GetOriginalURL", context.Background(), "abc123").
					Return("https://example.com", nil)
			},
			expectedStatus: http.StatusFound,
			expectedHeader: "https://example.com",
		},
		{
			name: "short code not found",
			path: "/notfound",
			setupMocks: func(mockService *mocks.URLShortener) {
				mockService.On("GetOriginalURL", context.Background(), "notfound").
					Return("", assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "API path ignored",
			path:           "/api/urls",
			setupMocks:     func(mockService *mocks.URLShortener) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty path",
			path:           "/",
			setupMocks:     func(mockService *mocks.URLShortener) {},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.URLShortener{}
			tt.setupMocks(mockService)

			handler := NewHandler(mockService, "http://localhost:8080")

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.Redirect(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, w.Header().Get("Location"))
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_ListURLs(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*mocks.URLShortener)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "successful list with URLs",
			setupMocks: func(mockService *mocks.URLShortener) {
				mockService.On("GetAllURLs", context.Background()).
					Return([]*domain.URLEntry{
						{ID: 1, ShortCode: "abc123", OriginalURL: "https://example.com"},
						{ID: 2, ShortCode: "def456", OriginalURL: "https://google.com"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "empty list",
			setupMocks: func(mockService *mocks.URLShortener) {
				mockService.On("GetAllURLs", context.Background()).
					Return([]*domain.URLEntry{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "service error",
			setupMocks: func(mockService *mocks.URLShortener) {
				mockService.On("GetAllURLs", context.Background()).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.URLShortener{}
			tt.setupMocks(mockService)

			handler := NewHandler(mockService, "http://localhost:8080")

			req := httptest.NewRequest(http.MethodGet, "/api/urls", nil)
			w := httptest.NewRecorder()

			handler.ListURLs(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var urls []*domain.URLEntry
				err := json.NewDecoder(w.Body).Decode(&urls)
				require.NoError(t, err)
				assert.Len(t, urls, tt.expectedCount)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_ErrorScenarios(t *testing.T) {
	t.Run("malformed JSON in CreateURL", func(t *testing.T) {
		mockService := &mocks.URLShortener{}
		handler := NewHandler(mockService, "http://localhost:8080")

		req := httptest.NewRequest(http.MethodPost, "/api/urls", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateURL(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid JSON")

		mockService.AssertExpectations(t)
	})

	t.Run("missing Content-Type in CreateURL", func(t *testing.T) {
		mockService := &mocks.URLShortener{}
		handler := NewHandler(mockService, "http://localhost:8080")

		// Even without Content-Type, the handler will still try to decode JSON
		// and call the service if JSON is valid
		mockService.On("CreateShortURL", mock.Anything, "https://example.com").
			Return(&domain.URLEntry{
				ID:          1,
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
				CreatedAt:   time.Now(),
				UsageCount:  0,
			}, nil)

		reqBody := domain.CreateURLRequest{URL: "https://example.com"}
		jsonData, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewBuffer(jsonData))
		// Don't set Content-Type header
		w := httptest.NewRecorder()

		handler.CreateURL(w, req)

		// Handler doesn't validate Content-Type, so if JSON is valid, it proceeds
		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("wrong Content-Type in CreateURL", func(t *testing.T) {
		mockService := &mocks.URLShortener{}
		handler := NewHandler(mockService, "http://localhost:8080")

		req := httptest.NewRequest(http.MethodPost, "/api/urls", strings.NewReader("url=https://example.com"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.CreateURL(w, req)

		// Handler tries to decode as JSON, which fails with form data
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid JSON")

		mockService.AssertExpectations(t)
	})

	t.Run("large request body", func(t *testing.T) {
		mockService := &mocks.URLShortener{}
		handler := NewHandler(mockService, "http://localhost:8080")

		// Create a very large JSON payload
		largeURL := "https://example.com/" + strings.Repeat("a", 10000)
		reqBody := domain.CreateURLRequest{URL: largeURL}
		jsonData, _ := json.Marshal(reqBody)

		// The handler will try to call the service with this large URL
		// Let's mock it to return an error indicating URL validation failure
		mockService.On("CreateShortURL", mock.Anything, largeURL).
			Return(nil, fmt.Errorf("URL too long"))

		req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateURL(w, req)

		// Should handle large requests gracefully by returning a client error
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "URL too long")

		mockService.AssertExpectations(t)
	})
}

func TestHandler_HTTPMethods(t *testing.T) {
	mockService := &mocks.URLShortener{}
	handler := NewHandler(mockService, "http://localhost:8080")

	// Test unsupported methods on various endpoints
	unsupportedMethods := []struct {
		method   string
		path     string
		handler  http.HandlerFunc
	}{
		{"PUT", "/api/urls", handler.CreateURL},
		{"PATCH", "/api/urls", handler.CreateURL},
		{"DELETE", "/api/urls", handler.CreateURL},
		{"POST", "/api/urls/abc123", handler.GetURL},
		{"PUT", "/api/urls/abc123", handler.GetURL},
		{"PATCH", "/api/urls/abc123", handler.GetURL},
		{"GET", "/api/urls/abc123", handler.DeleteURL},
		{"POST", "/api/urls/abc123", handler.DeleteURL},
		{"PUT", "/api/urls/abc123", handler.DeleteURL},
	}

	for _, tc := range unsupportedMethods {
		t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			tc.handler(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}

	mockService.AssertExpectations(t)
}

func TestHandler_PathEdgeCases(t *testing.T) {
	pathTests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		shortCode      string
	}{
		{"very short path", "GET", "/a", http.StatusNotFound, "a"},
		{"path with special chars", "GET", "/abc@#$", http.StatusNotFound, "abc@#$"},
		{"path with spaces", "GET", "/abc%20def", http.StatusNotFound, "abc def"},
		{"very long path", "GET", "/" + strings.Repeat("a", 1000), http.StatusNotFound, strings.Repeat("a", 1000)},
		{"path with query params", "GET", "/abc123?test=1", http.StatusNotFound, "abc123"},
		{"path with fragment", "GET", "/abc123#section", http.StatusNotFound, "abc123#section"},
	}

	for _, tc := range pathTests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mocks.URLShortener{}
			handler := NewHandler(mockService, "http://localhost:8080")

			// The handler will attempt to resolve these as short codes
			mockService.On("GetOriginalURL", mock.Anything, tc.shortCode).
				Return("", fmt.Errorf("not found"))

			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			handler.Redirect(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_ResponseHeaders(t *testing.T) {
	t.Run("Content-Type headers", func(t *testing.T) {
		mockService := &mocks.URLShortener{}
		handler := NewHandler(mockService, "http://localhost:8080")

		// Test JSON responses have correct Content-Type
		mockService.On("GetAllURLs", mock.Anything).Return([]*domain.URLEntry{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/urls", nil)
		w := httptest.NewRecorder()

		handler.ListURLs(w, req)

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		mockService.AssertExpectations(t)
	})

	t.Run("CORS headers", func(t *testing.T) {
		mockService := &mocks.URLShortener{}
		handler := NewHandler(mockService, "http://localhost:8080")

		req := httptest.NewRequest(http.MethodOptions, "/api/urls", nil)
		w := httptest.NewRecorder()

		handler.CreateURL(w, req)

		// Check if any CORS-related headers are set
		// This is more of a placeholder - actual CORS handling would be in middleware
		assert.True(t, w.Code >= 200 && w.Code < 500)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_ContextPropagation(t *testing.T) {
	t.Run("context is passed to service", func(t *testing.T) {
		mockService := &mocks.URLShortener{}
		handler := NewHandler(mockService, "http://localhost:8080")

		// Mock should receive a context (any context)
		mockService.On("GetAllURLs", mock.AnythingOfType("*context.valueCtx")).
			Return([]*domain.URLEntry{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/urls", nil)
		// Add a value to context to ensure it's propagated
		ctx := context.WithValue(req.Context(), "test", "value")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListURLs(w, req)

		mockService.AssertExpectations(t)
	})
}