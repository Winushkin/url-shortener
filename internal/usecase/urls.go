// Package usecase содержит реализацию бизнес логики
package usecase

import (
	"context"
	"errors"
	"fmt"
	"shortener/internal/repository/postgres"
	"shortener/pkg/base62"
)

type URLUseCase interface {
	Shorten(ctx context.Context, longURL string) (string, error)
	GetLongURL(ctx context.Context, shortCode string) (string, error)
}

// urlUseCase реализует бизнес-логику для работы с URL
type urlUseCase struct {
	repo postgres.Repository
}

// NewURLUseCase — конструктор слоя бизнес-логики
func NewURLUseCase(repo postgres.Repository) URLUseCase {
	return &urlUseCase{repo: repo}
}

// TODO Попробовать генерацию числа с помощью Sonyflake / Snowflake
func (uc *urlUseCase) Shorten(ctx context.Context, longURL string) (string, error) {
	id, err := uc.repo.InsertURL(ctx, longURL)
	if err != nil {
		return "", fmt.Errorf("usecase: failed to create url record: %w", err)
	}

	shortCode := base62.Encode(id)

	err = uc.repo.InsertShortCode(ctx, id, shortCode)
	if err != nil {
		return "", fmt.Errorf("usecase: failed to update short code: %w", err)
	}

	return shortCode, nil
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
