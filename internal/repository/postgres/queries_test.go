package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	longExample  = "https://example.com"
	shortExample = "abcde123"
	exampleID    = 42
)

func TestInsertURL(t *testing.T) {
	builder := insertURL(longExample)
	sqlStr, _, err := builder.ToSql()

	expectedSQL := "INSERT INTO urls (long_url) VALUES ($1) RETURNING id"

	assert.NoError(t, err)
	assert.Equal(t, expectedSQL, sqlStr)
}

func TestInsertShortURL(t *testing.T) {
	builder := insertShortURL(shortExample, exampleID)
	sqlStr, _, err := builder.ToSql()

	expectedSQL := "UPDATE urls SET short_code = $1 WHERE id = $2"

	assert.NoError(t, err)
	assert.Equal(t, expectedSQL, sqlStr)
}

func TestSelectLongURLByShort(t *testing.T) {
	builder := SelectLongURLByShort(shortExample)
	sqlStr, _, err := builder.ToSql()

	expectedSQL := "SELECT id, long_url, short_code, clicks_count, created_at FROM urls WHERE short_code = $1"

	assert.NoError(t, err)
	assert.Equal(t, expectedSQL, sqlStr)
}

func TestIncrementClicks(t *testing.T) {
	builder := incrementClicks(shortExample)
	sqlStr, _, err := builder.ToSql()

	expectedSQL := "UPDATE urls SET clicks_count = clicks_count + 1 WHERE short_code = $1"

	assert.NoError(t, err)
	assert.Equal(t, expectedSQL, sqlStr)
}
