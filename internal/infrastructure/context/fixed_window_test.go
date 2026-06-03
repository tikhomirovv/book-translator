package contextmgr_test

import (
	"context"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/domain"
	contextmgr "github.com/tikhomirovv/book-translator/internal/infrastructure/context"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/store"
)

func TestFixedWindow_replacesContext(t *testing.T) {
	t.Parallel()

	s := store.NewFilesystemStore(t.TempDir())
	ctx := context.Background()
	tr := domain.NewTranslation("", "/a.pdf", "/b.md", "ru", "nonfiction")
	if err := s.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}

	mgr := contextmgr.NewFixedWindow(s, tr.ID)
	if err := mgr.AddExtracted(ctx, 1, map[string]any{"summary": "first context"}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.AddExtracted(ctx, 2, map[string]any{"summary": "second context"}); err != nil {
		t.Fatal(err)
	}

	block, err := mgr.BuildPromptContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if block != "second context" {
		t.Fatalf("context = %q, want replacement not append", block)
	}
}

func TestFixedWindow_emptyContext(t *testing.T) {
	t.Parallel()

	mgr := contextmgr.NewFixedWindow(store.NewFilesystemStore(t.TempDir()), "id")
	block, err := mgr.BuildPromptContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if block != "" {
		t.Fatalf("empty context = %q", block)
	}
}
