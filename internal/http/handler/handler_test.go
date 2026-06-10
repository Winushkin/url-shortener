package handler_test

import (
	"net/http"
	"net/http/httptest"
	"shortener/internal/http/handler"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"valid URL", "https://example.com", true},
		{"valid URL", "ftp://192.168.1.1", true},
		{"invalid URL", "invalid-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handler.IsValidURL(tt.url); got != tt.want {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestHandler_ShortenURL(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(m *URLUseCaseMock)
		wantStatus     int
		wantInResponse string
	}{
		{
			name: "Успешное сокращение (201 Created)",
			body: `{"url": "https://example.com"}`,
			setupMock: func(m *URLUseCaseMock) {
				m.On("Shorten", mock.Anything, "https://example.com").Return("abc", nil)
			},
			wantStatus:     http.StatusCreated,
			wantInResponse: `"short_url":"localhost/abc"`,
		},
		{
			name:           "Ошибка: пустой или невалидный JSON (400 Bad Request)",
			body:           `{invalid json}`,
			setupMock:      func(m *URLUseCaseMock) {}, // мок не должен вызываться
			wantStatus:     http.StatusBadRequest,
			wantInResponse: "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(URLUseCaseMock)
			tt.setupMock(mockUC)

			h := handler.NewHandler(mockUC, "localhost")

			// Создаем фейковый HTTP запрос
			req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			// Создаем ResponseRecorder (заменяет ResponseWriter для записи ответа в память)
			rr := httptest.NewRecorder()

			// Вызываем метод хендлера напрямую
			h.ShortenURL(rr, req)

			// Проверяем статус-код и тело ответа
			assert.Equal(t, tt.wantStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.wantInResponse)
			mockUC.AssertExpectations(t)
		})
	}
}

