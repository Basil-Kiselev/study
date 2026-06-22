package kafka

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	logger *zerolog.Logger
}

func (c *Consumer) GetConfig() kafka.ReaderConfig {
	return c.reader.Config()
}

type ConsumerMessageHandler func(ctx context.Context, msg kafka.Message) error

func NewConsumer(cfg ConsumerConfig, logger *zerolog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           cfg.Brokers,
		GroupID:           cfg.GroupID,
		Topic:             cfg.Topics[0],
		MinBytes:          cfg.MinBytes,
		MaxBytes:          cfg.MaxBytes,
		MaxWait:           cfg.MaxWait,
		ReadBatchTimeout:  cfg.ReadBatchTimeout,
		HeartbeatInterval: cfg.HeartbeatInterval,
		CommitInterval:    cfg.CommitInterval,
	})

	return &Consumer{
		reader: reader,
		logger: logger,
	}
}

func (c *Consumer) Start(ctx context.Context, handler ConsumerMessageHandler) error {
	c.logger.Info().Str("topic", c.reader.Config().Topic).Msg("starting kafka consumer")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("stoping kafka consumer")
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error().Err(err).Msg("fail to read msg from kafka")
				time.Sleep(time.Second)
				continue
			}

			if err := handler(ctx, msg); err != nil {
				c.logger.Error().Err(err).
					Str("topic", msg.Topic).
					Int("partition", msg.Partition).
					Int64("offset", msg.Offset).
					Msg("failed to process message")
			} else {
				c.logger.Debug().
					Str("topic", msg.Topic).
					Int("partition", msg.Partition).
					Int64("offset", msg.Offset).
					Msg("message processed successfully")
			}
		}
	}
}

func (c *Consumer) Close() error {
	if c.reader != nil {
		c.reader.Close()
	}
	return nil
}
