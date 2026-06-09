// Package repository_test содержит тесты для слоя repository
package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"shortener/internal/entities"
	"shortener/internal/repository"
	_ "shortener/migrations"

	"github.com/Winushkin/go-toolkit/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	// testDB - глобальная переменная для хранения подключения к тестовой базе данных, которая будет использоваться во всех тестах.
	testDB *pgxpool.Pool

	// repo - глобальная переменная для хранения экземпляра BeerPostgres, который будет использоваться в тестах для взаимодействия с таблицей пива.
	repo repository.Repository
)

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

// TestMain запускает тестовую среду с помощью testcontainers, выполняет миграции и очищает ресурсы после тестов.
func TestMain(m *testing.M) {
	ctx := context.Background()
	ctx, err := logger.NewLoggerContext(ctx, true)
	if err != nil {
		panic(fmt.Errorf("failed to create logger context: %w", err))
	}

	dbContainer, err := pgContainer.Run(ctx,
		"postgres:17",
		pgContainer.WithDatabase("test_db"),
		pgContainer.WithUsername("test_user"),
		pgContainer.WithPassword("test_pswd"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("Failed to start container: %s", err)
	}

	args := []string{
		"sslmode=disable",
		"pool_min_conns=1",
		"pool_min_conns=2",
	}

	connStr, err := dbContainer.ConnectionString(ctx, args...)
	if err != nil {
		log.Fatalf("Failed create conn string: %s", err)
	}

	testDB, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s", err)
	}

	migrate(ctx, testDB)

	repo = repository.NewPostgres(testDB)

	code := m.Run()

	testDB.Close()
	if err = dbContainer.Terminate(ctx); err != nil {
		log.Printf("Failed to terminate container: %s", err)
	}

	os.Exit(code)
}

// cleanDB выполняет очистку указанной таблицы в базе данных, удаляя все записи и сбрасывая идентификаторы.
func cleanDB(t *testing.T, ctx context.Context) {
	_, err := testDB.Exec(ctx, "TRUNCATE TABLE urls RESTART IDENTITY CASCADE")
	if err != nil {
		t.Errorf("failed to clean db: %v", err)
	}
}

func TestRepo_InsertURL(t *testing.T) {
	ctx := t.Context()
	ctx, err := logger.NewLoggerContext(ctx, true)
	if err != nil {
		panic(fmt.Errorf("failed to create logger context: %w", err))
	}

	type args struct {
		url       string
		shortCode string
	}

	tests := []struct {
		name    string
		prepare func(t *testing.T)
		args    args
		wantErr bool
	}{
		{
			name: "Успешное сохранение новой ссылки",
			prepare: func(t *testing.T) {
				cleanDB(t, ctx)
			},
			args: args{
				url:       "https://google.com",
				shortCode: "ggl1",
			},
			wantErr: false,
		},
		{
			name: "Ошибка: дубликат short_code",
			prepare: func(t *testing.T) {
				cleanDB(t, ctx)
				_, err := testDB.Exec(ctx, "INSERT INTO urls (long_url, short_code) VALUES ($1, $2)", "https://yandex.ru", "duplicate")
				if err != nil {
					t.Fatal(err)
				}
			},
			args: args{
				url:       "https://apple.com",
				shortCode: "duplicate",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(t)

			err := repo.InsertURL(ctx, tt.args.url, tt.args.shortCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				var actualURL string
				err = testDB.QueryRow(ctx, "SELECT long_url FROM urls WHERE short_code = $1", tt.args.shortCode).Scan(&actualURL)
				assert.NoError(t, err)
				assert.Equal(t, tt.args.url, actualURL)
			}
		})
	}
}

func TestURLRepo_GetByShortCode(t *testing.T) {
	ctx := t.Context()
	ctx, err := logger.NewLoggerContext(ctx, true)
	if err != nil {
		panic(fmt.Errorf("failed to create logger context: %w", err))
	}

	tests := []struct {
		name      string
		prepare   func(t *testing.T)
		shortCode string
		want      *entities.URL
		wantErr   bool
		errType   error // если у вас в домене объявлена ошибка вроде ErrURLNotFound
	}{
		{
			name: "Успешный поиск существующего URL",
			prepare: func(t *testing.T) {
				cleanDB(t, ctx)
				_, err := testDB.Exec(ctx, "INSERT INTO urls (long_url, short_code, clicks_count) VALUES ($1, $2, $3)", "https://go.dev", "go123", 5)
				if err != nil {
					t.Fatal(err)
				}
			},
			shortCode: "go123",
			want: &entities.URL{
				LongURL:      "https://go.dev",
				ShortCode:    "go123",
				ClicksAmount: 5,
			},
			wantErr: false,
		},
		{
			name: "Ссылка не найдена (возвращаем nil или доменную ошибку)",
			prepare: func(t *testing.T) {
				cleanDB(t, ctx)
			},
			shortCode: "not_exist",
			want:      nil,
			wantErr:   true, // либо false, если ваш метод при отсутствии строк возвращает (nil, nil)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(t)

			res, err := repo.GetByShortCode(ctx, tt.shortCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, tt.want.LongURL, res.LongURL)
				assert.Equal(t, tt.want.ShortCode, res.ShortCode)
				assert.Equal(t, tt.want.ClicksAmount, res.ClicksAmount)
			}
		})
	}
}


func TestURLRepo_IncrementClicks(t *testing.T) {
	ctx := t.Context()
	ctx, err := logger.NewLoggerContext(ctx, true)
	if err != nil {
		panic(fmt.Errorf("failed to create logger context: %w", err))
	}

	tests := []struct {
		name       string
		prepare    func(t *testing.T)
		shortCode  string
		wantClicks int64
		wantErr    bool
	}{
		{
			name: "Инкремент счетчика с нуля",
			prepare: func(t *testing.T) {
				cleanDB(t, ctx)
				_, err := testDB.Exec(ctx, "INSERT INTO urls (long_url, short_code) VALUES ($1, $2)", "https://github.com", "abcdefghigk")
				if err != nil {
					t.Fatal(err)
				}
			},
			shortCode:  "abcdefghigk",
			wantClicks: 1,
			wantErr:    false,
		},
		{
			name: "Инкремент уже существующего счетчика",
			prepare: func(t *testing.T) {
				cleanDB(t, ctx)
				_, err := testDB.Exec(ctx, "INSERT INTO urls (long_url, short_code, clicks_count) VALUES ($1, $2, $3)", "https://habr.com", "habr7", 42)
				if err != nil {
					t.Fatal(err)
				}
			},
			shortCode:  "habr7",
			wantClicks: 43,
			wantErr:    false,
		},
		{
			name: "Попытка инкремента несуществующего кода (ошибки нет, просто 0 строк изменено)",
			prepare: func(t *testing.T) {
				cleanDB(t, ctx)
			},
			shortCode: "ghost",
			wantErr:   false, 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(t)

			err := repo.IncrementClicks(ctx, tt.shortCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tt.shortCode != "ghost" {
					var actualClicks int64
					err = testDB.QueryRow(ctx, "SELECT clicks_count FROM urls WHERE short_code = $1", tt.shortCode).Scan(&actualClicks)
					assert.NoError(t, err)
					assert.Equal(t, tt.wantClicks, actualClicks)
				}
			}
		})
	}
}

