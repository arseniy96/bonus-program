package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"github.com/arseniy96/bonus-program/internal/logger"
	"github.com/arseniy96/bonus-program/internal/services/mycrypto"
	"github.com/arseniy96/bonus-program/internal/store"
)

const (
	TokenExp = time.Hour * 24
)

type Claims struct {
	jwt.RegisteredClaims
	Login string
}

func (s *Server) SignUp(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var body SignUpRequest
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	hPass := mycrypto.HashFunc(body.Password)
	if err := s.Repository.CreateUser(ctx, body.Login, hPass); err != nil {
		if errors.Is(err, store.ErrConflict) {
			c.AbortWithError(http.StatusConflict, fmt.Errorf("user already exists"))
			return
		}
		logger.Log.Errorf("create user error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	token, err := mycrypto.CreateRandomToken(16)
	if err != nil {
		logger.Log.Errorf("build token error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	hashToken := mycrypto.HashFunc(token)
	tokenExp := time.Now().Add(TokenExp)
	if err := s.Repository.UpdateUserToken(ctx, body.Login, hashToken, tokenExp); err != nil {
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

	var body LoginRequest
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	hPass := mycrypto.HashFunc(body.Password)
	user, err := s.Repository.FindUserByLogin(ctx, body.Login)
	if err != nil {
		if errors.Is(err, store.ErrNowRows) {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if user.Password != hPass {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid password"))
		return
	}

	token, err := mycrypto.CreateRandomToken(16)
	if err != nil {
		logger.Log.Errorf("build token error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	hashToken := mycrypto.HashFunc(token)
	tokenExp := time.Now().Add(TokenExp)
	if err := s.Repository.UpdateUserToken(ctx, body.Login, hashToken, tokenExp); err != nil {
		logger.Log.Errorf("update user error: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Authorization", token)
	c.JSON(http.StatusOK, gin.H{"login": "success"})
}
