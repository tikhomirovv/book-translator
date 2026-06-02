package contextmgr

import (
	"fmt"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

// NewContextManager selects an implementation by strategy name.
func NewContextManager(strategy string, store ports.TranslationStore, translationID string, maxTokens int) (ports.ContextManager, error) {
	switch strategy {
	case "", "fixed_window":
		return NewFixedWindow(store, translationID, maxTokens), nil
	default:
		return nil, fmt.Errorf("%w: unknown context strategy %q", domain.ErrInvalidInput, strategy)
	}
}
