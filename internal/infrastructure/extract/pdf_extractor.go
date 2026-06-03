package extract

import (
	"context"
	"fmt"
	"os"

	gopdf "github.com/razvandimescu/gopdf/pdf"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

// plainTextReader loads raw text from a PDF path (overridable in tests).
type plainTextReader func(path string) (string, error)

// PDFExtractor implements TextExtractor using github.com/razvandimescu/gopdf.
// gopdf handles ToUnicode CMaps better than plain ledongthuc/pdf for many ebooks.
type PDFExtractor struct {
	readPlain plainTextReader
}

// NewPDFExtractor creates the default PDF extractor.
func NewPDFExtractor() *PDFExtractor {
	return &PDFExtractor{readPlain: readPDFPlainText}
}

// NewPDFExtractorWithReader is for tests and custom backends.
func NewPDFExtractorWithReader(reader plainTextReader) *PDFExtractor {
	return &PDFExtractor{readPlain: reader}
}

// NewLedongthucPDF is kept as an alias for older call sites.
func NewLedongthucPDF() *PDFExtractor {
	return NewPDFExtractor()
}

// NewLedongthucPDFWithReader is kept as an alias for older tests.
func NewLedongthucPDFWithReader(reader plainTextReader) *PDFExtractor {
	return NewPDFExtractorWithReader(reader)
}

var _ ports.TextExtractor = (*PDFExtractor)(nil)

// Extract reads a PDF file and returns normalized paragraphs.
func (e *PDFExtractor) Extract(ctx context.Context, path string) ([]domain.Paragraph, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if path == "" {
		return nil, domain.ErrInvalidInput
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", domain.ErrInvalidInput, path)
		}
		return nil, err
	}

	raw, err := e.readPlain(path)
	if err != nil {
		return nil, fmt.Errorf("read pdf %s: %w", path, err)
	}
	return NormalizeParagraphs(raw), nil
}

func readPDFPlainText(path string) (string, error) {
	doc, err := gopdf.OpenFile(path)
	if err != nil {
		return "", err
	}
	text, err := doc.Text()
	if err != nil {
		return "", err
	}
	return text, nil
}
