package benchmarks_test

import (
	"context"
	"shortener/internal/repository" 
	"shortener/internal/usecase" 
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v5/pgxpool"
)

const exampleURL = "https://example.com"

// Тестируем старый подход (INSERT + UPDATE)
func BenchmarkUseCase_Old_TwoQueries(b *testing.B) {
	ctx := context.Background()

	// Инициализируем пул один раз на весь запуск бенчмарка
	pool := setupTestPool(b, ctx)
	defer pool.Close()

	// Инициализируем репозиторий, который принимает *pgxpool.Pool
	oldRepo := NewLegacyPostgres(pool)
	uc := NewLegacyUseCase(oldRepo)

	b.ResetTimer() // Сбрасываем таймер, чтобы время подключения к БД не учитывалось
	for i := 0; i < b.N; i++ {
		_, err := uc.Shorten(ctx, exampleURL)
		if err != nil {
			b.Fatalf("error in old shorten: %v", err)
		}
	}
}

// Тестируем новый подход (Snowflake + 1 INSERT)
func BenchmarkUseCase_New_SnowflakeOneQuery(b *testing.B) {
	ctx := context.Background()

	pool := setupTestPool(b, ctx)
	defer pool.Close()

	repo := repository.NewPostgres(pool)

	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	deps := usecase.Dependencies{
		Repo: repo,
		Node: node,
	}

	uc := usecase.NewURLUseCase(deps)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := uc.Shorten(ctx, exampleURL)
		if err != nil {
			b.Fatalf("error in new shorten: %v", err)
		}
	}
}

// Вспомогательная функция для быстрой инициализации pgxpool
func setupTestPool(b *testing.B, ctx context.Context) *pgxpool.Pool {
	b.Helper() 

	// Строка подключения к вашей тестовой БД
	connStr := "postgres://postgres:1234@localhost:5432/test_urls?sslmode=disable&pool_min_conns=1&pool_max_conns=10"

	// Создаем пул соединений pgxpool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		b.Fatalf("failed to create pgxpool: %v", err)
	}

	// Проверяем доступность базы данных
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		b.Fatalf("failed to ping pgxpool: %v", err)
	}

	migration := `
		CREATE TABLE IF NOT EXISTS urls (
			id BIGSERIAL PRIMARY KEY,
			long_url TEXT NOT NULL,
			short_code VARCHAR(11) UNIQUE,
			clicks_count BIGINT DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);			
	`
	_, err = pool.Exec(ctx, migration)
	if err != nil {
		pool.Close()
		b.Fatalf("failed to magrate test table: %v", err)
	}

	// Очищаем таблицу перед тестом, чтобы бенчмарки были в равных условиях
	_, err = pool.Exec(ctx, "TRUNCATE TABLE urls RESTART IDENTITY CASCADE;")
	if err != nil {
		pool.Close()
		b.Fatalf("failed to truncate test table: %v", err)
	}

	return pool
}
