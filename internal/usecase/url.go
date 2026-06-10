// Package usecase содержит реализацию бизнес логики
package usecase

import (
	"context"
	"errors"
	"fmt"
	"shortener/internal/repository"
	"shortener/pkg/base62"
	"time"

	"github.com/Winushkin/go-toolkit/logger"
	"github.com/bwmarrin/snowflake"
	"github.com/redis/go-redis/v9"
)

const (
	day         = 24 * time.Hour
	ctxTimeout = 5 * time.Second
)

type Dependencies struct {
	Repo       repository.Repository
	Node       *snowflake.Node
	Rdb        *redis.Client
	ClicksChan chan string
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
	clickChan chan string
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
	if deps.ClicksChan != nil {
		uc.clickChan = deps.ClicksChan
		go uc.flushClicksWorker()
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

func (uc *urlUseCase) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	if uc.rdb != nil {
		longURL, err := uc.rdb.Get(ctx, shortCode).Result()
		if err == nil {
			err = uc.repo.IncrementClicks(ctx, shortCode)
			if err != nil {
				return "", fmt.Errorf("incrementClicks: %w", err)
			}

			return longURL, nil
		}

		if !errors.Is(err, redis.Nil) {
			log, ok := logger.GetLoggerFromCtx(ctx)
			if !ok {
				return "", errors.New("failed to get logger from ctx in rdb error")
			}
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

	uc.clickChan <- shortCode
	if err != nil {
		return "", fmt.Errorf("incrementClicks: %w", err)
	}

	return record.LongURL, nil
}

func (uc *urlUseCase) flushClicksWorker() {
	for code := range uc.clickChan{
		ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
		ctx, err := logger.NewLoggerContext(ctx, false)
		if err != nil {
			panic(fmt.Errorf("failed to create logger context: %w", err))
		}

		log, ok := logger.GetLoggerFromCtx(ctx)
		if !ok {
			panic("logger not found in context")
		}
		err = uc.repo.IncrementClicks(ctx, code)
		if err != nil{
			log.Error(ctx, err, "failed to increment clicks")
		}
		cancel()
	}
}
