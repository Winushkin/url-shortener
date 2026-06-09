// Package repository является слоем взаимодействия с PostgreSQL
package repository

import (
	"context"
	"errors"
	"fmt"
	"shortener/internal/entities"
	"shortener/internal/repository/queries"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	// Сохраняет длинный URL и возвращает сгенерированный ID
	InsertURL(ctx context.Context, url, shortCode string) error

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

func (p *postgres) InsertURL(ctx context.Context, url, shortCode string) error {
	sql, args, err := queries.InsertURL(url, shortCode).ToSql()
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	_, err = p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

func (p *postgres) GetByShortCode(ctx context.Context, shortCode string) (*entities.URL, error) {
	sql, args, err := queries.SelectLongURLByShort(shortCode).ToSql()
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
	sql, args, err := queries.IncrementClicks(shortCode).ToSql()
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	result, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	if result.RowsAffected() > 1 {
		return errors.New("more than 1 row affected")
	}

	return nil
}
