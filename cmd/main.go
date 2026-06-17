package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Winushkin/go-toolkit/config"
	"github.com/Winushkin/go-toolkit/logger"
	"github.com/Winushkin/go-toolkit/postgres"
	"github.com/Winushkin/go-toolkit/redis"
	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"shortener/internal/broker"
	"shortener/internal/http/handler"
	"shortener/internal/http/middleware"
	repository "shortener/internal/repository"
	"shortener/internal/usecase"
	"shortener/pkg/kafka"

	_ "shortener/migrations"
	// "github.com/prometheus/client_golang/prometheus/collectors"

	// "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	devMode           = true
	readHeaderTimeout = 3 * time.Second
)

func main() {
	// Логгер
	ctx, err := logger.NewLoggerContext(context.Background(), devMode)
	if err != nil {
		panic(fmt.Errorf("failed to create logger context: %w", err))
	}

	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		panic("logger not found in context")
	}

	// Конфиг
	cfg := config.NewAppConfig()

	// БД пул
	pool := connectPSQL(ctx, cfg.Postgres)
	defer pool.Close()
	repo := repository.NewPostgres(pool)

	// Миграции
	migrate(ctx, pool)

	// Клиент Кафка
	kafkaClient, kafkaCfg := initKafka(ctx)
	defer kafkaClient.Close()

	// UseCases
	uc := initUseCase(ctx, cfg.Redis, kafkaClient, kafkaCfg, repo)

	// Хендлеры
	protocol := config.GetEnv("PROTOCOL_NAME", "http")
	handler := handler.NewHandler(uc, cfg.DomainName, protocol)

	clickConsumer := broker.NewConsumer(kafkaClient, repo)
	log.Info(ctx, "Starting Kafka Consumer...")
	go clickConsumer.Start(ctx)

	// Prometheus
	promPort := ":" + config.GetEnv("PROMETHEUS_PORT", "2112")
	log.Debug(ctx, "prometheus", zap.String("port", promPort))
	go func() {
        http.Handle("/metrics", promhttp.Handler())
        if err := http.ListenAndServe(promPort, nil); err != nil {
            log.Error(ctx, err, "Failed to start Prometheus metrics server")
        }
    }()

	// Cервер
	server := registerServer(ctx, handler, cfg.Port)

	log.Info(ctx, "Server is running", zap.String("Port", cfg.Port))
	if err := server.ListenAndServe(); err != nil {
		log.Error(ctx, err, "Server failed")
	}
}

func registerServer(ctx context.Context, handler *handler.Handler, port string) *http.Server {
	mux := http.NewServeMux()
	handler.RegisterRouters(mux)
	wrappedMux := middleware.LoggingMiddleware(mux)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           wrappedMux,
		ReadHeaderTimeout: readHeaderTimeout,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	return server
}

func initKafka(ctx context.Context) (*kgo.Client, *kafka.Config) {
	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		panic("logger not found in context")
	}

	host := config.GetEnv("KAFKA_HOST", "localhost")
	port := config.GetEnv("KAFKA_PORT", "9092")
	topic := config.GetEnv("KAFKA_URLS_TOPIC", "topic")

	cfg := kafka.Config{
		Host:    host,
		Port:    port,
		GroupID: "url-clicks-processors",
		Topics: map[string]string{
			"url": topic,
		},
	}

	cl, err := kafka.NewClient(ctx, cfg)
	if err != nil {
		log.Error(ctx, err, "failed to create Kafka Client")
		return nil, nil
	}
	return cl, &cfg
}

func initUseCase(
	ctx context.Context,
	redisCfg redis.Config,
	kafkaClient *kgo.Client,
	kafkaCfg *kafka.Config,
	repo repository.Repository,
) usecase.URLUseCase {
	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		panic("logger not found in context")
	}

	rdb, err := redis.NewRedisClient(ctx, redisCfg)
	if err != nil {
		log.Error(ctx, err, "failed to create redis DB")
	}
	// newOpts := &redis.Options{
	// 	Addr:         net.JoinHostPort(redisCfg.Host, redisCfg.Port),
	// 	Password:     redisCfg.Password,
	// 	DB:           redisCfg.DB,
	// 	PoolSize:     1000, // Достаточно для 500+ параллельных горутин
	// 	MinIdleConns: 50,   // Держать прогретыми в фоне
	// 	ReadTimeout:  200 * time.Millisecond,
	// 	WriteTimeout: 200 * time.Millisecond,
	// }
	// rdb := redis.NewClient(newOpts)
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Error(ctx, err, "failed to create snowflake Node")
	}

	pb := broker.NewPublisher(kafkaClient, kafkaCfg.Topics["url"])

	deps := usecase.Dependencies{
		Repo:      repo,
		Rdb:       rdb,
		Node:      node,
		Publisher: pb,
	}
	uc := usecase.NewURLUseCase(deps)

	return uc
}

func connectPSQL(ctx context.Context, cfg postgres.Config) *pgxpool.Pool {
	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		panic("logger not found in context")
	}

	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		log.Error(ctx, err, "failed to create postgres pool")
	}
	log.Info(ctx, "Успешное подключение к базе данных!", zap.String("Port", cfg.Port))
	return pool
}

func migrate(ctx context.Context, pool *pgxpool.Pool) {
	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		panic("logger not found in context")
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Error(ctx, err, "Ошибка настройки диалекта")
	}
	db := stdlib.OpenDBFromPool(pool)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Error(ctx, err, "failed to close db")
		}
	}(db)

	log.Info(ctx, "Запуск миграций базы данных")
	if err := goose.Up(db, "."); err != nil {
		log.Error(ctx, err, "Ошибка выполнения миграций")
	}
	log.Info(ctx, "Миграции успешно применены!")
}

func initPrometheus(){

}