package actual

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *postgres {
	return &postgres{pool: pool}
}

func (p *postgres) InsertURL(ctx context.Context, url, shortCode string) error {
	sql, args, err := InsertURL(url, shortCode).ToSql()
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	_, err = p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

const (
	table     = "urls"
	longURL   = "long_url"
	shortCode = "short_code"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func InsertURL(url, code string) sq.InsertBuilder {
	data := map[string]any{
		longURL:   url,
		shortCode: code,
	}
	return psql.Insert(table).SetMap(data)
}
