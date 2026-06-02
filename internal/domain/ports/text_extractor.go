package ports

import (
	"context"

	"github.com/tikhomirovv/book-translater/internal/domain"
)

// TextExtractor reads a source file into normalized paragraphs.
type TextExtractor interface {
	Extract(ctx context.Context, path string) ([]domain.Paragraph, error)
}
