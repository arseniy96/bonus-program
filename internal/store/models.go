package store

import "time"

const (
	AccrualType           = "accrual"
	WithdrawalType        = "withdrawal"
	OrderStatusNew        = "NEW"
	OrderStatusWithdrawn  = "WITHDRAWN"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusProcessed  = "PROCESSED"
	OrderStatusInvalid    = "INVALID"
)

type User struct {
	ID         int
	Login      string
	Password   string
	Token      string
	Bonuses    int
	TokenExpAt time.Time
}

type Order struct {
	ID          int
	OrderNumber string
	Status      string
	UserID      int
	BonusAmount int
	CreatedAt   time.Time
}

type BonusTransaction struct {
	ID          int
	Amount      int
	Type        string
	UserID      int
	OrderID     int
	OrderNumber string
	CreatedAt   time.Time
}
