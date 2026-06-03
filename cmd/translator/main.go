// Command translator is the CLI entrypoint for book-translator.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/tikhomirovv/book-translator/internal/application/query"
	"github.com/tikhomirovv/book-translator/internal/application/resume"
	extracttext "github.com/tikhomirovv/book-translator/internal/application/extract"
	"github.com/tikhomirovv/book-translator/internal/application/translate"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
	contextmgr "github.com/tikhomirovv/book-translator/internal/infrastructure/context"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/chunk"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/config"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/extract"
	llminfra "github.com/tikhomirovv/book-translator/internal/infrastructure/llm"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/llm/openai"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/logging"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/prompt"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/store"
	"github.com/tikhomirovv/book-translator/internal/interfaces/cli"
)

func main() {
	cfg, err := config.Load("configs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		os.Exit(1)
	}

	logger := logging.NewLogger(cfg.LogLevel)

	renderer, err := prompt.NewYAMLRenderer(cfg.Prompts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: prompts: %v\n", err)
		os.Exit(1)
	}

	fs := store.NewFilesystemStore("")
	registry := extract.NewRegistry()

	baseLLM := openai.NewClient(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, nil)
	llm := &llminfra.RateLimitedLLM{
		Inner: &llminfra.RetryLLM{
			Inner:       baseLLM,
			MaxAttempts: 3,
			BaseDelay:   500 * time.Millisecond,
		},
		Delay: time.Duration(cfg.RequestDelayMs) * time.Millisecond,
	}

	llmCfg := translate.LLMConfig{
		Model:       cfg.LLM.Model,
		Temperature: cfg.LLM.Temperature,
		MaxTokens:   cfg.LLM.MaxTokens,
	}

	processChunk := &translate.ProcessChunk{
		LLM:     llm,
		Store:   fs,
		Prompts: renderer,
		LLMCfg:  llmCfg,
	}

	newContext := func(translationID string) ports.ContextManager {
		mgr, err := contextmgr.NewContextManager(cfg.Context.Strategy, fs, translationID, cfg.Context.MaxTokens)
		if err != nil {
			logger.Warn().Err(err).Msg("context manager fallback")
			return contextmgr.NewFixedWindow(fs, translationID, cfg.Context.MaxTokens)
		}
		return mgr
	}

	startUC := &translate.StartTranslation{
		Extractor:         registry,
		Store:             fs,
		ProcessChunk:      processChunk,
		Finalize:          &translate.FinalizeTranslation{Store: fs},
		NewContext:        newContext,
		BuildChunks:       chunk.BuildChunks,
		IsLanguageAllowed: cfg.IsLanguageAllowed,
		ChunkSize:         cfg.Chunk.SizeParagraphs,
		Overlap:           cfg.Chunk.OverlapParagraphs,
		ParagraphFrom:     cfg.Translation.ParagraphFrom,
		ParagraphTo:       cfg.Translation.ParagraphTo,
		DefaultPromptType: "nonfiction",
		Model:             cfg.LLM.Model,
		Provider:          "openai",
		LogDebug: func(msg string, kv ...any) {
			ev := logger.Debug()
			for i := 0; i+1 < len(kv); i += 2 {
				ev = ev.Interface(fmt.Sprint(kv[i]), kv[i+1])
			}
			ev.Msg(msg)
		},
	}

	resumeUC := &resume.ResumeTranslation{
		Extractor:    registry,
		Store:        fs,
		ProcessChunk: processChunk,
		Finalize:     &translate.FinalizeTranslation{Store: fs},
		NewContext:   newContext,
		BuildChunks:  chunk.BuildChunks,
		ChunkSize:     cfg.Chunk.SizeParagraphs,
		Overlap:       cfg.Chunk.OverlapParagraphs,
		ParagraphFrom: cfg.Translation.ParagraphFrom,
		ParagraphTo:   cfg.Translation.ParagraphTo,
		Model:         cfg.LLM.Model,
		Provider:     "openai",
	}

	cli.SetApp(&cli.App{
		Start:            startUC,
		Resume:           resumeUC,
		Extract:          &extracttext.ExtractSource{Extractor: registry},
		Status:           &query.GetStatus{Store: fs},
		List:             &query.ListTranslations{Store: fs},
		Logger:           logger,
		AllowedLanguages: cfg.AllowedLanguages,
	})

	if err := cli.Execute(); err != nil {
		logger.Error().Err(err).Msg("command failed")
		os.Exit(1)
	}
}
