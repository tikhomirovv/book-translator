package contextmgr_test

import (
	"errors"
	"testing"

	"github.com/tikhomirovv/book-translater/internal/domain"
	contextmgr "github.com/tikhomirovv/book-translater/internal/infrastructure/context"
	"github.com/tikhomirovv/book-translater/internal/infrastructure/store"
)

func TestNewContextManager_unknownStrategy(t *testing.T) {
	t.Parallel()

	_, err := contextmgr.NewContextManager("auto_summarize", store.NewFilesystemStore(t.TempDir()), "id", 100)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
