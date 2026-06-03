package config_test

import (
	"testing"

	"github.com/tikhomirovv/book-translator/internal/infrastructure/config"
)

func TestLLMConfig_ContextModelName(t *testing.T) {
	t.Parallel()

	cfg := config.LLMConfig{Model: "translate-model"}
	if got := cfg.ContextModelName(); got != "translate-model" {
		t.Fatalf("fallback: got %q, want translate-model", got)
	}

	cfg.ContextModel = "context-model"
	if got := cfg.ContextModelName(); got != "context-model" {
		t.Fatalf("override: got %q, want context-model", got)
	}
}
