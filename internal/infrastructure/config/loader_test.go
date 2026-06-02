package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tikhomirovv/book-translater/internal/infrastructure/config"
)

func TestLoadFromRepoConfigs(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Skip(err)
	}
	cfgDir := filepath.Join(root, "configs")

	cfg, err := config.Load(cfgDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Chunk.SizeParagraphs != 10 {
		t.Errorf("size_paragraphs = %d, want 10", cfg.Chunk.SizeParagraphs)
	}
	if !cfg.IsLanguageAllowed("ru") {
		t.Error("ru should be allowed")
	}
	if cfg.IsLanguageAllowed("xx") {
		t.Error("xx should not be allowed")
	}
	if _, ok := cfg.Prompts["nonfiction"]; !ok {
		t.Error("nonfiction prompt set missing")
	}
}

func TestLoadLocalOverride(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "config.yaml")
	local := filepath.Join(dir, "config.local.yaml")

	if err := os.WriteFile(base, []byte(`
chunk:
  size_paragraphs: 10
allowed_languages:
  - ru
prompts:
  nonfiction:
    system: "base"
    translation: "tr"
    context_extraction: "cx"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(local, []byte(`
chunk:
  size_paragraphs: 5
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Chunk.SizeParagraphs != 5 {
		t.Errorf("override size = %d, want 5", cfg.Chunk.SizeParagraphs)
	}
}

func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		dir = filepath.Dir(dir)
	}
	return "", os.ErrNotExist
}
