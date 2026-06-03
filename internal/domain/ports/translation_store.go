package ports

import (
	"context"
	"time"

	"github.com/tikhomirovv/book-translator/internal/domain"
)

// TranslationSummary is a list view row for CLI/API.
type TranslationSummary struct {
	ID                 string
	SourcePath         string
	TargetLang         string
	Status             domain.TranslationStatus
	LastCompletedChunk int
	TotalChunks        int
	UpdatedAt          time.Time
}

// TranslationStore persists translation progress on disk (MVP).
type TranslationStore interface {
	Create(ctx context.Context, t *domain.Translation) error
	Load(ctx context.Context, id string) (*domain.TranslationState, *domain.Translation, error)
	SaveState(ctx context.Context, id string, state *domain.TranslationState) error
	SaveChunk(ctx context.Context, id string, chunk domain.Chunk) error
	SaveExtractedSource(ctx context.Context, id string, paragraphs []domain.Paragraph) error
	LoadTranslatedChunks(ctx context.Context, id string) ([]domain.Chunk, error)
	WriteOutput(ctx context.Context, t *domain.Translation, markdown string) error
	UpdateTranslation(ctx context.Context, t *domain.Translation) error
	List(ctx context.Context) ([]TranslationSummary, error)
}
