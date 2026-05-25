package store

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/usage"
	log "github.com/sirupsen/logrus"
)

// MongoUsageStore implements usage.Plugin and stores usage records into MongoDB t_user_usage,
// enriching each record with a userId looked up from t_user_key by API key.
type MongoUsageStore struct {
	client     *mongo.Client
	database   *mongo.Database
	userKeyCol *mongo.Collection
	usageCol   *mongo.Collection
	userCol    *mongo.Collection
	modelCol   *mongo.Collection
	enabled    bool
	// cache maps API key to userId string, with expiry.
	cacheMu   sync.RWMutex
	cache     map[string]cachedUserId
}

// cachedUserId holds a userId and its expiry time.
type cachedUserId struct {
	userId   string
	expires  time.Time
}

// NewMongoUsageStore creates a new MongoUsageStore that reuses the MongoStore's
// MongoDB connection. If ms is nil or disabled, returns nil.
func NewMongoUsageStore(ms *MongoStore) (*MongoUsageStore, error) {
	if ms == nil || !ms.enabled {
		return nil, nil
	}

	db := ms.database
	store := &MongoUsageStore{
		client:     ms.client,
		database:   db,
		userKeyCol: db.Collection("t_user_key"),
		usageCol:   db.Collection("t_user_usage"),
		userCol:    db.Collection("t_user"),
		modelCol:   db.Collection("t_model"),
		enabled:    true,
		cache:      make(map[string]cachedUserId),
	}
	return store, nil
}

// HandleUsage implements usage.Plugin.HandleUsage.
func (s *MongoUsageStore) HandleUsage(ctx context.Context, record usage.Record) {
	if !s.enabled || s.usageCol == nil {
		return
	}

	// Use a short timeout context for DB operations to avoid being cancelled by the request ctx.
	dbCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Determine userId from API key (using cache).
	userID := s.lookupUserID(dbCtx, record.APIKey)

	// Calculate cost: cost = ceil(rate * (totalTokens + reasoningTokens) * 0.01)
	rate, err := s.getModelRate(dbCtx, record.Alias)
	if err != nil {
		log.WithError(err).Warn("mongodb usage: failed to get model rate, using default 1.0")
		rate = 1.0
	}
	totalTokens := record.Detail.TotalTokens + record.Detail.ReasoningTokens
	cost := int64(math.Ceil(rate * float64(totalTokens) * 0.01))

	// Build document for t_user_usage.
	doc := bson.M{
		"provider":     record.Provider,
		"model":        record.Model,
		"alias":        record.Alias,
		"apiKey":       record.APIKey,
		"authID":       record.AuthID,
		"authIndex":    record.AuthIndex,
		"authType":     record.AuthType,
		"source":       record.Source,
		"requestedAt":  record.RequestedAt,
		"latencyMs":    record.Latency.Milliseconds(),
		"failed":       record.Failed,
		"fail": bson.M{
			"statusCode": record.Fail.StatusCode,
			"body":       record.Fail.Body,
		},
		"detail": bson.M{
			"inputTokens":          record.Detail.InputTokens,
			"outputTokens":         record.Detail.OutputTokens,
			"reasoningTokens":      record.Detail.ReasoningTokens,
			"cachedTokens":         record.Detail.CachedTokens,
			"cacheReadTokens":      record.Detail.CacheReadTokens,
			"cacheCreationTokens":  record.Detail.CacheCreationTokens,
			"totalTokens":          record.Detail.TotalTokens,
		},
		"userId": userID,
		"cost":   cost,
	}

	_, err = s.usageCol.InsertOne(dbCtx, doc)
	if err != nil {
		log.WithError(err).Warn("mongodb usage: failed to insert usage record")
		return
	}

	// Deduct balance atomically (can be negative)
	if userID != "" && cost > 0 {
		userIDObj, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			log.WithError(err).Warnf("mongodb usage: invalid userId format: %s", userID)
			return
		}
		if err := s.updateUserBalance(dbCtx, userIDObj, cost); err != nil {
			log.WithError(err).Warnf("mongodb usage: failed to deduct balance for userId=%s cost=%d", userID, cost)
		}
	}
}

// lookupUserID returns the userId string for the given API key, using a simple cache.
// If not found, returns empty string.
func (s *MongoUsageStore) lookupUserID(ctx context.Context, apiKey string) string {
	if apiKey == "" {
		return ""
	}
	// Check cache.
	s.cacheMu.RLock()
	if entry, ok := s.cache[apiKey]; ok && time.Now().Before(entry.expires) {
		s.cacheMu.RUnlock()
		return entry.userId
	}
	s.cacheMu.RUnlock()

	// Not in cache or expired; query DB.
	var result struct {
		UserID interface{} `bson:"userId"`
	}
	filter := bson.M{"key": apiKey, "status": bson.M{"$exists": true}}
	// Optionally only active keys: {"status": 1}
	err := s.userKeyCol.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.WithError(err).Warnf("mongodb usage: failed to lookup userId for key %s", apiKey[:8])
		}
		// Miss: cache negative result for a short time to avoid hammering DB.
		s.cacheMu.Lock()
		s.cache[apiKey] = cachedUserId{userId: "", expires: time.Now().Add(time.Minute)}
		s.cacheMu.Unlock()
		return ""
	}
	// Convert userId to string.
	var userIdStr string
	switch v := result.UserID.(type) {
	case string:
		userIdStr = v
	case primitive.ObjectID:
		userIdStr = v.Hex()
	default:
		// Try to convert via fmt?
		userIdStr = fmt.Sprintf("%v", v)
	}
	// Cache successful lookup for longer (e.g., 1 hour).
	s.cacheMu.Lock()
	s.cache[apiKey] = cachedUserId{userId: userIdStr, expires: time.Now().Add(time.Hour)}
	s.cacheMu.Unlock()
	return userIdStr
}

// Close is a no-op; the underlying MongoDB connection is owned by MongoStore.
func (s *MongoUsageStore) Close() error { return nil }

func (s *MongoUsageStore) getModelRate(ctx context.Context, alias string) (float64, error) {
	if s.modelCol == nil {
		return 1.0, nil
	}
	if alias == "" {
		return 1.0, nil
	}

	cursor, err := s.modelCol.Find(ctx, bson.M{})
	if err != nil {
		return 1.0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc MongoModelDoc
		if err = cursor.Decode(&doc); err != nil {
			continue
		}
		for _, m := range doc.Models {
			if m.Alias == alias || m.Name == alias {
				if m.Rate > 0 {
					return m.Rate, nil
				}
				return 1.0, nil
			}
		}
	}
	return 1.0, cursor.Err()
}

func (s *MongoUsageStore) updateUserBalance(ctx context.Context, userID primitive.ObjectID, cost int64) error {
	if s.userCol == nil {
		return fmt.Errorf("user collection not available")
	}
	result, err := s.userCol.UpdateOne(ctx,
		bson.M{"_id": userID},
		bson.M{"$inc": bson.M{"balance": -cost}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}