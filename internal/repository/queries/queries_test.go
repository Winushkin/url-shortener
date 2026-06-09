package queries_test

import (
	"shortener/internal/repository/queries"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	longExample  = "https://example.com"
	shortExample = "abcde123"
	exampleID    = 42
)

func TestInsertURL(t *testing.T) {
	builder := queries.InsertURL(longExample, shortExample)
	sqlStr, _, err := builder.ToSql()

	expectedSQL := "INSERT INTO urls (long_url,short_code) VALUES ($1,$2)"

	require.NoError(t, err)
	assert.Equal(t, expectedSQL, sqlStr)
}

func TestSelectLongURLByShort(t *testing.T) {
	builder := queries.SelectLongURLByShort(shortExample)
	sqlStr, _, err := builder.ToSql()

	expectedSQL := "SELECT id, long_url, short_code, clicks_count, created_at FROM urls WHERE short_code = $1"

	require.NoError(t, err)
	assert.Equal(t, expectedSQL, sqlStr)
}

func TestIncrementClicks(t *testing.T) {
	builder := queries.IncrementClicks(shortExample)
	sqlStr, _, err := builder.ToSql()

	expectedSQL := "UPDATE urls SET clicks_count = clicks_count + 1 WHERE short_code = $1"

	require.NoError(t, err)
	assert.Equal(t, expectedSQL, sqlStr)
}
