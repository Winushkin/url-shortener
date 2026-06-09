package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upChangeShortCodeLength, downChangeShortCodeLength)
}

func upChangeShortCodeLength(ctx context.Context, tx *sql.Tx) error {
	query := `
		ALTER TABLE IF EXISTS urls
    	ALTER COLUMN short_code TYPE VARCHAR(11);

	`
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("ошибка создания таблицы urls или индекса: %w", err)
	}
	return nil
}

func downChangeShortCodeLength(ctx context.Context, tx *sql.Tx) error {
	query := `
		ALTER TABLE IF EXISTS urls
		ALTER COLUMN short_code TYPE VARCHAR(10);
	`
	if _, err := tx.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("ошибка удаления таблицы urls или индекса: %w", err)
	}
	return nil
}
