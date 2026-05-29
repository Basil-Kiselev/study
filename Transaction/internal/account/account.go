package account

import (
	"context"
	"fmt"
	accountpb "study/contracts/account"
)

type Service struct {
	client accountpb.AccountClient
}

func New(client accountpb.AccountClient) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) GetBalance(ctx context.Context, userID uint64) (int64, error) {
	res, err := s.client.GetBalance(ctx, &accountpb.GetBalanceRequest{UserId: userID})
	if err != nil {
		return 0, fmt.Errorf("fail to get balance: %w", err)
	}

	return res.Balance, nil
}

func (s *Service) Deposit(ctx context.Context, userID uint64, amount float32, operationID uint64) (int64, error) {
	res, err := s.client.Deposit(ctx, &accountpb.DepositRequest{
		UserId:      userID,
		Amount:      int64(amount),
		OperationId: operationID,
	})
	if err != nil {
		return 0, fmt.Errorf("fail to deposit: %w", err)
	}

	return res.Balance, nil
}

func (s *Service) Withdraw(ctx context.Context, userID uint64, amount float32, operationID uint64) (int64, error) {
	res, err := s.client.Withdraw(ctx, &accountpb.WithdrawRequest{
		UserId:      userID,
		Amount:      int64(amount),
		OperationId: operationID,
	})
	if err != nil {
		return 0, fmt.Errorf("fail to withdraw: %w", err)
	}

	return res.Balance, nil
}

func (s *Service) Transfer(ctx context.Context, userID, recipientID uint64, amount float32, operationID uint64) (string, error) {
	res, err := s.client.Transfer(ctx, &accountpb.TransferRequest{
		UserId:      userID,
		RecipientId: recipientID,
		Amount:      int64(amount),
		OperationId: operationID,
	})

	return res.Status, err
}
