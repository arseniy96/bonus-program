package server

import (
	"context"
	"time"

	"github.com/arseniy96/bonus-program/internal/config"
	"github.com/arseniy96/bonus-program/internal/store"
)

type Server struct {
	Repository  Repository
	Config      *config.Settings
	OrdersQueue chan *store.Order
}

type Repository interface {
	CreateUser(context.Context, string, string) error
	UpdateUserToken(context.Context, string, string, time.Time) error
	FindUserByLogin(context.Context, string) (*store.User, error)
	FindUserByToken(context.Context, string) (*store.User, error)
	FindOrdersByUserID(context.Context, int) ([]store.Order, error)
	FindBonusTransactionsByUserID(context.Context, int) ([]store.BonusTransaction, error)
	GetWithdrawalSumByUserID(context.Context, int) (int, error)
	SaveWithdrawBonuses(context.Context, int, string, float64) error
	FindOrderByOrderNumber(context.Context, string) (*store.Order, error)
	CreateOrder(context.Context, int, string, string) (*store.Order, error)
	UpdateOrderStatus(context.Context, *store.Order, string, int) error
}

func NewServer(r Repository, c *config.Settings) *Server {
	server := &Server{
		Repository:  r,
		Config:      c,
		OrdersQueue: make(chan *store.Order, 10),
	}

	go server.OrdersWorker()

	return server
}
