package extract_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/tikhomirovv/book-translater/internal/domain"
	"github.com/tikhomirovv/book-translater/internal/infrastructure/extract"
)

func TestRegistry_ExtractPDF_mocked(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "book.pdf")
	if err := os.WriteFile(path, []byte("%PDF"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := extract.NewRegistry()
	r.Register(".pdf", extract.NewLedongthucPDFWithReader(func(string) (string, error) {
		return "Alpha.\n\nBeta.", nil
	}))

	paras, err := r.Extract(context.Background(), path)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(paras) != 2 || paras[0].Text != "Alpha." {
		t.Fatalf("unexpected paragraphs: %+v", paras)
	}
}

func TestRegistry_UnsupportedExtension(t *testing.T) {
	t.Parallel()

	r := extract.NewRegistry()
	_, err := r.Extract(context.Background(), "/tmp/book.epub")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
