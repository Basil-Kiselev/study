package server

import (
	"context"
	"study/Auth/internal/mapper"
	"study/Auth/internal/model"
	authpb "study/contracts/auth"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	authpb.UnimplementedAuthServer
	logger      *zerolog.Logger
	authService AuthService
}

type AuthService interface {
	Register(context.Context, model.RegisterUser) error
	Login(context.Context, model.LoginUser) (*model.TokenPair, error)
	Refresh(context.Context, string) (*model.TokenPair, error)
	Logout(context.Context, string) error
	Validate(context.Context, string) (uint64, error)
}

func NewServer(service AuthService, logger *zerolog.Logger) *Server {
	return &Server{
		logger:      logger,
		authService: service,
	}
}

func (s *Server) Register(ctx context.Context, req *authpb.RegisterRequest) (*emptypb.Empty, error) {
	regUser := mapper.PbToRegisterUser(req)
	err := s.authService.Register(ctx, regUser)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.TokenPair, error) {
	loginUser := mapper.PbToLoginUser(req)
	tokenPair, err := s.authService.Login(ctx, loginUser)
	if err != nil {
		return nil, err
	}

	return mapper.TokenPairTopb(*tokenPair), nil
}

func (s *Server) Refresh(ctx context.Context, req *authpb.RefreshRequest) (*authpb.TokenPair, error) {
	tokenPair, err := s.authService.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return mapper.TokenPairTopb(*tokenPair), nil
}

func (s *Server) Logout(ctx context.Context, req *authpb.RefreshRequest) (*emptypb.Empty, error) {
	if err := s.authService.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Validate(ctx context.Context, req *authpb.ValidateRequest) (*authpb.ValidateResponse, error) {
	userID, err := s.authService.Validate(ctx, req.GetAccessToken())
	if err != nil {
		return nil, err
	}

	return &authpb.ValidateResponse{
		UserId: userID,
	}, nil
}
