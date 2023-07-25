package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) PingHandler(c *gin.Context) {
	c.String(http.StatusOK, "PONG")
}
