package service

import (
	"context"
	"fmt"
	"study/Transaction/internal/model"

	"github.com/rs/zerolog"
)

type TransactionService struct {
	repo   Repository
	logger *zerolog.Logger
}

func New(repo Repository, logger *zerolog.Logger) *TransactionService {
	return &TransactionService{
		repo:   repo,
		logger: logger,
	}
}

type Repository interface {
	GetTransactions(context.Context, model.GetTransactionsParams) ([]model.Transaction, error)
	GetTransactionDetails(context.Context, uint64) (model.TransactionDetails, error)
	Deposit(context.Context, model.DepositParams) (model.TransactionDetails, error)
	Transfer(context.Context, model.TransferParams) (model.TransactionDetails, error)
	Withdraw(context.Context, model.WithdrawParams) (model.TransactionDetails, error)
}

func (s *TransactionService) GetTransactionsWithDetails(ctx context.Context, params model.GetTransactionsParams) ([]model.TransactionDetails, error) {
	transactions, err := s.repo.GetTransactions(ctx, params)
	if err != nil {
		return []model.TransactionDetails{}, fmt.Errorf("fail to get transactions: %w", err)
	}

	details := make([]model.TransactionDetails, 0, len(transactions))
	for _, transaction := range transactions {
		detail, err := s.repo.GetTransactionDetails(ctx, transaction.ID)
		if err != nil {
			s.logger.Error().Err(err).Msg("fail to get transaction details")
			continue
		}
		details = append(details, detail)
	}
	return details, nil
}

func (s *TransactionService) Deposit(ctx context.Context, userId uint64, amount float64) (model.TransactionDetails, error) {
	params := model.DepositParams{
		UserID: userId,
		Amount: amount,
	}

	return s.repo.Deposit(ctx, params)
}

func (s *TransactionService) Withdraw(ctx context.Context, userId uint64, amount float64) (model.TransactionDetails, error) {
	params := model.WithdrawParams{
		AccountID: userId,
		Amount:    amount,
	}

	return s.repo.Withdraw(ctx, params)
}

func (s *TransactionService) Transfer(ctx context.Context, userId uint64, amount float64, recipient uint64) (model.TransactionDetails, error) {
	params := model.TransferParams{
		UserID:    userId,
		Amount:    amount,
		Recipient: recipient,
	}

	return s.repo.Transfer(ctx, params)
}
