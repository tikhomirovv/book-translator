package domain_test

import (
	"testing"

	"github.com/tikhomirovv/book-translater/internal/domain"
)

func TestNewTranslation(t *testing.T) {
	t.Run("pending status", func(t *testing.T) {
		tr := domain.NewTranslation("id", "/in.pdf", "/out.md", "ru", "nonfiction")
		if tr.Status != domain.StatusPending {
			t.Errorf("status = %q", tr.Status)
		}
		if tr.TargetLang != "ru" {
			t.Errorf("lang = %q", tr.TargetLang)
		}
	})
}

func TestUsageAdd(t *testing.T) {
	cost := 0.01
	u := domain.Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30}
	u.Add(domain.Usage{PromptTokens: 5, CompletionTokens: 5, TotalTokens: 10, EstimatedCost: &cost})
	if u.TotalTokens != 40 {
		t.Errorf("total = %d", u.TotalTokens)
	}
}
