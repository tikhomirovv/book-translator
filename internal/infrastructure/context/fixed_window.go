package contextmgr

import (
	"context"
	"strings"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

// FixedWindow stores a single consolidated context block produced by the context LLM.
type FixedWindow struct {
	store         ports.TranslationStore
	translationID string
	state         *domain.TranslationState
}

// NewFixedWindow creates a context manager backed by TranslationStore state.
func NewFixedWindow(store ports.TranslationStore, translationID string) *FixedWindow {
	return &FixedWindow{
		store:         store,
		translationID: translationID,
		state: &domain.TranslationState{
			TranslationID: translationID,
			Glossary:      map[string]any{},
		},
	}
}

var _ ports.ContextManager = (*FixedWindow)(nil)

// AddExtracted replaces stored context with the LLM's consolidated output.
func (f *FixedWindow) AddExtracted(ctx context.Context, _ int, data map[string]any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if data == nil {
		return nil
	}

	summary, ok := data["summary"].(string)
	if !ok || strings.TrimSpace(summary) == "" {
		return nil
	}

	f.state.ContextSummary = strings.TrimSpace(summary)
	f.state.Glossary = map[string]any{}
	return nil
}

// BuildPromptContext returns the consolidated context for translation prompts.
func (f *FixedWindow) BuildPromptContext(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return strings.TrimSpace(f.state.ContextSummary), nil
}

// Save persists in-memory state.
func (f *FixedWindow) Save(ctx context.Context, translationID string) error {
	if translationID != "" {
		f.translationID = translationID
		f.state.TranslationID = translationID
	}
	return f.store.SaveState(ctx, f.translationID, f.state)
}

// Load restores state from the store.
func (f *FixedWindow) Load(ctx context.Context, translationID string) error {
	state, _, err := f.store.Load(ctx, translationID)
	if err != nil {
		return err
	}
	f.translationID = translationID
	f.state = state
	if f.state.Glossary == nil {
		f.state.Glossary = map[string]any{}
	}
	return nil
}
