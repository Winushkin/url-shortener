package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request){
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !IsValidURL(req.URL){
		http.Error(w, "URL is Invalid", http.StatusBadRequest)
		return
	}

	shortCode, err := h.useCase.Shorten(r.Context(), req.URL)
	if err != nil {
		http.Error(w, "Failed to shorten URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fullShortURL := fmt.Sprintf("http://%s/%s", h.domain, shortCode)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(shortenResponse{ShortURL: fullShortURL})
}	

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request){
	shortCode := r.PathValue("code")
	if shortCode == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	longURL, err := h.useCase.GetLongURL(r.Context(), shortCode)
	if err != nil {
		// В реальном приложении тут должна быть проверка на ошибку domain.ErrNotFound
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}


	http.Redirect(w, r, longURL, http.StatusFound)
}