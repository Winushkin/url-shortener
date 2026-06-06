// Package migrations содержит goose миграции для PostgreSQL 
package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitShortenerTable, downInitShortenerTable)
}

func upInitShortenerTable(ctx context.Context, tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			id BIGSERIAL PRIMARY KEY,
			long_url TEXT NOT NULL,
			short_code VARCHAR(10) UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			clicks_count BIGINT DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);			
	`
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("ошибка создания таблицы urls или индекса: %w", err)
	}
	
	return nil
}

func downInitShortenerTable(ctx context.Context, tx *sql.Tx) error {
	query := `
		DROP INDEX IF EXISTS idx_urls_short_code;
		DROP TABLE IF EXISTS urls
	`
		if _, err := tx.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("ошибка удаления таблицы urls или индекса: %w", err)
	}
	
	return nil
}
