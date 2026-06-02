package chunk_test

import (
	"fmt"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/chunk"
)

func TestBuildChunks_overlap(t *testing.T) {
	t.Parallel()

	var paras []domain.Paragraph
	for i := 0; i < 12; i++ {
		paras = append(paras, domain.Paragraph{Index: i, Text: fmt.Sprintf("p%d", i)})
	}

	chunks := chunk.BuildChunks(paras, 10, 2)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0].Index != 1 || chunks[1].Index != 2 {
		t.Fatalf("chunk indices: %+v", chunks)
	}
	if chunks[1].OverlapFromPrev == "" {
		t.Fatal("expected overlap text on second chunk")
	}
	if chunks[0].ParagraphStart != 0 || chunks[0].ParagraphEnd != 10 {
		t.Fatalf("first range: %d-%d", chunks[0].ParagraphStart, chunks[0].ParagraphEnd)
	}
}

func TestBuildChunks_single(t *testing.T) {
	t.Parallel()

	paras := []domain.Paragraph{{Index: 0, Text: "only"}}
	chunks := chunk.BuildChunks(paras, 10, 2)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
}
