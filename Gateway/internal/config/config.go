package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName     string `env:"SERVICE_NAME" envDefault:"gateway"`
	AppEnv          string `env:"APP_ENV" envDefault:"development"`
	Host            string `env:"GRPC_HOST" envDefault:"localhost"`
	Port            string `env:"GRPC_PORT" envDefault:"5003"`
	LogLvl          string `env:"LOG_LEVEL" envDefault:"info"`
	AccountGrpcHost string `env:"ACCOUNT_GRPC_HOST" envDefault:"localhost:50051"`
	AuthGrpcHost    string `env:"AUTH_GRPC_HOST" envDefault:"localhost:50052"`
	JWTSecret       string `env:"JWT_SECRET" envDefault:"secret"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("fail to load .env")
	}
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
