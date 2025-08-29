package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/joshdurbin/url-shortener/internal/domain"
	"github.com/joshdurbin/url-shortener/internal/service"
)

// Handler holds the HTTP handlers for the URL shortener
type Handler struct {
	shortener service.URLShortener
	serverURL string
}

// NewHandler creates a new HTTP handler
func NewHandler(shortener service.URLShortener, serverURL string) *Handler {
	return &Handler{
		shortener: shortener,
		serverURL: serverURL,
	}
}

// CreateURL handles POST /api/urls
func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid JSON in create URL request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		log.Printf("[ERROR] Empty URL provided in create request")
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	entry, err := h.shortener.CreateShortURL(r.Context(), req.URL)
	if err != nil {
		log.Printf("[ERROR] Failed to create short URL for '%s': %v", req.URL, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := domain.CreateURLResponse{
		ShortCode:   entry.ShortCode,
		ShortURL:    h.serverURL + "/" + entry.ShortCode,
		OriginalURL: entry.OriginalURL,
		CreatedAt:   entry.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetURL handles GET /api/urls/{shortCode}
func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := strings.TrimPrefix(r.URL.Path, "/api/urls/")
	if shortCode == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	entry, err := h.shortener.GetURLInfo(r.Context(), shortCode)
	if err != nil {
		log.Printf("[ERROR] Failed to get URL info for code '%s': %v", shortCode, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entry); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// DeleteURL handles DELETE /api/urls/{shortCode}
func (h *Handler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := strings.TrimPrefix(r.URL.Path, "/api/urls/")
	if shortCode == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	err := h.shortener.DeleteShortURL(r.Context(), shortCode)
	if err != nil {
		log.Printf("[ERROR] Failed to delete URL with code '%s': %v", shortCode, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListURLs handles GET /api/urls
func (h *Handler) ListURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries, err := h.shortener.GetAllURLs(r.Context())
	if err != nil {
		log.Printf("Error getting all URLs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Redirect handles GET /{shortCode} - redirects to original URL
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == "" || shortCode == "api/urls" || strings.HasPrefix(shortCode, "api/") {
		http.NotFound(w, r)
		return
	}

	originalURL, err := h.shortener.GetOriginalURL(r.Context(), shortCode)
	if err != nil {
		log.Printf("[ERROR] Failed to get original URL for code '%s': %v", shortCode, err)
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// URLsHandler handles both POST /api/urls and GET /api/urls
func (h *Handler) URLsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateURL(w, r)
	case http.MethodGet:
		h.ListURLs(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// URLsDetailHandler handles GET /api/urls/{shortCode} and DELETE /api/urls/{shortCode}
func (h *Handler) URLsDetailHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetURL(w, r)
	case http.MethodDelete:
		h.DeleteURL(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}