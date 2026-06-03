package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

const (
	templateSystem            = "system"
	templateTranslation       = "translation"
	templateContextExtraction = "context_extraction"
)

// LLMConfig holds non-secret model parameters for chat calls.
type LLMConfig struct {
	Model        string
	ContextModel string
	Temperature  float64
	MaxTokens    int
}

func (c LLMConfig) contextModel() string {
	if c.ContextModel != "" {
		return c.ContextModel
	}
	return c.Model
}

// ProcessChunk translates one chunk and extracts context in parallel.
type ProcessChunk struct {
	LLM     ports.LLMProvider
	Store   ports.TranslationStore
	Prompts ports.PromptRenderer
	Context ports.ContextManager
	LLMCfg  LLMConfig
}

// ProcessChunkRequest describes one chunk iteration.
// Caller must pass chunks sequentially: Chunk.Index == LastCompletedChunk + 1.
type ProcessChunkRequest struct {
	TranslationID string
	PromptType    string
	TargetLang    string
	Chunk         domain.Chunk
}

// Execute runs translation and context extraction, then persists progress.
func (uc *ProcessChunk) Execute(ctx context.Context, req ProcessChunkRequest) error {
	if uc == nil || uc.LLM == nil || uc.Store == nil || uc.Prompts == nil || uc.Context == nil {
		return domain.ErrInvalidInput
	}
	if req.TranslationID == "" || req.PromptType == "" || req.Chunk.Index <= 0 {
		return domain.ErrInvalidInput
	}

	state, tr, err := uc.Store.Load(ctx, req.TranslationID)
	if err != nil {
		return err
	}

	targetLang := req.TargetLang
	if targetLang == "" {
		targetLang = tr.TargetLang
	}

	expected := state.LastCompletedChunk + 1
	if req.Chunk.Index != expected {
		return fmt.Errorf("%w: expected chunk index %d, got %d", domain.ErrInvalidInput, expected, req.Chunk.Index)
	}

	contextBlock, err := uc.Context.BuildPromptContext(ctx)
	if err != nil {
		return fmt.Errorf("build prompt context: %w", err)
	}

	promptData := ports.PromptData{
		TargetLang:   targetLang,
		ContextBlock: contextBlock,
		ChunkText:    req.Chunk.SourceText,
		OverlapText:  req.Chunk.OverlapFromPrev,
	}

	systemPrompt, err := uc.Prompts.Render(req.PromptType, templateSystem, promptData)
	if err != nil {
		return fmt.Errorf("render system prompt: %w", err)
	}
	translationPrompt, err := uc.Prompts.Render(req.PromptType, templateTranslation, promptData)
	if err != nil {
		return fmt.Errorf("render translation prompt: %w", err)
	}
	contextPrompt, err := uc.Prompts.Render(req.PromptType, templateContextExtraction, promptData)
	if err != nil {
		return fmt.Errorf("render context extraction prompt: %w", err)
	}

	var (
		translated     string
		extracted      string
		usageTranslate domain.Usage
		usageExtract   domain.Usage
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		resp, err := uc.LLM.Chat(gctx, ports.ChatRequest{
			Model: uc.LLMCfg.Model,
			Messages: []ports.ChatMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: translationPrompt},
			},
			Temperature: uc.LLMCfg.Temperature,
			MaxTokens:   uc.LLMCfg.MaxTokens,
		})
		if err != nil {
			return fmt.Errorf("translate chunk %d: %w", req.Chunk.Index, err)
		}
		translated = strings.TrimSpace(resp.Content)
		usageTranslate = usageFromChat(resp.Usage)
		return nil
	})

	g.Go(func() error {
		resp, err := uc.LLM.Chat(gctx, ports.ChatRequest{
			Model: uc.LLMCfg.contextModel(),
			Messages: []ports.ChatMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: contextPrompt},
			},
			Temperature: uc.LLMCfg.Temperature,
			MaxTokens:   uc.LLMCfg.MaxTokens,
		})
		if err != nil {
			return fmt.Errorf("extract context chunk %d: %w", req.Chunk.Index, err)
		}
		extracted = strings.TrimSpace(resp.Content)
		usageExtract = usageFromChat(resp.Usage)
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if err := uc.Context.AddExtracted(ctx, req.Chunk.Index, parseContextExtraction(extracted)); err != nil {
		return fmt.Errorf("add extracted context: %w", err)
	}

	chunk := req.Chunk
	chunk.TranslatedText = translated
	if err := uc.Store.SaveChunk(ctx, req.TranslationID, chunk); err != nil {
		return fmt.Errorf("save chunk: %w", err)
	}

	if err := uc.Context.Save(ctx, req.TranslationID); err != nil {
		return fmt.Errorf("save context: %w", err)
	}

	state, _, err = uc.Store.Load(ctx, req.TranslationID)
	if err != nil {
		return fmt.Errorf("reload state: %w", err)
	}
	state.LastCompletedChunk = req.Chunk.Index
	state.Usage.Add(usageTranslate)
	state.Usage.Add(usageExtract)
	if err := uc.Store.SaveState(ctx, req.TranslationID, state); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	tr.LastCompletedChunk = req.Chunk.Index
	if tr.Status == domain.StatusPending {
		tr.Status = domain.StatusRunning
	}
	if err := uc.Store.UpdateTranslation(ctx, tr); err != nil {
		return fmt.Errorf("update translation: %w", err)
	}

	return nil
}

func usageFromChat(u ports.ChatUsage) domain.Usage {
	return domain.Usage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	}
}

// parseContextExtraction turns LLM output into ContextManager data.
// JSON with summary/glossary is preferred; plain text becomes summary.
func parseContextExtraction(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var parsed struct {
		Summary  string         `json:"summary"`
		Glossary map[string]any `json:"glossary"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		out := map[string]any{}
		if strings.TrimSpace(parsed.Summary) != "" {
			out["summary"] = strings.TrimSpace(parsed.Summary)
		}
		if len(parsed.Glossary) > 0 {
			out["glossary"] = parsed.Glossary
		}
		if len(out) > 0 {
			return out
		}
	}

	return map[string]any{"summary": raw}
}
