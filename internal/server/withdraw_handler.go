package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/logger"
)

func (s *Server) WithdrawHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	user, err := s.repository.FindUserByToken(ctx, authHeader)
	if err != nil {
		logger.Log.Errorf("find user error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var body WithdrawRequest
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if body.Sum > float64(user.Bonuses)/100 {
		c.AbortWithError(http.StatusPaymentRequired, fmt.Errorf("insufficient funds"))
		return
	}

	err = s.repository.SaveWithdrawBonuses(ctx, user.ID, body.Order, body.Sum)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}