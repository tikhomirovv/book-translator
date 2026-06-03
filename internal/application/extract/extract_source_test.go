package extracttext_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	extracttext "github.com/tikhomirovv/book-translator/internal/application/extract"
	"github.com/tikhomirovv/book-translator/internal/domain"
)

type stubExtractor struct {
	paragraphs []domain.Paragraph
}

func (s stubExtractor) Extract(ctx context.Context, path string) ([]domain.Paragraph, error) {
	return s.paragraphs, nil
}

func TestExtractSource_writesParagraphFile(t *testing.T) {
	t.Parallel()

	outPath := filepath.Join(t.TempDir(), "out", "source.txt")
	uc := &extracttext.ExtractSource{
		Extractor: stubExtractor{
			paragraphs: []domain.Paragraph{
				{Index: 0, Text: "First"},
				{Index: 1, Text: "Second"},
			},
		},
	}

	result, err := uc.Execute(context.Background(), extracttext.ExtractSourceRequest{
		InputPath:  "book.pdf",
		OutputPath: outPath,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.ParagraphCount != 2 {
		t.Fatalf("paragraphs = %d", result.ParagraphCount)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "# paragraph 0\nFirst") {
		t.Fatalf("unexpected content:\n%s", content)
	}
}
