package extract

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

// Registry routes extraction by file extension.
type Registry struct {
	byExt map[string]ports.TextExtractor
}

// NewRegistry builds a registry with MVP defaults (.pdf).
func NewRegistry() *Registry {
	r := &Registry{byExt: make(map[string]ports.TextExtractor)}
	r.Register(".pdf", NewLedongthucPDF())
	return r
}

// Register adds an extractor for a file extension (e.g. ".pdf").
func (r *Registry) Register(ext string, extractor ports.TextExtractor) {
	ext = normalizeExt(ext)
	r.byExt[ext] = extractor
}

// Extract delegates to the extractor for path's extension.
func (r *Registry) Extract(ctx context.Context, path string) ([]domain.Paragraph, error) {
	ext := normalizeExt(filepath.Ext(path))
	ex, ok := r.byExt[ext]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported file type %q", domain.ErrInvalidInput, ext)
	}
	return ex.Extract(ctx, path)
}

func normalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}
