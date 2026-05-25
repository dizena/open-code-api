package cliproxy

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/sdk/config"
)

func TestMergeMongoRuntimeConfig_PreservesBaseAndAppendsMongoModels(t *testing.T) {
	base := &config.Config{
		Host: "127.0.0.1",
		Port: 8317,
		OpenAICompatibility: []config.OpenAICompatibility{
			{
				Name:    "local-provider",
				BaseURL: "https://local.example/v1",
				Models: []config.OpenAICompatibilityModel{
					{Name: "local-upstream", Alias: "local-alias"},
				},
			},
		},
	}
	mongoCfg := &config.Config{
		OpenAICompatibility: []config.OpenAICompatibility{
			{
				Name:    "mongo-provider",
				BaseURL: "https://mongo.example/v1",
				Models: []config.OpenAICompatibilityModel{
					{Name: "MiniMax-M2.5", Alias: "geo-pro"},
				},
			},
		},
	}

	mergeMongoRuntimeConfig(base, mongoCfg)

	if base.Host != "127.0.0.1" || base.Port != 8317 {
		t.Fatalf("base server config unexpectedly changed: host=%q port=%d", base.Host, base.Port)
	}
	if len(base.OpenAICompatibility) != 2 {
		t.Fatalf("openai compatibility len = %d, want 2", len(base.OpenAICompatibility))
	}
	if base.OpenAICompatibility[0].Name != "local-provider" {
		t.Fatalf("first provider = %q, want local-provider", base.OpenAICompatibility[0].Name)
	}
	if base.OpenAICompatibility[1].Name != "mongo-provider" {
		t.Fatalf("second provider = %q, want mongo-provider", base.OpenAICompatibility[1].Name)
	}
	if got := base.OpenAICompatibility[1].Models[0].Alias; got != "geo-pro" {
		t.Fatalf("mongo alias = %q, want geo-pro", got)
	}
}
