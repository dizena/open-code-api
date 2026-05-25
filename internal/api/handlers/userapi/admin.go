package userapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/auth"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminHandler struct {
	store             *store.MongoStore
	keyGen            *auth.SnowflakeGenerator
	refreshModelState func(context.Context) error
}

func NewAdminHandler(s *store.MongoStore, refreshModelState func(context.Context) error) *AdminHandler {
	return &AdminHandler{
		store:             s,
		keyGen:            auth.NewSnowflakeGenerator(time.Now().UnixMilli()%1024 + 1),
		refreshModelState: refreshModelState,
	}
}

func (h *AdminHandler) refreshModels(ctx context.Context) error {
	if h == nil || h.refreshModelState == nil {
		return nil
	}
	return h.refreshModelState(ctx)
}

// --- User Management ---

type CreateUserRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
	Balance  int64  `json:"balance"`
	Status   *int   `json:"status"`
}

type UpdateUserRequest struct {
	Role    *string `json:"role"`
	Balance *int64  `json:"balance"`
	Status  *int    `json:"status"`
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	input := store.ListUsersInput{
		Page:     1,
		PageSize: 20,
	}
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			input.Page = n
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 {
			input.PageSize = n
		}
	}
	if r := c.Query("role"); r != "" {
		role := store.UserRole(r)
		input.Role = &role
	}
	if s := c.Query("status"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			st := store.UserStatus(n)
			input.Status = &st
		}
	}

	result, err := h.store.ListUsers(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	type UserItem struct {
		ID       string    `json:"id"`
		Role     string    `json:"role"`
		Account  string    `json:"account"`
		Balance  int64     `json:"balance"`
		Status   int       `json:"status"`
		CreateAt time.Time `json:"createAt"`
		UpdateAt time.Time `json:"updateAt"`
	}

	items := make([]UserItem, 0, len(result.Users))
	for _, u := range result.Users {
		items = append(items, UserItem{
			ID:       u.ID.Hex(),
			Role:     string(u.Role),
			Account:  u.Account,
			Balance:  u.Balance,
			Status:   int(u.Status),
			CreateAt: u.CreateAt,
			UpdateAt: u.UpdateAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": items, "total": result.Total, "page": input.Page, "pageSize": input.PageSize})
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	role := store.UserRole(req.Role)
	if role == "" {
		role = store.RoleUser
	}
	status := store.UserStatusActive
	if req.Status != nil {
		status = store.UserStatus(*req.Status)
	}

	user, err := h.store.CreateUser(c.Request.Context(), store.CreateUserInput{
		Role:     role,
		Account:  strings.TrimSpace(req.Account),
		Password: hashed,
		Balance:  req.Balance,
		Status:   status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      user.ID.Hex(),
		"account": user.Account,
		"role":    string(user.Role),
		"status":  int(user.Status),
	})
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	account := c.Param("account")
	if account == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account is required"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	input := store.UpdateUserInput{
		Balance: req.Balance,
		Status:  ptrStatus(req.Status),
	}
	if req.Role != nil {
		r := store.UserRole(*req.Role)
		input.Role = &r
	}

	if err := h.store.UpdateUser(c.Request.Context(), account, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user updated"})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	account := c.Param("account")
	if account == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account is required"})
		return
	}

	if err := h.store.DeleteUser(c.Request.Context(), account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// --- Model Management ---

type CreateModelRequest struct {
	Type     string             `json:"type" binding:"required"`
	Name     string             `json:"name" binding:"required"`
	Priority int                `json:"priority"`
	BaseURL  string             `json:"baseUrl"`
	Models   []store.MongoModel `json:"models"`
	Headers  map[string]string  `json:"headers"`
	APIKey   []string           `json:"apiKey"`
}

type UpdateModelRequest struct {
	Type     *string             `json:"type"`
	Name     *string             `json:"name"`
	Priority *int                `json:"priority"`
	BaseURL  *string             `json:"baseUrl"`
	Models   *[]store.MongoModel `json:"models"`
	Headers  *map[string]string  `json:"headers"`
	APIKey   *[]string           `json:"apiKey"`
}

func (h *AdminHandler) ListModels(c *gin.Context) {
	input := store.ListModelsInput{
		Page:     1,
		PageSize: 50,
	}
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			input.Page = n
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 {
			input.PageSize = n
		}
	}
	if t := c.Query("type"); t != "" {
		input.Type = &t
	}

	result, err := h.store.ListModels(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list models"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": result.Models, "total": result.Total, "page": input.Page, "pageSize": input.PageSize})
}

func (h *AdminHandler) CreateModel(c *gin.Context) {
	var req CreateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	model, err := h.store.CreateModel(c.Request.Context(), store.CreateModelInput{
		Type:     req.Type,
		Name:     req.Name,
		Priority: req.Priority,
		BaseURL:  req.BaseURL,
		Models:   req.Models,
		Headers:  req.Headers,
		APIKey:   req.APIKey,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create model: " + err.Error()})
		return
	}
	if err := h.refreshModels(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model created but failed to refresh runtime: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model)
}

func (h *AdminHandler) UpdateModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var req UpdateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	input := store.UpdateModelInput{
		Type:     req.Type,
		Name:     req.Name,
		Priority: req.Priority,
		BaseURL:  req.BaseURL,
		Models:   req.Models,
		Headers:  req.Headers,
		APIKey:   req.APIKey,
	}

	if err := h.store.UpdateModel(c.Request.Context(), id, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update model: " + err.Error()})
		return
	}
	if err := h.refreshModels(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model updated but failed to refresh runtime: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "model updated"})
}

func (h *AdminHandler) DeleteModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	if err := h.store.DeleteModel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete model: " + err.Error()})
		return
	}
	if err := h.refreshModels(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model deleted but failed to refresh runtime: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "model deleted"})
}

