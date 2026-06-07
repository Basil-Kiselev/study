package server

import (
	"context"
	"study/Transaction/internal/mapper"
	"study/Transaction/internal/model"
	transactionpb "study/contracts/transaction"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	transactionpb.UnimplementedTransactionServiceServer
	logger             *zerolog.Logger
	transactionService TransactionService
}

func NewServer(transactionService TransactionService, logger *zerolog.Logger) *Server {
	return &Server{
		logger:             logger,
		transactionService: transactionService,
	}
}

type TransactionService interface {
	GetTransactionsWithDetails(context.Context, model.GetTransactionsParams) ([]model.TransactionDetails, error)
	Deposit(context.Context, uint64, float64) (model.TransactionDetails, error)
	Withdraw(context.Context, uint64, float64) (model.TransactionDetails, error)
	Transfer(context.Context, uint64, float64, uint64) (model.TransactionDetails, error)
}

func (s *Server) Deposit(ctx context.Context, req *transactionpb.DepositRequest) (*emptypb.Empty, error) {
	_, err := s.transactionService.Deposit(ctx, req.UserId, float64(req.Amount))
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Withdraw(ctx context.Context, req *transactionpb.WithdrawRequest) (*emptypb.Empty, error) {
	_, err := s.transactionService.Withdraw(ctx, req.UserId, float64(req.Amount))
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Transfer(ctx context.Context, req *transactionpb.TransferRequest) (*emptypb.Empty, error) {
	_, err := s.transactionService.Transfer(ctx, req.UserdId, float64(req.Amount), req.Recipient)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetTransactions(ctx context.Context, req *transactionpb.GetTransactionsRequest) (*transactionpb.GetTransactionsResponse, error) {
	transDetails, err := s.transactionService.GetTransactionsWithDetails(ctx, mapper.PbToGetTransRequest(req))
	if err != nil {
		return nil, err
	}

	res := mapper.TransDetailsToPb(transDetails)

	return &transactionpb.GetTransactionsResponse{
		Transactions: res,
	}, nil
}
