package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/domain/ports"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/llm/openai"
)

func TestClient_Chat_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("auth header missing")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "translated"}},
			},
			"usage": map[string]int{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		})
	}))
	defer srv.Close()

	client := openai.NewClient("test-key", srv.URL+"/v1", srv.Client())
	resp, err := client.Chat(context.Background(), ports.ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []ports.ChatMessage{
			{Role: "user", Content: "hello"},
		},
		Temperature: 0.2,
		MaxTokens:   100,
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Content != "translated" {
		t.Fatalf("content: %q", resp.Content)
	}
	if resp.Usage.TotalTokens != 30 {
		t.Fatalf("usage: %+v", resp.Usage)
	}
}

func TestClient_Chat_missingUsage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "ok"}},
			},
		})
	}))
	defer srv.Close()

	client := openai.NewClient("test-key", srv.URL+"/v1", srv.Client())
	resp, err := client.Chat(context.Background(), ports.ChatRequest{
		Model:    "gpt-4o-mini",
		Messages: []ports.ChatMessage{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Usage.TotalTokens != 0 {
		t.Fatalf("expected zero usage, got %+v", resp.Usage)
	}
}
