package userapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *UserHandler) ListUsage(c *gin.Context) {
	userIDStr, _ := c.Get("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	input := store.ListUsageInput{
		UserID:   &userID,
		Page:     1,
		PageSize: 20,
	}

	// Default to today
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	input.StartDate = &startOfDay
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	input.EndDate = &endOfDay

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
		APIKey          string    `json:"apiKey"`
		Alias           string    `json:"alias"`
		LatencyMs       int64     `json:"latencyMs"`
		InputTokens     int64     `json:"inputTokens"`
		OutputTokens    int64     `json:"outputTokens"`
		ReasoningTokens int64     `json:"reasoningTokens"`
		Cost            int64     `json:"cost"`
		RequestedAt     time.Time `json:"requestedAt"`
	}

	items := make([]UsageItem, 0, len(result.Records))
	for _, r := range result.Records {
		items = append(items, UsageItem{
			APIKey:          maskKey(r.APIKey),
			Alias:           r.Alias,
			LatencyMs:       r.LatencyMs,
			InputTokens:     r.Detail.InputTokens,
			OutputTokens:    r.Detail.OutputTokens,
			ReasoningTokens: r.Detail.ReasoningTokens,
			Cost:            r.Cost,
			RequestedAt:     r.RequestedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"records": items,
		"total":   result.Total,
		"page":    input.Page,
		"pageSize": input.PageSize,
	})
}
