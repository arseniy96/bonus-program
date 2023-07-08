package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Router interface {
	Run(addr ...string) error
}

func NewRouter() Router {
	g := gin.Default()
	// TODO: написать миддлварю, которая логгирует запрос/ответ
	g.GET("/", mock)
	return g
}

func mock(c *gin.Context) {
	cookie, err := c.Cookie("test")
	if err != nil {
		c.Writer.Write([]byte("missing cookie"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"cookie": cookie})
}
