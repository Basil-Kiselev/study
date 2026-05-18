package main

import (
	"context"
	"log"
	"study/Auth/internal/app"
	"study/Auth/internal/config"
	"study/Auth/internal/logger"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Fail to load config: %v", err)
	}

	logger := logger.New()

	application := app.New(cfg, &logger)

	if err := application.Run(ctx); err != nil {
		logger.Fatal().Err(err).Msg("error")
	}
}
