// Package middleware содержит мидлварю для логгирования запроса
package middleware

import (
	"fmt"
	"sync/atomic"

	"net/http"
	"time"

	"github.com/Winushkin/go-toolkit/logger"

	"go.uber.org/zap"
)

// responseWriterInterceptor перехватывает HTTP-статус ответа
type responseWriterInterceptor struct {
	http.ResponseWriter

	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterInterceptor) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// LoggingMiddleware адаптирована под ваш кастомный пакет logger
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		// 1. Пытаемся достать логгер, который инициализировался при старте приложения.
		// Если его нет в контексте, мидлварь ничего не сделает (защита от panic).
		log, ok := logger.GetLoggerFromCtx(ctx)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Генерируем уникальный Request ID для текущего HTTP-запроса
		reqID := generateRequestID()

		// 3. Обогащаем контекст с помощью вашей функции WithRequestID.
		// Теперь методы l.Info(ctx, ...) автоматически увидят этот ID.
		ctx = logger.WithRequestID(ctx, reqID)
		r = r.WithContext(ctx)

		// 4. Подменяем ResponseWriter для перехвата статус-кода
		interceptor := &responseWriterInterceptor{ResponseWriter: w}

		// 5. Передаем управление хендлерам дальше по цепочке
		next.ServeHTTP(interceptor, r)

		// 6. Вычисляем время обработки запроса
		duration := time.Since(start)

		statusCode := interceptor.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		// 7. Пишем структурированный лог, используя ваш метод Info.
		// request_id добавится автоматически внутри вашего метода l.withRequestID(ctx)
		log.Info(ctx, "http request processed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
			zap.String("remote_addr", r.RemoteAddr),
		)
	})
}

var requestCounter atomic.Uint64

func generateRequestID() string {
	id := requestCounter.Add(1) 
	return fmt.Sprintf("%016x", id)
}