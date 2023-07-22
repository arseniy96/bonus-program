package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/logger"
	"github.com/arseniy96/bonus-program/internal/store"
)

func (s *Server) UploadOrderHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	user, err := s.Repository.FindUserByToken(ctx, authHeader)
	if err != nil {
		logger.Log.Errorf("find user error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	orderNumber, err := io.ReadAll(c.Request.Body)
	if err != nil || len(orderNumber) == 0 {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("request error"))
		return
	}

	order, err := s.Repository.FindOrderByOrderNumber(ctx, string(orderNumber))
	if err != nil {
		if err == store.ErrNowRows {
			order, err = s.Repository.CreateOrder(ctx, user.ID, string(orderNumber), store.OrderStatusNew)
			if err != nil {
				logger.Log.Errorf("create order error: %v", err)
				c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("save order error %v", err))
				return
			}
			s.OrdersQueue <- order // кладём в очередь для фоновой обработки

			c.String(http.StatusAccepted, "order saved")
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	logger.Log.Infow("order already exists",
		"user_id", order.UserID,
		"order_number", order.OrderNumber)

	if order.UserID == user.ID {
		c.String(http.StatusOK, "order already exists")
		return
	}

	c.AbortWithError(http.StatusConflict, fmt.Errorf("order already exists"))
}
