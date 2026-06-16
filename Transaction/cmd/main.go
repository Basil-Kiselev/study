package main

import (
	"context"
	"log"
	"study/Transaction/internal/app"
	"study/Transaction/internal/config"
	"study/Transaction/internal/logger"
	_ "study/Transaction/internal/migrations"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Fail to load config: %v", err)
	}

	logger := logger.New()

	application := app.New(&logger, cfg)

	if err := application.Run(ctx); err != nil {
		logger.Fatal().Err(err).Msg("error")
	}
}
