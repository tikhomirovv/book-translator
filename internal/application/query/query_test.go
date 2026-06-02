package query_test

import (
	"context"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/application/query"
	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/store"
)

func TestGetStatus_returnsProgressAndUsage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fs := store.NewFilesystemStore(t.TempDir())

	tr := domain.NewTranslation("", "/books/a.pdf", "/out/a.md", "ru", "nonfiction")
	if err := fs.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}

	state, _, err := fs.Load(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	state.LastCompletedChunk = 2
	state.TotalChunks = 5
	state.Usage = domain.Usage{TotalTokens: 42}
	state.LastError = "previous failure"
	if err := fs.SaveState(ctx, tr.ID, state); err != nil {
		t.Fatal(err)
	}

	uc := &query.GetStatus{Store: fs}
	view, err := uc.Execute(ctx, tr.ID)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if view.CompletedChunks != 2 || view.TotalChunks != 5 {
		t.Fatalf("progress = %d/%d", view.CompletedChunks, view.TotalChunks)
	}
	if view.Usage.TotalTokens != 42 {
		t.Fatalf("usage = %d", view.Usage.TotalTokens)
	}
	if view.LastError != "previous failure" {
		t.Fatalf("last error = %q", view.LastError)
	}
}

func TestListTranslations_returnsRows(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fs := store.NewFilesystemStore(t.TempDir())

	tr := domain.NewTranslation("", "/books/a.pdf", "/out/a.md", "de", "fiction")
	if err := fs.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}
	state, _, err := fs.Load(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	state.LastCompletedChunk = 1
	state.TotalChunks = 3
	if err := fs.SaveState(ctx, tr.ID, state); err != nil {
		t.Fatal(err)
	}

	uc := &query.ListTranslations{Store: fs}
	items, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %d, want 1", len(items))
	}
	if items[0].Progress != "1/3" {
		t.Fatalf("progress = %q", items[0].Progress)
	}
	if items[0].TargetLang != "de" {
		t.Fatalf("lang = %q", items[0].TargetLang)
	}
}
