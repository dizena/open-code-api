package userapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/auth"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/mail"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/store"
	log "github.com/sirupsen/logrus"
)

type AuthHandler struct {
	store      *store.MongoStore
	secret     string
	mailSvc    *mail.Service
}

func NewAuthHandler(s *store.MongoStore, secret string, mailSvc *mail.Service) *AuthHandler {
	return &AuthHandler{store: s, secret: secret, mailSvc: mailSvc}
}

type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Account string `json:"account"`
	Role    string `json:"role"`
	UserID  string `json:"userId"`
	Balance int64  `json:"balance"`
}

type SendCodeRequest struct {
	Account string `json:"account" binding:"required"`
	Purpose string `json:"purpose"` // "register" or "reset_password"
}

type VerifyCodeRequest struct {
	Account string `json:"account" binding:"required"`
	Code    string `json:"code" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	user, err := h.store.GetUserByAccount(c.Request.Context(), strings.TrimSpace(req.Account))
	if err != nil {
		log.WithError(err).Error("auth: get user error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid account or password"})
		return
	}
	if user.Status == store.UserStatusDisabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		return
	}
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid account or password"})
		return
	}

	token, err := auth.GenerateToken(h.secret, user.ID.Hex(), user.Account, string(user.Role))
	if err != nil {
		log.WithError(err).Error("auth: generate token error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token:   token,
		Account: user.Account,
		Role:    string(user.Role),
		UserID:  user.ID.Hex(),
		Balance: user.Balance,
	})
}

func (h *AuthHandler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	account := strings.TrimSpace(req.Account)
	if account == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account is required"})
		return
	}

	// Check if user already exists
	user, err := h.store.GetUserByAccount(c.Request.Context(), account)
	if err != nil {
		log.WithError(err).Error("auth: get user error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// If user exists and purpose is register, return error
	if user != nil && req.Purpose == "register" {
		c.JSON(http.StatusConflict, gin.H{"error": "account already exists", "exists": true})
		return
	}

	// Generate 8-digit code
	code, err := mail.GenerateCode()
	if err != nil {
		log.WithError(err).Error("auth: generate code error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Send email
	if h.mailSvc != nil && h.mailSvc.IsEnabled() {
		if err := h.mailSvc.SendVerificationCode(account, code); err != nil {
			log.WithError(err).Error("auth: send code error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send verification code"})
			return
		}
	} else {
		// Fallback: log code for development
		log.Warnf("auth: SMTP not configured, verification code for %s: %s", account, code)
	}

	// Save code to t_user_check with 10 minute expiry
	expiresAt := time.Now().Add(10 * time.Minute)
	if _, err := h.store.SaveVerificationCode(c.Request.Context(), account, code, expiresAt); err != nil {
		log.WithError(err).Error("auth: save code error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "verification code sent"})
}

func (h *AuthHandler) VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	account := strings.TrimSpace(req.Account)
	code := strings.TrimSpace(req.Code)

	// Verify code
	validCode, err := h.store.GetValidVerificationCode(c.Request.Context(), account, code)
	if err != nil {
		log.WithError(err).Error("auth: verify code error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if validCode == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired verification code"})
		return
	}

	// Check if user exists
	user, err := h.store.GetUserByAccount(c.Request.Context(), account)
	if err != nil {
		log.WithError(err).Error("auth: get user error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if user == nil {
		// Register new user
		hashed, err := auth.HashPassword(code)
		if err != nil {
			log.WithError(err).Error("auth: hash password error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		user, err = h.store.CreateUser(c.Request.Context(), store.CreateUserInput{
			Role:     store.RoleUser,
			Account:  account,
			Password: hashed,
			Balance:  0,
			Status:   store.UserStatusActive,
		})
		if err != nil {
			log.WithError(err).Error("auth: create user error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
			return
		}
	} else {
		// Reset password
		hashed, err := auth.HashPassword(code)
		if err != nil {
			log.WithError(err).Error("auth: hash password error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		if err := h.store.UpdateUserPassword(c.Request.Context(), account, hashed); err != nil {
			log.WithError(err).Error("auth: update password error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
			return
		}
	}

	// Generate token
	token, err := auth.GenerateToken(h.secret, user.ID.Hex(), user.Account, string(user.Role))
	if err != nil {
		log.WithError(err).Error("auth: generate token error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token:   token,
		Account: user.Account,
		Role:    string(user.Role),
		UserID:  user.ID.Hex(),
		Balance: user.Balance,
	})
}

func (h *AuthHandler) CheckAccount(c *gin.Context) {
	account := c.Query("account")
	if account == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account is required"})
		return
	}

	user, err := h.store.GetUserByAccount(c.Request.Context(), strings.TrimSpace(account))
	if err != nil {
		log.WithError(err).Error("auth: get user error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if user != nil {
		c.JSON(http.StatusOK, gin.H{"exists": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"exists": false})
	}
}
