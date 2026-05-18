package server

import (
	"context"
	"study/Gateway/internal/interceptor"
	"study/Gateway/internal/mapper"
	"study/Gateway/internal/model"
	gatewaypb "study/contracts/gateway"
	"study/contracts/pagination"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	gatewaypb.UnimplementedGatewayServer
	logger         *zerolog.Logger
	gatewayService GatewayService
}

func New(svc GatewayService, logger *zerolog.Logger) *Server {
	return &Server{
		logger:         logger,
		gatewayService: svc,
	}
}

type GatewayService interface {
	Register(context.Context, model.CreateUser, string) (model.User, model.TokenPair, error)
	Login(context.Context, string, string) (model.User, model.TokenPair, error)
	Logout(context.Context, string) error
	Refresh(context.Context, string) (model.TokenPair, error)
	Validate(context.Context, string) (uint64, bool, error)
	CreateUser(context.Context, model.CreateUser) error
	GetUsers(context.Context, uint32, uint32) ([]model.User, error)
	GetUser(context.Context, uint64) (model.User, error)
	DeleteUser(context.Context, uint64) error
	UpdateUser(context.Context, uint64, model.UpdateUser) error
}

func (s *Server) New(service GatewayService, logger *zerolog.Logger) *Server {
	return &Server{
		logger:         logger,
		gatewayService: service,
	}
}

func (s *Server) Register(ctx context.Context, req *gatewaypb.RegisterRequest) (*gatewaypb.RegisterResponse, error) {
	user, tokens, err := s.gatewayService.Register(ctx, mapper.PbToUserCreate(req.User), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &gatewaypb.RegisterResponse{
		User:   mapper.UserToPb(user),
		Tokens: mapper.TokenPairTopb(tokens),
	}, nil
}

func (s *Server) Login(ctx context.Context, req *gatewaypb.LoginRequest) (*gatewaypb.LoginResponse, error) {
	user, tokens, err := s.gatewayService.Login(ctx, req.GetLoginOrEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &gatewaypb.LoginResponse{
		User:   mapper.UserToPb(user),
		Tokens: mapper.TokenPairTopb(tokens),
	}, nil
}

func (s *Server) Refresh(ctx context.Context, req *gatewaypb.RefreshRequest) (*gatewaypb.RefreshResponse, error) {
	tokens, err := s.gatewayService.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return &gatewaypb.RefreshResponse{Tokens: mapper.TokenPairTopb(tokens)}, nil
}

func (s *Server) Logout(ctx context.Context, req *gatewaypb.LogoutRequest) (*emptypb.Empty, error) {
	err := s.gatewayService.Logout(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *gatewaypb.ValidateTokenRequest) (*gatewaypb.ValidateTokenResponse, error) {
	userID, isValid, err := s.gatewayService.Validate(ctx, req.GetAccessToken())
	if err != nil {
		return nil, err
	}

	return &gatewaypb.ValidateTokenResponse{
		UserId:  userID,
		IsValid: isValid,
	}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *gatewaypb.CreateUserRequest) (*emptypb.Empty, error) {
	err := s.gatewayService.CreateUser(ctx, mapper.PbToUserCreate(req.User))
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetUsers(ctx context.Context, req *gatewaypb.GetUsersRequest) (*gatewaypb.GetUsersResponse, error) {
	users, err := s.gatewayService.GetUsers(ctx, req.Pagination.GetLimit(), req.Pagination.GetOffset())
	if err != nil {
		return nil, err
	}

	return &gatewaypb.GetUsersResponse{
		Users: mapper.UsersToPb(users),
		Pagination: &pagination.Pagination{
			Limit:  req.Pagination.GetLimit(),
			Offset: req.Pagination.GetOffset(),
		},
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *gatewaypb.GetUserRequest) (*gatewaypb.GetUserResponse, error) {
	user, err := s.gatewayService.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &gatewaypb.GetUserResponse{User: mapper.UserToPb(user)}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *gatewaypb.DeleteUserRequest) (*emptypb.Empty, error) {
	err := s.gatewayService.DeleteUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *gatewaypb.UpdateUserRequest) (*emptypb.Empty, error) {
	err := s.gatewayService.UpdateUser(ctx, req.GetUserId(), mapper.PbToUserUpdate(req.User))
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetCurrentUser(ctx context.Context, req *emptypb.Empty) (*gatewaypb.GetCurrentUserResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.gatewayService.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &gatewaypb.GetCurrentUserResponse{User: mapper.UserToPb(user)}, nil
}

func (s *Server) DeleteCurrentUser(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.gatewayService.DeleteUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateCurrentUser(ctx context.Context, req *gatewaypb.UpdateCurrentUserRequest) (*emptypb.Empty, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.gatewayService.UpdateUser(ctx, userID, mapper.PbToUserUpdate(req.User))
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
