package server

import (
	"context"

	"github.com/arseniy96/bonus-program/internal/config"
	"github.com/arseniy96/bonus-program/internal/store"
)

type Server struct {
	repository  Repository
	Config      *config.Settings
	OrdersQueue chan OrderWithTTL
}

type Repository interface {
	CreateUser(context.Context, string, string) error
	UpdateUserToken(context.Context, string, string) error
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
		repository:  r,
		Config:      c,
		OrdersQueue: make(chan OrderWithTTL, 1),
	}

	go server.OrdersWorker()

	return server
}
