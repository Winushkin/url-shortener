// Package broker содержит операции брокеров сообщений
package broker

import (
	"context"
	"encoding/json"
	"shortener/internal/repository"

	"github.com/Winushkin/go-toolkit/logger"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type ClickConsumer struct {
	client *kgo.Client
	repo   repository.Repository
}

func NewConsumer(client *kgo.Client, repo repository.Repository) *ClickConsumer {
	client.AddConsumeTopics("url.clicks")
	return &ClickConsumer{
		client: client,
		repo:   repo,
	}
}

func (c *ClickConsumer) Start(ctx context.Context) {
	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		panic("logger not found in context")
	}
	log.Info(ctx, "click Consumer started")

	for {
		fetches := c.client.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, fe := range errs {
				log.Error(ctx, fe.Err, "Kafka Error", zap.String("Topic", fe.Topic), zap.Int32("Partition", fe.Partition))
			}
		}
		log.Debug(ctx, "Новый клик пришел")
		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()

			var event ClickEvent
			if err := json.Unmarshal(record.Value, &event); err != nil {
				log.Error(ctx, err, "failed to unmarshal record")
				continue
			}

			log.Debug(ctx, "Записываем новый клик")
			err := c.repo.IncrementClicks(ctx, event.URLCode)
			if err != nil {
				log.Error(ctx, err, "failed to save click to db")
				// TODO: DLQ error handling
			}
			log.Debug(ctx, "Записали новый клик")
		}
	}
}
