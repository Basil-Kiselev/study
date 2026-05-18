package repository

import (
	"context"
	"fmt"
	"study/Auth/internal/model"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	logger *zerolog.Logger
}

func NewRepo(db *gorm.DB, logger *zerolog.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user model.User) error {
	res := r.db.WithContext(ctx).Create(&user)
	if res.Error != nil {
		return fmt.Errorf("create user by login %s: %w", user.Login, res.Error)
	}

	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID uint64) (*model.User, error) {
	var user model.User
	res := r.db.WithContext(ctx).Where("id = ?", userID).First(&user)
	if res.Error != nil {
		return nil, fmt.Errorf("get user by id %d: %w", userID, res.Error)
	}

	return &user, nil
}

func (r *Repository) GetUserByLoginOrEmail(ctx context.Context, loe string) (*model.User, error) {
	var user model.User
	res := r.db.WithContext(ctx).Where("login = ?", loe).Or("email = ?", loe).First(&user)
	if res.Error != nil {
		return nil, fmt.Errorf("get user by login or email %s: %w", loe, res.Error)
	}

	return &user, nil
}

func (r *Repository) SaveRefreshToken(ctx context.Context, token model.RefreshToken) error {
	res := r.db.WithContext(ctx).Create(&token)
	if res.Error != nil {
		return fmt.Errorf("save token: %w", res.Error)
	}

	return nil
}

func (r *Repository) GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var rToken model.RefreshToken
	res := r.db.WithContext(ctx).Where("token = ? AND revoked_at IS NULL", token).First(&rToken)
	if res.Error != nil {
		return nil, fmt.Errorf("get rToken %s: %w", token, res.Error)
	}

	return &rToken, nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, token string) error {
	res := r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked_at", time.Now())
	if res.Error != nil {
		return fmt.Errorf("revoke token %s: %w", token, res.Error)
	}

	return nil
}
