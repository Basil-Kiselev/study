package service

import (
	"context"
	"fmt"
	"study/Transaction/internal/account"
	"study/Transaction/internal/model"

	"github.com/rs/zerolog"
)

type TransactionService struct {
	repo           Repository
	logger         *zerolog.Logger
	accountService *account.Service
}

func New(repo Repository, logger *zerolog.Logger, accountService *account.Service) *TransactionService {
	return &TransactionService{
		repo:           repo,
		logger:         logger,
		accountService: accountService,
	}
}

type Repository interface {
	GetTransactions(context.Context, model.GetTransactionsParams) ([]model.Transaction, error)
	GetTransactionDetails(context.Context, uint64) (model.TransactionDetails, error)
	Deposit(context.Context, model.DepositParams) (model.TransactionDetails, error)
	Transfer(context.Context, model.TransferParams) (model.TransactionDetails, error)
	Withdraw(context.Context, model.WithdrawParams) (model.TransactionDetails, error)
	UpdateTransactionsStatus(context.Context, uint64, model.TransactionStatus) error
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

	res, err := s.repo.Deposit(ctx, params)
	if err != nil {
		return model.TransactionDetails{}, fmt.Errorf("fail to create deposit transsaction: %w", err)
	}

	_, err = s.accountService.Deposit(ctx, userId, float32(amount), res.Transaction.ID)
	if err != nil {
		s.logger.Error().Uint64("userID", userId).Uint64("transactionID", res.Transaction.ID).Msg("fail create deposit transaction in acc service")
		err := s.repo.UpdateTransactionsStatus(ctx, res.Transaction.ID, model.TransactionStatusFailed)
		if err != nil {
			return model.TransactionDetails{}, fmt.Errorf("fail to updated transaction status to failed: %w", err)
		}
		return model.TransactionDetails{}, fmt.Errorf("fail to create deposit transaction in account service: %w", err)
	}

	err = s.repo.UpdateTransactionsStatus(ctx, res.Transaction.ID, model.TransactionStatusCompleted)
	if err != nil {
		return model.TransactionDetails{}, fmt.Errorf("fail to update transaction status: %w", err)
	}

	res.Transaction.Status = model.TransactionStatusCompleted
	return res, nil
}

func (s *TransactionService) Withdraw(ctx context.Context, userId uint64, amount float64) (model.TransactionDetails, error) {
	params := model.WithdrawParams{
		AccountID: userId,
		Amount:    amount,
	}

	res, err := s.repo.Withdraw(ctx, params)
	if err != nil {
		return model.TransactionDetails{}, fmt.Errorf("fail to create withdraw transaction: %w", err)
	}

	_, err = s.accountService.Withdraw(ctx, userId, float32(amount), res.Transaction.ID)
	if err != nil {
		s.logger.Error().Uint64("userID", userId).Uint64("transactionID", res.Transaction.ID).Msg("fail to create withdraw transaction in account service")
		err := s.repo.UpdateTransactionsStatus(ctx, res.Transaction.ID, model.TransactionStatusFailed)
		if err != nil {
			return model.TransactionDetails{}, fmt.Errorf("fail to update transaction failed status: %w", err)
		}
		return model.TransactionDetails{}, fmt.Errorf("fail to create withdraw transaction in account setvice: %w", err)
	}

	err = s.repo.UpdateTransactionsStatus(ctx, res.Transaction.ID, model.TransactionStatusCompleted)
	if err != nil {
		return model.TransactionDetails{}, fmt.Errorf("fail to update transaction status to completed: %w", err)
	}

	res.Transaction.Status = model.TransactionStatusCompleted

	return res, nil
}

func (s *TransactionService) Transfer(ctx context.Context, userId uint64, amount float64, recipient uint64) (model.TransactionDetails, error) {
	params := model.TransferParams{
		UserID:    userId,
		Amount:    amount,
		Recipient: recipient,
	}

	res, err := s.repo.Transfer(ctx, params)
	if err != nil {
		return model.TransactionDetails{}, fmt.Errorf("fail to create transfer transaction: %w", err)
	}

	_, err = s.accountService.Transfer(ctx, userId, recipient, float32(amount), res.Transaction.ID)
	if err != nil {
		s.logger.Error().Uint64("userID", userId).Uint64("recipientID", recipient).Uint64("transactionID", res.Transaction.ID).Msg("fail to create transfer transaction in account service")
		err = s.repo.UpdateTransactionsStatus(ctx, res.Transaction.ID, model.TransactionStatusFailed)
		if err != nil {
			return model.TransactionDetails{}, fmt.Errorf("fail to update transfer transaction status: %w", err)
		}
		return model.TransactionDetails{}, fmt.Errorf("fail to create transfer transaction in account service")
	}

	err = s.repo.UpdateTransactionsStatus(ctx, res.Transaction.ID, model.TransactionStatusCompleted)
	if err != nil {
		return model.TransactionDetails{}, fmt.Errorf("fail to update transaction status to completed: %w", err)
	}

	res.Transaction.Status = model.TransactionStatusCompleted

	return res, nil
}
