package resume

import (
	"context"
	"fmt"

	"github.com/tikhomirovv/book-translator/internal/application/translate"
	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

// ResumeTranslation continues an interrupted translation idempotently.
type ResumeTranslation struct {
	Extractor    ports.TextExtractor
	Store        ports.TranslationStore
	ProcessChunk *translate.ProcessChunk
	Finalize     *translate.FinalizeTranslation
	NewContext   translate.ContextFactory
	BuildChunks  translate.ChunkSplitter
	ChunkSize    int
	Overlap      int
	Model        string
	Provider     string
	OnProgress   func(completed, total int)
}

// ResumeTranslationRequest identifies the job to resume.
type ResumeTranslationRequest struct {
	TranslationID string
}

// Execute skips completed chunks and processes the remainder.
func (uc *ResumeTranslation) Execute(ctx context.Context, req ResumeTranslationRequest) error {
	if uc == nil || uc.Extractor == nil || uc.Store == nil || uc.ProcessChunk == nil || uc.Finalize == nil || uc.NewContext == nil || uc.BuildChunks == nil {
		return domain.ErrInvalidInput
	}
	if req.TranslationID == "" {
		return domain.ErrInvalidInput
	}

	state, tr, err := uc.Store.Load(ctx, req.TranslationID)
	if err != nil {
		return err
	}
	if tr.Status == domain.StatusCompleted {
		return nil
	}

	paragraphs, err := uc.Extractor.Extract(ctx, tr.SourcePath)
	if err != nil {
		return fmt.Errorf("extract text: %w", err)
	}
	chunks := uc.BuildChunks(paragraphs, uc.ChunkSize, uc.Overlap)
	if len(chunks) == 0 {
		return fmt.Errorf("%w: no chunks built from source", domain.ErrInvalidInput)
	}

	state.TotalChunks = len(chunks)
	tr.TotalChunks = len(chunks)
	tr.Status = domain.StatusRunning
	if err := uc.Store.SaveState(ctx, tr.ID, state); err != nil {
		return fmt.Errorf("save state: %w", err)
	}
	if err := uc.Store.UpdateTranslation(ctx, tr); err != nil {
		return fmt.Errorf("update translation: %w", err)
	}

	mgr := uc.NewContext(tr.ID)
	if err := mgr.Load(ctx, tr.ID); err != nil {
		return fmt.Errorf("load context: %w", err)
	}
	uc.ProcessChunk.Context = mgr

	for _, ch := range chunks {
		if ch.Index <= state.LastCompletedChunk {
			continue
		}
		if err := uc.ProcessChunk.Execute(ctx, translate.ProcessChunkRequest{
			TranslationID: tr.ID,
			PromptType:    tr.PromptType,
			TargetLang:    tr.TargetLang,
			Chunk:         ch,
		}); err != nil {
			state.LastError = err.Error()
			tr.Status = domain.StatusFailed
			_ = uc.Store.SaveState(ctx, tr.ID, state)
			_ = uc.Store.UpdateTranslation(ctx, tr)
			return fmt.Errorf("process chunk %d: %w", ch.Index, err)
		}
		if uc.OnProgress != nil {
			uc.OnProgress(ch.Index, len(chunks))
		}
	}

	state, _, err = uc.Store.Load(ctx, tr.ID)
	if err != nil {
		return err
	}
	state.LastError = ""
	if err := uc.Store.SaveState(ctx, tr.ID, state); err != nil {
		return fmt.Errorf("clear last error: %w", err)
	}

	if err := uc.Finalize.Execute(ctx, translate.FinalizeTranslationRequest{
		TranslationID: tr.ID,
		Model:         uc.Model,
		Provider:      uc.Provider,
	}); err != nil {
		return fmt.Errorf("finalize translation: %w", err)
	}
	return nil
}
