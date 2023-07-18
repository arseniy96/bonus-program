package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/logger"
)

func (s *Server) GetOrders(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	user, err := s.repository.FindUserByToken(ctx, authHeader)
	if err != nil {
		logger.Log.Errorf("find user error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	orders, err := s.repository.FindOrdersByUserID(ctx, user.ID)
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
			Accrual:    float64(order.BonusAmount) / 100,
			UploadedAt: order.CreatedAt.Format(time.RFC3339),
		}
		response = append(response, orderResp)
	}
	c.JSON(http.StatusOK, response)
}
