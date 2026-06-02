package query

import (
	"context"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

// TranslationStatusView is detailed progress for one translation.
type TranslationStatusView struct {
	ID              string
	SourcePath      string
	TargetLang      string
	Status          domain.TranslationStatus
	CompletedChunks int
	TotalChunks     int
	Usage           domain.Usage
	LastError       string
}

// GetStatus returns progress and metadata for a translation id.
type GetStatus struct {
	Store ports.TranslationStore
}

// Execute loads status fields for CLI/API display.
func (uc *GetStatus) Execute(ctx context.Context, translationID string) (*TranslationStatusView, error) {
	if uc == nil || uc.Store == nil {
		return nil, domain.ErrInvalidInput
	}
	if translationID == "" {
		return nil, domain.ErrInvalidInput
	}

	state, tr, err := uc.Store.Load(ctx, translationID)
	if err != nil {
		return nil, err
	}

	return &TranslationStatusView{
		ID:              tr.ID,
		SourcePath:      tr.SourcePath,
		TargetLang:      tr.TargetLang,
		Status:          tr.Status,
		CompletedChunks: state.LastCompletedChunk,
		TotalChunks:     state.TotalChunks,
		Usage:           state.Usage,
		LastError:       state.LastError,
	}, nil
}
