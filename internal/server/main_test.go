package server

import (
	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/services/mycrypto"
)

func SetUpRouter() *gin.Engine {
	router := gin.Default()
	return router
}

var (
	allowedToken      = "some_auth_token"
	allowedToken2     = "some_auth_token_zero_orders"
	wrongToken        = "wrong_auth_token"
	allowedTokenHash  = mycrypto.HashFunc(allowedToken)
	allowedToken2Hash = mycrypto.HashFunc(allowedToken2)
	wrongTokenHash    = mycrypto.HashFunc(wrongToken)
)
