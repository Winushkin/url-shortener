// Package middleware содержит мидлварю для логгирования запроса
package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/Winushkin/go-toolkit/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	requestCounter atomic.Uint64

	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Количество входящих HTTP-запросов.",
		},
		[]string{"path", "method", "code"}, // Лейблы, которые используются в наших графиках Grafana
	)

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Длительность обработки запросов в секундах.",
			Buckets: prometheus.DefBuckets, // Бакеты по умолчанию (от 0.005 до 10 секунд)
		},
		[]string{"path", "method"},
	)
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

func generateRequestID() string {
	id := requestCounter.Add(1)
	return fmt.Sprintf("%016x", id)
}

func getURLPattern(r *http.Request) string {
	if r.Pattern == "" {
		return "Not Found"
	}
	return r.Pattern
}

func TelemetryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		log, ok := logger.GetLoggerFromCtx(ctx)
		if !ok {
			// Если логгера нет, то запишем запрос просто в Prometheus
			interceptor := &responseWriterInterceptor{ResponseWriter: w}
			next.ServeHTTP(interceptor, r)

			duration := time.Since(start).Seconds()
			statusCode := interceptor.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			pathPattern := getURLPattern(r)
			httpRequestsTotal.WithLabelValues(pathPattern, r.Method, strconv.Itoa(statusCode)).Inc()
			httpDuration.WithLabelValues(pathPattern, r.Method).Observe(duration)
			return
		}

		reqID := generateRequestID()
		ctx = logger.WithRequestID(ctx, reqID)
		r = r.WithContext(ctx)

		interceptor := &responseWriterInterceptor{ResponseWriter: w}

		next.ServeHTTP(interceptor, r)

		duration := time.Since(start)

		statusCode := interceptor.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		statusCodeStr := strconv.Itoa(statusCode)

		pathPattern := getURLPattern(r)
		httpRequestsTotal.WithLabelValues(pathPattern, r.Method, statusCodeStr).Inc()
		httpDuration.WithLabelValues(pathPattern, r.Method).Observe(duration.Seconds())

		log.Info(ctx, "http request processed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
			zap.String("remote_addr", r.RemoteAddr),
		)
	})
}
