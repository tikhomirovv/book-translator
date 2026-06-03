package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

const (
	defaultBaseURL        = "https://api.openai.com/v1"
	DefaultRequestTimeout = 120 * time.Second
)

// Client calls an OpenAI-compatible chat completions API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewClient creates an API client.
func NewClient(apiKey, baseURL string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultRequestTimeout}
	}
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
		logger:     slog.Default(),
	}
}

var _ ports.LLMProvider = (*Client)(nil)

type chatRequestPayload struct {
	Model       string              `json:"model"`
	Messages    []chatMessagePayload `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
}

type chatMessagePayload struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponsePayload struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Chat sends a chat completion request.
func (c *Client) Chat(ctx context.Context, req ports.ChatRequest) (*ports.ChatResponse, error) {
	if c.apiKey == "" && requiresCloudAPIKey(c.baseURL) {
		return nil, fmt.Errorf("openai api key is not configured")
	}

	payload := chatRequestPayload{
		Model:       req.Model,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}
	for _, m := range req.Messages {
		payload.Messages = append(payload.Messages, chatMessagePayload{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai api status %d: %s", resp.StatusCode, truncate(string(respBody), 512))
	}

	var parsed chatResponsePayload
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return nil, fmt.Errorf("openai api error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("openai api returned no choices")
	}

	out := &ports.ChatResponse{
		Content: parsed.Choices[0].Message.Content,
	}
	if parsed.Usage != nil {
		out.Usage = ports.ChatUsage{
			PromptTokens:     parsed.Usage.PromptTokens,
			CompletionTokens: parsed.Usage.CompletionTokens,
			TotalTokens:      parsed.Usage.TotalTokens,
		}
	} else {
		c.logger.Warn("llm response missing usage; token accounting will be zero",
			"model", req.Model,
		)
	}
	return out, nil
}

// requiresCloudAPIKey is true for the official OpenAI API endpoint.
func requiresCloudAPIKey(baseURL string) bool {
	return strings.Contains(strings.ToLower(baseURL), "api.openai.com")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
