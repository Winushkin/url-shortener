package postgres

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

const (
	// Константы с именами таблиц, используемых в запросах
	table       = "urls"
	ID          = "id"
	longURL     = "long_url"
	shortCode   = "short_code"
	clicksCount = "clicks_count"
	createdAt   = "created_at"
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
func insertShortURL(shortURL string, id uint64) sq.UpdateBuilder {
	return psql.Update(table).
		Set(shortCode, shortURL).
		Where(sq.Eq{ID: id})
}

// SelectLongURLByShort возвращает запрос для селекта длинной ссылки по короткой
func SelectLongURLByShort(shortURL string) sq.SelectBuilder {
	return psql.Select(
		ID,
		longURL,
		shortCode,
		clicksCount,
		createdAt,
	).From(table).
		Where(sq.Eq{shortCode: shortURL})
}

// incrementClicks возвращает запрос для обновления счетчика перехода по ссылке
func incrementClicks(shortURL string) sq.UpdateBuilder {
	return psql.Update(table).
		Set(clicksCount, sq.Expr(fmt.Sprintf("%s + 1", clicksCount))).
		Where(sq.Eq{shortCode: shortURL})
}
