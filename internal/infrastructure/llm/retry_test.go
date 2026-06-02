package llm_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tikhomirovv/book-translator/internal/domain/ports"
	llminfra "github.com/tikhomirovv/book-translator/internal/infrastructure/llm"
)

type stubLLM struct {
	attempts int
	failures int
}

func (s *stubLLM) Chat(ctx context.Context, req ports.ChatRequest) (*ports.ChatResponse, error) {
	s.attempts++
	if s.attempts <= s.failures {
		return nil, errors.New("temporary")
	}
	return &ports.ChatResponse{Content: "ok"}, nil
}

func TestRetryLLM_eventualSuccess(t *testing.T) {
	t.Parallel()

	inner := &stubLLM{failures: 2}
	retry := &llminfra.RetryLLM{Inner: inner, MaxAttempts: 3, BaseDelay: time.Millisecond}
	resp, err := retry.Chat(context.Background(), ports.ChatRequest{})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Content != "ok" || inner.attempts != 3 {
		t.Fatalf("attempts=%d content=%q", inner.attempts, resp.Content)
	}
}