// --- Admin Key Management ---

type CreateAdminKeyRequest struct {
	Account string `json:"account" binding:"required"`
	Name    string `json:"name" binding:"required"`
}

func (h *AdminHandler) ListUserKeys(c *gin.Context) {
	account := c.Query("account")

	input := store.ListUserKeysInput{
		Account:  account,
		Page:     1,
		PageSize: 20,
	}
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			input.Page = n
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 {
			input.PageSize = n
		}
	}
	if s := c.Query("start"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
			input.StartDate = &start
		}
	}
	if e := c.Query("end"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local)
			input.EndDate = &end
		}
	}

	result, err := h.store.ListUserKeysByAccount(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list keys: " + err.Error()})
		return
	}

	resp := make([]KeyResponse, 0, len(result.Keys))
	for _, k := range result.Keys {
		resp = append(resp, KeyResponse{
			ID:         k.ID.Hex(),
			Name:       k.Name,
			Key:        k.Key,
			Status:     int(k.Status),
			CreatedAt:  k.CreatedAt,
			LastUsedAt: k.LastUsedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"keys": resp, "total": result.Total, "page": input.Page, "pageSize": input.PageSize})
}

func (h *AdminHandler) CreateKeyForUser(c *gin.Context) {
	var req CreateAdminKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.store.GetUserByAccount(c.Request.Context(), req.Account)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	apiKey := h.keyGen.Generate()

	key, err := h.store.CreateUserKey(c.Request.Context(), store.CreateUserKeyInput{
		UserID: user.ID,
		Name:   req.Name,
		Key:    apiKey,
		Status: store.UserKeyActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   key.ID.Hex(),
		"key":  apiKey,
		"name": key.Name,
	})
}

func (h *AdminHandler) DeleteUserKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	if err := h.store.DeleteUserKey(c.Request.Context(), key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete key: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key deleted"})
}

func (h *AdminHandler) DisableUserKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	if err := h.store.UpdateUserKeyStatus(c.Request.Context(), key, store.UserKeyDisabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disable key: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key disabled"})
}

