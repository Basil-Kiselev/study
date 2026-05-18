package service

import (
	"context"
	"fmt"
	"study/Gateway/internal/model"

	"github.com/rs/zerolog"
)

type GatewayService struct {
	logger  *zerolog.Logger
	authSrv AuthService
	accSrv  AccountService
}

func New(logger *zerolog.Logger, authSrv AuthService, accSrv AccountService) *GatewayService {
	return &GatewayService{
		logger:  logger,
		authSrv: authSrv,
		accSrv:  accSrv,
	}
}

type AuthService interface {
	Login(context.Context, string, string) (model.TokenPair, error)
	Refresh(context.Context, string) (model.TokenPair, error)
	Logout(context.Context, string) error
	Register(context.Context, string, string, string) error
	Validate(context.Context, string) (uint64, error)
	DeleteUser(context.Context, uint64) error
}

type AccountService interface {
	GetUser(context.Context, uint64) (model.User, error)
	CreateUser(context.Context, model.CreateUser) (model.User, error)
	GetUsers(context.Context, uint32, uint32) ([]model.User, error)
	DeleteUser(context.Context, uint64) error
	UpdateUser(context.Context, uint64, model.UpdateUser) error
}

func (s *GatewayService) Register(ctx context.Context, newUser model.CreateUser, password string) (model.User, model.TokenPair, error) {
	user, err := s.accSrv.CreateUser(ctx, newUser)
	if err != nil {
		return model.User{}, model.TokenPair{}, fmt.Errorf("fail to create user: %w", err)
	}

	err = s.authSrv.Register(ctx, user.Login, user.Email, password)
	if err != nil {
		return model.User{}, model.TokenPair{}, fmt.Errorf("fail to register user: %w", err)
	}

	tokenPair, err := s.authSrv.Login(ctx, newUser.Login, password)
	if err != nil {
		return model.User{}, model.TokenPair{}, fmt.Errorf("fail to login user: %w", err)
	}

	return user, tokenPair, nil
}

func (s *GatewayService) Login(ctx context.Context, loginOrEmail string, password string) (model.User, model.TokenPair, error) {
	tokenPair, err := s.authSrv.Login(ctx, loginOrEmail, password)
	if err != nil {
		return model.User{}, model.TokenPair{}, fmt.Errorf("fail to login user: %w", err)
	}

	user := model.User{}

	return user, tokenPair, nil
}

func (s *GatewayService) Logout(ctx context.Context, refresh string) error {
	err := s.authSrv.Logout(ctx, refresh)
	if err != nil {
		return fmt.Errorf("fail to logout: %w", err)
	}

	return nil
}

func (s *GatewayService) Refresh(ctx context.Context, refresh string) (model.TokenPair, error) {
	return s.authSrv.Refresh(ctx, refresh)
}

func (s *GatewayService) Validate(ctx context.Context, access string) (uint64, bool, error) {
	userID, err := s.authSrv.Validate(ctx, access)
	if err != nil {
		return 0, false, err
	}

	return userID, true, nil
}

func (s *GatewayService) CreateUser(ctx context.Context, newUser model.CreateUser) error {
	if _, err := s.accSrv.CreateUser(ctx, newUser); err != nil {
		return err
	}

	return nil
}

func (s *GatewayService) GetUsers(ctx context.Context, limit uint32, offset uint32) ([]model.User, error) {
	return s.accSrv.GetUsers(ctx, limit, offset)
}

func (s *GatewayService) GetUser(ctx context.Context, userID uint64) (model.User, error) {
	return s.accSrv.GetUser(ctx, userID)
}

func (s *GatewayService) DeleteUser(ctx context.Context, userID uint64) error {
	err := s.accSrv.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	err = s.authSrv.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *GatewayService) UpdateUser(ctx context.Context, userID uint64, user model.UpdateUser) error {
	return s.accSrv.UpdateUser(ctx, userID, user)
}
