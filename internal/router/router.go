package router

import (
	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/middlewares"
	"github.com/arseniy96/bonus-program/internal/server"
)

type Router interface {
	Run(addr ...string) error
}

func NewRouter(s *server.Server) Router {
	g := gin.Default()
	g.Use(middlewares.AuthMiddleware())
	// TODO: написать миддлварю, которая логгирует запрос/ответ
	g.GET("/ping", s.PingHandler)
	g.POST("/api/user/register", s.SignUp)
	g.POST("/api/user/login", s.Login)
	g.GET("/api/user/orders", s.GetOrders)
	g.GET("/api/user/balance", s.GetUserBalance)
	return g
}
