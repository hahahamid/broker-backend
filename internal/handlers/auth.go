package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hahahamid/broker-backend/config"
	"github.com/hahahamid/broker-backend/internal/repository"
	"github.com/hahahamid/broker-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo repository.UserRepo
	cfg  *config.Config
}

func NewAuthHandler(r repository.UserRepo, c *config.Config) *AuthHandler {
	return &AuthHandler{repo: r, cfg: c}
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.CreateUser(context.Background(), req.Email, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.repo.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	at, rt, err := utils.GenerateTokens(user.ID.Hex(), h.cfg.JWTSecret, h.cfg.RefreshSecret, h.cfg.AccessTokenExpireMin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate tokens"})
		return
	}
	_ = h.repo.SaveRefreshToken(context.Background(), user.ID.Hex(), rt)
	c.JSON(http.StatusOK, gin.H{"access_token": at, "refresh_token": rt})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := utils.ValidateToken(req.RefreshToken, h.cfg.RefreshSecret)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)

	newAt, newRt, err := utils.GenerateTokens(userID, h.cfg.JWTSecret, h.cfg.RefreshSecret, h.cfg.AccessTokenExpireMin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate tokens"})
		return
	}
	_ = h.repo.SaveRefreshToken(context.Background(), userID, newRt)
	c.JSON(http.StatusOK, gin.H{"access_token": newAt, "refresh_token": newRt})
}
