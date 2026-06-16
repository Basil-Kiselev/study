package service

import (
	"context"
	"encoding/json"
	"fmt"
	"study/Transaction/internal/account"
	"study/Transaction/internal/model"
	"time"

	"github.com/rs/zerolog"
)

type TransactionService struct {
	repo           Repository
	logger         *zerolog.Logger
	accountService *account.Service
	kafka          KafkaPublisher
}

type AccountResponse struct {
	RequestType string `json:"request_type"`
	UserID      uint64 `json:"user_id"`
	OperationID uint64 `json:"operation_id"`
	Success     bool   `json:"success"`
}

func New(repo Repository, logger *zerolog.Logger, accountService *account.Service, kafka KafkaPublisher) *TransactionService {
	return &TransactionService{
		repo:           repo,
		logger:         logger,
		accountService: accountService,
		kafka:          kafka,
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

type KafkaPublisher interface {
	Publish(ctx context.Context, topic string, key string, data any) error
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

	request := map[string]any{
		"request_type":  "deposit",
		"user_id":       userId,
		"amount":        int64(amount * 100),
		"opetartion_id": res.Transaction.ID,
		"timestamp":     time.Now().UTC(),
	}

	err = s.kafka.Publish(ctx, "transaction_data", fmt.Sprintf("%d", userId), request)
	if err != nil {
		s.logger.Error().Err(err).Uint64("transaction_id", res.Transaction.ID).Uint64("user_id", userId).Msg("fail to publish message to kafka")
		updErr := s.repo.UpdateTransactionsStatus(ctx, res.Transaction.ID, model.TransactionStatusFailed)
		if updErr != nil {
			return model.TransactionDetails{}, fmt.Errorf("fail to updated transaction status to failed: %w", err)
		}
		return model.TransactionDetails{}, err
	}

	s.logger.Info().Uint64("transaction_id", res.Transaction.ID).Uint64("user_id", userId).Msg("successfully send deposit transaction msg to kafka")
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

	request := map[string]any{
		"request_type": "withdraw",
		"user_id":      userId,
		"amount":       int64(amount * 100),
		"operation_id": res.Transaction.ID,
		"timestamp":    time.Now().UTC(),
	}

	err = s.kafka.Publish(ctx, "transaction_data", fmt.Sprintf("%d", userId), request)
	if err != nil {
		s.logger.Error().Err(err).Uint64("transaction_id", res.Transaction.ID).Uint64("user_id", userId).Msg("fail to send transaction withdraw msg to kafka")
		updErr := s.repo.UpdateTransactionsStatus(ctx, res.Transaction.ID, model.TransactionStatusFailed)
		if updErr != nil {
			return model.TransactionDetails{}, fmt.Errorf("fail to update transaction status failed: %w", updErr)
		}
		return model.TransactionDetails{}, err
	}

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

	request := map[string]any{
		"request_type": "transfer",
		"user_id":      userId,
		"recipient_id": recipient,
		"amount":       int64(amount * 100),
		"operation_id": res.Transaction.ID,
		"timestamp":    time.Now().UTC(),
	}

	err = s.kafka.Publish(ctx, "transaction_data", fmt.Sprintf("%d", userId), request)
	if err != nil {
		s.logger.Error().Err(err).Uint64("transaction_id", res.Transaction.ID).Uint64("user_id", userId).Msg("fail to send msg transfer to kafka")
		updErr := s.repo.UpdateTransactionsStatus(ctx, res.Transaction.ID, model.TransactionStatusFailed)
		if updErr != nil {
			return model.TransactionDetails{}, err
		}

		return model.TransactionDetails{}, fmt.Errorf("fail to publish transaction transfer msg to kafka")
	}

	s.logger.Debug().Uint64("transaction_id", res.Transaction.ID).Uint64("user_id", userId).Msg("successfully send transfer transaction msg to kafka")

	return res, nil
}

func (s *TransactionService) HandleAccountResponse(ctx context.Context, topic string, key string, data []byte) error {
	var response AccountResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("fail to unmarshal account response json: %w", err)
	}

	if response.RequestType == "" || response.UserID == 0 || response.OperationID == 0 {
		return fmt.Errorf("invalid account response")
	}

	s.logger.Info().
		Str("request_type", response.RequestType).
		Uint64("user_id", response.UserID).
		Uint64("operation_id", response.OperationID).
		Msg("received account response from kafka")

	var status model.TransactionStatus
	var logMsg string

	if response.Success {
		status = model.TransactionStatusCompleted
		logMsg = "transcastion successfully"
	} else {
		status = model.TransactionStatusFailed
		logMsg = "transaction failed"
	}

	err := s.repo.UpdateTransactionsStatus(ctx, response.OperationID, status)
	if err != nil {
		s.logger.Error().Err(err).Uint64("transaction_id", response.OperationID).Uint64("user_id", response.UserID).Msg("fail to update trans status")
		return fmt.Errorf("fail to update transaction status: %w", err)
	}
	s.logger.Info().Uint64("transaction_id", response.OperationID).Uint64("user_id", response.UserID).Msg(logMsg)

	return nil
}
