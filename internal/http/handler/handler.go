// Package handler содержит реализацию http-обработчиков
package handler

import (
	"net/http"
	"net/url"
	"shortener/internal/usecase"
)

type Handler struct {
	useCase usecase.URLUseCase
	domain  string
}

func NewHandler(uc usecase.URLUseCase, domainName string) *Handler {
	return &Handler{
		useCase: uc,
		domain:  domainName,
	}
}

func (h *Handler) RegisterRouters(mux *http.ServeMux) {
	mux.HandleFunc("GET /{code}", h.Redirect)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	mux.HandleFunc("POST /shorten", h.ShortenURL)
}

func IsValidURL(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}

	return parsedURL.Scheme != "" && parsedURL.Host != ""
}
