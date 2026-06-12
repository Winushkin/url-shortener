package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Winushkin/go-toolkit/logger"
	"github.com/twmb/franz-go/pkg/kgo"
)

type ClickEvent struct {
	URLCode   string `json:"url_code"`
	ClickedAt int64  `json:"clicked_at"`
}

type ClickPublisher interface {
	PublishClick(ctx context.Context, event ClickEvent) error
}

type Publisher struct {
	client *kgo.Client
	topic  string
}

func NewPublisher(client *kgo.Client, topic string) *Publisher {
	return &Publisher{
		client: client,
		topic:  topic,
	}
}

func (p *Publisher) PublishClick(ctx context.Context, event ClickEvent) error {
	log, ok := logger.GetLoggerFromCtx(ctx)
	if !ok {
		panic("logger not found in context")
	}

	log.Debug(ctx, "kafka is writing the click")
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal click event: %w", err)
	}

	record := &kgo.Record{
		Topic: p.topic,
		Value: payload,
		Key:   []byte(event.URLCode),
	}

	p.client.Produce(ctx, record, func(r *kgo.Record, err error) {
		if err != nil {
			log.Error(ctx, err, "failed to send message")
		}
	})

	return nil
}
