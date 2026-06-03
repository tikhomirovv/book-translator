package resume_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/application/query"
	"github.com/tikhomirovv/book-translator/internal/application/resume"
	"github.com/tikhomirovv/book-translator/internal/application/translate"
	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
	contextmgr "github.com/tikhomirovv/book-translator/internal/infrastructure/context"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/config"
	chunkinfra "github.com/tikhomirovv/book-translator/internal/infrastructure/chunk"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/prompt"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/store"
)

type fixtureExtractor struct {
	paragraphs []domain.Paragraph
}

func (f *fixtureExtractor) Extract(ctx context.Context, path string) ([]domain.Paragraph, error) {
	return f.paragraphs, nil
}

type resumeMockLLM struct{}

func (m *resumeMockLLM) Chat(ctx context.Context, req ports.ChatRequest) (*ports.ChatResponse, error) {
	user := ""
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			user = msg.Content
		}
	}
	if strings.Contains(user, "You maintain rolling") {
		return &ports.ChatResponse{Content: `{"summary":"note"}`, Usage: ports.ChatUsage{TotalTokens: 1}}, nil
	}
	return &ports.ChatResponse{Content: "translated", Usage: ports.ChatUsage{TotalTokens: 2}}, nil
}

func TestResumeTranslation_skipsCompletedChunks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base := t.TempDir()
	fs := store.NewFilesystemStore(base)
	outPath := filepath.Join(base, "out.md")

	cfg, err := config.Load(filepath.Join(repoRoot(t), "configs"))
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := prompt.NewYAMLRenderer(cfg.Prompts)
	if err != nil {
		t.Fatal(err)
	}

	tr := domain.NewTranslation("", "/book.pdf", outPath, "ru", "nonfiction")
	if err := fs.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}
	state, _, err := fs.Load(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	state.TotalChunks = 2
	state.LastCompletedChunk = 1
	if err := fs.SaveState(ctx, tr.ID, state); err != nil {
		t.Fatal(err)
	}
	if err := fs.SaveChunk(ctx, tr.ID, domain.Chunk{Index: 1, TranslatedText: "chunk one"}); err != nil {
		t.Fatal(err)
	}

	process := &translate.ProcessChunk{
		LLM:            &resumeMockLLM{},
		Store:          fs,
		Prompts:        renderer,
		TranslationLLM: translate.LLMCallParams{Model: "test"},
		ContextLLM:     translate.LLMCallParams{Model: "test"},
	}

	uc := &resume.ResumeTranslation{
		Extractor: &fixtureExtractor{
			paragraphs: []domain.Paragraph{
				{Index: 0, Text: "p1"},
				{Index: 1, Text: "p2"},
			},
		},
		Store:        fs,
		ProcessChunk: process,
		Finalize:     &translate.FinalizeTranslation{Store: fs},
		NewContext: func(id string) ports.ContextManager {
			return contextmgr.NewFixedWindow(fs, id)
		},
		BuildChunks: chunkinfra.BuildChunks,
		ChunkSize:   1,
		Overlap:     0,
		ParagraphFrom: -1,
		ParagraphTo:   -1,
		Model:       "test",
	}

	if err := uc.Execute(ctx, resume.ResumeTranslationRequest{TranslationID: tr.ID}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	status, err := (&query.GetStatus{Store: fs}).Execute(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if status.CompletedChunks != 2 {
		t.Fatalf("completed = %d, want 2", status.CompletedChunks)
	}
	if status.Status != domain.StatusCompleted {
		t.Fatalf("status = %s", status.Status)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "chunk one") || !strings.Contains(string(data), "translated") {
		t.Fatalf("output = %s", string(data))
	}
}

func TestResumeTranslation_idempotentWhenCompleted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fs := store.NewFilesystemStore(t.TempDir())
	tr := domain.NewTranslation("", "/a.pdf", "/b.md", "ru", "nonfiction")
	tr.Status = domain.StatusCompleted
	if err := fs.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}

	uc := &resume.ResumeTranslation{
		Extractor:    &fixtureExtractor{},
		Store:        fs,
		ProcessChunk: &translate.ProcessChunk{},
		Finalize:     &translate.FinalizeTranslation{Store: fs},
		NewContext:   func(id string) ports.ContextManager { return contextmgr.NewFixedWindow(fs, id) },
		BuildChunks:  chunkinfra.BuildChunks,
	}

	if err := uc.Execute(ctx, resume.ResumeTranslationRequest{TranslationID: tr.ID}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Skip("repo root not found")
	return ""
}
