package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/logger"
	"github.com/arseniy96/bonus-program/internal/services/converter"
	"github.com/arseniy96/bonus-program/internal/services/mycrypto"
)

func (s *Server) GetUserBalance(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	token := mycrypto.HashFunc(authHeader)
	user, err := s.Repository.FindUserByToken(ctx, token)
	if err != nil {
		logger.Log.Errorf("find user error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	withdrawalSum, err := s.Repository.GetWithdrawalSumByUserID(ctx, user.ID)
	if err != nil {
		logger.Log.Errorf("find bonus_transactions error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, GetUserBalanceResponse{
		Current:   converter.ConvertFromCent(user.Bonuses),  // в БД храним в копейках
		Withdrawn: converter.ConvertFromCent(withdrawalSum), // в БД храним в копейках
	})
}
