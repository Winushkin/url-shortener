package benchmarks

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	sq "github.com/Masterminds/squirrel"
)

type legacyPostgres struct {
	pool *pgxpool.Pool
}

func NewLegacyPostgres(pool *pgxpool.Pool) *legacyPostgres {
	return &legacyPostgres{pool: pool}
}

func (p *legacyPostgres) InsertURL(ctx context.Context, longURL string) (int64, error) {
	sql, args, err := insertURL(longURL).ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to parse query: %w", err)
	}
	var id int64
	err = p.pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("queryRow: %w", err)
	}

	return id, nil
}

func (p *legacyPostgres) InsertShortCode(ctx context.Context, id int64, shortCode string) error {
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

const (
	// Константы с именами таблиц, используемых в запросах
	table       = "urls"
	ID          = "id"
	longURL     = "long_url"
	shortCode   = "short_code"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// insertUrl возвращает запрос для вставки новой ссылки
func insertURL(URL string) sq.InsertBuilder {
	data := map[string]any{
		longURL: URL,
	}
	return psql.Insert(table).
		SetMap(data).
		Suffix("RETURNING id")
}

// insertShortURL возвращает запрос для вставки сокращенной ссылки
func insertShortURL(shortURL string, id int64) sq.UpdateBuilder {
	return psql.Update(table).
		Set(shortCode, shortURL).
		Where(sq.Eq{ID: id})
}
