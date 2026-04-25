package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName string `env:"SERVICE_NAME" required:"true" default:"acount-service"`
	AppEnv      string `env:"APP_ENV" required:"true" default:"development"`
	LogLvl      string `env:"LOG_LEVEL" required:"true" default:"info"`
	Host        string `env:"HTTP_HOST" required:"true" default:"localhost"`
	Port        int    `env:"HTTP_PORT" required:"true" default:"9000"`
	DB          string `env:"DB_DSN" required:"true"`
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
