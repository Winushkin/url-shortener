// Package usecase содержит реализацию бизнес логики
package usecase

import (
	"context"
	"errors"
	"fmt"
	"shortener/internal/repository/postgres"
	"shortener/pkg/base62"

	"github.com/bwmarrin/snowflake"
)

type URLUseCase interface {
	// Shorten - обрабатывает запрос укращения ссылки
	Shorten(ctx context.Context, longURL string) (string, error)

	// GetLongURL - обрабатывает запрос получения оригинальной ссылки по коду
	GetLongURL(ctx context.Context, shortCode string) (string, error)
}

// urlUseCase реализует бизнес-логику для работы с URL
type urlUseCase struct {
	repo postgres.Repository
	node *snowflake.Node
}

// NewURLUseCase — конструктор слоя бизнес-логики
func NewURLUseCase(repo postgres.Repository, node *snowflake.Node) URLUseCase {
	return &urlUseCase{
		repo: repo,
		node: node,
	}
}

func (uc *urlUseCase) Shorten(ctx context.Context, longURL string) (string, error) {
	id := uc.node.Generate().Int64()
	code := base62.Encode(id)

	err := uc.repo.InsertURL(ctx, longURL, code)
	if err != nil {
		return "", fmt.Errorf("usecase: failed to create url record: %w", err)
	}

	return code, nil
}

func (uc *urlUseCase) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	record, err := uc.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to get url by code: %w", err)
	}

	if record == nil {
		return "", errors.New("url not found")
	}

	// TODO Атомарно увеличиваем счетчик кликов (запускаем в фоне, чтобы не тормозить редирект для юзера). В продакшене лучше делать это через очередь или Redis
	err = uc.repo.IncrementClicks(ctx, shortCode)

	if err != nil {
		return "", fmt.Errorf("incrementClicks: %w", err)
	}

	return record.LongURL, nil
}
