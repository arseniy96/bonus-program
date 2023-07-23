package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/logger"
	"github.com/arseniy96/bonus-program/internal/services/converter"
	"github.com/arseniy96/bonus-program/internal/services/mycrypto"
)

func (s *Server) WithdrawHandler(c *gin.Context) {
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

	var body WithdrawRequest
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if body.Sum > converter.ConvertFromCent(user.Bonuses) {
		c.AbortWithError(http.StatusPaymentRequired, fmt.Errorf("insufficient funds"))
		return
	}
	if body.Sum == 0 {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid amount"))
		return
	}

	err = s.Repository.SaveWithdrawBonuses(ctx, user.ID, body.Order, converter.ConvertToCent(body.Sum))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
