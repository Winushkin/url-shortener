// package main

// import (
// 	_ "shortener/migrations"
// 	"github.com/jackc/pgx/v5/stdlib"
// 	"shortener/internal/postgres"
// )

// func main(

// 	postgres.Config

// )

package main

import (
	"context"
	"fmt"

	"github.com/Winushkin/go-toolkit/config"
	"github.com/Winushkin/go-toolkit/logger"
	"github.com/Winushkin/go-toolkit/postgres"
	"go.uber.org/zap"
)

const devMode = true

func main() {
	// 1. Инициализация логгера
	ctx, err := logger.NewLoggerContext(context.Background(), devMode)
	if err != nil {
		panic(fmt.Errorf("failed to create logger context: %w", err))
	}

	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		panic("logger not found in context")
	}

	// 2. Инициализация конфигурации PostgreSQL
	pgCfg := postgres.Config{
		Host: config.GetEnv("POSTGRES_HOST", "localhost"),
		Port: config.GetEnv("POSTGRES_PORT", "5432"),
		// ...
		MaxConns: config.GetEnv("POSTGRES_MAXCONNS", "1"),
	}

	// 3. Подключение к PostgreSQL
	pool, err := postgres.NewPool(ctx, pgCfg)
	if err != nil {
		panic(fmt.Errorf("failed to create postgres pool: %w", err))
	}

	log.Info(ctx, "Успешное подключение к базе данных!", zap.String("Port", pgCfg.Port))
}
