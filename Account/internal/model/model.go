package model

import "time"

type OperationType string

const (
	OperationTypeDeposit OperationType = "deposit"
	OperationTypeCredit  OperationType = "credit"
)

type User struct {
	ID         uint64
	Login      string
	Email      string
	Phone      string
	FirstName  string
	LastName   string
	MiddleName string
	Age        uint32
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Balance    float64
	IsDeleted  bool
}

type CreateUser struct {
	Login      string
	Email      string
	Phone      string
	FirstName  string
	LastName   string
	MiddleName string
	Age        uint32
	Balance    float64
	IsDeleted  bool
}

type UpdateUser struct {
	Email      string
	Phone      string
	FirstName  string
	LastName   string
	MiddleName string
	Age        uint32
	Balance    float64
	IsDeleted  bool
}

type GetBalanceResponse struct {
	UserID  uint64
	Balance float32
}

type UpdateBalanceRequest struct {
	UserID uint64
	Amount float32
	Type   OperationType
}

type UpdateBalanceResponse struct {
	UserID     uint64
	OldBalance float32
	NewBalance float32
	Amount     float32
	Type       OperationType
}
