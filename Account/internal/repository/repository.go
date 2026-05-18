package repository

import (
	"context"
	"errors"
	"fmt"
	"study/Account/internal/model"
	"study/Account/internal/repository/mapper"
	repomodel "study/Account/internal/repository/model"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db     *gorm.DB
	logger *zerolog.Logger
}

func NewRepository(db *gorm.DB, logger *zerolog.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	userRepo := mapper.UserToRepoUser(user)
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&userRepo)
	if res.Error != nil {
		r.logger.Error().Err(res.Error).Msg("failed to create user")
		return model.User{}, res.Error
	}
	return mapper.RepoUserToUser(userRepo), nil
}

func (r *Repository) GetUser(ctx context.Context, id uint64) (model.User, error) {
	var user repomodel.User
	res := r.db.WithContext(ctx).Model(&repomodel.User{}).
		Where("id = ?", id).
		First(&user)
	if res.Error != nil {
		return model.User{}, res.Error
	}
	return mapper.RepoUserToUser(user), nil
}

func (r *Repository) GetUsers(ctx context.Context, limit int, offset int) ([]model.User, error) {
	var users []repomodel.User
	res := r.db.WithContext(ctx).Model(&repomodel.User{}).Offset(offset).Limit(limit).Find(&users)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("users not found")
	} else if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get users")
		return nil, res.Error
	}
	return mapper.RepoUsersToUsers(users), nil
}

func (r *Repository) UpdateUser(ctx context.Context, userId uint64, user model.UpdateUser) error {
	res := r.db.WithContext(ctx).Table("users").Where("id = ?", userId).Updates(user)
	if res.Error != nil {
		r.logger.Err(res.Error).Msg("fail to update user")
		return fmt.Errorf("fail to update user")
	}

	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID uint64) error {
	var user repomodel.User
	if res := r.db.WithContext(ctx).Where("id = ?", userID).Delete(&user); res.Error != nil {
		return res.Error
	}
	return nil
}
