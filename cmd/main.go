package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"

	"github.com/Winushkin/go-toolkit/config"
	"github.com/Winushkin/go-toolkit/logger"
	"github.com/Winushkin/go-toolkit/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"shortener/internal/http/handler"
	"shortener/internal/http/middleware"
	repository "shortener/internal/repository/postgres"
	"shortener/internal/usecase"
	_ "shortener/migrations"
)

const devMode = true

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
	pool := connectDB(ctx, cfg.Postgres)
	defer pool.Close()

	// Миграции
	migrate(ctx, pool)

	repo := repository.NewPostgres(pool)
	uc := usecase.NewURLUseCase(repo)
	handler := handler.NewHandler(uc, cfg.DomainName)

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
		Addr:    ":" + port,
		Handler: wrappedMux,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	return server
}

func connectDB(ctx context.Context, cfg postgres.Config) *pgxpool.Pool {
	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		panic("logger not found in context")
	}

	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		panic(fmt.Errorf("failed to create postgres pool: %w", err))
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
