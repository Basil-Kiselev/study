package service

import (
	"context"
	"encoding/json"
	"errors"
	"study/Transaction/internal/model"
	"study/Transaction/internal/service/mocks"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTransactionService_Deposit(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint64
		amount         float64
		setupMocks     func(*mocks.MockRepository, *mocks.MockKafkaPublisher)
		expectedErr    bool
		expectedResult model.TransactionDetails
	}{
		{
			name:   "successful deposit",
			userID: 1,
			amount: 100.50,
			setupMocks: func(mockRepo *mocks.MockRepository, mockKafka *mocks.MockKafkaPublisher) {
				expectedParams := model.DepositParams{
					UserID: 1,
					Amount: 100.50,
				}

				expectedResult := model.TransactionDetails{
					Transaction: model.Transaction{
						ID:        1,
						UserID:    1,
						Amount:    10050,
						Status:    model.TransactionStatusPending,
						Type:      model.TransactionTypeDeposit,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Entries: []model.TransactionEntry{},
				}

				mockRepo.EXPECT().
					Deposit(gomock.Any(), expectedParams).
					Return(expectedResult, nil)

				mockKafka.EXPECT().
					Publish(gomock.Any(), "transaction_data", "1", gomock.Any()).
					Return(nil)
			},
			expectedErr: false,
			expectedResult: model.TransactionDetails{
				Transaction: model.Transaction{
					ID:        1,
					UserID:    1,
					Amount:    10050,
					Status:    model.TransactionStatusPending,
					Type:      model.TransactionTypeDeposit,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Entries: []model.TransactionEntry{},
			},
		},
		{
			name:   "repository error",
			userID: 1,
			amount: 100.50,
			setupMocks: func(mr *mocks.MockRepository, mkp *mocks.MockKafkaPublisher) {
				expectedParams := model.DepositParams{
					UserID: 1,
					Amount: 100.50,
				}

				mr.EXPECT().
					Deposit(gomock.Any(), expectedParams).
					Return(model.TransactionDetails{}, errors.New("database error"))
			},
			expectedErr:    true,
			expectedResult: model.TransactionDetails{},
		},
		{
			name:   "kafka error",
			userID: 1,
			amount: 100.50,
			setupMocks: func(mr *mocks.MockRepository, mkp *mocks.MockKafkaPublisher) {
				expectedParams := model.DepositParams{
					UserID: 1,
					Amount: 100.50,
				}

				expectedResult := model.TransactionDetails{
					Transaction: model.Transaction{
						ID:        1,
						UserID:    1,
						Amount:    10050,
						Status:    model.TransactionStatusPending,
						Type:      model.TransactionTypeDeposit,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Entries: []model.TransactionEntry{},
				}

				mr.EXPECT().
					Deposit(gomock.Any(), expectedParams).
					Return(expectedResult, nil)

				mkp.EXPECT().
					Publish(gomock.Any(), "transaction_data", "1", gomock.Any()).
					Return(errors.New("kafka error"))

				mr.EXPECT().
					UpdateTransactionsStatus(gomock.Any(), uint64(1), model.TransactionStatusFailed).
					Return(nil)
			},
			expectedErr:    true,
			expectedResult: model.TransactionDetails{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockAccService := mocks.NewMockAccountService(ctrl)
			mockKafka := mocks.NewMockKafkaPublisher(ctrl)
			logger := zerolog.Nop()

			tt.setupMocks(mockRepo, mockKafka)

			service := New(mockRepo, &logger, mockAccService, mockKafka)
			res, err := service.Deposit(context.Background(), tt.userID, tt.amount)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Transaction.ID, res.Transaction.ID)
				assert.Equal(t, tt.expectedResult.Transaction.UserID, res.Transaction.UserID)
				assert.Equal(t, tt.expectedResult.Transaction.Amount, res.Transaction.Amount)
				assert.Equal(t, tt.expectedResult.Transaction.Status, res.Transaction.Status)
				assert.Equal(t, tt.expectedResult.Transaction.Type, res.Transaction.Type)
			}
		})
	}
}

func TestTransactionService_Withdraw(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint64
		amount         float64
		setupMocks     func(*mocks.MockRepository, *mocks.MockKafkaPublisher)
		expectedErr    bool
		expectedResult model.TransactionDetails
	}{
		{
			name:   "successful withdraw",
			userID: 1,
			amount: 50.50,
			setupMocks: func(mr *mocks.MockRepository, mkp *mocks.MockKafkaPublisher) {
				expectedParams := model.WithdrawParams{
					AccountID: 1,
					Amount:    50.50,
				}

				expectedResult := model.TransactionDetails{
					Transaction: model.Transaction{
						ID:        2,
						UserID:    1,
						Amount:    5050,
						Status:    model.TransactionStatusPending,
						Type:      model.TransactionTypeWithdraw,
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					},
					Entries: []model.TransactionEntry{},
				}

				mr.EXPECT().Withdraw(gomock.Any(), expectedParams).Return(expectedResult, nil)
				mkp.EXPECT().Publish(gomock.Any(), "transaction_data", "1", gomock.Any()).Return(nil)
			},
			expectedErr: false,
			expectedResult: model.TransactionDetails{
				Transaction: model.Transaction{
					ID:        2,
					UserID:    1,
					Amount:    5050,
					Status:    model.TransactionStatusPending,
					Type:      model.TransactionTypeWithdraw,
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				},
				Entries: []model.TransactionEntry{},
			},
		},
		{
			name:   "repo error",
			userID: 1,
			amount: 50.50,
			setupMocks: func(mr *mocks.MockRepository, mkp *mocks.MockKafkaPublisher) {
				expectedParams := model.WithdrawParams{
					AccountID: 1,
					Amount:    50.50,
				}

				mr.EXPECT().Withdraw(gomock.Any(), expectedParams).Return(model.TransactionDetails{}, errors.New("database error"))
			},
			expectedErr:    true,
			expectedResult: model.TransactionDetails{},
		},
		{
			name:   "kafka error",
			userID: 1,
			amount: 50.50,
			setupMocks: func(mr *mocks.MockRepository, mkp *mocks.MockKafkaPublisher) {
				expectedParams := model.WithdrawParams{
					AccountID: 1,
					Amount:    50.50,
				}

				expectedResult := model.TransactionDetails{
					Transaction: model.Transaction{
						ID:        2,
						UserID:    1,
						Amount:    5050,
						Type:      model.TransactionTypeWithdraw,
						Status:    model.TransactionStatusPending,
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					},
					Entries: []model.TransactionEntry{},
				}

				mr.EXPECT().Withdraw(gomock.Any(), expectedParams).Return(expectedResult, nil)
				mkp.EXPECT().Publish(gomock.Any(), "transaction_data", "1", gomock.Any()).Return(errors.New("kafka error"))
				mr.EXPECT().UpdateTransactionsStatus(gomock.Any(), uint64(2), model.TransactionStatusFailed).Return(nil)
			},
			expectedErr:    true,
			expectedResult: model.TransactionDetails{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := mocks.NewMockRepository(ctrl)
			mockKafka := mocks.NewMockKafkaPublisher(ctrl)
			mockAccService := mocks.NewMockAccountService(ctrl)

			logger := zerolog.Nop()

			tt.setupMocks(mockRepo, mockKafka)

			service := New(mockRepo, &logger, mockAccService, mockKafka)

			result, err := service.Withdraw(context.Background(), tt.userID, tt.amount)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Transaction.ID, result.Transaction.ID)
				assert.Equal(t, tt.expectedResult.Transaction.Status, result.Transaction.Status)
				assert.Equal(t, tt.expectedResult.Transaction.Type, result.Transaction.Type)
				assert.Equal(t, tt.expectedResult.Transaction.Amount, result.Transaction.Amount)
				assert.Equal(t, tt.expectedResult.Transaction.UserID, result.Transaction.UserID)
			}
		})
	}
}

