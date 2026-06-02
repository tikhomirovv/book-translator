package ports

import "context"

// ChatMessage is a single message in a chat completion request.
type ChatMessage struct {
	Role    string
	Content string
}

// ChatRequest is sent to an OpenAI-compatible API.
type ChatRequest struct {
	Model       string
	Messages    []ChatMessage
	Temperature float64
	MaxTokens   int
}

// ChatResponse is returned from the LLM.
type ChatResponse struct {
	Content string
	Usage   ChatUsage
}

// ChatUsage mirrors API usage fields when present.
type ChatUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// LLMProvider calls a chat completion API.
type LLMProvider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
