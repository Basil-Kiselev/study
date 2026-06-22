package kafka

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

type Client interface {
	Publish(ctx context.Context, topic string, key string, data any) error
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Close() error
}

type MessageHandler func(ctx context.Context, topic string, key string, data []byte) error

type Kafka struct {
	producer *Producer
	brokers  []string
	groupID  string
	logger   *zerolog.Logger
}

func New(producer *Producer, brokers []string, groupID string, logger *zerolog.Logger) *Kafka {
	return &Kafka{
		producer: producer,
		brokers:  brokers,
		groupID:  groupID,
		logger:   logger,
	}
}

func (k *Kafka) Publish(ctx context.Context, topic string, key string, data any) error {
	return k.producer.SendJSONMessage(ctx, topic, key, data)
}

func (k *Kafka) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	uniqueGroupID := fmt.Sprintf("%s_%s", k.groupID, topic)
	cfg := DefaultConsumerConfig(k.brokers, uniqueGroupID, []string{topic})
	consumer := NewConsumer(cfg, k.logger)

	go func() {
		k.logger.Info().Str("topic", topic).Str("groupID", uniqueGroupID).Msg("subscribing to topic")
		messageHandler := ConsumerMessageHandler(func(ctx context.Context, msg kafka.Message) error {
			return handler(ctx, msg.Topic, string(msg.Key), msg.Value)
		})
		consumer.Start(ctx, messageHandler)
	}()

	return nil
}

func (k *Kafka) Close() error {
	if k.producer != nil {
		if err := k.producer.Close(); err != nil {
			return fmt.Errorf("fail to close producer: %w", err)
		}
	}
	return nil
}
