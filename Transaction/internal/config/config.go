package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName           string   `env:"SERVICE_NAME" required:"true" default:"transaction-service"`
	AppEnv                string   `env:"APP_ENV" required:"true" default:"development"`
	LogLvl                string   `env:"LOG_LEVEL" required:"true" default:"info"`
	Host                  string   `env:"GRPC_HOST" required:"true" default:"localhost"`
	Port                  int      `env:"GRPC_PORT" required:"true" default:"50054"`
	DB                    string   `env:"DB_DSN" required:"true"`
	AccountGrpcHost       string   `env:"ACCOUNT_GRPC_HOST" required:"true" default:"localhost:50051"`
	KafkaBrokers          []string `env:"KAFKA_BROKER_HOST" envSeparator:"," json:"kafka_brokers" required:"true"`
	KafkaGroupID          string   `env:"KAFKA_COMSUMER_GROUP" json:"kafka_consumer_group" required:"true"`
	KafkaTransactionTopic string   `env:"KAFKA_TRANSACTION_TOPIC" json:"kafka_transaction_topic" required:"true"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	var conf Config

	err := env.Parse(&conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}
