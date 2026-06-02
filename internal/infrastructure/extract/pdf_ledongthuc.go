package extract

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ledongthuc/pdf"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

// plainTextReader loads raw text from a PDF path (overridable in tests).
type plainTextReader func(path string) (string, error)

// LedongthucPDF implements TextExtractor using github.com/ledongthuc/pdf.
type LedongthucPDF struct {
	readPlain plainTextReader
}

// NewLedongthucPDF creates the default PDF extractor.
func NewLedongthucPDF() *LedongthucPDF {
	return &LedongthucPDF{readPlain: readPDFPlainText}
}

// NewLedongthucPDFWithReader is for tests and custom backends.
func NewLedongthucPDFWithReader(reader plainTextReader) *LedongthucPDF {
	return &LedongthucPDF{readPlain: reader}
}

var _ ports.TextExtractor = (*LedongthucPDF)(nil)

// Extract reads a PDF file and returns normalized paragraphs.
func (e *LedongthucPDF) Extract(ctx context.Context, path string) ([]domain.Paragraph, error) {
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
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	reader, err := r.GetPlainText()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return "", err
	}
	return buf.String(), nil
}
