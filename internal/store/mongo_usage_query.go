package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

type UsageDetail struct {
	InputTokens         int64 `bson:"inputTokens" json:"inputTokens"`
	OutputTokens        int64 `bson:"outputTokens" json:"outputTokens"`
	ReasoningTokens     int64 `bson:"reasoningTokens" json:"reasoningTokens"`
	CachedTokens        int64 `bson:"cachedTokens" json:"cachedTokens"`
	CacheReadTokens     int64 `bson:"cacheReadTokens" json:"cacheReadTokens"`
	CacheCreationTokens int64 `bson:"cacheCreationTokens" json:"cacheCreationTokens"`
	TotalTokens         int64 `bson:"totalTokens" json:"totalTokens"`
}

type UsageFailInfo struct {
	StatusCode int    `bson:"statusCode" json:"statusCode"`
	Body       string `bson:"body" json:"body"`
}

type MongoUsageRecord struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"userId" json:"userId"`
	Provider    string             `bson:"provider" json:"provider"`
	Model       string             `bson:"model" json:"model"`
	Alias       string             `bson:"alias" json:"alias"`
	APIKey      string             `bson:"apiKey" json:"apiKey"`
	AuthID      string             `bson:"authID" json:"authID"`
	AuthIndex   string             `bson:"authIndex" json:"authIndex"`
	AuthType    string             `bson:"authType" json:"authType"`
	Source      string             `bson:"source" json:"source"`
	RequestedAt time.Time          `bson:"requestedAt" json:"requestedAt"`
	LatencyMs   int64              `bson:"latencyMs" json:"latencyMs"`
	Failed      bool               `bson:"failed" json:"failed"`
	Fail        UsageFailInfo      `bson:"fail" json:"fail"`
	Detail      UsageDetail        `bson:"detail" json:"detail"`
	Cost        int64              `bson:"cost" json:"cost"`
}

type ListUsageInput struct {
	UserID   *primitive.ObjectID
	APIKey   string
	Page     int
	PageSize int
	StartDate *time.Time
	EndDate   *time.Time
}

type UsageListResult struct {
	Records []MongoUsageRecord `json:"records"`
	Total   int64              `json:"total"`
}

func (s *MongoStore) ListUsage(ctx context.Context, input ListUsageInput) (*UsageListResult, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}

	filter := bson.M{}
	if input.UserID != nil {
		filter["userId"] = input.UserID.Hex()
	}
	if input.APIKey != "" {
		filter["apiKey"] = input.APIKey
	}
	if input.StartDate != nil || input.EndDate != nil {
		dateFilter := bson.M{}
		if input.StartDate != nil {
			dateFilter["$gte"] = *input.StartDate
		}
		if input.EndDate != nil {
			dateFilter["$lte"] = *input.EndDate
		}
		filter["requestedAt"] = dateFilter
	}

	total, err := s.usageCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("mongodb: count usage: %w", err)
	}

	skip := int64(input.Page-1) * int64(input.PageSize)
	opts := options.Find().SetSkip(skip).SetLimit(int64(input.PageSize)).SetSort(bson.M{"requestedAt": -1})

	cursor, err := s.usageCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("mongodb: list usage: %w", err)
	}
	defer cursor.Close(ctx)

	var records []MongoUsageRecord
	for cursor.Next(ctx) {
		var r MongoUsageRecord
		if err = cursor.Decode(&r); err != nil {
			log.WithError(err).Warn("mongodb: decode usage record")
			continue
		}
		records = append(records, r)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("mongodb: cursor: %w", err)
	}

	return &UsageListResult{Records: records, Total: total}, nil
}

func (s *MongoStore) usageCollection() *mongo.Collection {
	return s.database.Collection("t_user_usage")
}
