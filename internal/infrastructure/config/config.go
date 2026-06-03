package config

// Config holds application settings loaded from YAML and env.
type Config struct {
	Chunk            ChunkConfig            `mapstructure:"chunk"`
	LLM              LLMConfig              `mapstructure:"llm"`
	Translation         TranslationConfig      `mapstructure:"translation"`
	RequestDelayMs      int                    `mapstructure:"request_delay_ms"`
	RequestTimeoutSeconds int                  `mapstructure:"request_timeout_seconds"`
	AllowedLanguages    []string               `mapstructure:"allowed_languages"`
	LogLevel         string                 `mapstructure:"log_level"`
	Prompts          map[string]PromptSet   `mapstructure:"prompts"`
	// Secrets from env (not in yaml)
	OpenAIAPIKey  string `mapstructure:"-"`
	OpenAIBaseURL string `mapstructure:"-"`
}

// ChunkConfig controls text splitting.
type ChunkConfig struct {
	SizeParagraphs    int `mapstructure:"size_paragraphs"`
	OverlapParagraphs int `mapstructure:"overlap_paragraphs"`
}

// LLMCallConfig holds model parameters for one chat completion role.
type LLMCallConfig struct {
	Model       string  `mapstructure:"model"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
}

// LLMConfig holds independent settings for translation and context extraction.
type LLMConfig struct {
	Translation LLMCallConfig `mapstructure:"translation"`
	Context     LLMCallConfig `mapstructure:"context"`
}

// TranslationConfig controls optional dev/test limits on the translation pipeline.
type TranslationConfig struct {
	// ParagraphFrom/To filter source paragraphs by Index (inclusive). -1 = no bound.
	ParagraphFrom int `mapstructure:"paragraph_from"`
	ParagraphTo   int `mapstructure:"paragraph_to"`
}

// ParagraphRange returns the configured paragraph filter for chunking.
func (c TranslationConfig) ParagraphRange() (from, to int) {
	return c.ParagraphFrom, c.ParagraphTo
}

// PromptSet holds templates for one prompt-type id.
type PromptSet struct {
	System            string `mapstructure:"system"`
	Translation       string `mapstructure:"translation"`
	ContextExtraction string `mapstructure:"context_extraction"`
}

// IsLanguageAllowed reports whether target language is in the whitelist.
func (c *Config) IsLanguageAllowed(lang string) bool {
	for _, l := range c.AllowedLanguages {
		if l == lang {
			return true
		}
	}
	return false
}
