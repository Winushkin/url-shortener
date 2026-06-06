package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Winushkin/go-toolkit/config"
	"github.com/Winushkin/go-toolkit/logger"
	"github.com/Winushkin/go-toolkit/postgres"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

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
	pool, err := postgres.NewPool(ctx, cfg.Postgres)
	if err != nil {
		panic(fmt.Errorf("failed to create postgres pool: %w", err))
	}
	defer pool.Close()

	log.Info(ctx, "Успешное подключение к базе данных!", zap.String("Port", cfg.Postgres.Port))

	// Миграции
	if err := goose.SetDialect("postgres"); err != nil {
		log.Error(ctx, "Ошибка настройки диалекта: %v", zap.Error(err))
	}
	db := stdlib.OpenDBFromPool(pool)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	log.Info(ctx, "Запуск миграций базы данных")
	if err := goose.Up(db, ""); err != nil {
		log.Error(ctx, "Ошибка выполнения миграций: %v", zap.Error(err))
	}
	log.Info(ctx, "Миграции успешно применены!")

}
