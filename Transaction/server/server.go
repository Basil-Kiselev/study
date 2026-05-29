package server

import (
	"context"
	"study/Transaction/internal/model"
	transactionpb "study/contracts/transaction"

	"github.com/rs/zerolog"
)

type Server struct {
	transactionpb.UnimplementedTransactionServiceServer
	logger             *zerolog.Logger
	transactionService TransactionService
}

func NewServer(transactionService)

type TransactionService interface {
	GetTransactionsWithDetails(context.Context, model.GetTransactionsParams) ([]model.TransactionDetails, error)
	Deposit(context.Context, uint64, float32) (model.TransactionDetails, error)
	Withdraw(context.Context, uint64, float32) (model.TransactionDetails, error)
	Transfer(context.Context, uint64, float64, uint64) (model.TransactionDetails, error)
}
