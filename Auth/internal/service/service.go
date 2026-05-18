package service

import (
	"context"
	"fmt"
	"study/Auth/internal/config"
	"study/Auth/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo   Repository
	logger *zerolog.Logger
	cfg    config.Config
}

func NewAuthService(repo Repository, logger *zerolog.Logger, cfg config.Config) *AuthService {
	return &AuthService{
		repo:   repo,
		logger: logger,
		cfg:    cfg,
	}
}

type Repository interface {
	CreateUser(context.Context, model.User) error
	GetUserByID(context.Context, uint64) (*model.User, error)
	GetUserByLoginOrEmail(context.Context, string) (*model.User, error)
	SaveRefreshToken(context.Context, model.RefreshToken) error
	GetRefreshToken(context.Context, string) (*model.RefreshToken, error)
	RevokeRefreshToken(context.Context, string) error
}

func (s *AuthService) Register(ctx context.Context, regUser model.RegisterUser) error {
	pass, err := bcrypt.GenerateFromPassword([]byte(regUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("generate password hash: %w", err)
	}
	user := model.User{
		Login:        regUser.Login,
		Email:        regUser.Email,
		PasswordHash: string(pass),
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, creds model.LoginUser) (*model.TokenPair, error) {
	user, err := s.repo.GetUserByLoginOrEmail(ctx, creds.LoginOrMail)
	if err != nil {
		return nil, fmt.Errorf("get user by login or email: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password))
	if err != nil {
		return nil, fmt.Errorf("compare password: %w", err)
	}

	return s.generateTokenPair(ctx, user.ID)
}

func (s *AuthService) generateTokenPair(ctx context.Context, userID uint64) (*model.TokenPair, error) {
	now := time.Now()

	expAccessTime := now.Add(time.Duration(s.cfg.AccessTokenTTLMinutes) * time.Minute)
	accessToken, err := s.generateToken(userID, now, expAccessTime)
	if err != nil {
		return nil, fmt.Errorf("fail generated token: %w", err)
	}

	expRefreshTime := now.Add(time.Duration(s.cfg.RefreshTokenTTLDays) * time.Hour * 24)
	refreshToken, err := s.generateToken(userID, now, expRefreshTime)
	if err != nil {
		return nil, fmt.Errorf("fail generated token: %w", err)
	}

	refresh := model.RefreshToken{
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: expRefreshTime,
	}

	if err = s.repo.SaveRefreshToken(ctx, refresh); err != nil {
		return nil, fmt.Errorf("fail to save refresh token: %w", err)
	}

	tokenPair := model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return &tokenPair, nil
}

func (s *AuthService) generateToken(userID uint64, now, exp time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": jwt.NewNumericDate(exp),
		"iat": jwt.NewNumericDate(now),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JwtSecret))
	if err != nil {
		return "", fmt.Errorf("fail to create jwt token: %w", err)
	}

	return token, nil
}

func (s *AuthService) Logout(ctx context.Context, refresh string) error {
	return s.repo.RevokeRefreshToken(ctx, refresh)
}

func (s *AuthService) Refresh(ctx context.Context, refresh string) (*model.TokenPair, error) {
	rt, err := s.repo.GetRefreshToken(ctx, refresh)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if rt.RevokeAt != nil || time.Now().After(rt.ExpiresAt) {
		return nil, fmt.Errorf("refresh token revoked or expired")
	}

	return s.generateTokenPair(ctx, rt.UserID)
}

func (s *AuthService) Validate(ctx context.Context, accessToken string) (uint64, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (any, error) { return []byte(s.cfg.JwtSecret), nil })
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	uidFloat, ok := claims["sub"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid subject")
	}
	return uint64(uidFloat), nil
}
