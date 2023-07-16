package store

import "time"

type User struct {
	ID       int
	Login    string
	Password string
	Token    string
	Bonuses  int
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
