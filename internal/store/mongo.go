package store

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoTimeout    = 10 * time.Second
	mongoCollection = "t_model"
)

type MongoStore struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
	enabled    bool

	registryMu               sync.Mutex
	registeredModelClientIDs map[string]struct{}
}

type MongoModelConfig struct {
	Type     string            `bson:"type" json:"type"`
	Name     string            `bson:"name" json:"name"`
	Priority int               `bson:"priority" json:"priority"`
	BaseURL  string            `bson:"baseUrl" json:"baseUrl"`
	Models   []MongoModel      `bson:"models" json:"models"`
	Headers  map[string]string `bson:"headers" json:"headers"`
	APIKey   []string          `bson:"apiKey" json:"apiKey"`
}

type MongoModel struct {
	Name  string  `bson:"name" json:"name"`
	Alias string  `bson:"alias" json:"alias"`
	Rate  float64 `bson:"rate" json:"rate"`
}

type MongoAPIKeyEntry struct {
	APIKey string `bson:"apiKey" json:"apiKey"`
}

var mongoStore *MongoStore

func NewMongoStore(cfg config.MongoConfig) (*MongoStore, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	trimmedURI := strings.TrimSpace(cfg.URI)
	if trimmedURI == "" {
		return nil, fmt.Errorf("mongodb: uri is required when enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	clientOpts := options.Client().ApplyURI(trimmedURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("mongodb: connect: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongodb: ping: %w", err)
	}

	db := client.Database(cfg.GetDatabase())
	col := db.Collection(mongoCollection)

	store := &MongoStore{
		client:                   client,
		database:                 db,
		collection:               col,
		enabled:                  true,
		registeredModelClientIDs: make(map[string]struct{}),
	}

	mongoStore = store
	return store, nil
}

func (s *MongoStore) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.client.Disconnect(ctx)
}

func (s *MongoStore) IsEnabled() bool {
	if s == nil {
		return false
	}
	return s.enabled
}

func (s *MongoStore) GetModelConfigs(ctx context.Context, types []string) (map[string][]MongoModelConfig, error) {
	if s == nil || !s.enabled {
		return nil, fmt.Errorf("mongodb: store not initialized or disabled")
	}

	if len(types) == 0 {
		return nil, nil
	}

	filter := bson.M{"type": bson.M{"$in": types}}
	log.Infof("mongodb: querying collection=%s, filter=%v", s.collection.Name(), filter)

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("mongodb: find: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[string][]MongoModelConfig)
	count := 0
	for cursor.Next(ctx) {
		count++
		var cfg MongoModelConfig
		if err = cursor.Decode(&cfg); err != nil {
			log.WithError(err).Warn("mongodb: decode model config")
			continue
		}
		if cfg.Type == "" {
			continue
		}
		result[cfg.Type] = append(result[cfg.Type], cfg)
	}

	log.Infof("mongodb: found %d documents, types=%v", count, result)

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("mongodb: cursor: %w", err)
	}

	return result, nil
}

