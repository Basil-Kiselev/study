package repository

import (
	"context"
	"errors"
	"fmt"
	"study/Transaction/internal/model"
	"study/Transaction/internal/repository/mapper"
	repomodel "study/Transaction/internal/repository/model"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *Repository) GetTransactions(ctx context.Context, params model.GetTransactionsParams) ([]model.Transaction, error) {
	var transactions []repomodel.Transaction

	query := r.db.WithContext(ctx).Model(&repomodel.Transaction{})
	if params.UserID != nil {
		query = query.Where("user_id = ?", *params.UserID)
	}

	if params.Type != nil {
		query = query.Where("type = ?", *params.Type)
	}

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	if params.DateFrom != nil {
		query = query.Where("created_at >= ?", *params.DateFrom)
	}

	if params.DateTo != nil {
		query = query.Where("created_at <= ?", *params.DateTo)
	}

	res := query.
		Offset(params.Offset).
		Limit(params.Limit).
		Order("created_at DESC").
		Find(&transactions)

	if res.Error != nil {
		r.logger.Error().Err(res.Error).Msg("fail to get transactions")
		return nil, fmt.Errorf("fail to get transactions: %w", res.Error)
	}

	return mapper.RepoTransactionsToModel(transactions), nil
}

func (r *Repository) GetTransactionDetails(ctx context.Context, transactionID uint64) (model.TransactionDetails, error) {
	var transaction repomodel.Transaction
	var entries []repomodel.TransactionEntry

	res := r.db.WithContext(ctx).
		Model(&repomodel.Transaction{}).
		Where("id = ?", transactionID).
		First(&transaction)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return model.TransactionDetails{}, fmt.Errorf("transaction not found")
	} else if res.Error != nil {
		return model.TransactionDetails{}, fmt.Errorf("fail to find transaction: %w", res.Error)
	}

	res = r.db.WithContext(ctx).
		Model(&repomodel.TransactionEntry{}).
		Where("transaction_id = ?", transaction.ID).
		Find(&entries)
	if res.Error != nil {
		return model.TransactionDetails{}, fmt.Errorf("fail to find transaction entries: %w", res.Error)
	}

	return model.TransactionDetails{
		Transaction: mapper.RepoTranToModel(transaction),
		Entries:     mapper.RepoTrEntriesToModel(entries),
	}, nil
}

func (r *Repository) Deposit(ctx context.Context, params model.DepositParams) (model.TransactionDetails, error) {
	var result model.TransactionDetails

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transaction := model.Transaction{
			UserID: params.UserID,
			Amount: int64(params.Amount * 100), // в копейки
			Status: model.TransactionStatusPending,
			Type:   model.TransactionTypeDeposit,
		}

		repoTran := mapper.TransactionToRepo(transaction)

		res := tx.Table("transactions").Create(&repoTran)
		if res.Error != nil {
			r.logger.Error().Err(res.Error).Msg("fail to create transaction")
			return fmt.Errorf("fail to create transaction: %w", res.Error)
		}

		entry := model.TransactionEntry{
			TransactionID: repoTran.ID,
			AccountID:     repoTran.UserID,
			Direction:     model.TransactionEntryDirectionDebit,
			Amount:        float64(repoTran.Amount),
		}

		entryRepo := mapper.TransactionEntryToRepo(entry)
		res = tx.Table("transaction_entries").Create(&entryRepo)
		if res.Error != nil {
			r.logger.Error().Err(res.Error).Msg("fail to create transaction entry")
			return fmt.Errorf("fail to create transaction entry: %w", res.Error)
		}

		res = tx.Table("transactions").Where("id = ?", repoTran.ID).Update("status", model.TransactionStatusCompleted)
		if res.Error != nil {
			r.logger.Error().Err(res.Error).Msg("fail to update transaction status")
			return fmt.Errorf("fail to update transaction status: %w", res.Error)
		}

		result = model.TransactionDetails{
			Transaction: mapper.RepoTranToModel(repoTran),
			Entries:     mapper.RepoTrEntriesToModel([]repomodel.TransactionEntry{entryRepo}),
		}
		return nil
	})
	if err != nil {
		return model.TransactionDetails{}, fmt.Errorf("transaction failed: %w", err)
	}

	return result, nil
}

