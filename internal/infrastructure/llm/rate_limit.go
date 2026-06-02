package llm

import (
	"context"
	"time"

	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

// RateLimitedLLM waits before each request to respect a simple delay limit.
type RateLimitedLLM struct {
	Inner ports.LLMProvider
	Delay time.Duration
}

// Chat applies delay then delegates to the inner provider.
func (r *RateLimitedLLM) Chat(ctx context.Context, req ports.ChatRequest) (*ports.ChatResponse, error) {
	if r.Delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(r.Delay):
		}
	}
	return r.Inner.Chat(ctx, req)
}
