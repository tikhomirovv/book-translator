package contextmgr

import (
	"context"
	"fmt"
	"strings"

	"github.com/tikhomirovv/book-translater/internal/domain"
	"github.com/tikhomirovv/book-translater/internal/domain/ports"
)

// FixedWindow keeps glossary and rolling summary within an approximate token budget.
type FixedWindow struct {
	store         ports.TranslationStore
	translationID string
	maxTokens     int
	state         *domain.TranslationState
}

// NewFixedWindow creates a context manager backed by TranslationStore state.
func NewFixedWindow(store ports.TranslationStore, translationID string, maxTokens int) *FixedWindow {
	if maxTokens <= 0 {
		maxTokens = 2000
	}
	return &FixedWindow{
		store:         store,
		translationID: translationID,
		maxTokens:     maxTokens,
		state: &domain.TranslationState{
			TranslationID: translationID,
			Glossary:      map[string]any{},
		},
	}
}

var _ ports.ContextManager = (*FixedWindow)(nil)

// AddExtracted merges structured extraction output into memory.
func (f *FixedWindow) AddExtracted(ctx context.Context, chunkIndex int, data map[string]any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if data == nil {
		return nil
	}
	if f.state.Glossary == nil {
		f.state.Glossary = map[string]any{}
	}

	if summary, ok := data["summary"].(string); ok && strings.TrimSpace(summary) != "" {
		if f.state.ContextSummary != "" {
			f.state.ContextSummary += "\n"
		}
		f.state.ContextSummary += strings.TrimSpace(summary)
	}

	if glossary, ok := data["glossary"].(map[string]any); ok {
		for k, v := range glossary {
			f.state.Glossary[fmt.Sprintf("%d:%s", chunkIndex, k)] = v
		}
	}

	f.enforceBudget()
	return nil
}

// BuildPromptContext returns a trimmed context block for prompts.
func (f *FixedWindow) BuildPromptContext(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	f.enforceBudget()

	var parts []string
	if summary := strings.TrimSpace(f.state.ContextSummary); summary != "" {
		parts = append(parts, "Summary:\n"+summary)
	}
	if len(f.state.Glossary) > 0 {
		var terms []string
		for k, v := range f.state.Glossary {
			terms = append(terms, fmt.Sprintf("- %s: %v", k, v))
		}
		parts = append(parts, "Glossary:\n"+strings.Join(terms, "\n"))
	}
	return strings.Join(parts, "\n\n"), nil
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

func (f *FixedWindow) enforceBudget() {
	for estimateTokens(f.state.ContextSummary, f.state.Glossary) > f.maxTokens {
		if f.trimGlossaryOne() {
			continue
		}
		f.state.ContextSummary = trimToTokenBudget(f.state.ContextSummary, f.maxTokens/2)
		if estimateTokens(f.state.ContextSummary, f.state.Glossary) <= f.maxTokens {
			return
		}
		// Last resort: clear summary.
		f.state.ContextSummary = ""
		return
	}
}

func (f *FixedWindow) trimGlossaryOne() bool {
	if len(f.state.Glossary) == 0 {
		return false
	}
	var firstKey string
	for k := range f.state.Glossary {
		firstKey = k
		break
	}
	delete(f.state.Glossary, firstKey)
	return true
}

// estimateTokens uses a simple chars/4 heuristic for MVP budgeting.
func estimateTokens(summary string, glossary map[string]any) int {
	total := len(summary) / 4
	for k, v := range glossary {
		total += (len(k) + len(fmt.Sprint(v))) / 4
	}
	return total
}

func trimToTokenBudget(text string, maxTokens int) string {
	maxChars := maxTokens * 4
	if len(text) <= maxChars {
		return text
	}
	return text[len(text)-maxChars:]
}
