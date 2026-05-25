package userapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/auth"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/store"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserHandler struct {
	store   *store.MongoStore
	keyGen  *auth.SnowflakeGenerator
}

func NewUserHandler(s *store.MongoStore) *UserHandler {
	return &UserHandler{
		store:  s,
		keyGen: auth.NewSnowflakeGenerator(time.Now().UnixMilli() % 1024),
	}
}

type CreateKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

type KeyResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Key        string    `json:"key"`
	Status     int       `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
}

func (h *UserHandler) ListKeys(c *gin.Context) {
	userIDStr, _ := c.Get("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	keys, err := h.store.GetUserKeysByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list keys"})
		return
	}

	resp := make([]KeyResponse, 0, len(keys))
	for _, k := range keys {
		resp = append(resp, KeyResponse{
			ID:         k.ID.Hex(),
			Name:       k.Name,
			Key:        k.Key,
			Status:     int(k.Status),
			CreatedAt:  k.CreatedAt,
			LastUsedAt: k.LastUsedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"keys": resp})
}

func (h *UserHandler) CreateKey(c *gin.Context) {
	var req CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userIDStr, _ := c.Get("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	existingKeys, err := h.store.GetUserKeysByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check key limit"})
		return
	}
	if len(existingKeys) >= 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key limit reached (max 5)"})
		return
	}

	apiKey := h.keyGen.Generate()

	key, err := h.store.CreateUserKey(c.Request.Context(), store.CreateUserKeyInput{
		UserID: userID,
		Name:   req.Name,
		Key:    apiKey,
		Status: store.UserKeyActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create key"})
		return
	}

	c.JSON(http.StatusCreated, KeyResponse{
		ID:         key.ID.Hex(),
		Name:       key.Name,
		Key:        apiKey,
		Status:     int(key.Status),
		CreatedAt:  key.CreatedAt,
		LastUsedAt: key.LastUsedAt,
	})
}

func (h *UserHandler) DeleteKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	userIDStr, _ := c.Get("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	existing, err := h.store.GetUserKeyByKey(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find key"})
		return
	}
	if existing == nil || existing.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	if err := h.store.DeleteUserKey(c.Request.Context(), key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key deleted"})
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func (h *UserHandler) ListAvailableModels(c *gin.Context) {
	aliases, err := h.store.GetAvailableModelAliases(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list available models"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"models": aliases})
}

func (h *UserHandler) GetBalance(c *gin.Context) {
	userIDStr, _ := c.Get("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.store.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": user.Balance})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be at least 6 characters"})
		return
	}

	account, _ := c.Get("account")

	user, err := h.store.GetUserByAccount(c.Request.Context(), account.(string))
	if err != nil {
		log.WithError(err).Error("change password: get user error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if !auth.CheckPassword(req.OldPassword, user.PasswordHash) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "old password is incorrect"})
		return
	}

	hashed, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		log.WithError(err).Error("change password: hash password error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if err := h.store.UpdateUserPassword(c.Request.Context(), account.(string), hashed); err != nil {
		log.WithError(err).Error("change password: update password error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
