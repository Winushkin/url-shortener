// Package handler содержит реализацию http-обработчиков
package handler

import (
	"net/http"
	"net/url"
	"shortener/internal/usecase"
	"strings"
)

type Handler struct {
	useCase  usecase.URLUseCase
	domain   string
	protocol string
}

func NewHandler(uc usecase.URLUseCase, domainName, protocolName string) *Handler {
	return &Handler{
		useCase:  uc,
		domain:   domainName,
		protocol: protocolName,
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
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return false
	}

	// Отсекаем абсолютные пути сервера
	if strings.HasPrefix(rawURL, "/") {
		return false
	}

	if strings.HasPrefix(rawURL, "://") {
		rawURL = rawURL[3:]
	}

	hasStrictScheme := strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://")
	if !hasStrictScheme {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Хост не должен быть пустым и должен содержать точку (признак домена)
	return parsedURL.Host != "" && strings.Contains(parsedURL.Host, ".")
}
