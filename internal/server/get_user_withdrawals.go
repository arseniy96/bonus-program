package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/logger"
)

func (s *Server) GetUserWithdrawals(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	user, err := s.repository.FindUserByToken(ctx, authHeader)
	if err != nil {
		logger.Log.Errorf("find user error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	bonusTransactions, err := s.repository.FindBonusTransactionsByUserID(ctx, user.ID)
	if err != nil {
		logger.Log.Errorf("find bonus_transactions error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	var response GetUserWithdrawalsResponse
	for _, tr := range bonusTransactions {
		if tr.Type == "withdrawal" { // TODO: вынести в константу
			response = append(response, WithdrawalsResponse{
				Order:       tr.OrderNumber,
				Sum:         float64(tr.Amount) / 100,
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
