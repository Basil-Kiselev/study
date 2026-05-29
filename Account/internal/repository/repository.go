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
		Where("id = ? AND is_deleted = ?", id, false).
		First(&user)
	if res.Error != nil {
		return model.User{}, res.Error
	}
	return mapper.RepoUserToUser(user), nil
}

func (r *Repository) GetUsers(ctx context.Context, limit int, offset int) ([]model.User, error) {
	var users []repomodel.User
	res := r.db.WithContext(ctx).Where("is_deleted = ?", false).Model(&repomodel.User{}).Offset(offset).Limit(limit).Find(&users)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("users not found")
	} else if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get users")
		return nil, res.Error
	}
	return mapper.RepoUsersToUsers(users), nil
}

func (r *Repository) UpdateUser(ctx context.Context, userId uint64, user model.UpdateUser) error {
	res := r.db.WithContext(ctx).Table("users").Where("id = ? AND is_deleted = ?", userId, false).Updates(user)
	if res.Error != nil {
		r.logger.Err(res.Error).Msg("fail to update user")
		return fmt.Errorf("fail to update user")
	}

	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID uint64) error {
	if res := r.db.WithContext(ctx).Where("id = ?", userID).Update("is_deleted", true); res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *Repository) GetBalance(ctx context.Context, userID uint64) (float32, error) {
	var balance float32
	res := r.db.WithContext(ctx).
		Model(&repomodel.User{}).
		Select("balance").
		Where("id = ? and is_deleted = ?", userID, false).
		Scan(&balance)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("user not found")
	} else if res.Error != nil {
		r.logger.Error().Err(res.Error).Msg("fail to get user balance")
		return 0, fmt.Errorf("fail to get user balance: %w", res.Error)
	}

	return balance, nil
}
func (r *Repository) UpdateBalance(ctx context.Context, userID uint64, amount float32, operationType model.OperationType) (float32, float32, error) {
	var oldBalance float32
	var newBalance float32

	res := r.db.WithContext(ctx).
		Select("balance").
		Where("id = ? and is_deleted = ?", userID, false).
		Model(&repomodel.User{}).
		Scan(&oldBalance)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return 0, 0, fmt.Errorf("user not found")
	} else if res.Error != nil {
		r.logger.Error().Err(res.Error).Msg("fail to get user balance on update")
		return 0, 0, fmt.Errorf("fail to get user balance on update: %w", res.Error)
	}

	switch operationType {
	case model.OperationTypeCredit:
		newBalance = oldBalance - amount
		if newBalance < 0 {
			return 0, 0, fmt.Errorf("insufficient balance: current balance %.2f, requested amount %.2f", oldBalance, amount)
		}
	case model.OperationTypeDeposit:
		newBalance = oldBalance + amount
	default:
		return 0, 0, fmt.Errorf("invalid operation type")
	}

	res = r.db.WithContext(ctx).Model(&repomodel.User{}).
		Where("id = ? and is_deleted = ?", userID, false).
		Update("balance", newBalance)

	if res.Error != nil {
		return 0, 0, fmt.Errorf("fail to update user balance: %w", res.Error)
	}

	return oldBalance, newBalance, nil
}

func (r *Repository) TransferBalance(ctx context.Context, fromUserID, toUserID uint64, amount float32) error {
	if amount <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	if fromUserID == toUserID {
		return fmt.Errorf("users id must be difference")
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var toUser repomodel.User
		res := tx.Model(&repomodel.User{}).Where("id = ? and is_deleted = ?", toUserID, false).Find(&toUser)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("recipient user not found")
		} else if res.Error != nil {
			return fmt.Errorf("fail to get recipient user: %w", res.Error)
		}

		var fromUser repomodel.User
		res = tx.Model(&repomodel.User{}).Where("id = ? and is_deleted = ?", fromUserID, false).Find(&fromUser)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("sender user not found")
		} else if res.Error != nil {
			return fmt.Errorf("fail to get sender user: %w", res.Error)
		}

		if fromUser.Balance < float64(amount) {
			return fmt.Errorf("insufficient balance")
		}

		res = tx.Model(&repomodel.User{}).Where("id = ? and is_deleted = ?", fromUserID, false).Update("balance", fromUser.Balance-float64(amount))
		if res.Error != nil {
			return fmt.Errorf("fail to update balance fromUser: %w", res.Error)
		}

		res = tx.Model(&repomodel.User{}).Where("id = ? and is_deleted = ?", toUserID, false).Update("balance", toUser.Balance+float64(amount))
		if res.Error != nil {
			return fmt.Errorf("fail to update balance toUser: %w", res.Error)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("fail transfer balance: %w", err)
	}

	return nil
}
