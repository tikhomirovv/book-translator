package translate_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/application/translate"
	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/store"
)

func TestFinalizeTranslation_writesFrontmatter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base := t.TempDir()
	fs := store.NewFilesystemStore(base)
	outPath := filepath.Join(base, "final.md")

	tr := domain.NewTranslation("", "/src.pdf", outPath, "de", "nonfiction")
	if err := fs.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}

	state, _, err := fs.Load(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	state.TotalChunks = 2
	state.LastCompletedChunk = 2
	state.Usage = domain.Usage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30}
	if err := fs.SaveState(ctx, tr.ID, state); err != nil {
		t.Fatal(err)
	}

	for i, text := range []string{"# Part one", "# Part two"} {
		if err := fs.SaveChunk(ctx, tr.ID, domain.Chunk{
			Index:          i + 1,
			ParagraphStart: i,
			ParagraphEnd:   i + 1,
			TranslatedText: text,
		}); err != nil {
			t.Fatal(err)
		}
	}

	uc := &translate.FinalizeTranslation{Store: fs}
	if err := uc.Execute(ctx, translate.FinalizeTranslationRequest{
		TranslationID: tr.ID,
		Model:         "gpt-test",
		Provider:      "openai",
	}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	content := string(data)
	for _, want := range []string{
		"target_lang: de",
		"model: gpt-test",
		"provider: openai",
		"total_tokens: 30",
		"# Part one",
		"# Part two",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("output missing %q:\n%s", want, content)
		}
	}

	_, tr, err = fs.Load(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if tr.Status != domain.StatusCompleted {
		t.Fatalf("status = %s, want completed", tr.Status)
	}
}
