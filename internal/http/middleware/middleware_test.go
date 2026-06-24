package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"shortener/internal/http/middleware"
	"testing"

	"github.com/Winushkin/go-toolkit/logger"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	ctx := context.Background()
	ctx, _ = logger.NewLoggerContext(ctx, true)

	nextHandlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true

		log, ok := logger.GetLoggerFromCtx(r.Context())
		assert.True(t, ok)
		assert.NotNil(t, log)

		w.WriteHeader(http.StatusTeapot)
	})

	wrappedHandler := middleware.TelemetryMiddleware(nextHandler)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/test-path", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.True(t, nextHandlerCalled, "Мидлварь должна была вызвать следующий хендлер в цепочке")
	assert.Equal(t, http.StatusTeapot, rr.Code, "Мидлварь не должна менять статус ответа, установленный хендлером")
}