func (s *MongoStore) ConvertToConfig() (*config.Config, error) {
	if s == nil || !s.enabled {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	types := []string{
		"gemini-api-key",
		"codex-api-key",
		"claude-api-key",
		"openai-compatibility",
	}

	configs, err := s.GetModelConfigs(ctx, types)
	if err != nil {
		return nil, fmt.Errorf("mongodb: get model configs: %w", err)
	}

	cfg := &config.Config{}

	for typeName, cfgs := range configs {
		for _, mongoCfg := range cfgs {
			switch strings.ToLower(typeName) {
			case "gemini-api-key":
				cfg.GeminiKey = append(cfg.GeminiKey, s.convertToGeminiKey(mongoCfg)...)
			case "codex-api-key":
				cfg.CodexKey = append(cfg.CodexKey, s.convertToCodexKey(mongoCfg)...)
			case "claude-api-key":
				cfg.ClaudeKey = append(cfg.ClaudeKey, s.convertToClaudeKey(mongoCfg)...)
			case "openai-compatibility":
				cfg.OpenAICompatibility = append(cfg.OpenAICompatibility, s.convertToOpenAICompat(mongoCfg))
			}
		}
	}

	return cfg, nil
}

func (s *MongoStore) convertToGeminiKey(cfg MongoModelConfig) []config.GeminiKey {
	if len(cfg.APIKey) == 0 {
		return nil
	}

	keys := make([]config.GeminiKey, 0, len(cfg.APIKey))
	for _, apiKey := range cfg.APIKey {
		if apiKey == "" {
			continue
		}

		models := make([]config.GeminiModel, 0, len(cfg.Models))
		for _, m := range cfg.Models {
			model := config.GeminiModel{
				Name:  m.Name,
				Alias: m.Alias,
			}
			if model.Alias == "" {
				model.Alias = model.Name
			}
			models = append(models, model)
		}

		key := config.GeminiKey{
			APIKey:   apiKey,
			Priority: cfg.Priority,
			BaseURL:  cfg.BaseURL,
			Headers:  cfg.Headers,
			Models:   models,
		}
		keys = append(keys, key)
	}

	return keys
}

func (s *MongoStore) convertToCodexKey(cfg MongoModelConfig) []config.CodexKey {
	if len(cfg.APIKey) == 0 {
		return nil
	}

	keys := make([]config.CodexKey, 0, len(cfg.APIKey))
	for _, apiKey := range cfg.APIKey {
		if apiKey == "" {
			continue
		}

		models := make([]config.CodexModel, 0, len(cfg.Models))
		for _, m := range cfg.Models {
			model := config.CodexModel{
				Name:  m.Name,
				Alias: m.Alias,
			}
			if model.Alias == "" {
				model.Alias = model.Name
			}
			models = append(models, model)
		}

		key := config.CodexKey{
			APIKey:   apiKey,
			Priority: cfg.Priority,
			BaseURL:  cfg.BaseURL,
			Headers:  cfg.Headers,
			Models:   models,
		}
		keys = append(keys, key)
	}

	return keys
}

func (s *MongoStore) convertToClaudeKey(cfg MongoModelConfig) []config.ClaudeKey {
	if len(cfg.APIKey) == 0 {
		return nil
	}

	keys := make([]config.ClaudeKey, 0, len(cfg.APIKey))
	for _, apiKey := range cfg.APIKey {
		if apiKey == "" {
			continue
		}

		models := make([]config.ClaudeModel, 0, len(cfg.Models))
		for _, m := range cfg.Models {
			model := config.ClaudeModel{
				Name:  m.Name,
				Alias: m.Alias,
			}
			if model.Alias == "" {
				model.Alias = model.Name
			}
			models = append(models, model)
		}

		key := config.ClaudeKey{
			APIKey:   apiKey,
			Priority: cfg.Priority,
			BaseURL:  cfg.BaseURL,
			Headers:  cfg.Headers,
			Models:   models,
		}
		keys = append(keys, key)
	}

	return keys
}

func (s *MongoStore) convertToOpenAICompat(cfg MongoModelConfig) config.OpenAICompatibility {
	apiKeyEntries := make([]config.OpenAICompatibilityAPIKey, 0, len(cfg.APIKey))
	for _, apiKey := range cfg.APIKey {
		if apiKey == "" {
			continue
		}
		apiKeyEntries = append(apiKeyEntries, config.OpenAICompatibilityAPIKey{
			APIKey: apiKey,
		})
	}

	models := make([]config.OpenAICompatibilityModel, 0, len(cfg.Models))
	for _, m := range cfg.Models {
		model := config.OpenAICompatibilityModel{
			Name:  m.Name,
			Alias: m.Alias,
		}
		if model.Alias == "" {
			model.Alias = model.Name
		}
		models = append(models, model)
	}

	return config.OpenAICompatibility{
		Name:          cfg.Name,
		Priority:      cfg.Priority,
		BaseURL:       cfg.BaseURL,
		Headers:       cfg.Headers,
		APIKeyEntries: apiKeyEntries,
		Models:        models,
	}
}

func (s *MongoStore) RegisterModelsFromConfig(ctx context.Context) error {
	if s == nil || !s.enabled {
		return nil
	}

	types := []string{
		"gemini-api-key",
		"codex-api-key",
		"claude-api-key",
		"openai-compatibility",
	}

	configs, err := s.GetModelConfigs(ctx, types)
	if err != nil {
		return fmt.Errorf("mongodb: get model configs: %w", err)
	}

	registry := registry.GetGlobalRegistry()
	desiredIDs := make(map[string]struct{})

	s.registryMu.Lock()
	for clientID := range s.registeredModelClientIDs {
		registry.UnregisterClient(clientID)
	}
	s.registeredModelClientIDs = make(map[string]struct{})
	s.registryMu.Unlock()

	for typeName, cfgs := range configs {
		for _, cfg := range cfgs {
			models := s.convertToModelInfos(cfg)
			if len(models) == 0 {
				continue
			}

			clientID := fmt.Sprintf("mongo-%s-%s", typeName, cfg.Name)
			provider := s.getProviderForType(typeName)

			registry.RegisterClient(clientID, provider, models)
			desiredIDs[clientID] = struct{}{}
			log.Infof("mongodb: registered %d models for type=%s, name=%s, provider=%s",
				len(models), typeName, cfg.Name, provider)
		}
	}

	s.registryMu.Lock()
	s.registeredModelClientIDs = desiredIDs
	s.registryMu.Unlock()

	return nil
}

func (s *MongoStore) convertToModelInfos(cfg MongoModelConfig) []*registry.ModelInfo {
	if len(cfg.Models) == 0 {
		return nil
	}

	models := make([]*registry.ModelInfo, 0, len(cfg.Models))
	now := time.Now().Unix()

	for _, m := range cfg.Models {
		modelID := m.Alias
		if modelID == "" {
			modelID = m.Name
		}
		if modelID == "" {
			continue
		}

		models = append(models, &registry.ModelInfo{
			ID:          modelID,
			Object:      "model",
			Created:     now,
			OwnedBy:     cfg.Name,
			Type:        cfg.Type,
			DisplayName: modelID,
			UserDefined: true,
		})
	}

	return models
}

func (s *MongoStore) getProviderForType(typeName string) string {
	switch strings.ToLower(typeName) {
	case "gemini-api-key":
		return "gemini"
	case "codex-api-key":
		return "codex"
	case "claude-api-key":
		return "claude"
	case "openai-compatibility":
		return "openai-compatibility"
	default:
		return typeName
	}
}

func GetMongoStore() *MongoStore {
	return mongoStore
}
