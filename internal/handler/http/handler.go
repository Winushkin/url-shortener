// Package http содержит реализацию http-обработчиков
package http

import (
	"net/http"
	"shortener/internal/usecase"
)

type Handler struct {
	useCase    usecase.URLUseCase
	domain string
}

func NewHandler(uc usecase.URLUseCase, domainName string) *Handler {
	return &Handler{
		useCase: uc,
		domain: domainName,
	}
}

func (h *Handler) RegisterRouters(mux *http.ServeMux) {
	mux.HandleFunc("POST /shorten", h.ShortenURL)
	mux.HandleFunc("GET /{code}", h.Redirect)
}