// --- Admin Usage Query ---

func (h *AdminHandler) ListUsage(c *gin.Context) {
	input := store.ListUsageInput{
		Page:     1,
		PageSize: 20,
	}

	if account := c.Query("account"); account != "" {
		user, err := h.store.GetUserByAccount(c.Request.Context(), account)
		if err != nil || user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		input.UserID = &user.ID
	}

	if s := c.Query("start"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
			input.StartDate = &start
		}
	}
	if e := c.Query("end"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local)
			input.EndDate = &end
		}
	}
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			input.Page = n
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 {
			input.PageSize = n
		}
	}

	result, err := h.store.ListUsage(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list usage"})
		return
	}

	type UsageItem struct {
		UserID          string    `json:"userId"`
		APIKey          string    `json:"apiKey"`
		Alias           string    `json:"alias"`
		Model           string    `json:"model"`
		Provider        string    `json:"provider"`
		LatencyMs       int64     `json:"latencyMs"`
		InputTokens     int64     `json:"inputTokens"`
		OutputTokens    int64     `json:"outputTokens"`
		ReasoningTokens int64     `json:"reasoningTokens"`
		TotalTokens     int64     `json:"totalTokens"`
		Failed          bool      `json:"failed"`
		Cost            int64     `json:"cost"`
		RequestedAt     time.Time `json:"requestedAt"`
	}

	items := make([]UsageItem, 0, len(result.Records))
	for _, r := range result.Records {
		items = append(items, UsageItem{
			UserID:          r.UserID,
			APIKey:          maskKey(r.APIKey),
			Alias:           r.Alias,
			Model:           r.Model,
			Provider:        r.Provider,
			LatencyMs:       r.LatencyMs,
			InputTokens:     r.Detail.InputTokens,
			OutputTokens:    r.Detail.OutputTokens,
			ReasoningTokens: r.Detail.ReasoningTokens,
			TotalTokens:     r.Detail.TotalTokens,
			Failed:          r.Failed,
			Cost:            r.Cost,
			RequestedAt:     r.RequestedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"records":  items,
		"total":    result.Total,
		"page":     input.Page,
		"pageSize": input.PageSize,
	})
}
// --- Alias Management ---

type AliasItem struct {
	DocID    string  `json:"docId"`
	Provider string  `json:"provider"`
	URL      string  `json:"url"`
	Name     string  `json:"name"`
	Alias    string  `json:"alias"`
	Rate     float64 `json:"rate"`
	Index    int     `json:"index"`
}

func (h *AdminHandler) ListAliases(c *gin.Context) {
	result, err := h.store.ListModels(c.Request.Context(), store.ListModelsInput{Page: 1, PageSize: 100, SortBy: "alias", SortOrder: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list models"})
		return
	}

	var items []AliasItem
	for _, doc := range result.Models {
		for idx, m := range doc.Models {
			items = append(items, AliasItem{
				DocID:    doc.ID.Hex(),
				Provider: doc.Type,
				URL:      doc.BaseURL,
				Name:     m.Name,
				Alias:    m.Alias,
				Rate:     m.Rate,
				Index:    idx,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"aliases": items, "total": len(items)})
}

type UpdateAliasRequest struct {
	Index *int    `json:"index"`
	Alias string  `json:"alias" binding:"required"`
	Rate  float64 `json:"rate"`
}

func (h *AdminHandler) UpdateModelAlias(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var req UpdateAliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	idx := 0
	if req.Index != nil {
		idx = *req.Index
	}

	if err := h.store.UpdateModelAlias(c.Request.Context(), id, idx, req.Alias, req.Rate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update alias: " + err.Error()})
		return
	}
	if err := h.refreshModels(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "alias updated but failed to refresh runtime: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alias updated"})
}


func ptrStatus(s *int) *store.UserStatus {
	if s == nil {
		return nil
	}
	st := store.UserStatus(*s)
	return &st
}