func TestTransactionService_Transfer(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint64
		resipientID    uint64
		amount         float64
		setupMocks     func(*mocks.MockRepository, *mocks.MockKafkaPublisher)
		expectedErr    bool
		expectedResult model.TransactionDetails
	}{
		{
			name:        "successful transfer",
			userID:      1,
			resipientID: 2,
			amount:      20.0,
			setupMocks: func(mr *mocks.MockRepository, mkp *mocks.MockKafkaPublisher) {
				expectedParams := model.TransferParams{
					UserID:    1,
					Recipient: 2,
					Amount:    20.0,
				}

				expectedResult := model.TransactionDetails{
					Transaction: model.Transaction{
						ID:        3,
						UserID:    1,
						Amount:    2000,
						Status:    model.TransactionStatusPending,
						Type:      model.TransactionTypeTransfer,
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					},
					Entries: []model.TransactionEntry{},
				}

				mr.EXPECT().Transfer(gomock.Any(), expectedParams).Return(expectedResult, nil)
				mkp.EXPECT().Publish(gomock.Any(), "transaction_data", "1", gomock.Any()).Return(nil)
			},
			expectedErr: false,
			expectedResult: model.TransactionDetails{
				Transaction: model.Transaction{
					ID:        3,
					UserID:    1,
					Amount:    2000,
					Status:    model.TransactionStatusPending,
					Type:      model.TransactionTypeTransfer,
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				},
				Entries: []model.TransactionEntry{},
			},
		},
		{
			name:        "repo error",
			userID:      1,
			amount:      20.0,
			resipientID: 2,
			setupMocks: func(mr *mocks.MockRepository, mkp *mocks.MockKafkaPublisher) {
				expectedParams := model.TransferParams{
					UserID:    1,
					Recipient: 2,
					Amount:    20.0,
				}

				mr.EXPECT().Transfer(gomock.Any(), expectedParams).Return(model.TransactionDetails{}, errors.New("database error"))
			},
			expectedErr:    true,
			expectedResult: model.TransactionDetails{},
		},
		{
			name:        "kafka error",
			userID:      1,
			resipientID: 2,
			amount:      20.0,
			setupMocks: func(mr *mocks.MockRepository, mkp *mocks.MockKafkaPublisher) {
				expectedParams := model.TransferParams{
					UserID:    1,
					Recipient: 2,
					Amount:    20.0,
				}

				expectedResult := model.TransactionDetails{
					Transaction: model.Transaction{
						ID:        3,
						UserID:    1,
						Amount:    2000,
						Status:    model.TransactionStatusPending,
						Type:      model.TransactionTypeTransfer,
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					},
				}

				mr.EXPECT().Transfer(gomock.Any(), expectedParams).Return(expectedResult, nil)
				mkp.EXPECT().Publish(gomock.Any(), "transaction_data", "1", gomock.Any()).Return(errors.New("kafka error"))
				mr.EXPECT().UpdateTransactionsStatus(gomock.Any(), uint64(3), model.TransactionStatusFailed).Return(nil)
			},
			expectedErr:    true,
			expectedResult: model.TransactionDetails{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := mocks.NewMockRepository(ctrl)
			mockKafka := mocks.NewMockKafkaPublisher(ctrl)
			mockAccService := mocks.NewMockAccountService(ctrl)

			logger := zerolog.Nop()

			tt.setupMocks(mockRepo, mockKafka)

			service := New(mockRepo, &logger, mockAccService, mockKafka)
			res, err := service.Transfer(context.Background(), tt.userID, tt.amount, tt.resipientID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Transaction.ID, res.Transaction.ID)
				assert.Equal(t, tt.expectedResult.Transaction.UserID, res.Transaction.UserID)
				assert.Equal(t, tt.expectedResult.Transaction.Amount, res.Transaction.Amount)
				assert.Equal(t, tt.expectedResult.Transaction.Status, res.Transaction.Status)
				assert.Equal(t, tt.expectedResult.Transaction.Type, res.Transaction.Type)
			}
		})
	}
}

