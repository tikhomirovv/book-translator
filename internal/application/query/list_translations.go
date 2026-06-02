package query

import (
	"context"
	"fmt"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

// TranslationListItem is one row for the list command.
type TranslationListItem struct {
	ID         string
	SourcePath string
	TargetLang string
	Progress   string
	Status     domain.TranslationStatus
	UpdatedAt  string
}

// ListTranslations returns all translations for list display.
type ListTranslations struct {
	Store ports.TranslationStore
}

// Execute returns table rows ordered by updated time (newest first).
func (uc *ListTranslations) Execute(ctx context.Context) ([]TranslationListItem, error) {
	if uc == nil || uc.Store == nil {
		return nil, domain.ErrInvalidInput
	}

	summaries, err := uc.Store.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]TranslationListItem, 0, len(summaries))
	for _, s := range summaries {
		progress := fmt.Sprintf("%d/%d", s.LastCompletedChunk, s.TotalChunks)
		items = append(items, TranslationListItem{
			ID:         s.ID,
			SourcePath: s.SourcePath,
			TargetLang: s.TargetLang,
			Progress:   progress,
			Status:     s.Status,
			UpdatedAt:  s.UpdatedAt.UTC().Format("2006-01-02 15:04"),
		})
	}
	return items, nil
}
