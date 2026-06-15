// Package actual содержить обновленую версию для сравнения с legacy кодом
package actual

import (
	"context"
	"fmt"
	"shortener/pkg/base62"

	"github.com/bwmarrin/snowflake"
)

type UseCase struct {
	repo *postgres
	node *snowflake.Node
}

func NewUseCase(repo *postgres, node *snowflake.Node) *UseCase {
	return &UseCase{
		repo: repo,
		node: node,
	}
}

func (uc *UseCase) Shorten(ctx context.Context, longURL string) (string, error) {
	id := uc.node.Generate().Int64()
	code := base62.Encode(id)

	err := uc.repo.InsertURL(ctx, longURL, code)
	if err != nil {
		return "", fmt.Errorf("failed to create url record: %w", err)
	}

	return code, nil
}
