package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"bitly/internal/domain"

	"github.com/gorilla/mux"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type Handler struct {
	Service domain.ShortenerService
}

func NewHandler(service domain.ShortenerService) *Handler {
	return &Handler{
		Service: service,
	}
}

func respondWithError(w http.ResponseWriter, code int, message string){
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}){
	response, _ := json.MarshalIndent(payload, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (h *Handler) commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions { 
			w.WriteHeader(http.StatusOK)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) CreateShortURL(w http.ResponseWriter,r *http.Request){
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
		if strings.Contains(err.Error(), domain.ErrConflict){
			respondWithJSON(w, http.StatusConflict, url)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal Server error")
		return
	}

	respondWithJSON(w, http.StatusCreated, url)
}

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

func (h *Handler) DeleteURL(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]
	
	err := h.Service.Delete(r.Context(), shortCode)

	if err != nil {
		if strings.Contains(err.Error(), domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, domain.ErrNotFound)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	url, err := h.Service.GetStats(r.Context(), shortCode)

	if err != nil {
		if strings.Contains(err.Error(), domain.ErrNotFound){
			respondWithError(w, http.StatusNotFound, domain.ErrNotFound)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
	return 
	}
	respondWithJSON(w, http.StatusOK, url)
}

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

func (h *Handler) Router() *mux.Router{
	r := mux.NewRouter()

	apiRouter := r.PathPrefix("/shorten").Subrouter()
	apiRouter.Use(h.commonMiddleware)

	apiRouter.HandleFunc("", h.CreateShortURL).Methods("POST")
	apiRouter.HandleFunc("/{shortenCode}", h.GetURL).Methods("POST", "PUT", "DELETE")
	apiRouter.HandleFunc("/{shortenCode}/stats", h.GetStats).Methods("GET")

	r.HandlerFunc("/s/{shortenCode}", h.Redirect).Methods("GET")
}