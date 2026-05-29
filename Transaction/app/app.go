package app

import (
	"study/Transaction/internal/account"
	"study/Transaction/internal/config"
	"study/Transaction/internal/repository"
	"study/Transaction/internal/service"

	"github.com/rs/zerolog"
)

type App struct {
	cfg                *config.Config
	logger             *zerolog.Logger
	transactionRepo    *repository.Repository
	transactionService *service.TransactionService
	accountService     *account.Service
}
