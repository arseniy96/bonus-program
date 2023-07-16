package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"github.com/arseniy96/bonus-program/internal/server"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == `/api/user/register` || path == `/api/user/login` {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if err := checkAuthHeader(authHeader); err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		c.Next()
	}
}

func checkAuthHeader(authHeader string) error {
	if len(authHeader) == 0 {
		return fmt.Errorf("missing Authorization header")
	}

	// TODO: насколько ОК из миддлвари обращаться к сервису, чтобы что-то подтянуть?
	// 		 Кажется, такое надо переносить в отдельное место
	claims := &server.Claims{}
	token, err := jwt.ParseWithClaims(authHeader, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(server.SecretKey), nil
		})
	if err != nil {
		return fmt.Errorf("parse jwt error: %v", err)
	}
	if !token.Valid {
		return fmt.Errorf("token is not valid")
	}
	// TODO: Проверять токен в БД

	return nil
}
