package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/logger"
)

func (s *Server) GetUserBalance(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	user, err := s.repository.FindUserByToken(ctx, authHeader)
	if err != nil {
		logger.Log.Errorf("find user error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	withdrawalSum, err := s.repository.GetWithdrawalSumByUserID(ctx, user.ID)
	if err != nil {
		logger.Log.Errorf("find bonus_transactions error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, GetUserBalanceResponse{
		Current:   float64(user.Bonuses) / 100,  // в БД храним в копейках (*100)
		Withdrawn: float64(withdrawalSum) / 100, // в БД храним в копейках (*100)
	})
}