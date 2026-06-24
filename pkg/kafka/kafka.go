// Package kafka содержит создание клиента kafka
package kafka

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Config struct {
	Host    string `env:"KAFKA_HOST"`
	Port    string `env:"KAFKA_PORT"`
	GroupID string `env:"GROUP_ID"`
	Topics  map[string]string
	Options map[string]any
}

const (
	recordRetries = 5
	pingTimeout   = 5 * time.Second
)

// NewClient инициализирует и проверяет подключение к Kafka
func NewClient(ctx context.Context, cfg Config) (*kgo.Client, error) {
	// Собираем опции конфигурации клиента
	opts := []kgo.Opt{
		kgo.SeedBrokers(net.JoinHostPort(cfg.Host, cfg.Port)),

		// Настройки Продюсера для гарантии At-Least-Once
		kgo.RequiredAcks(kgo.AllISRAcks()), // Ждем подтверждения от всех реплик
		kgo.RecordRetries(recordRetries),   // Количество попыток переотправки при сетевом сбое

		// Настройки Консумера
		kgo.ConsumerGroup(cfg.GroupID),
		kgo.ConsumeTopics("url.clicks"),
		kgo.DisableAutoCommit(), // Отключаем автокоммит (будем коммитить вручную после записи в БД)
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	// Делаем обязательный Ping (проверку связи с брокером) перед стартом приложения
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := client.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("failed to ping kafka broker: %w", err)
	}

	return client, nil
}
