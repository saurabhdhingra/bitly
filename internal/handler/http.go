package handler

import (
	"context"
	"encoding/json"

	"net/http"
	"strings"
	"time"

	"bitly/internal/domain"

	"github.com/gorilla/mux"
)

// ErrorResponse defines the structure for API error messages.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Handler holds the service dependency and implements HTTP handlers.
type Handler struct {
	Service domain.ShortenerService
}

// NewHandler creates a new Handler instance.
func NewHandler(service domain.ShortenerService) *Handler {
	return &Handler{
		Service: service,
	}
}

// respondWithError writes a JSON error response.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

// respondWithJSON writes a JSON successful response.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.MarshalIndent(payload, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Common middleware for CORS and context timeout
func (h *Handler) commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Preflight CORS handler
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Set a context timeout for API requests
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// CreateShortURL handles POST /shorten
func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	url, err := h.Service.Create(r.Context(), req.URL)

	if err != nil {
		if strings.Contains(err.Error(), domain.ErrInvalidURL) {
			respondWithError(w, http.StatusBadRequest, domain.ErrInvalidURL)
			return
		}
		if strings.Contains(err.Error(), domain.ErrConflict) {
			respondWithJSON(w, http.StatusConflict, url)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusCreated, url)
}

// GetURL handles GET /shorten/{code}
func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	url, err := h.Service.Get(r.Context(), shortCode)

	if err != nil {
		if strings.Contains(err.Error(), domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, domain.ErrNotFound)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, url)
}

// UpdateURL handles PUT /shorten/{code}
func (h *Handler) UpdateURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	var req domain.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	url, err := h.Service.Update(r.Context(), shortCode, req.URL)

	if err != nil {
		if strings.Contains(err.Error(), domain.ErrInvalidURL) {
			respondWithError(w, http.StatusBadRequest, domain.ErrInvalidURL)
			return
		}
		if strings.Contains(err.Error(), domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, domain.ErrNotFound)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, url)
}

// DeleteURL handles DELETE /shorten/{code}
func (h *Handler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	err := h.Service.Delete(r.Context(), shortCode)

	if err != nil {
		if strings.Contains(err.Error(), domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, domain.ErrNotFound)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetStats handles GET /shorten/{code}/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	url, err := h.Service.GetStats(r.Context(), shortCode)

	if err != nil {
		if strings.Contains(err.Error(), domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, domain.ErrNotFound)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, url)
}

// Redirect handles GET /s/{code}
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	longURL, err := h.Service.Redirect(r.Context(), shortCode)

	if err != nil {
		if strings.Contains(err.Error(), domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, domain.ErrNotFound)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect) // 307
}

// Router sets up all the application routes.
func (h *Handler) Router() *mux.Router {
	r := mux.NewRouter()
	
	// Apply common middleware to all API endpoints
	apiRouter := r.PathPrefix("/shorten").Subrouter()
	apiRouter.Use(h.commonMiddleware)

	// API Endpoints (CRUD and Stats)
	apiRouter.HandleFunc("", h.CreateShortURL).Methods("POST")
	apiRouter.HandleFunc("/{shortCode}", h.GetURL).Methods("GET", "PUT", "DELETE")
	apiRouter.HandleFunc("/{shortCode}/stats", h.GetStats).Methods("GET")

	// Redirection Endpoint (No middleware applied to keep it fast)
	r.HandleFunc("/s/{shortCode}", h.Redirect).Methods("GET")
	
	return r
}
