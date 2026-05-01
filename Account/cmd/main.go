package main

import (
	"context"
	"log"
	_ "study/Account/docs"
	"study/Account/internal/app"
	"study/Account/internal/config"
	"study/Account/internal/logger"
	_ "study/Account/internal/migrations"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

// @title Account Service
// @version 1.0
// @description developer desc
// @host localhost:8080
// @BasePath /
// @schemes http
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

// PingExample godoc
// @Summary Хэлсчек сервиса
// @Description Возсращает pong
// @Tags health
// @Success 200 {string} string "pong"
// @Router /ping [get]
func PingExample(c *gin.Context) {
	c.String(200, "pong")

}
