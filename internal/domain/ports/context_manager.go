package ports

import "context"

// ContextManager maintains translation memory (glossary, summary, overlap).
type ContextManager interface {
	AddExtracted(ctx context.Context, chunkIndex int, data map[string]any) error
	BuildPromptContext(ctx context.Context) (string, error)
	Save(ctx context.Context, translationID string) error
	Load(ctx context.Context, translationID string) error
}
