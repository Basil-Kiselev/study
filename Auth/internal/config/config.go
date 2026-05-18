package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName           string `env:"SERVICE_NAME" required:"true" default:"auth-service"`
	AppEnv                string `env:"APP_ENV" required:"true" default:"development"`
	LogLvl                string `env:"LOG_LEVEL" required:"true" default:"info"`
	Host                  string `env:"GRPC_HOST" required:"true" default:"localhost"`
	Port                  int    `env:"GRPC_PORT" required:"true" default:"50052"`
	DB                    string `env:"DB_DSN" required:"true"`
	JwtSecret             string `env:"JWT_SECRET" json:"jwt_secret" required:"true"`
	AccessTokenTTLMinutes int    `env:"ACCESS_TOKEN_TTL_MINUTES" json:"access_token_ttl_minutes" required:"true" default:"60"`
	RefreshTokenTTLDays   int    `env:"REFRESH_TOKEN_TTL_DAYS" json:"refresh_token_ttl_days" required:"true" default:"30"`
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
