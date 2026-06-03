package translate

import (
	"context"
	"fmt"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
	chunkinfra "github.com/tikhomirovv/book-translator/internal/infrastructure/chunk"
)

// LanguageValidator checks whether a target language is allowed.
type LanguageValidator func(lang string) bool

// ContextFactory builds a ContextManager for a translation id.
type ContextFactory func(translationID string) ports.ContextManager

// ChunkSplitter groups paragraphs into translation chunks.
type ChunkSplitter func(paragraphs []domain.Paragraph, size, overlap int) []domain.Chunk

// StartTranslation orchestrates extract → chunk → process → finalize.
type StartTranslation struct {
	Extractor         ports.TextExtractor
	Store             ports.TranslationStore
	ProcessChunk      *ProcessChunk
	Finalize          *FinalizeTranslation
	NewContext        ContextFactory
	BuildChunks       ChunkSplitter
	IsLanguageAllowed LanguageValidator
	ChunkSize         int
	Overlap           int
	ParagraphFrom     int
	ParagraphTo       int
	DefaultPromptType string
	Model             string
	Provider          string
	OnProgress        func(completed, total int)
	LogDebug          func(msg string, kv ...any)
}

// StartTranslationRequest starts a new translation job.
type StartTranslationRequest struct {
	SourcePath string
	OutputPath string
	TargetLang string
	PromptType string
}

// StartTranslationResult returns the created translation id.
type StartTranslationResult struct {
	TranslationID string
}

// Execute runs the full translation pipeline for a new book.
func (uc *StartTranslation) Execute(ctx context.Context, req StartTranslationRequest) (*StartTranslationResult, error) {
	if uc == nil || uc.Extractor == nil || uc.Store == nil || uc.ProcessChunk == nil || uc.Finalize == nil || uc.NewContext == nil || uc.BuildChunks == nil {
		return nil, domain.ErrInvalidInput
	}
	if req.SourcePath == "" || req.OutputPath == "" || req.TargetLang == "" {
		return nil, domain.ErrInvalidInput
	}
	if uc.IsLanguageAllowed != nil && !uc.IsLanguageAllowed(req.TargetLang) {
		return nil, fmt.Errorf("%w: target language %q is not allowed", domain.ErrInvalidLanguage, req.TargetLang)
	}

	promptType := req.PromptType
	if promptType == "" {
		promptType = uc.DefaultPromptType
	}
	if promptType == "" {
		promptType = "nonfiction"
	}

	tr := domain.NewTranslation("", req.SourcePath, req.OutputPath, req.TargetLang, promptType)
	tr.Status = domain.StatusRunning
	if err := uc.Store.Create(ctx, tr); err != nil {
		return nil, fmt.Errorf("create translation: %w", err)
	}

	paragraphs, err := uc.Extractor.Extract(ctx, req.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("extract text: %w", err)
	}
	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("%w: no text extracted from source", domain.ErrInvalidInput)
	}

	paragraphs = chunkinfra.FilterParagraphs(paragraphs, chunkinfra.ParagraphRange{
		From: uc.ParagraphFrom,
		To:   uc.ParagraphTo,
	})
	if len(paragraphs) == 0 {
		return nil, fmt.Errorf("%w: no paragraphs in configured range", domain.ErrInvalidInput)
	}

	if err := uc.Store.SaveExtractedSource(ctx, tr.ID, paragraphs); err != nil {
		return nil, fmt.Errorf("save extracted source: %w", err)
	}

	chunks := uc.BuildChunks(paragraphs, uc.ChunkSize, uc.Overlap)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("%w: no chunks built from source", domain.ErrInvalidInput)
	}
	if uc.LogDebug != nil {
		uc.LogDebug("extract and chunk plan",
			"paragraphs", len(paragraphs),
			"chunks", len(chunks),
			"chunk_size", uc.ChunkSize,
			"paragraph_from", uc.ParagraphFrom,
			"paragraph_to", uc.ParagraphTo,
		)
	}

	state, _, err := uc.Store.Load(ctx, tr.ID)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}
	state.TotalChunks = len(chunks)
	if err := uc.Store.SaveState(ctx, tr.ID, state); err != nil {
		return nil, fmt.Errorf("save total chunks: %w", err)
	}

	tr.TotalChunks = len(chunks)
	if err := uc.Store.UpdateTranslation(ctx, tr); err != nil {
		return nil, fmt.Errorf("update translation: %w", err)
	}

	mgr := uc.NewContext(tr.ID)
	if err := mgr.Load(ctx, tr.ID); err != nil {
		return nil, fmt.Errorf("load context: %w", err)
	}
	uc.ProcessChunk.Context = mgr

	for _, ch := range chunks {
		if err := uc.ProcessChunk.Execute(ctx, ProcessChunkRequest{
			TranslationID: tr.ID,
			PromptType:    promptType,
			TargetLang:    req.TargetLang,
			Chunk:         ch,
		}); err != nil {
			state, _, loadErr := uc.Store.Load(ctx, tr.ID)
			if loadErr == nil {
				state.LastError = err.Error()
				_ = uc.Store.SaveState(ctx, tr.ID, state)
			}
			tr.Status = domain.StatusFailed
			_ = uc.Store.UpdateTranslation(ctx, tr)
			return nil, fmt.Errorf("process chunk %d: %w", ch.Index, err)
		}
		if uc.OnProgress != nil {
			uc.OnProgress(ch.Index, len(chunks))
		}
	}

	if err := uc.Finalize.Execute(ctx, FinalizeTranslationRequest{
		TranslationID: tr.ID,
		Model:         uc.Model,
		Provider:      uc.Provider,
	}); err != nil {
		return nil, fmt.Errorf("finalize translation: %w", err)
	}

	return &StartTranslationResult{TranslationID: tr.ID}, nil
}
