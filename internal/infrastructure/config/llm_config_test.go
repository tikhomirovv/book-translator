package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/infrastructure/config"
)

func TestLLMConfig_Normalize_nested(t *testing.T) {
	t.Parallel()

	cfg := config.LLMConfig{
		Translation: config.LLMCallConfig{
			Model:       "translate-model",
			Temperature: 0.4,
			MaxTokens:   1000,
		},
		Context: config.LLMCallConfig{
			Model:       "context-model",
			Temperature: 0.1,
			MaxTokens:   500,
		},
	}
	cfg.Normalize()

	if cfg.Translation.Model != "translate-model" || cfg.Context.Model != "context-model" {
		t.Fatalf("models changed: %+v", cfg)
	}
	if cfg.Translation.Temperature != 0.4 || cfg.Context.Temperature != 0.1 {
		t.Fatalf("temperature changed: %+v", cfg)
	}
	if cfg.Translation.MaxTokens != 1000 || cfg.Context.MaxTokens != 500 {
		t.Fatalf("max_tokens changed: %+v", cfg)
	}
}

func TestLLMConfig_Normalize_legacyFlat(t *testing.T) {
	t.Parallel()

	cfg := config.LLMConfig{
		LegacyModel:        "legacy-translate",
		LegacyContextModel: "legacy-context",
		LegacyTemperature:  0.35,
		LegacyMaxTokens:    2048,
	}
	cfg.Normalize()

	if cfg.Translation.Model != "legacy-translate" {
		t.Fatalf("translation model = %q", cfg.Translation.Model)
	}
	if cfg.Context.Model != "legacy-context" {
		t.Fatalf("context model = %q", cfg.Context.Model)
	}
	if cfg.Translation.Temperature != 0.35 {
		t.Fatalf("translation temperature = %v", cfg.Translation.Temperature)
	}
	if cfg.Translation.MaxTokens != 2048 {
		t.Fatalf("translation max_tokens = %d", cfg.Translation.MaxTokens)
	}
	if cfg.Context.Temperature != 0.2 {
		t.Fatalf("context temperature = %v, want default 0.2", cfg.Context.Temperature)
	}
	if cfg.Context.MaxTokens != 8192 {
		t.Fatalf("context max_tokens = %d, want default 8192", cfg.Context.MaxTokens)
	}
}

func TestLoad_nestedLLMConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
llm:
  translation:
    model: big-model
    temperature: 0.5
    max_tokens: 12000
  context:
    model: small-model
    temperature: 0.1
    max_tokens: 3000
allowed_languages:
  - ru
prompts:
  nonfiction:
    system: "s"
    translation: "t"
    context_extraction: "c"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LLM.Translation.Model != "big-model" {
		t.Fatalf("translation model = %q", cfg.LLM.Translation.Model)
	}
	if cfg.LLM.Context.MaxTokens != 3000 {
		t.Fatalf("context max_tokens = %d", cfg.LLM.Context.MaxTokens)
	}
}
