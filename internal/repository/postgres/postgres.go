// Package postgres является слоем взаимодействия с PostgreSQL
package postgres

import (
	"context"
	"errors"
	"fmt"
	"shortener/internal/entities"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	// Сохраняет длинный URL и возвращает сгенерированный ID
	InsertURL(ctx context.Context, longURL string) (uint64, error)

	// Обновляет короткий код для записи по её ID (после генерации Base62)
	UpdateShortCode(ctx context.Context, id uint64, shortCode string) error

	// Находит полную запись по короткому коду (используется при редиректе)
	GetByShortCode(ctx context.Context, shortCode string) (*entities.URL, error)

	// Атомарно увеличивает счетчик кликов на 1
	IncrementClicks(ctx context.Context, shortCode string) error
}

type postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) Repository {
	return &postgres{pool: pool}
}

func (p *postgres) InsertURL(ctx context.Context, longURL string) (uint64, error) {
	sql, args, err := insertURL(longURL).ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to parse query: %w", err)
	}
	var id uint64
	err = p.pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("queryRow: %w", err)
	}

	return id, nil
}

func (p *postgres) UpdateShortCode(ctx context.Context, id uint64, shortCode string) error {
	sql, args, err := insertShortURL(shortCode, id).ToSql()
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	result, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	if result.RowsAffected() != 1 {
		return errors.New("more than 1 row affected")
	}

	return nil
}

func (p *postgres) GetByShortCode(ctx context.Context, shortCode string) (*entities.URL, error) {
	sql, args, err := SelectLongURLByShort(shortCode).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}
	var url entities.URL
	err = p.pool.QueryRow(ctx, sql, args...).Scan(
		&url.ID,
		&url.LongURL,
		&url.ShortCode,
		&url.ClicksAmount,
		&url.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("queryRow: %w", err)
	}

	return &url, nil
}

func (p *postgres) IncrementClicks(ctx context.Context, shortCode string) error {
	sql, args, err := incrementClicks(shortCode).ToSql()
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	result, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	if result.RowsAffected() != 1 {
		return errors.New("more than 1 row affected")
	}

	return nil
}
