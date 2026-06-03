package chunk_test

import (
	"testing"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/chunk"
)

func TestFilterParagraphs_noRange(t *testing.T) {
	t.Parallel()

	paras := []domain.Paragraph{{Index: 0, Text: "a"}, {Index: 1, Text: "b"}}
	got := chunk.FilterParagraphs(paras, chunk.ParagraphRange{From: -1, To: -1})
	if len(got) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", len(got))
	}
}

func TestFilterParagraphs_inclusiveRange(t *testing.T) {
	t.Parallel()

	var paras []domain.Paragraph
	for i := 0; i < 10; i++ {
		paras = append(paras, domain.Paragraph{Index: i, Text: "p"})
	}

	got := chunk.FilterParagraphs(paras, chunk.ParagraphRange{From: 3, To: 5})
	if len(got) != 3 {
		t.Fatalf("expected 3 paragraphs, got %d", len(got))
	}
	if got[0].Index != 3 || got[2].Index != 5 {
		t.Fatalf("unexpected indices: %+v", got)
	}
}

func TestFilterParagraphs_openEnd(t *testing.T) {
	t.Parallel()

	paras := []domain.Paragraph{
		{Index: 0, Text: "a"},
		{Index: 5, Text: "b"},
		{Index: 9, Text: "c"},
	}

	got := chunk.FilterParagraphs(paras, chunk.ParagraphRange{From: 5, To: -1})
	if len(got) != 2 || got[0].Index != 5 || got[1].Index != 9 {
		t.Fatalf("unexpected filter result: %+v", got)
	}
}
