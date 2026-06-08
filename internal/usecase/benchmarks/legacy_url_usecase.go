// Package benchmarks содержит бенчмарк тесты для usecase
package benchmarks

import (
	"context"
	"fmt"
	"shortener/pkg/base62"
)

type legacyUseCase struct {
	repo *legacyPostgres
}

func NewLegacyUseCase(repo *legacyPostgres) *legacyUseCase {
	return &legacyUseCase{repo: repo}
}

func (uc *legacyUseCase) Shorten(ctx context.Context, longURL string) (string, error) {
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
