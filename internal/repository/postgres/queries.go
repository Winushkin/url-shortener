// Package queries содержит функции для сборки запросов к базе данных
package postgres

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

const (
	// Константы с именами таблиц, используемых в запросах
	table        = "urls"
	ID           = "id"
	long_url     = "long_url"
	short_code   = "short_code"
	clicks_count = "clicks_count"
	created_at   = "created_at"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// insertUrl возвращает запрос для вставки новой ссылки
func insertURL(longURL string) sq.InsertBuilder {
	data := map[string]any{
		long_url: longURL,
	}
	return psql.Insert(table).SetMap(data)
}

// insertShortURL возвращает запрос для вставки сокращенной ссылки
func insertShortURL(shortURL string, id uint64) sq.UpdateBuilder {
	return psql.Update(table).
		Set(clicks_count, shortURL).
		Where(sq.Eq{ID: id})
}

// insertShortURL возвращает запрос для вставки сокращенной ссылки
func SelectLongURLByShort(shortURL string) sq.SelectBuilder {
	return psql.Select(table).
		Where(sq.Eq{short_code: shortURL})
}

// incrementClicks возвращает запрос для обновления счетчика перехода по ссылке
func incrementClicks(shortCode string) sq.UpdateBuilder {
	return psql.Update(table).
		Set(clicks_count, sq.Expr(fmt.Sprintf("%s + 1", clicks_count))).
		Where(sq.Eq{short_code: shortCode})
}
