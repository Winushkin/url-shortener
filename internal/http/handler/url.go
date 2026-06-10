package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Winushkin/go-toolkit/logger"
)

type CtxKey string

const (
	ProtocolKey CtxKey = "protocol"
	DomainKey   CtxKey = "domain"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ViewURL  string `json:"view_url"`
	ShortURL string `json:"short_url"`
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		panic("logger not in ctx")
	}

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
	viewURL := fmt.Sprintf("%s/%s", h.domain, shortCode)
	fullShortURL := h.protocol + "://" + viewURL

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(shortenResponse{
		ShortURL: fullShortURL,
		ViewURL:  viewURL,
	})
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		panic("logger not in ctx")
	}

	shortCode := r.PathValue("code")
	if shortCode == ""{
		log.Error(ctx, nil, "Empty shortCode")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 2. Игнорируем запрос иконки, чтобы он не шел в БД
    if shortCode == "favicon.ico" {
        w.WriteHeader(http.StatusNoContent) // Возвращаем 204 статус
        return
    }

	longURL, err := h.useCase.GetLongURL(r.Context(), shortCode)
	if err != nil {
		log.Error(ctx, err, "GetLongURL error")
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
        longURL = "https://" + longURL
    }

	http.Redirect(w, r, longURL, http.StatusFound)
}
