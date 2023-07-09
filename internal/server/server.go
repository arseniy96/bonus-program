package server

import (
	"context"

	"github.com/arseniy96/bonus-program/internal/config"
)

type Server struct {
	repository Repository
	Config     *config.Settings
}

type Repository interface {
	CreateUser(context.Context, string, string) error
}

func NewServer(r Repository, c *config.Settings) *Server {
	return &Server{
		repository: r,
		Config:     c,
	}
}
