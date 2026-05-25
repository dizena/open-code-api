package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const userCheckCollection = "t_user_check"

type MongoUserCheck struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Account   string             `bson:"account" json:"account"`
	Code      string             `bson:"code" json:"code"`
	ExpiresAt time.Time          `bson:"expiresAt" json:"expiresAt"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

func (s *MongoStore) SaveVerificationCode(ctx context.Context, account, code string, expiresAt time.Time) (*MongoUserCheck, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}

	doc := MongoUserCheck{
		Account:   account,
		Code:      code,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	result, err := s.userCheckCollection().InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("mongodb: save verification code: %w", err)
	}
	doc.ID = result.InsertedID.(primitive.ObjectID)
	return &doc, nil
}

func (s *MongoStore) GetValidVerificationCode(ctx context.Context, account, code string) (*MongoUserCheck, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}

	filter := bson.M{
		"account":   account,
		"code":      code,
		"expiresAt": bson.M{"$gt": time.Now()},
	}

	var doc MongoUserCheck
	err := s.userCheckCollection().FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("mongodb: get verification code: %w", err)
	}
	return &doc, nil
}

func (s *MongoStore) CleanupExpiredVerificationCodes(ctx context.Context) error {
	if s == nil || !s.enabled {
		return nil
	}

	_, err := s.userCheckCollection().DeleteMany(ctx, bson.M{"expiresAt": bson.M{"$lte": time.Now()}})
	return err
}

func (s *MongoStore) userCheckCollection() *mongo.Collection {
	return s.database.Collection(userCheckCollection)
}
