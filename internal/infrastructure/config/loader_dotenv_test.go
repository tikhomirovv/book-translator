package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/infrastructure/config"
)

func TestLoad_readsDotEnv(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(`
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

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	if err := os.Unsetenv("OPENAI_API_KEY"); err != nil {
		t.Fatal(err)
	}
	if err := os.Unsetenv("OPENAI_BASE_URL"); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(".env", []byte(`
OPENAI_BASE_URL=http://192.168.1.50:1234/v1
OPENAI_API_KEY=sk-your-key
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(cfgDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.OpenAIBaseURL != "http://192.168.1.50:1234/v1" {
		t.Fatalf("base url = %q", cfg.OpenAIBaseURL)
	}
	if cfg.OpenAIAPIKey != "" {
		t.Fatalf("placeholder key should normalize to empty, got %q", cfg.OpenAIAPIKey)
	}
}
