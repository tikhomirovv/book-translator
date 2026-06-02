package prompt_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/config"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/prompt"
)

func TestYAMLRenderer_fromRepoConfig(t *testing.T) {
	t.Parallel()

	root, err := repoRoot()
	if err != nil {
		t.Skip(err)
	}

	cfg, err := config.Load(filepath.Join(root, "configs"))
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}

	renderer, err := prompt.NewYAMLRenderer(cfg.Prompts)
	if err != nil {
		t.Fatalf("NewYAMLRenderer: %v", err)
	}

	data := ports.PromptData{
		TargetLang:   "ru",
		ContextBlock: "Earlier glossary: foo = bar",
		ChunkText:    "Hello world.",
	}

	got, err := renderer.Render("nonfiction", prompt.TemplateTranslation, data)
	if err != nil {
		t.Fatalf("Render translation: %v", err)
	}

	for _, want := range []string{"ru", "Earlier glossary", "Hello world."} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered prompt missing %q:\n%s", want, got)
		}
	}
}

func TestYAMLRenderer_allTemplateKeys(t *testing.T) {
	t.Parallel()

	renderer, err := prompt.NewYAMLRenderer(map[string]config.PromptSet{
		"test": {
			System:            "system {{.TargetLang}}",
			Translation:       "translate {{.ChunkText}} to {{.TargetLang}}",
			ContextExtraction: "extract from {{.ChunkText}} overlap={{.OverlapText}}",
		},
	})
	if err != nil {
		t.Fatalf("NewYAMLRenderer: %v", err)
	}

	data := ports.PromptData{
		TargetLang:  "de",
		ChunkText:   "sample",
		OverlapText: "prev",
	}

	tests := []struct {
		key  string
		want string
	}{
		{prompt.TemplateSystem, "system de"},
		{prompt.TemplateTranslation, "translate sample to de"},
		{prompt.TemplateContextExtraction, "extract from sample overlap=prev"},
	}

	for _, tc := range tests {
		got, err := renderer.Render("test", tc.key, data)
		if err != nil {
			t.Fatalf("Render(%q): %v", tc.key, err)
		}
		if got != tc.want {
			t.Errorf("Render(%q) = %q, want %q", tc.key, got, tc.want)
		}
	}
}

func TestYAMLRenderer_unknownPromptType(t *testing.T) {
	t.Parallel()

	renderer, err := prompt.NewYAMLRenderer(map[string]config.PromptSet{
		"known": {System: "ok"},
	})
	if err != nil {
		t.Fatalf("NewYAMLRenderer: %v", err)
	}

	_, err = renderer.Render("missing", prompt.TemplateSystem, ports.PromptData{})
	if err == nil {
		t.Fatal("expected error for unknown prompt type")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
	if !strings.Contains(err.Error(), "unknown prompt type") {
		t.Errorf("error = %q, want mention of unknown prompt type", err.Error())
	}
}

func TestYAMLRenderer_unknownTemplateKey(t *testing.T) {
	t.Parallel()

	renderer, err := prompt.NewYAMLRenderer(map[string]config.PromptSet{
		"known": {System: "ok"},
	})
	if err != nil {
		t.Fatalf("NewYAMLRenderer: %v", err)
	}

	_, err = renderer.Render("known", "not_a_template", ports.PromptData{})
	if err == nil {
		t.Fatal("expected error for unknown template key")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
	if !strings.Contains(err.Error(), "unknown template key") {
		t.Errorf("error = %q, want mention of unknown template key", err.Error())
	}
}

func TestYAMLRenderer_invalidTemplateSyntax(t *testing.T) {
	t.Parallel()

	_, err := prompt.NewYAMLRenderer(map[string]config.PromptSet{
		"bad": {System: "{{.TargetLang"},
	})
	if err == nil {
		t.Fatal("expected parse error for invalid template")
	}
	if !strings.Contains(err.Error(), "parse prompt template") {
		t.Errorf("error = %q, want parse failure message", err.Error())
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
