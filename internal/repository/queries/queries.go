package queries

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

// InsertURL возвращает запрос на вставку url и short кода
func InsertURL(url, code string) sq.InsertBuilder {
	data := map[string]any{
		longURL:   url,
		shortCode: code,
	}
	return psql.Insert(table).SetMap(data)
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
func IncrementClicks(shortURL string) sq.UpdateBuilder {
	return psql.Update(table).
		Set(clicksCount, sq.Expr(fmt.Sprintf("%s + 1", clicksCount))).
		Where(sq.Eq{shortCode: shortURL})
}


