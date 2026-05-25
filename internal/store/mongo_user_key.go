package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

const (
	userKeyCollection = "t_user_key"
)

type UserKeyStatus int

const (
	UserKeyDisabled UserKeyStatus = 0
	UserKeyActive   UserKeyStatus = 1
)

type MongoUserKey struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"userId" json:"userId"`
	Name       string             `bson:"name" json:"name"`
	Key        string             `bson:"key" json:"key"`
	Status     UserKeyStatus      `bson:"status" json:"status"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	LastUsedAt time.Time          `bson:"lastUsedAt" json:"lastUsedAt"`
}

type CreateUserKeyInput struct {
	UserID primitive.ObjectID
	Name   string
	Key    string
	Status UserKeyStatus
}

func (s *MongoStore) CreateUserKey(ctx context.Context, input CreateUserKeyInput) (*MongoUserKey, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}
	now := time.Now()
	key := MongoUserKey{
		UserID:     input.UserID,
		Name:       strings.TrimSpace(input.Name),
		Key:        strings.TrimSpace(input.Key),
		Status:     input.Status,
		CreatedAt:  now,
		LastUsedAt: time.Time{},
	}
	if key.Status == 0 {
		key.Status = UserKeyActive
	}

	result, err := s.userKeyCollection().InsertOne(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("mongodb: create user key: %w", err)
	}
	key.ID = result.InsertedID.(primitive.ObjectID)
	log.Infof("mongodb: created user key name=%s userId=%s", key.Name, key.UserID.Hex())
	return &key, nil
}

func (s *MongoStore) GetUserKeyByKey(ctx context.Context, key string) (*MongoUserKey, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}
	var uk MongoUserKey
	filter := bson.M{"key": strings.TrimSpace(key)}
	err := s.userKeyCollection().FindOne(ctx, filter).Decode(&uk)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("mongodb: get user key: %w", err)
	}
	return &uk, nil
}

func (s *MongoStore) GetUserKeysByUserID(ctx context.Context, userID primitive.ObjectID) ([]MongoUserKey, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}
	cursor, err := s.userKeyCollection().Find(ctx, bson.M{"userId": userID}, options.Find().SetSort(bson.M{"createdAt": -1}))
	if err != nil {
		return nil, fmt.Errorf("mongodb: list user keys: %w", err)
	}
	defer cursor.Close(ctx)

	var keys []MongoUserKey
	for cursor.Next(ctx) {
		var k MongoUserKey
		if err = cursor.Decode(&k); err != nil {
			log.WithError(err).Warn("mongodb: decode user key")
			continue
		}
		keys = append(keys, k)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("mongodb: cursor: %w", err)
	}
	return keys, nil
}

type ListUserKeysInput struct {
	Account  string
	Page     int
	PageSize int
	StartDate *time.Time
	EndDate   *time.Time
}

type UserKeyListResult struct {
	Keys  []MongoUserKey `json:"keys"`
	Total int64          `json:"total"`
}

func (s *MongoStore) ListUserKeysByAccount(ctx context.Context, input ListUserKeysInput) (*UserKeyListResult, error) {
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
	if input.Account != "" {
		user, err := s.GetUserByAccount(ctx, input.Account)
		if err != nil || user == nil {
			return &UserKeyListResult{Keys: []MongoUserKey{}, Total: 0}, nil
		}
		filter["userId"] = user.ID
	}
	if input.StartDate != nil || input.EndDate != nil {
		dateFilter := bson.M{}
		if input.StartDate != nil {
			dateFilter["$gte"] = *input.StartDate
		}
		if input.EndDate != nil {
			dateFilter["$lte"] = *input.EndDate
		}
		filter["createdAt"] = dateFilter
	}

	total, err := s.userKeyCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("mongodb: count user keys: %w", err)
	}

	skip := int64(input.Page-1) * int64(input.PageSize)
	opts := options.Find().SetSkip(skip).SetLimit(int64(input.PageSize)).SetSort(bson.M{"createdAt": -1})

	cursor, err := s.userKeyCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("mongodb: list user keys: %w", err)
	}
	defer cursor.Close(ctx)

	var keys []MongoUserKey
	for cursor.Next(ctx) {
		var k MongoUserKey
		if err = cursor.Decode(&k); err != nil {
			log.WithError(err).Warn("mongodb: decode user key")
			continue
		}
		keys = append(keys, k)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("mongodb: cursor: %w", err)
	}

	return &UserKeyListResult{Keys: keys, Total: total}, nil
}

func (s *MongoStore) GetUserKeysByAccount(ctx context.Context, account string) ([]MongoUserKey, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}
	user, err := s.GetUserByAccount(ctx, account)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return s.GetUserKeysByUserID(ctx, user.ID)
}

func (s *MongoStore) DeleteUserKey(ctx context.Context, key string) error {
	if s == nil || !s.enabled {
		return fmt.Errorf("mongodb: store not initialized")
	}
	result, err := s.userKeyCollection().DeleteOne(ctx, bson.M{"key": strings.TrimSpace(key)})
	if err != nil {
		return fmt.Errorf("mongodb: delete user key: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("key not found")
	}
	log.Infof("mongodb: deleted user key key=%s", key[:8])
	return nil
}

func (s *MongoStore) UpdateUserKeyStatus(ctx context.Context, key string, status UserKeyStatus) error {
	if s == nil || !s.enabled {
		return fmt.Errorf("mongodb: store not initialized")
	}
	result, err := s.userKeyCollection().UpdateOne(ctx, bson.M{"key": strings.TrimSpace(key)}, bson.M{"$set": bson.M{"status": status}})
	if err != nil {
		return fmt.Errorf("mongodb: update user key status: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("key not found")
	}
	log.Infof("mongodb: updated user key status key=%s status=%d", key[:8], status)
	return nil
}

func (s *MongoStore) TouchUserKeyLastUsed(ctx context.Context, key string) error {
	if s == nil || !s.enabled {
		return nil
	}
	_, err := s.userKeyCollection().UpdateOne(ctx, bson.M{"key": strings.TrimSpace(key)}, bson.M{"$set": bson.M{"lastUsedAt": time.Now()}})
	return err
}

func (s *MongoStore) userKeyCollection() *mongo.Collection {
	return s.database.Collection(userKeyCollection)
}
