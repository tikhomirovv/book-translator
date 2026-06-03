package integration_test

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
	"github.com/tikhomirovv/book-translator/internal/infrastructure/chunk"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/config"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/extract"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/prompt"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/store"
)

// integrationLLM is an offline stub — no HTTP calls.
type integrationLLM struct{}

func (integrationLLM) Chat(ctx context.Context, req ports.ChatRequest) (*ports.ChatResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	user := ""
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			user = msg.Content
		}
	}
	if strings.Contains(user, "You maintain rolling") {
		return &ports.ChatResponse{
			Content: `integration context note`,
			Usage:   ports.ChatUsage{PromptTokens: 3, CompletionTokens: 2, TotalTokens: 5},
		}, nil
	}
	return &ports.ChatResponse{
		Content: "Translated paragraph.",
		Usage:   ports.ChatUsage{PromptTokens: 8, CompletionTokens: 4, TotalTokens: 12},
	}, nil
}

func TestE2E_startTranslation_writesOutput(t *testing.T) {
	stack := newTestStack(t)

	pdfPath := filepath.Join(stack.dir, "sample.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4 stub"), 0o644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(stack.dir, "book.ru.md")

	result, err := stack.start.Execute(context.Background(), translate.StartTranslationRequest{
		SourcePath: pdfPath,
		OutputPath: outPath,
		TargetLang: "ru",
		PromptType: "nonfiction",
	})
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	content := string(data)
	for _, want := range []string{"target_lang: ru", "Translated paragraph.", result.TranslationID} {
		if !strings.Contains(content, want) {
			t.Errorf("output missing %q", want)
		}
	}

	chunksDir := filepath.Join(stack.dir, "translations", result.TranslationID, "chunks")
	entries, err := os.ReadDir(chunksDir)
	if err != nil {
		t.Fatalf("chunks dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected chunk files on disk")
	}
}

func TestE2E_resumeTranslation_completesPartialJob(t *testing.T) {
	stack := newTestStack(t)
	ctx := context.Background()

	pdfPath := filepath.Join(stack.dir, "resume.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4 stub"), 0o644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(stack.dir, "resume-out.md")

	tr := domain.NewTranslation("", pdfPath, outPath, "ru", "nonfiction")
	if err := stack.store.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}
	state, _, err := stack.store.Load(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	state.TotalChunks = 2
	state.LastCompletedChunk = 1
	if err := stack.store.SaveState(ctx, tr.ID, state); err != nil {
		t.Fatal(err)
	}
	if err := stack.store.SaveChunk(ctx, tr.ID, domain.Chunk{
		Index:          1,
		TranslatedText: "First chunk done.",
	}); err != nil {
		t.Fatal(err)
	}

	if err := stack.resume.Execute(ctx, resume.ResumeTranslationRequest{TranslationID: tr.ID}); err != nil {
		t.Fatalf("resume: %v", err)
	}

	view, err := stack.status.Execute(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != domain.StatusCompleted {
		t.Fatalf("status = %s", view.Status)
	}
	if view.CompletedChunks != 2 {
		t.Fatalf("completed = %d, want 2", view.CompletedChunks)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "First chunk done.") || !strings.Contains(content, "Translated paragraph.") {
		t.Fatalf("output = %s", content)
	}
}

type testStack struct {
	dir     string
	store   *store.FilesystemStore
	start   *translate.StartTranslation
	resume  *resume.ResumeTranslation
	status  *query.GetStatus
}

func newTestStack(t *testing.T) *testStack {
	t.Helper()

	dir := t.TempDir()
	cfg, err := config.Load(filepath.Join(repoRoot(t), "configs"))
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := prompt.NewYAMLRenderer(cfg.Prompts)
	if err != nil {
		t.Fatal(err)
	}

	fs := store.NewFilesystemStore(filepath.Join(dir, "translations"))
	registry := extract.NewRegistry()
	registry.Register(".pdf", extract.NewLedongthucPDFWithReader(func(string) (string, error) {
		return "Alpha paragraph.\n\nBeta paragraph.", nil
	}))

	llm := integrationLLM{}
	process := &translate.ProcessChunk{
		LLM:     llm,
		Store:   fs,
		Prompts: renderer,
		TranslationLLM: translate.LLMCallParams{
			Model:       "mock",
			Temperature: 0.3,
			MaxTokens:   512,
		},
		ContextLLM: translate.LLMCallParams{
			Model:       "mock",
			Temperature: 0.3,
			MaxTokens:   512,
		},
	}

	newContext := func(id string) ports.ContextManager {
		return contextmgr.NewFixedWindow(fs, id)
	}

	start := &translate.StartTranslation{
		Extractor:         registry,
		Store:             fs,
		ProcessChunk:      process,
		Finalize:          &translate.FinalizeTranslation{Store: fs},
		NewContext:        newContext,
		BuildChunks:       chunk.BuildChunks,
		IsLanguageAllowed: cfg.IsLanguageAllowed,
		ChunkSize:         1,
		Overlap:           0,
		ParagraphFrom:     -1,
		ParagraphTo:       -1,
		DefaultPromptType: "nonfiction",
		Model:             "mock",
		Provider:          "mock",
	}

	resumeUC := &resume.ResumeTranslation{
		Extractor:    registry,
		Store:        fs,
		ProcessChunk: process,
		Finalize:     &translate.FinalizeTranslation{Store: fs},
		NewContext:   newContext,
		BuildChunks:  chunk.BuildChunks,
		ChunkSize:    1,
		Overlap:      0,
		ParagraphFrom: -1,
		ParagraphTo:   -1,
		Model:        "mock",
		Provider:     "mock",
	}

	return &testStack{
		dir:    dir,
		store:  fs,
		start:  start,
		resume: resumeUC,
		status: &query.GetStatus{Store: fs},
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
	t.Fatal("repo root not found")
	return ""
}
