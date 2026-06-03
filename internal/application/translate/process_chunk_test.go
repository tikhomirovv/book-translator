package translate_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/application/translate"
	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
	contextmgr "github.com/tikhomirovv/book-translator/internal/infrastructure/context"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/config"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/prompt"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/store"
)

type mockLLM struct {
	calls int
}

type trackingLLM struct {
	translateModel       string
	contextModel         string
	translateTemperature float64
	contextTemperature   float64
	translateMaxTokens   int
	contextMaxTokens     int
	models               []string
}

func (m *trackingLLM) Chat(ctx context.Context, req ports.ChatRequest) (*ports.ChatResponse, error) {
	m.models = append(m.models, req.Model)

	userContent := ""
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			userContent = msg.Content
		}
	}

	switch {
	case strings.Contains(userContent, "You maintain rolling translation memory"):
		if req.Model != m.contextModel {
			return nil, errors.New("unexpected context model")
		}
		if req.Temperature != m.contextTemperature {
			return nil, errors.New("unexpected context temperature")
		}
		if req.MaxTokens != m.contextMaxTokens {
			return nil, errors.New("unexpected context max_tokens")
		}
		return &ports.ChatResponse{
			Content: `{"summary":"notes"}`,
			Usage:   ports.ChatUsage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2},
		}, nil
	case strings.Contains(userContent, "Translate into") || strings.Contains(userContent, "Translate to"):
		if req.Model != m.translateModel {
			return nil, errors.New("unexpected translation model")
		}
		if req.Temperature != m.translateTemperature {
			return nil, errors.New("unexpected translation temperature")
		}
		if req.MaxTokens != m.translateMaxTokens {
			return nil, errors.New("unexpected translation max_tokens")
		}
		return &ports.ChatResponse{
			Content: "Translated.",
			Usage:   ports.ChatUsage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2},
		}, nil
	default:
		return nil, errors.New("unexpected prompt")
	}
}

func (m *mockLLM) Chat(ctx context.Context, req ports.ChatRequest) (*ports.ChatResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.calls++

	userContent := ""
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			userContent = msg.Content
		}
	}

	// Route by rendered prompt content from config templates.
	switch {
	case strings.Contains(userContent, "You maintain rolling translation memory"):
		return &ports.ChatResponse{
			Content: `updated translation memory`,
			Usage:   ports.ChatUsage{PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8},
		}, nil
	case strings.Contains(userContent, "Translate into") || strings.Contains(userContent, "Translate to"):
		return &ports.ChatResponse{
			Content: "Переведённый текст.",
			Usage:   ports.ChatUsage{PromptTokens: 10, CompletionTokens: 4, TotalTokens: 14},
		}, nil
	default:
		return nil, errors.New("unexpected prompt")
	}
}

