package server

import (
	"context"
	"study/Account/internal/mapper"
	"study/Account/internal/model"
	accountpb "study/contracts/account"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	accountpb.UnimplementedAccountServer

	accountService AccountService
	logger         *zerolog.Logger
}

func NewServer(accountService AccountService, logger *zerolog.Logger) *Server {
	return &Server{
		accountService: accountService,
		logger:         logger,
	}
}

type AccountService interface {
	CreateUser(context.Context, model.CreateUser) (model.User, error)
	GetUser(context.Context, uint64) (model.User, error)
	GetUsers(context.Context, int, int) ([]model.User, error)
	DeleteUser(context.Context, uint64) error
	UpdateUser(context.Context, uint64, model.UpdateUser) error
}

func (s *Server) CreateUser(ctx context.Context, req *accountpb.CreateUserRequest) (*accountpb.CreateUserResponse, error) {
	user, err := s.accountService.CreateUser(ctx, mapper.PbToUserCreate(req.GetUser()))
	if err != nil {
		return nil, err
	}

	return &accountpb.CreateUserResponse{User: mapper.UserToPb(user)}, nil
}

func (s *Server) GetUser(ctx context.Context, req *accountpb.GetUserRequest) (*accountpb.GetUserResponse, error) {
	user, err := s.accountService.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &accountpb.GetUserResponse{
		User: mapper.UserToPb(user),
	}, nil
}

func (s *Server) GetUsers(ctx context.Context, req *accountpb.GetUsersRequest) (*accountpb.GetUsersResponse, error) {
	users, err := s.accountService.GetUsers(ctx, int(req.GetPagination().GetLimit()), int(req.GetPagination().GetOffset()))
	if err != nil {
		return nil, err
	}

	return &accountpb.GetUsersResponse{
		Users: mapper.UsersToPb(users),
	}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *accountpb.DeleteUserRequest) (*emptypb.Empty, error) {
	if err := s.accountService.DeleteUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *accountpb.UpdateUserRequest) (*emptypb.Empty, error) {
	if err := s.accountService.UpdateUser(ctx, req.GetUserId(), mapper.PbToUserUpdate(req.GetUser())); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
