package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers []string
	GroupID string
	Topics  []string
}

type ProducerConfig struct {
	Brokers      []string
	BatchSize    int
	BatchTimeout time.Duration
	RequiredAcks kafka.RequiredAcks
}

type ConsumerConfig struct {
	Brokers           []string
	GroupID           string
	Topics            []string
	MinBytes          int
	MaxBytes          int
	MaxWait           time.Duration
	ReadBatchTimeout  time.Duration
	HeartbeatInterval time.Duration
	CommitInterval    time.Duration
}

func DefaultProducerConfig(brokers []string) ProducerConfig {
	return ProducerConfig{
		Brokers:      brokers,
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}
}

func DefaultConsumerConfig(brokers []string, groupID string, topics []string) ConsumerConfig {
	return ConsumerConfig{
		Brokers:           brokers,
		GroupID:           groupID,
		Topics:            topics,
		MinBytes:          10e3,
		MaxBytes:          10e6,
		MaxWait:           1 * time.Second,
		ReadBatchTimeout:  1 * time.Second,
		HeartbeatInterval: 1 * time.Second,
		CommitInterval:    1 * time.Second,
	}
}
