package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"shortener/internal/http/handler"
	"strings"
	"testing"

	"github.com/Winushkin/go-toolkit/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"valid with protocol", "https://example.com", true},
		{"valid without protocol", "example.com", true},
		{"valid starts with ://", "://example.com", true},

		{"invalid, only path", "/only/path", false},
		{"invalid random text", "random-text", false},
		{"empty url", "", false},
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
			wantInResponse: `{"view_url":"localhost/abc","short_url":"http://localhost/abc"}` + "\n",
		},
		{
			name:       "Ошибка: пустой или невалидный JSON (400 Bad Request)",
			body:       `{invalid json}`,
			setupMock:  func(m *URLUseCaseMock) {},
			wantStatus: http.StatusBadRequest,
			// Убрали обратный слэш из строки, чтобы не было ошибки компиляции
			wantInResponse: "Invalid request body\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := logger.NewLoggerContext(context.Background(), false)
			if err != nil {
				t.Fatalf("failed to create logger context: %v", err) // Избегайте panic в тестах
			}

			mockUC := new(URLUseCaseMock)
			tt.setupMock(mockUC)

			h := handler.NewHandler(mockUC, "localhost", "http")

			req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/shorten", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.ShortenURL(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			// JSON может иметь пробелы, поэтому надежнее использовать Contains или JSONEq
			assert.Contains(t, rr.Body.String(), tt.wantInResponse)
			mockUC.AssertExpectations(t)
		})
	}
}
