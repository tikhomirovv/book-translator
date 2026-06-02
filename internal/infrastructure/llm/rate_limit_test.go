package llm_test

import (
	"context"
	"testing"
	"time"

	"github.com/tikhomirovv/book-translater/internal/domain/ports"
	llminfra "github.com/tikhomirovv/book-translater/internal/infrastructure/llm"
)

func TestRateLimitedLLM_delay(t *testing.T) {
	t.Parallel()

	inner := &stubLLM{}
	limited := &llminfra.RateLimitedLLM{Inner: inner, Delay: 50 * time.Millisecond}
	start := time.Now()
	if _, err := limited.Chat(context.Background(), ports.ChatRequest{}); err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if time.Since(start) < 40*time.Millisecond {
		t.Fatal("expected delay before call")
	}
}
