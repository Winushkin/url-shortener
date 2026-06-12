// Package usecase содержит реализацию бизнес логики
package usecase

import (
	"context"
	"errors"
	"fmt"
	"shortener/internal/broker"
	"shortener/internal/repository"
	"shortener/pkg/base62"
	"time"

	"github.com/Winushkin/go-toolkit/logger"
	"github.com/bwmarrin/snowflake"
	"github.com/redis/go-redis/v9"
)

const (
	day        = 24 * time.Hour
	ctxTimeout = 5 * time.Second
)

type Dependencies struct {
	Repo      repository.Repository
	Node      *snowflake.Node
	Rdb       *redis.Client
	Publisher broker.ClickPublisher
}

type URLUseCase interface {
	// Shorten - обрабатывает запрос укращения ссылки
	Shorten(ctx context.Context, longURL string) (string, error)

	// GetLongURL - обрабатывает запрос получения оригинальной ссылки по коду
	GetLongURL(ctx context.Context, shortCode string) (string, error)
}

// urlUseCase реализует бизнес-логику для работы с URL
type urlUseCase struct {
	repo      repository.Repository
	node      *snowflake.Node
	rdb       *redis.Client
	publisher broker.ClickPublisher
}

// NewURLUseCase — конструктор слоя бизнес-логики
func NewURLUseCase(deps Dependencies) URLUseCase {
	uc := &urlUseCase{}
	if deps.Repo != nil {
		uc.repo = deps.Repo
	}
	if deps.Node != nil {
		uc.node = deps.Node
	}
	if deps.Rdb != nil {
		uc.rdb = deps.Rdb
	}
	if deps.Publisher != nil {
		uc.publisher = deps.Publisher
	}
	return uc
}

func (uc *urlUseCase) Shorten(ctx context.Context, longURL string) (string, error) {
	id := uc.node.Generate().Int64()
	code := base62.Encode(id)

	err := uc.repo.InsertURL(ctx, longURL, code)
	if err != nil {
		return "", fmt.Errorf("failed to create url record: %w", err)
	}

	if uc.rdb != nil {
		err := uc.rdb.Set(ctx, code, longURL, day).Err()
		if err != nil {
			log, ok := logger.GetLoggerFromCtx(ctx)
			if !ok {
				return "", errors.New("failed to get logger from ctx in rdb error")
			}
			log.Error(ctx, err, "failed to cache url")
		}
	}

	return code, nil
}

//nolint:gosec // G118: Изолированный контекст используется намеренно
func (uc *urlUseCase) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		return "", errors.New("failed to get logger from ctx")
	}
	if uc.rdb != nil {
		longURL, err := uc.rdb.Get(ctx, shortCode).Result()
		if err == nil {
			go func() {
				pubCtx := logger.NewContextWithLogger(context.Background(), log)
				event := broker.ClickEvent{
					URLCode:   shortCode,
					ClickedAt: time.Now().Unix(),
				}
				err = uc.publisher.PublishClick(pubCtx, event)
				if err != nil {
					log.Error(ctx, err, "failed to count Click")
				}
			}()
			return longURL, nil
		}

		if !errors.Is(err, redis.Nil) {
			log.Error(ctx, err, "failed to get cache")
		}
	}
	record, err := uc.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to get url by code: %w", err)
	}

	if record == nil {
		return "", errors.New("url not found")
	}
	if uc.rdb != nil {
		err = uc.rdb.Set(ctx, shortCode, record.LongURL, day).Err()
		if err != nil {
			log, ok := logger.GetLoggerFromCtx(ctx)
			if !ok {
				return "", errors.New("failed to get logger from ctx in rdb error")
			}
			log.Error(ctx, err, "failed to cache url")
		}
	}

	go func() {
		pubCtx := logger.NewContextWithLogger(context.Background(), log)
		event := broker.ClickEvent{
			URLCode:   shortCode,
			ClickedAt: time.Now().Unix(),
		}
		err = uc.publisher.PublishClick(pubCtx, event)
		if err != nil {
			log.Error(ctx, err, "failed to count Click")
		}
	}()

	return record.LongURL, nil
}