func TestTransactionService_HandleAccountResponse(t *testing.T) {
	tests := []struct {
		name        string
		response    AccountResponse
		setupMocks  func(*mocks.MockRepository)
		expectedErr bool
	}{
		{
			name: "successful response - transaction completed",
			response: AccountResponse{
				RequestType: "deposit",
				UserID:      1,
				OperationID: 1,
				Result:      true,
			},
			setupMocks: func(mr *mocks.MockRepository) {
				mr.EXPECT().UpdateTransactionsStatus(gomock.Any(), uint64(1), model.TransactionStatusCompleted).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "failed response - transaction failed",
			response: AccountResponse{
				RequestType: "withdraw",
				UserID:      1,
				OperationID: 2,
				Result:      false,
			},
			setupMocks: func(mr *mocks.MockRepository) {
				mr.EXPECT().UpdateTransactionsStatus(gomock.Any(), uint64(2), model.TransactionStatusFailed).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "repo error",
			response: AccountResponse{
				RequestType: "deposit",
				UserID:      1,
				OperationID: 1,
				Result:      true,
			},
			setupMocks: func(mr *mocks.MockRepository) {
				mr.EXPECT().UpdateTransactionsStatus(gomock.Any(), uint64(1), model.TransactionStatusCompleted).Return(errors.New("db error"))
			},
			expectedErr: true,
		},
		{
			name: "invalid response - missing request type",
			response: AccountResponse{
				RequestType: "",
				UserID:      1,
				OperationID: 1,
				Result:      true,
			},
			setupMocks: func(mr *mocks.MockRepository) {

			},
			expectedErr: true,
		},
		{
			name: "invalid response - missing user id",
			response: AccountResponse{
				RequestType: "deposit",
				UserID:      0,
				OperationID: 1,
				Result:      true,
			},
			setupMocks: func(mr *mocks.MockRepository) {

			},
			expectedErr: true,
		},
		{
			name: "invalid response - missing operation id",
			response: AccountResponse{
				RequestType: "deposit",
				UserID:      1,
				OperationID: 0,
				Result:      true,
			},
			setupMocks:  func(mr *mocks.MockRepository) {},
			expectedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crtl := gomock.NewController(t)
			defer crtl.Finish()
			mockRepo := mocks.NewMockRepository(crtl)
			mockAccService := mocks.NewMockAccountService(crtl)
			mockKafka := mocks.NewMockKafkaPublisher(crtl)

			logger := zerolog.Nop()
			tt.setupMocks(mockRepo)

			service := New(mockRepo, &logger, mockAccService, mockKafka)

			responseByte, err := json.Marshal(tt.response)
			assert.NoError(t, err)

			err = service.HandleAccountResponse(context.Background(), "test_topic", "test_key", responseByte)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransactionService_GetTransactionsWithDetails(t *testing.T) {
	tests := []struct {
		name          string
		params        model.GetTransactionsParams
		setupMocks    func(*mocks.MockRepository)
		expectedErr   bool
		expectedCount int
	}{
		{
			name: "successful get transactions",
			params: model.GetTransactionsParams{
				UserID: func() *uint64 { id := uint64(1); return &id }(),
				Limit:  10,
				Offset: 0,
			},
			setupMocks: func(mr *mocks.MockRepository) {
				transactions := []model.Transaction{
					{
						ID:        1,
						UserID:    1,
						Amount:    1000,
						Status:    model.TransactionStatusCompleted,
						Type:      model.TransactionTypeDeposit,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					{
						ID:        2,
						UserID:    1,
						Amount:    500,
						Status:    model.TransactionStatusCompleted,
						Type:      model.TransactionTypeWithdraw,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
				mr.EXPECT().GetTransactions(gomock.Any(), gomock.Any()).Return(transactions, nil)
				mr.EXPECT().GetTransactionDetails(gomock.Any(), uint64(1)).Return(model.TransactionDetails{
					Transaction: transactions[0],
				}, nil)
				mr.EXPECT().GetTransactionDetails(gomock.Any(), uint64(2)).Return(model.TransactionDetails{
					Transaction: transactions[1],
				}, nil)
			},
			expectedErr:   false,
			expectedCount: 2,
		},
		{
			name: "repo error",
			params: model.GetTransactionsParams{
				UserID: func() *uint64 { id := uint64(1); return &id }(),
				Limit:  10,
				Offset: 0,
			},
			setupMocks: func(mr *mocks.MockRepository) {
				mr.EXPECT().GetTransactions(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			expectedErr:   true,
			expectedCount: 0,
		},
		{
			name: "get transactions details error",
			params: model.GetTransactionsParams{
				UserID: func() *uint64 { id := uint64(1); return &id }(),
				Limit:  10,
				Offset: 0,
			},
			setupMocks: func(mr *mocks.MockRepository) {
				transactions := []model.Transaction{
					{
						ID:        1,
						UserID:    1,
						Amount:    1000,
						Status:    model.TransactionStatusCompleted,
						Type:      model.TransactionTypeDeposit,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
				mr.EXPECT().GetTransactions(gomock.Any(), gomock.Any()).Return(transactions, nil)
				mr.EXPECT().GetTransactionDetails(gomock.Any(), uint64(1)).Return(model.TransactionDetails{}, errors.New("details error"))
			},
			expectedErr:   false,
			expectedCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockAccService := mocks.NewMockAccountService(ctrl)
			mockKafka := mocks.NewMockKafkaPublisher(ctrl)
			logger := zerolog.Nop()

			tt.setupMocks(mockRepo)

			service := New(mockRepo, &logger, mockAccService, mockKafka)
			result, err := service.GetTransactionsWithDetails(context.Background(), tt.params)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}
