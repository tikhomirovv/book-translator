package translate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

// FinalizeTranslation assembles translated chunks into output Markdown.
type FinalizeTranslation struct {
	Store ports.TranslationStore
}

// FinalizeTranslationRequest holds metadata for the output frontmatter.
type FinalizeTranslationRequest struct {
	TranslationID string
	Model         string
}

// Execute writes output Markdown with YAML frontmatter and marks translation completed.
func (uc *FinalizeTranslation) Execute(ctx context.Context, req FinalizeTranslationRequest) error {
	if uc == nil || uc.Store == nil {
		return domain.ErrInvalidInput
	}
	if req.TranslationID == "" {
		return domain.ErrInvalidInput
	}

	state, tr, err := uc.Store.Load(ctx, req.TranslationID)
	if err != nil {
		return err
	}
	if state.TotalChunks > 0 && state.LastCompletedChunk < state.TotalChunks {
		return fmt.Errorf("%w: translation incomplete (%d/%d chunks)", domain.ErrInvalidInput, state.LastCompletedChunk, state.TotalChunks)
	}

	chunks, err := uc.Store.LoadTranslatedChunks(ctx, req.TranslationID)
	if err != nil {
		return fmt.Errorf("load chunks: %w", err)
	}
	if len(chunks) == 0 {
		return fmt.Errorf("%w: no translated chunks found", domain.ErrInvalidInput)
	}

	markdown := buildOutputMarkdown(tr, req, state.Usage, chunks)
	if err := uc.Store.WriteOutput(ctx, tr, markdown); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	tr.Status = domain.StatusCompleted
	tr.LastCompletedChunk = state.LastCompletedChunk
	tr.TotalChunks = state.TotalChunks
	return uc.Store.UpdateTranslation(ctx, tr)
}

func buildOutputMarkdown(tr *domain.Translation, req FinalizeTranslationRequest, usage domain.Usage, chunks []domain.Chunk) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("translation_id: %s\n", tr.ID))
	b.WriteString(fmt.Sprintf("target_lang: %s\n", tr.TargetLang))
	b.WriteString(fmt.Sprintf("date: %s\n", time.Now().UTC().Format(time.RFC3339)))
	if req.Model != "" {
		b.WriteString(fmt.Sprintf("model: %s\n", req.Model))
	}
	b.WriteString(fmt.Sprintf("input_tokens: %d\n", usage.PromptTokens))
	b.WriteString(fmt.Sprintf("output_tokens: %d\n", usage.CompletionTokens))
	b.WriteString(fmt.Sprintf("total_tokens: %d\n", usage.TotalTokens))
	if usage.EstimatedCost != nil {
		b.WriteString(fmt.Sprintf("estimated_cost: %.6f\n", *usage.EstimatedCost))
	}
	b.WriteString("---\n")

	for i, chunk := range chunks {
		if i > 0 && chunk.TranslatedText != "" {
			b.WriteString("\n\n")
		}
		b.WriteString(strings.TrimSpace(chunk.TranslatedText))
	}
	if len(chunks) > 0 && chunks[len(chunks)-1].TranslatedText != "" {
		b.WriteByte('\n')
	}
	return b.String()
}
