package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"study/Account/internal/model"
	"time"

	"github.com/rs/zerolog"
)

type AccountService struct {
	repo   Repository
	logger *zerolog.Logger
	kafka  KafkaPublisher
}

func NewAccountService(repo Repository, logger *zerolog.Logger, kafka KafkaPublisher) *AccountService {
	return &AccountService{
		repo:   repo,
		logger: logger,
		kafka:  kafka,
	}
}

type KafkaPublisher interface {
	Publish(ctx context.Context, topic string, key string, data any) error
}
type Repository interface {
	CreateUser(context.Context, model.User) (model.User, error)
	GetUser(context.Context, uint64) (model.User, error)
	GetUsers(context.Context, int, int) ([]model.User, error)
	DeleteUser(context.Context, uint64) error
	UpdateUser(context.Context, uint64, model.UpdateUser) error
	GetBalance(context.Context, uint64) (float32, error)
	UpdateBalance(context.Context, uint64, float32, model.OperationType) (float32, float32, error)
	TransferBalance(context.Context, uint64, uint64, float32) error
}

type TransactionRequest struct {
	RequestType string `json:"request_type"`
	UserID      uint64 `json:"user_id"`
	OperationID uint64 `json:"operation_id"`
	Amount      int64  `json:"amount"`
	RecipientID uint64 `json:"recipient_id"`
}

type TransactionResponse struct {
	RequestType string                 `json:"request_type"`
	UserID      uint64                 `json:"user_id"`
	OperationID uint64                 `json:"operation_id"`
	Result      map[string]interface{} `json:"result"`
}

func (s *AccountService) CreateUser(ctx context.Context, newUser model.CreateUser) (model.User, error) {
	user := model.User{
		Login:      newUser.Login,
		Email:      newUser.Email,
		Phone:      newUser.Phone,
		FirstName:  newUser.FirstName,
		LastName:   newUser.LastName,
		MiddleName: newUser.MiddleName,
		Age:        newUser.Age,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *AccountService) GetUsers(ctx context.Context, limit int, offset int) ([]model.User, error) {
	return s.repo.GetUsers(ctx, limit, offset)
}

func (s *AccountService) GetUser(ctx context.Context, userID uint64) (model.User, error) {
	return s.repo.GetUser(ctx, userID)
}

func (s *AccountService) DeleteUser(ctx context.Context, userID uint64) error {
	return s.repo.DeleteUser(ctx, userID)
}

func (s *AccountService) UpdateUser(ctx context.Context, userID uint64, user model.UpdateUser) error {
	return s.repo.UpdateUser(ctx, userID, user)
}

func (s *AccountService) GetBalance(ctx context.Context, userID uint64) (model.GetBalanceResponse, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Error().Uint64("userID", userID).Msg("fail to get balance")
		return model.GetBalanceResponse{}, err
	}

	return model.GetBalanceResponse{
		Balance: balance,
		UserID:  userID,
	}, nil
}

func (s *AccountService) UpdateBalance(ctx context.Context, req model.UpdateBalanceRequest) (model.UpdateBalanceResponse, error) {
	if req.Amount < 0 {
		return model.UpdateBalanceResponse{}, fmt.Errorf("amount cannot be negative")
	}

	oldBalance, newBalance, err := s.repo.UpdateBalance(ctx, req.UserID, req.Amount, req.Type)
	if err != nil {
		return model.UpdateBalanceResponse{}, fmt.Errorf("fail to update balance: %w", err)
	}

	s.logger.Info().Uint64("user_id", req.UserID).
		Float32("amount", req.Amount).
		Float32("old_balance", oldBalance).
		Float32("new_balance", newBalance).
		Str("type", string(req.Type)).Msg("update balance successfully")

	return model.UpdateBalanceResponse{
		UserID:     req.UserID,
		Amount:     req.Amount,
		OldBalance: oldBalance,
		NewBalance: newBalance,
		Type:       req.Type,
	}, nil
}

