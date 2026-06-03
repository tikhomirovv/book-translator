package config

// Config holds application settings loaded from YAML and env.
type Config struct {
	Chunk          ChunkConfig            `mapstructure:"chunk"`
	Context        ContextConfig          `mapstructure:"context"`
	LLM            LLMConfig              `mapstructure:"llm"`
	RequestDelayMs int                    `mapstructure:"request_delay_ms"`
	AllowedLanguages []string             `mapstructure:"allowed_languages"`
	LogLevel         string               `mapstructure:"log_level"`
	Prompts        map[string]PromptSet   `mapstructure:"prompts"`
	// Secrets from env (not in yaml)
	OpenAIAPIKey  string `mapstructure:"-"`
	OpenAIBaseURL string `mapstructure:"-"`
}

// ChunkConfig controls text splitting.
type ChunkConfig struct {
	SizeParagraphs    int `mapstructure:"size_paragraphs"`
	OverlapParagraphs int `mapstructure:"overlap_paragraphs"`
}

// ContextConfig selects memory strategy.
type ContextConfig struct {
	Strategy  string `mapstructure:"strategy"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

// LLMConfig is passed to the provider (non-secret).
type LLMConfig struct {
	Model       string  `mapstructure:"model"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
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
