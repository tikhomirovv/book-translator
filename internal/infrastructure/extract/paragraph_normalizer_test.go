package extract_test

import (
	"testing"

	"github.com/tikhomirovv/book-translator/internal/infrastructure/extract"
)

func TestNormalizeParagraphs(t *testing.T) {
	t.Parallel()

	raw := "  First paragraph.  \n\nSecond.\r\n\r\n\n\nThird."
	paras := extract.NormalizeParagraphs(raw)
	if len(paras) != 3 {
		t.Fatalf("expected 3 paragraphs, got %d", len(paras))
	}
	if paras[0].Index != 0 || paras[0].Text != "First paragraph." {
		t.Fatalf("first: %+v", paras[0])
	}
	if paras[2].Text != "Third." {
		t.Fatalf("third: %+v", paras[2])
	}
}

func TestNormalizeParagraphs_doubleNewline(t *testing.T) {
	t.Parallel()

	// Covered in paragraph_reflow_test.go with richer fixtures.
	raw := "Alpha.\n\nBeta."
	paras := extract.NormalizeParagraphs(raw)
	if len(paras) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", len(paras))
	}
}
