package main

import (
	_ "account/docs"
	"account/internal/config"
	"account/internal/logger"
	"account/repository"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Account Service
// @version 1.0
// @description developer desc
// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Fail to load config: $v", err)
	}

	logger := logger.New()

	db, err := gorm.Open(postgres.Open(cfg.DB), &gorm.Config{})
	if err != nil {
		logger.Error().Msgf("Fail to connect to database: %v", err)
		return
	}
	logger.Info().Msg("Database connected")

	repo := repository.NewRepository(db, &logger)
	_ = repo

	router := gin.Default()

	router.GET("/ping", PingExample)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
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