func (s *AccountService) TransferBalance(ctx context.Context, fromUserID, toUserID uint64, amount float32) error {
	if amount < 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	if toUserID == fromUserID {
		return fmt.Errorf("user ids dont be equal")
	}

	err := s.repo.TransferBalance(ctx, fromUserID, toUserID, amount)
	if err != nil {
		return fmt.Errorf("fail to transfer: %w", err)
	}

	s.logger.Info().Uint64("from_user_id", fromUserID).
		Uint64("to_user_id", toUserID).
		Float32("amount", amount).
		Msg("transfer successfully")

	return nil
}

func (s *AccountService) Deposit(ctx context.Context, userID uint64, amount float32) (model.UpdateBalanceResponse, error) {
	req := model.UpdateBalanceRequest{
		UserID: userID,
		Amount: amount,
		Type:   model.OperationTypeDeposit,
	}

	res, err := s.UpdateBalance(ctx, req)
	if err != nil {
		return model.UpdateBalanceResponse{}, fmt.Errorf("fail deposit update user balance: %w", err)
	}

	s.logger.Info().Uint64("user_id", userID).
		Str("type", string(model.OperationTypeDeposit)).
		Float32("amount", amount).
		Msg("update balance successfully")

	return res, nil
}

func (s *AccountService) Withdraw(ctx context.Context, userID uint64, amount float32) (model.UpdateBalanceResponse, error) {
	req := model.UpdateBalanceRequest{
		UserID: userID,
		Amount: amount,
		Type:   model.OperationTypeCredit,
	}

	res, err := s.UpdateBalance(ctx, req)
	if err != nil {
		return model.UpdateBalanceResponse{}, fmt.Errorf("fail update credit user balance: %w", err)
	}

	s.logger.Info().Uint64("user_id", userID).
		Str("type", string(model.OperationTypeCredit)).
		Float32("amount", amount).
		Msg("update balance successfully")

	return res, nil
}

func (s *AccountService) Transfer(ctx context.Context, fromUserID, toUserID uint64, amount float32) (model.GetBalanceResponse, model.GetBalanceResponse, error) {
	err := s.TransferBalance(ctx, fromUserID, toUserID, amount)
	if err != nil {
		return model.GetBalanceResponse{}, model.GetBalanceResponse{}, fmt.Errorf("fail to transfer balance: %w", err)
	}

	fromUserBalance, err := s.GetBalance(ctx, fromUserID)
	if err != nil {
		return model.GetBalanceResponse{}, model.GetBalanceResponse{}, fmt.Errorf("fail to get fromUser balance: %w", err)
	}

	toUserBalance, err := s.GetBalance(ctx, toUserID)
	if err != nil {
		return model.GetBalanceResponse{}, model.GetBalanceResponse{}, fmt.Errorf("fail to get toUser balance: %w", err)
	}

	s.logger.Info().Uint64("from_user_id", fromUserID).
		Uint64("to_user_id", toUserID).
		Float32("amount", amount).Msg("transfer successfully")

	return fromUserBalance, toUserBalance, nil
}

func (s *AccountService) HandleTransaction(ctx context.Context, topic string, key string, data []byte) error {
	s.logger.Info().Msg("handling transaction request")
	var request TransactionRequest
	var err error

	err = json.Unmarshal(data, &request)
	if err != nil {
		return fmt.Errorf("fail to unmarshal json: %w", err)
	}

	result := map[string]any{
		"request_type": request.RequestType,
		"user_id":      request.UserID,
		"success":      false,
		"operation_id": request.OperationID,
	}

	switch request.RequestType {
	case "deposit":
		_, err = s.Deposit(ctx, request.UserID, float32(request.Amount/100.0))
	case "withdraw":
		_, err = s.Withdraw(ctx, request.UserID, float32(request.Amount/100.0))
	case "transfer":
		_, _, err = s.Transfer(ctx, request.UserID, request.RecipientID, float32(request.Amount/100.0))
	default:
		err = fmt.Errorf("invalid transaction request type: %s", request.RequestType)
	}

	if err != nil {
		s.logger.Error().Err(err).Uint64("transaction_id", request.OperationID).Uint64("user_id", request.UserID).Msg("fail to handle transaction request")
		s.kafka.Publish(ctx, "transaction_response", "key", result)
		return fmt.Errorf("fail to handle transaction request: %w", err)
	}

	result["success"] = true
	s.kafka.Publish(ctx, "transaction_response", strconv.FormatUint(request.UserID, 10), result)

	return nil
}
