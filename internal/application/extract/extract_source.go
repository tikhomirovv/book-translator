package extracttext

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/extract"
)

// ExtractSource runs PDF/text extraction without calling the LLM.
type ExtractSource struct {
	Extractor ports.TextExtractor
}

// ExtractSourceRequest identifies input file and output path.
type ExtractSourceRequest struct {
	InputPath  string
	OutputPath string
}

// ExtractSourceResult summarizes extraction output.
type ExtractSourceResult struct {
	ParagraphCount int
	OutputPath     string
}

// Execute extracts paragraphs and writes them to a text file for inspection.
func (uc *ExtractSource) Execute(ctx context.Context, req ExtractSourceRequest) (*ExtractSourceResult, error) {
	if uc == nil || uc.Extractor == nil {
		return nil, domain.ErrInvalidInput
	}
	if req.InputPath == "" || req.OutputPath == "" {
		return nil, domain.ErrInvalidInput
	}

	paragraphs, err := uc.Extractor.Extract(ctx, req.InputPath)
	if err != nil {
		return nil, fmt.Errorf("extract text: %w", err)
	}
	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("%w: no text extracted from source", domain.ErrInvalidInput)
	}

	outDir := filepath.Dir(req.OutputPath)
	if outDir != "." && outDir != "" {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return nil, fmt.Errorf("create output dir: %w", err)
		}
	}

	content := extract.FormatParagraphs(paragraphs)
	if err := os.WriteFile(req.OutputPath, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("write output: %w", err)
	}

	return &ExtractSourceResult{
		ParagraphCount: len(paragraphs),
		OutputPath:     req.OutputPath,
	}, nil
}
