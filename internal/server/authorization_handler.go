package server

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"github.com/arseniy96/bonus-program/internal/logger"
	"github.com/arseniy96/bonus-program/internal/server/models"
	"github.com/arseniy96/bonus-program/internal/store"
)

const (
	SecretKey = "8ha37nlpa4"
	TokenExp  = time.Hour * 24
)

type Claims struct {
	jwt.RegisteredClaims
	Login string
}

func (s *Server) SignUp(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var body models.SignUpRequest
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	hPassword := hashPassword(body.Password)
	if err := s.repository.CreateUser(ctx, body.Login, hPassword); err != nil {
		if errors.Is(err, store.ErrConflict) {
			c.AbortWithError(http.StatusConflict, fmt.Errorf("user already exists"))
			return
		}
		logger.Log.Errorf("create user error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	token, err := buildJWTString(body.Login)
	if err != nil {
		logger.Log.Errorf("build json web token error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if err := s.repository.UpdateUserToken(ctx, body.Login, token); err != nil {
		logger.Log.Errorf("update user error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Authorization", token)
	c.JSON(http.StatusOK, gin.H{"login": "success"})
}

func (s *Server) Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var body models.LoginRequest
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	hPassword := hashPassword(body.Password)
	user, err := s.repository.FindUserByLogin(ctx, body.Login)
	if err != nil {
		if errors.Is(err, store.ErrInvalidLogin) {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if user.Password != hPassword {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid password"))
		return
	}

	token, err := buildJWTString(body.Login)
	if err != nil {
		logger.Log.Errorf("build json web token error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if err := s.repository.UpdateUserToken(ctx, body.Login, token); err != nil {
		logger.Log.Errorf("update user error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Authorization", token)
	c.JSON(http.StatusOK, gin.H{"login": "success"})
}

func hashPassword(password string) string {
	// не будем усложнять – просто возьмём хэш 1 раз
	initString := fmt.Sprintf("%v:%v", password, SecretKey)

	return fmt.Sprintf("%x", md5.Sum([]byte(initString)))
}

func buildJWTString(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		Login: login,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}