func (r *Repository) Transfer(ctx context.Context, params model.TransferParams) (model.TransactionDetails, error) {
	var result model.TransactionDetails

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transaction := model.Transaction{
			UserID: params.UserID,
			Amount: int64(params.Amount * 100),
			Type:   model.TransactionTypeTransfer,
			Status: model.TransactionStatusPending,
		}

		repoTran := mapper.TransactionToRepo(transaction)
		res := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&repoTran)
		if res.Error != nil {
			r.logger.Error().Err(res.Error).Msg("fail to create transaction in transfer")
			return fmt.Errorf("fail to create transaction: %w", res.Error)
		}

		entryCredit := model.TransactionEntry{
			TransactionID: repoTran.ID,
			AccountID:     repoTran.UserID,
			Direction:     model.TransactionEntryDirectionCredit,
			Amount:        params.Amount,
		}

		repoEntryCredit := mapper.TransactionEntryToRepo(entryCredit)
		res = tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&repoEntryCredit)
		if res.Error != nil {
			r.logger.Error().Err(res.Error).Msg("fail to create transaction entry credit in transfer")
			return fmt.Errorf("fail to create transaction entry: %w", res.Error)
		}

		entryDebit := model.TransactionEntry{
			TransactionID: repoTran.ID,
			AccountID:     params.Recipient,
			Direction:     model.TransactionEntryDirectionDebit,
			Amount:        params.Amount,
		}

		repoEntryDebit := mapper.TransactionEntryToRepo(entryDebit)
		res = tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&repoEntryDebit)
		if res.Error != nil {
			r.logger.Error().Err(res.Error).Msg("fail to create transaction entry debit in transfer")
			return fmt.Errorf("fail to create transaction entry: %w", res.Error)
		}

		updateData := mapper.UpdateTransactionToRepo(model.UpdateTransaction{Status: model.TransactionStatusCompleted})
		res = tx.Where("id = ?", repoTran.ID).Updates(updateData)
		if res.Error != nil {
			r.logger.Error().Err(res.Error).Msg("fail to update transaction in transfer")
			return fmt.Errorf("fail to update transaction: %w", res.Error)
		}

		result = model.TransactionDetails{
			Transaction: mapper.RepoTranToModel(repoTran),
			Entries: []model.TransactionEntry{
				mapper.RepoTransEntryToModel(repoEntryDebit),
				mapper.RepoTransEntryToModel(repoEntryCredit),
			},
		}

		return nil
	})

	if err != nil {
		return model.TransactionDetails{}, fmt.Errorf("transaction failed: %w", err)
	}

	return result, nil
}

func (r *Repository) Withdraw(ctx context.Context, params model.WithdrawParams) (model.TransactionDetails, error) {
	var result model.TransactionDetails

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transaction := model.Transaction{
			UserID: params.AccountID,
			Amount: int64(params.Amount * 100),
			Type:   model.TransactionTypeWithdraw,
			Status: model.TransactionStatusPending,
		}

		repoTr := mapper.TransactionToRepo(transaction)
		res := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&repoTr)
		if res.Error != nil {
			r.logger.Error().Err(res.Error).Msg("fail to create transaction in withdraw")
			return fmt.Errorf("fail to create transaction: %w", res.Error)
		}

		entry := model.TransactionEntry{
			TransactionID: repoTr.ID,
			AccountID:     repoTr.UserID,
			Amount:        params.Amount,
			Direction:     model.TransactionEntryDirectionCredit,
		}

		repoEntry := mapper.TransactionEntryToRepo(entry)
		res = tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&repoEntry)
		if res.Error != nil {
			r.logger.Error().Err(res.Error).Msg("fail to create transaction entry in withdraw")
			return fmt.Errorf("fail to create transaction entry: %w", res.Error)
		}

		updateData := mapper.UpdateTransactionToRepo(model.UpdateTransaction{Status: model.TransactionStatusCompleted})
		res = tx.Where("id = ?", repoTr.ID).Updates(updateData)
		if res.Error != nil {
			r.logger.Error().Err(res.Error).Msg("fail to update transaction in withdraw")
			return fmt.Errorf("fail to update transaction: %w", res.Error)
		}

		result = model.TransactionDetails{
			Transaction: mapper.RepoTranToModel(repoTr),
			Entries: []model.TransactionEntry{
				mapper.RepoTransEntryToModel(repoEntry),
			},
		}
		return nil
	})

	if err != nil {
		return model.TransactionDetails{}, err
	}

	return result, nil
}

func (r *Repository) UpdateTransactionsStatus(ctx context.Context, transactionID uint64, status model.TransactionStatus) error {
	res := r.db.WithContext(ctx).
		Model(&repomodel.Transaction{}).
		Where("id = ?", transactionID).
		Update("status", status)

	if res.Error != nil {
		return fmt.Errorf("fail to update transaction status")
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}
