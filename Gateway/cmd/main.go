package main

import (
	"context"
	"os"
	"os/signal"
	"study/Gateway/internal/app"
	"study/Gateway/internal/config"
	"syscall"

	"github.com/rs/zerolog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Logger()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a := app.New(&logger, cfg)
	if err := a.Run(ctx); err != nil {
		logger.Error().Err(err).Msg("app stopped")
	}
}