func TestProcessChunk_success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base := t.TempDir()
	fs := store.NewFilesystemStore(base)

	tr := domain.NewTranslation("", "/book.pdf", "/out.md", "ru", "nonfiction")
	if err := fs.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}

	state, _, err := fs.Load(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	state.TotalChunks = 2
	if err := fs.SaveState(ctx, tr.ID, state); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(filepath.Join(repoRoot(t), "configs"))
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := prompt.NewYAMLRenderer(cfg.Prompts)
	if err != nil {
		t.Fatal(err)
	}

	mgr := contextmgr.NewFixedWindow(fs, tr.ID)
	if err := mgr.Load(ctx, tr.ID); err != nil {
		t.Fatal(err)
	}

	llm := &mockLLM{}
	uc := &translate.ProcessChunk{
		LLM:     llm,
		Store:   fs,
		Prompts: renderer,
		Context: mgr,
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

	chunk := domain.Chunk{
		Index:           1,
		ParagraphStart:  0,
		ParagraphEnd:    2,
		SourceText:      "Alice went to the market.",
		OverlapFromPrev: "",
	}

	if err := uc.Execute(ctx, translate.ProcessChunkRequest{
		TranslationID: tr.ID,
		PromptType:    "nonfiction",
		Chunk:         chunk,
	}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if llm.calls != 2 {
		t.Fatalf("LLM calls = %d, want 2 parallel requests", llm.calls)
	}

	reloaded, loadedTr, err := fs.Load(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.LastCompletedChunk != 1 {
		t.Fatalf("LastCompletedChunk = %d, want 1", reloaded.LastCompletedChunk)
	}
	if reloaded.Usage.TotalTokens != 22 {
		t.Fatalf("usage tokens = %d, want 22", reloaded.Usage.TotalTokens)
	}
	if reloaded.ContextSummary != "updated translation memory" {
		t.Fatalf("context summary = %q, want consolidated LLM output", reloaded.ContextSummary)
	}
	if loadedTr.LastCompletedChunk != 1 {
		t.Fatalf("translation LastCompletedChunk = %d", loadedTr.LastCompletedChunk)
	}
}

func TestProcessChunk_independentLLMProfiles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fs := store.NewFilesystemStore(t.TempDir())
	tr := domain.NewTranslation("", "/book.pdf", "/out.md", "ru", "nonfiction")
	if err := fs.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}

	state, _, err := fs.Load(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	state.TotalChunks = 1
	if err := fs.SaveState(ctx, tr.ID, state); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(filepath.Join(repoRoot(t), "configs"))
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := prompt.NewYAMLRenderer(cfg.Prompts)
	if err != nil {
		t.Fatal(err)
	}

	mgr := contextmgr.NewFixedWindow(fs, tr.ID)
	if err := mgr.Load(ctx, tr.ID); err != nil {
		t.Fatal(err)
	}

	llm := &trackingLLM{
		translateModel:       "big-translate",
		contextModel:         "small-context",
		translateTemperature: 0.4,
		contextTemperature:   0.1,
		translateMaxTokens:   16000,
		contextMaxTokens:     2048,
	}
	uc := &translate.ProcessChunk{
		LLM:     llm,
		Store:   fs,
		Prompts: renderer,
		Context: mgr,
		TranslationLLM: translate.LLMCallParams{
			Model:       "big-translate",
			Temperature: 0.4,
			MaxTokens:   16000,
		},
		ContextLLM: translate.LLMCallParams{
			Model:       "small-context",
			Temperature: 0.1,
			MaxTokens:   2048,
		},
	}

	if err := uc.Execute(ctx, translate.ProcessChunkRequest{
		TranslationID: tr.ID,
		PromptType:    "nonfiction",
		Chunk: domain.Chunk{
			Index:      1,
			SourceText: "Sample paragraph.",
		},
	}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(llm.models) != 2 {
		t.Fatalf("models used: %v, want 2 calls", llm.models)
	}
}

func TestProcessChunk_enforcesSequentialIndex(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fs := store.NewFilesystemStore(t.TempDir())
	tr := domain.NewTranslation("", "/book.pdf", "/out.md", "ru", "nonfiction")
	if err := fs.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(filepath.Join(repoRoot(t), "configs"))
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := prompt.NewYAMLRenderer(cfg.Prompts)
	if err != nil {
		t.Fatal(err)
	}

	mgr := contextmgr.NewFixedWindow(fs, tr.ID)
	if err := mgr.Load(ctx, tr.ID); err != nil {
		t.Fatal(err)
	}

	uc := &translate.ProcessChunk{
		LLM:     &mockLLM{},
		Store:   fs,
		Prompts: renderer,
		Context: mgr,
		TranslationLLM: translate.LLMCallParams{Model: "test"},
		ContextLLM:     translate.LLMCallParams{Model: "test"},
	}

	err = uc.Execute(ctx, translate.ProcessChunkRequest{
		TranslationID: tr.ID,
		PromptType:    "nonfiction",
		Chunk: domain.Chunk{
			Index:      2,
			SourceText: "skip first",
		},
	})
	if err == nil {
		t.Fatal("expected sequential index error")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
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
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Skip("repo root not found")
	return ""
}
