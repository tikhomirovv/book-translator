package contextmgr_test

import (
	"context"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/domain"
	contextmgr "github.com/tikhomirovv/book-translator/internal/infrastructure/context"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/store"
)

func TestFixedWindow_enforceBudget(t *testing.T) {
	t.Parallel()

	s := store.NewFilesystemStore(t.TempDir())
	ctx := context.Background()
	tr := domain.NewTranslation("", "/a.pdf", "/b.md", "ru", "nonfiction")
	if err := s.Create(ctx, tr); err != nil {
		t.Fatal(err)
	}

	mgr := contextmgr.NewFixedWindow(s, tr.ID, 20)
	long := stringsRepeat("word ", 200)
	if err := mgr.AddExtracted(ctx, 1, map[string]any{
		"summary": long,
		"glossary": map[string]any{
			"term": "definition",
		},
	}); err != nil {
		t.Fatal(err)
	}

	block, err := mgr.BuildPromptContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(block) >= len(long) {
		t.Fatalf("expected trimmed context, got len=%d", len(block))
	}
}

func stringsRepeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
