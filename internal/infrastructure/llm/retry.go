package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/tikhomirovv/book-translater/internal/domain/ports"
)

// RetryLLM retries transient failures with exponential backoff.
type RetryLLM struct {
	Inner       ports.LLMProvider
	MaxAttempts int
	BaseDelay   time.Duration
}

// Chat calls the inner provider with retries.
func (r *RetryLLM) Chat(ctx context.Context, req ports.ChatRequest) (*ports.ChatResponse, error) {
	attempts := r.MaxAttempts
	if attempts <= 0 {
		attempts = 3
	}
	delay := r.BaseDelay
	if delay <= 0 {
		delay = 500 * time.Millisecond
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		resp, err := r.Inner.Chat(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if attempt == attempts || ctx.Err() != nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
		delay *= 2
	}
	return nil, fmt.Errorf("llm chat failed after %d attempts: %w", attempts, lastErr)
}
