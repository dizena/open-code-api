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

type MongoModelDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type      string             `bson:"type" json:"type"`
	Name      string             `bson:"name" json:"name"`
	Priority  int                `bson:"priority" json:"priority"`
	BaseURL   string             `bson:"baseUrl" json:"baseUrl"`
	Models    []MongoModel       `bson:"models" json:"models"`
	Headers   map[string]string  `bson:"headers" json:"headers"`
	APIKey    []string           `bson:"apiKey" json:"apiKey"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdateAt  time.Time          `bson:"updateAt" json:"updateAt"`
}

type CreateModelInput struct {
	Type     string
	Name     string
	Priority int
	BaseURL  string
	Models   []MongoModel
	Headers  map[string]string
	APIKey   []string
}

type UpdateModelInput struct {
	Type     *string
	Name     *string
	Priority *int
	BaseURL  *string
	Models   *[]MongoModel
	Headers  *map[string]string
	APIKey   *[]string
}

type ListModelsInput struct {
	Page      int
	PageSize  int
	Type      *string
	SortBy    string // 排序字段，如 "name" 或 "createdAt"
	SortOrder int    // 1 升序，-1 降序
}


type ModelListResult struct {
	Models []MongoModelDoc `json:"models"`
	Total  int64           `json:"total"`
}

func (s *MongoStore) CreateModel(ctx context.Context, input CreateModelInput) (*MongoModelDoc, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}
	now := time.Now()
	doc := MongoModelDoc{
		Type:      strings.TrimSpace(input.Type),
		Name:      strings.TrimSpace(input.Name),
		Priority:  input.Priority,
		BaseURL:   strings.TrimSpace(input.BaseURL),
		Models:    input.Models,
		Headers:   input.Headers,
		APIKey:    input.APIKey,
		CreatedAt: now,
		UpdateAt:  now,
	}

	result, err := s.modelCollection().InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("mongodb: create model: %w", err)
	}
	doc.ID = result.InsertedID.(primitive.ObjectID)
	log.Infof("mongodb: created model type=%s name=%s", doc.Type, doc.Name)
	return &doc, nil
}

func (s *MongoStore) ListModels(ctx context.Context, input ListModelsInput) (*ModelListResult, error) {
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
	if input.Type != nil {
		filter["type"] = *input.Type
	}

	total, err := s.modelCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("mongodb: count models: %w", err)
	}

	skip := int64(input.Page-1) * int64(input.PageSize)
	// opts := options.Find().SetSkip(skip).SetLimit(int64(input.PageSize)).SetSort(bson.M{"priority": -1, "createdAt": -1})
	opts := options.Find().SetSkip(skip).SetLimit(int64(input.PageSize)).SetSort(bson.M{"priority": 1})

	cursor, err := s.modelCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("mongodb: list models: %w", err)
	}
	defer cursor.Close(ctx)

	var models []MongoModelDoc
	for cursor.Next(ctx) {
		var m MongoModelDoc
		if err = cursor.Decode(&m); err != nil {
			log.WithError(err).Warn("mongodb: decode model")
			continue
		}
		models = append(models, m)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("mongodb: cursor: %w", err)
	}

	return &ModelListResult{Models: models, Total: total}, nil
}

func (s *MongoStore) GetModelByID(ctx context.Context, id primitive.ObjectID) (*MongoModelDoc, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}
	var doc MongoModelDoc
	err := s.modelCollection().FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("mongodb: get model: %w", err)
	}
	return &doc, nil
}

func (s *MongoStore) UpdateModel(ctx context.Context, id primitive.ObjectID, input UpdateModelInput) error {
	if s == nil || !s.enabled {
		return fmt.Errorf("mongodb: store not initialized")
	}
	update := bson.M{"$set": bson.M{"updateAt": time.Now()}}
	if input.Type != nil {
		update["$set"].(bson.M)["type"] = *input.Type
	}
	if input.Name != nil {
		update["$set"].(bson.M)["name"] = *input.Name
	}
	if input.Priority != nil {
		update["$set"].(bson.M)["priority"] = *input.Priority
	}
	if input.BaseURL != nil {
		update["$set"].(bson.M)["baseUrl"] = *input.BaseURL
	}
	if input.Models != nil {
		update["$set"].(bson.M)["models"] = *input.Models
	}
	if input.Headers != nil {
		update["$set"].(bson.M)["headers"] = *input.Headers
	}
	if input.APIKey != nil {
		update["$set"].(bson.M)["apiKey"] = *input.APIKey
	}

	result, err := s.modelCollection().UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("mongodb: update model: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("model not found")
	}
	log.Infof("mongodb: updated model id=%s", id.Hex())
	return nil
}

func (s *MongoStore) DeleteModel(ctx context.Context, id primitive.ObjectID) error {
	if s == nil || !s.enabled {
		return fmt.Errorf("mongodb: store not initialized")
	}
	result, err := s.modelCollection().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("mongodb: delete model: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("model not found")
	}
	log.Infof("mongodb: deleted model id=%s", id.Hex())
	return nil
}

func (s *MongoStore) modelCollection() *mongo.Collection {
	return s.collection
}

func (s *MongoStore) GetAvailableModelAliases(ctx context.Context) ([]string, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized")
	}

	cursor, err := s.modelCollection().Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("mongodb: find models: %w", err)
	}
	defer cursor.Close(ctx)

	aliasSet := make(map[string]struct{})
	for cursor.Next(ctx) {
		var doc MongoModelDoc
		if err = cursor.Decode(&doc); err != nil {
			log.WithError(err).Warn("mongodb: decode model for aliases")
			continue
		}
		for _, m := range doc.Models {
			if m.Alias != "" {
				aliasSet[m.Alias] = struct{}{}
			}
		}
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("mongodb: cursor: %w", err)
	}

	aliases := make([]string, 0, len(aliasSet))
	for alias := range aliasSet {
		aliases = append(aliases, alias)
	}

	// Sort alphabetically
	for i := 0; i < len(aliases); i++ {
		for j := i + 1; j < len(aliases); j++ {
			if aliases[i] > aliases[j] {
				aliases[i], aliases[j] = aliases[j], aliases[i]
			}
		}
	}

	return aliases, nil
}

func (s *MongoStore) GetModelRateByAlias(ctx context.Context, alias string) (float64, error) {
	if s == nil || !s.enabled {
		return 1.0, nil
	}
	if alias == "" {
		return 1.0, nil
	}

	cursor, err := s.modelCollection().Find(ctx, bson.M{})
	if err != nil {
		return 1.0, fmt.Errorf("mongodb: find models for rate: %w", err)
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
	if err = cursor.Err(); err != nil {
		return 1.0, fmt.Errorf("mongodb: cursor error for rate: %w", err)
	}

	return 1.0, nil
}

// UpdateModelAlias updates the alias and rate of a specific model within a MongoModelDoc.
func (s *MongoStore) UpdateModelAlias(ctx context.Context, id primitive.ObjectID, index int, alias string, rate float64) error {
	if s == nil || !s.enabled {
		return fmt.Errorf("mongodb: store not initialized")
	}

	if index < 0 {
		return fmt.Errorf("model index out of range")
	}

	// Use dot notation for atomic array element update
	aliasKey := fmt.Sprintf("models.%d.alias", index)
	rateKey := fmt.Sprintf("models.%d.rate", index)

	update := bson.M{
		"$set": bson.M{
			aliasKey:   strings.TrimSpace(alias),
			rateKey:    rate,
			"updateAt": time.Now(),
		},
	}

	result, err := s.modelCollection().UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		// Check if error is due to invalid array index
		if strings.Contains(err.Error(), "path") || strings.Contains(err.Error(), "element") {
			return fmt.Errorf("model index out of range")
		}
		return fmt.Errorf("mongodb: update model alias: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("model not found")
	}
	log.Infof("mongodb: updated model alias id=%s index=%d", id.Hex(), index)
	return nil
}
