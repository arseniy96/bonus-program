package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/logger"
	"github.com/arseniy96/bonus-program/internal/services/converter"
	"github.com/arseniy96/bonus-program/internal/services/mycrypto"
	"github.com/arseniy96/bonus-program/internal/store"
)

func (s *Server) GetUserWithdrawals(c *gin.Context) {
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

	bonusTransactions, err := s.Repository.FindBonusTransactionsByUserID(ctx, user.ID)
	if err != nil {
		logger.Log.Errorf("find bonus_transactions error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	var response GetUserWithdrawalsResponse
	for _, tr := range bonusTransactions {
		if tr.Type == store.WithdrawalType {
			response = append(response, WithdrawalsResponse{
				Order:       tr.OrderNumber,
				Sum:         converter.ConvertFromCent(tr.Amount),
				ProcessedAt: tr.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	if len(response) == 0 {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	c.JSON(http.StatusOK, response)
}
