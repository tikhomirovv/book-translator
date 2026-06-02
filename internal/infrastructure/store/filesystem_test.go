package store_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/store"
)

func TestFilesystemStore_CreateLoadSave(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	s := store.NewFilesystemStore(base)
	ctx := context.Background()

	tr := domain.NewTranslation("", "/books/sample.pdf", "/out/sample.md", "ru", "nonfiction")
	if err := s.Create(ctx, tr); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tr.ID == "" {
		t.Fatal("expected generated uuid id")
	}
	if _, err := uuid.Parse(tr.ID); err != nil {
		t.Fatalf("id is not uuid: %v", err)
	}

	state, loaded, err := s.Load(ctx, tr.ID)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.SourcePath != tr.SourcePath || loaded.TargetLang != "ru" {
		t.Fatalf("unexpected translation: %+v", loaded)
	}
	if state.TranslationID != tr.ID {
		t.Fatalf("state id mismatch: %s", state.TranslationID)
	}

	state.ContextSummary = "book about habits"
	state.LastCompletedChunk = 1
	state.TotalChunks = 10
	state.Glossary = map[string]any{"term": "определение"}
	state.Usage.PromptTokens = 100
	if err := s.SaveState(ctx, tr.ID, state); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	chunk := domain.Chunk{
		Index:           1,
		ParagraphStart:  0,
		ParagraphEnd:    10,
		SourceText:      "hello",
		TranslatedText:  "# Привет\n\nТекст.",
		OverlapFromPrev: "prev",
	}
	if err := s.SaveChunk(ctx, tr.ID, chunk); err != nil {
		t.Fatalf("SaveChunk: %v", err)
	}

	chunkPath := filepath.Join(base, tr.ID, "chunks", "0001.md")
	if _, err := os.Stat(chunkPath); err != nil {
		t.Fatalf("chunk file missing: %v", err)
	}

	tr.Status = domain.StatusRunning
	tr.LastCompletedChunk = 1
	tr.TotalChunks = 10
	if err := s.UpdateTranslation(ctx, tr); err != nil {
		t.Fatalf("UpdateTranslation: %v", err)
	}

	reloadedState, reloadedTr, err := s.Load(ctx, tr.ID)
	if err != nil {
		t.Fatalf("Load after updates: %v", err)
	}
	if reloadedTr.Status != domain.StatusRunning {
		t.Fatalf("status: %s", reloadedTr.Status)
	}
	if reloadedState.ContextSummary != "book about habits" {
		t.Fatalf("summary: %q", reloadedState.ContextSummary)
	}
	if reloadedState.Usage.PromptTokens != 100 {
		t.Fatalf("usage tokens: %d", reloadedState.Usage.PromptTokens)
	}
}

func TestFilesystemStore_List(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	s := store.NewFilesystemStore(base)
	ctx := context.Background()

	for _, lang := range []string{"ru", "de"} {
		tr := domain.NewTranslation("", "/a.pdf", "/b.md", lang, "fiction")
		if err := s.Create(ctx, tr); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	list, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(list))
	}
}

func TestFilesystemStore_LoadNotFound(t *testing.T) {
	t.Parallel()

	s := store.NewFilesystemStore(t.TempDir())
	_, _, err := s.Load(context.Background(), uuid.NewString())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestFilesystemStore_CreateDuplicate(t *testing.T) {
	t.Parallel()

	s := store.NewFilesystemStore(t.TempDir())
	ctx := context.Background()
	id := store.NewTranslationID()
	tr := domain.NewTranslation(id, "/a.pdf", "/b.md", "ru", "nonfiction")
	if err := s.Create(ctx, tr); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if err := s.Create(ctx, tr); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestNewTranslationID_IsUUIDv4(t *testing.T) {
	t.Parallel()

	id := store.NewTranslationID()
	parsed, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Version() != 4 {
		t.Fatalf("expected uuid v4, got version %d", parsed.Version())
	}
}
