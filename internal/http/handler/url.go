package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Winushkin/go-toolkit/logger"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, _ := logger.GetLoggerFromCtx(ctx)

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(ctx, nil, "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !IsValidURL(req.URL) {
		log.Error(ctx, nil, "URL is Invalid")
		http.Error(w, "URL is Invalid", http.StatusBadRequest)
		return
	}

	shortCode, err := h.useCase.Shorten(r.Context(), req.URL)
	if err != nil {
		log.Error(ctx, err, "Failed to shorten URL")
		http.Error(w, "Failed to shorten URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fullShortURL := fmt.Sprintf("%s/%s", h.domain, shortCode)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(shortenResponse{ShortURL: fullShortURL})
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, _ := logger.GetLoggerFromCtx(ctx)

	shortCode := r.PathValue("code")
	if shortCode == "" {
		log.Error(ctx, nil, "Empty shortCode")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	longURL, err := h.useCase.GetLongURL(r.Context(), shortCode)
	if err != nil {
		log.Error(ctx, err, "GetLongURL error")
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}
