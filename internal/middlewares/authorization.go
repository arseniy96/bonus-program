package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/arseniy96/bonus-program/internal/logger"
	"github.com/arseniy96/bonus-program/internal/services/mycrypto"
	"github.com/arseniy96/bonus-program/internal/store"
)

type repository interface {
	FindUserByToken(context.Context, string) (*store.User, error)
}

func AuthMiddleware(r repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == `/api/user/register` || path == `/api/user/login` {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if err := checkHeader(r, authHeader); err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		c.Next()
	}
}

func checkHeader(r repository, header string) error {
	if len(header) == 0 {
		return fmt.Errorf("missing Authorization header")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	token := mycrypto.HashFunc(header)

	user, err := r.FindUserByToken(ctx, token)
	if err != nil {
		logger.Log.Errorf("find user error: %v", err)
		return fmt.Errorf("invalid token")
	}

	if user.TokenExpAt.Before(time.Now()) {
		return fmt.Errorf("token expired")
	}
	return nil
}
