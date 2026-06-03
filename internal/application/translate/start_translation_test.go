package translate_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/application/translate"
	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
	contextmgr "github.com/tikhomirovv/book-translator/internal/infrastructure/context"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/config"
	chunkinfra "github.com/tikhomirovv/book-translator/internal/infrastructure/chunk"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/prompt"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/store"
)

type mockExtractor struct {
	paragraphs []domain.Paragraph
}

func (m *mockExtractor) Extract(ctx context.Context, path string) ([]domain.Paragraph, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return m.paragraphs, nil
}

func TestStartTranslation_endToEndWithMocks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base := t.TempDir()
	fs := store.NewFilesystemStore(base)
	outPath := filepath.Join(base, "book.ru.md")

	cfg, err := config.Load(filepath.Join(repoRoot(t), "configs"))
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := prompt.NewYAMLRenderer(cfg.Prompts)
	if err != nil {
		t.Fatal(err)
	}

	process := &translate.ProcessChunk{
		LLM:     &mockLLM{},
		Store:   fs,
		Prompts: renderer,
		TranslationLLM: translate.LLMCallParams{
			Model:       "test-model",
			Temperature: 0.3,
			MaxTokens:   1024,
		},
		ContextLLM: translate.LLMCallParams{
			Model:       "test-model",
			Temperature: 0.3,
			MaxTokens:   1024,
		},
	}

	start := &translate.StartTranslation{
		Extractor: &mockExtractor{
			paragraphs: []domain.Paragraph{
				{Index: 0, Text: "First paragraph."},
				{Index: 1, Text: "Second paragraph."},
			},
		},
		Store:        fs,
		ProcessChunk: process,
		Finalize:     &translate.FinalizeTranslation{Store: fs},
		NewContext: func(id string) ports.ContextManager {
			return contextmgr.NewFixedWindow(fs, id)
		},
		BuildChunks:       chunkinfra.BuildChunks,
		IsLanguageAllowed: cfg.IsLanguageAllowed,
		ChunkSize:         1,
		Overlap:           0,
		ParagraphFrom:     -1,
		ParagraphTo:       -1,
		DefaultPromptType: "nonfiction",
		Model:             "test-model",
		Provider:          "mock",
	}

	result, err := start.Execute(ctx, translate.StartTranslationRequest{
		SourcePath: "/books/sample.pdf",
		OutputPath: outPath,
		TargetLang: "ru",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	_, tr, err := fs.Load(ctx, result.TranslationID)
	if err != nil {
		t.Fatal(err)
	}
	if tr.Status != domain.StatusCompleted {
		t.Fatalf("status = %s, want completed", tr.Status)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	content := string(data)
	for _, want := range []string{"target_lang: ru", "translation_id:", "model: test-model", "Переведённый текст."} {
		if !strings.Contains(content, want) {
			t.Errorf("output missing %q:\n%s", want, content)
		}
	}
}

func TestStartTranslation_rejectsInvalidLanguage(t *testing.T) {
	t.Parallel()

	start := &translate.StartTranslation{
		Extractor:         &mockExtractor{},
		Store:             store.NewFilesystemStore(t.TempDir()),
		ProcessChunk:      &translate.ProcessChunk{},
		Finalize:          &translate.FinalizeTranslation{},
		NewContext:        func(id string) ports.ContextManager { return contextmgr.NewFixedWindow(nil, id) },
		BuildChunks:       chunkinfra.BuildChunks,
		IsLanguageAllowed: func(lang string) bool { return lang == "ru" },
	}

	_, err := start.Execute(context.Background(), translate.StartTranslationRequest{
		SourcePath: "/a.pdf",
		OutputPath: "/b.md",
		TargetLang: "xx",
	})
	if err == nil {
		t.Fatal("expected invalid language error")
	}
	if !errors.Is(err, domain.ErrInvalidLanguage) {
		t.Fatalf("error = %v, want ErrInvalidLanguage", err)
	}
}

func TestStartTranslation_paragraphRange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fs := store.NewFilesystemStore(t.TempDir())
	outPath := filepath.Join(t.TempDir(), "slice.md")

	cfg, err := config.Load(filepath.Join(repoRoot(t), "configs"))
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := prompt.NewYAMLRenderer(cfg.Prompts)
	if err != nil {
		t.Fatal(err)
	}

	var paras []domain.Paragraph
	for i := 0; i < 10; i++ {
		paras = append(paras, domain.Paragraph{Index: i, Text: fmt.Sprintf("Paragraph %d.", i)})
	}

	process := &translate.ProcessChunk{
		LLM:            &mockLLM{},
		Store:          fs,
		Prompts:        renderer,
		TranslationLLM: translate.LLMCallParams{Model: "test-model", MaxTokens: 1024},
		ContextLLM:     translate.LLMCallParams{Model: "test-model", MaxTokens: 1024},
	}

	start := &translate.StartTranslation{
		Extractor:         &mockExtractor{paragraphs: paras},
		Store:             fs,
		ProcessChunk:      process,
		Finalize:          &translate.FinalizeTranslation{Store: fs},
		NewContext:        func(id string) ports.ContextManager { return contextmgr.NewFixedWindow(fs, id) },
		BuildChunks:       chunkinfra.BuildChunks,
		IsLanguageAllowed: cfg.IsLanguageAllowed,
		ChunkSize:         10,
		Overlap:           0,
		ParagraphFrom:     3,
		ParagraphTo:       5,
		DefaultPromptType: "nonfiction",
		Model:             "test-model",
	}

	result, err := start.Execute(ctx, translate.StartTranslationRequest{
		SourcePath: "/books/sample.pdf",
		OutputPath: outPath,
		TargetLang: "ru",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	state, _, err := fs.Load(ctx, result.TranslationID)
	if err != nil {
		t.Fatal(err)
	}
	if state.TotalChunks != 1 {
		t.Fatalf("total chunks = %d, want 1 for 3-paragraph slice", state.TotalChunks)
	}
}
