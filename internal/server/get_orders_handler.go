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

func (s *Server) GetOrders(c *gin.Context) {
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
	orders, err := s.Repository.FindOrdersByUserID(ctx, user.ID)
	if err != nil {
		logger.Log.Errorf("find orders error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if len(orders) == 0 {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	var response GetOrdersResponse
	for _, order := range orders {
		orderResp := OrderResponse{
			Number:     order.OrderNumber,
			Status:     order.Status,
			UploadedAt: order.CreatedAt.Format(time.RFC3339),
		}

		ac := converter.ConvertFromCent(order.BonusAmount) // FIXME: Не знаю, как сделать по-другому
		if ac != 0 {
			orderResp.Accrual = ac
		}

		response = append(response, orderResp)
	}
	c.JSON(http.StatusOK, response)
}
