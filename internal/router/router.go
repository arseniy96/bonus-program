package router

import (
	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/server"
)

type Router interface {
	Run(addr ...string) error
}

func NewRouter(s *server.Server) Router {
	g := gin.Default()
	// TODO: написать миддлварю, которая логгирует запрос/ответ
	g.GET("/ping", s.PingHandler)
	return g
}
