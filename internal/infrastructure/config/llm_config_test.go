package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/infrastructure/config"
)

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
	if cfg.LLM.Context.Model != "small-model" {
		t.Fatalf("context model = %q", cfg.LLM.Context.Model)
	}
	if cfg.LLM.Context.MaxTokens != 3000 {
		t.Fatalf("context max_tokens = %d", cfg.LLM.Context.MaxTokens)
	}
}
