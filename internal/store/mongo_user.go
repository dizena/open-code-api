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
	userCollection = "t_user"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type UserStatus int

const (
	UserStatusDisabled UserStatus = 0
	UserStatusActive   UserStatus = 1
)

type MongoUser struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Role         UserRole           `bson:"role" json:"role"`
	Account      string             `bson:"account" json:"account"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	Balance      int64              `bson:"balance" json:"balance"`
	Status       UserStatus         `bson:"status" json:"status"`
	CreateAt     time.Time          `bson:"createAt" json:"createAt"`
	UpdateAt     time.Time          `bson:"updateAt" json:"updateAt"`
}

type CreateUserInput struct {
	Role     UserRole
	Account  string
	Password string // plain text, will be hashed
	Balance  int64
	Status   UserStatus
}

type UpdateUserInput struct {
	Role    *UserRole
	Balance *int64
	Status  *UserStatus
}

type ListUsersInput struct {
	Page     int
	PageSize int
	Role     *UserRole
	Status   *UserStatus
}

type UserListResult struct {
	Users []MongoUser `json:"users"`
	Total int64       `json:"total"`
}

func (s *MongoStore) CreateUser(ctx context.Context, input CreateUserInput) (*MongoUser, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}
	now := time.Now()
	user := MongoUser{
		Role:         input.Role,
		Account:      strings.TrimSpace(input.Account),
		PasswordHash: input.Password, // already hashed by caller
		Balance:      input.Balance,
		Status:       input.Status,
		CreateAt:     now,
		UpdateAt:     now,
	}
	if user.Role == "" {
		user.Role = RoleUser
	}
	if user.Status == 0 {
		user.Status = UserStatusActive
	}

	result, err := s.userCollection().InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("mongodb: create user: %w", err)
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	log.Infof("mongodb: created user account=%s role=%s", user.Account, user.Role)
	return &user, nil
}

func (s *MongoStore) GetUserByAccount(ctx context.Context, account string) (*MongoUser, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}
	var user MongoUser
	filter := bson.M{"account": strings.TrimSpace(account)}
	err := s.userCollection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("mongodb: get user: %w", err)
	}
	return &user, nil
}

func (s *MongoStore) GetUserByID(ctx context.Context, id primitive.ObjectID) (*MongoUser, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}
	var user MongoUser
	filter := bson.M{"_id": id}
	err := s.userCollection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("mongodb: get user by id: %w", err)
	}
	return &user, nil
}

func (s *MongoStore) ListUsers(ctx context.Context, input ListUsersInput) (*UserListResult, error) {
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
	if input.Role != nil {
		filter["role"] = *input.Role
	}
	if input.Status != nil {
		filter["status"] = *input.Status
	}

	total, err := s.userCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("mongodb: count users: %w", err)
	}

	skip := int64(input.Page-1) * int64(input.PageSize)
	opts := options.Find().SetSkip(skip).SetLimit(int64(input.PageSize)).SetSort(bson.M{"createAt": -1})

	cursor, err := s.userCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("mongodb: list users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []MongoUser
	for cursor.Next(ctx) {
		var u MongoUser
		if err = cursor.Decode(&u); err != nil {
			log.WithError(err).Warn("mongodb: decode user")
			continue
		}
		users = append(users, u)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("mongodb: cursor: %w", err)
	}

	return &UserListResult{Users: users, Total: total}, nil
}

func (s *MongoStore) UpdateUser(ctx context.Context, account string, input UpdateUserInput) error {
	if s == nil || !s.enabled {
		return fmt.Errorf("mongodb: store not initialized")
	}
	update := bson.M{"$set": bson.M{"updateAt": time.Now()}}
	if input.Role != nil {
		update["$set"].(bson.M)["role"] = *input.Role
	}
	if input.Balance != nil {
		update["$set"].(bson.M)["balance"] = *input.Balance
	}
	if input.Status != nil {
		update["$set"].(bson.M)["status"] = *input.Status
	}

	result, err := s.userCollection().UpdateOne(ctx, bson.M{"account": strings.TrimSpace(account)}, update)
	if err != nil {
		return fmt.Errorf("mongodb: update user: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}
	log.Infof("mongodb: updated user account=%s", account)
	return nil
}

func (s *MongoStore) DeleteUser(ctx context.Context, account string) error {
	if s == nil || !s.enabled {
		return fmt.Errorf("mongodb: store not initialized")
	}
	result, err := s.userCollection().DeleteOne(ctx, bson.M{"account": strings.TrimSpace(account)})
	if err != nil {
		return fmt.Errorf("mongodb: delete user: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found")
	}
	log.Infof("mongodb: deleted user account=%s", account)
	return nil
}

func (s *MongoStore) UpdateUserBalance(ctx context.Context, userID primitive.ObjectID, delta int64) error {
	if s == nil || !s.enabled {
		return fmt.Errorf("mongodb: store not initialized")
	}
	result, err := s.userCollection().UpdateOne(ctx,
		bson.M{"_id": userID},
		bson.M{"$inc": bson.M{"balance": -delta}},
	)
	if err != nil {
		return fmt.Errorf("mongodb: update user balance: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (s *MongoStore) UpdateUserPassword(ctx context.Context, account string, hashedPassword string) error {
	if s == nil || !s.enabled {
		return fmt.Errorf("mongodb: store not initialized")
	}
	result, err := s.userCollection().UpdateOne(ctx,
		bson.M{"account": strings.TrimSpace(account)},
		bson.M{"$set": bson.M{"passwordHash": hashedPassword, "updateAt": time.Now()}},
	)
	if err != nil {
		return fmt.Errorf("mongodb: update user password: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (s *MongoStore) userCollection() *mongo.Collection {
	return s.database.Collection(userCollection)
}
