package auth

import (
	"context"
	"fmt"
	"study/Gateway/internal/model"
	authpb "study/contracts/auth"
)

type Service struct {
	client authpb.AuthClient
}

func NewService(client authpb.AuthClient) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) Login(ctx context.Context, loginOrEmail string, password string) (model.TokenPair, error) {
	pair, err := s.client.Login(ctx, &authpb.LoginRequest{LoginOrEmail: loginOrEmail, Password: password})
	if err != nil {
		return model.TokenPair{}, fmt.Errorf("fail to login: %w", err)
	}

	return PbToTokenPair(pair), nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (model.TokenPair, error) {
	pair, err := s.client.Refresh(ctx, &authpb.RefreshRequest{RefreshToken: refreshToken})
	if err != nil {
		return model.TokenPair{}, fmt.Errorf("fail to refresh token: %w", err)
	}

	return PbToTokenPair(pair), nil
}

func (s *Service) Logout(ctx context.Context, refresh string) error {
	_, err := s.client.Logout(ctx, &authpb.RefreshRequest{RefreshToken: refresh})
	if err != nil {
		return fmt.Errorf("fail to logout: %w", err)
	}

	return nil
}

func (s *Service) Validate(ctx context.Context, accessTok string) (uint64, error) {
	response, err := s.client.Validate(ctx, &authpb.ValidateRequest{AccessToken: accessTok})
	if err != nil {
		return 0, fmt.Errorf("validate fail: %w", err)
	}

	return response.UserId, nil
}

func (s *Service) Register(ctx context.Context, login, email, password string) error {
	_, err := s.client.Register(ctx, &authpb.RegisterRequest{
		Login:    login,
		Email:    email,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("fail to register user: %w", err)
	}
	return nil
}

func (s *Service) DeleteUser(ctx context.Context, userID uint64) error {
	_, err := s.client.DeleteUser(ctx, &authpb.DeleteUserRequest{UserId: userID})
	if err != nil {
		return fmt.Errorf("fail to delete user: %w", err)
	}
	return nil
}

func PbToTokenPair(pbPair *authpb.TokenPair) model.TokenPair {
	return model.TokenPair{
		AccessToken:  pbPair.AccessToken,
		RefreshToken: pbPair.RefreshToken,
	}
}
