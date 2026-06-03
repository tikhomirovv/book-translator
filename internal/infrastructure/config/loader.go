package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Load reads configuration: .env → defaults → config.yaml → config.local.yaml → env.
func Load(configDir string) (*Config, error) {
	// Optional secrets file; missing .env is fine.
	_ = gotenv.Load()

	v := viper.New()
	v.SetConfigType("yaml")

	setDefaults(v)

	if configDir != "" {
		v.AddConfigPath(configDir)
	}
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Optional local overrides (not in git).
	v.SetConfigName("config.local")
	_ = v.MergeInConfig()

	v.SetEnvPrefix("BOOK_TRANSLATOR")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Common secret env names without prefix.
	v.BindEnv("openai_api_key", "OPENAI_API_KEY")
	v.BindEnv("openai_base_url", "OPENAI_BASE_URL")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.OpenAIAPIKey = normalizeAPIKey(firstNonEmpty(
		v.GetString("openai_api_key"),
		v.GetString("BOOK_TRANSLATOR_OPENAI_API_KEY"),
	))
	cfg.OpenAIBaseURL = firstNonEmpty(
		v.GetString("openai_base_url"),
		v.GetString("BOOK_TRANSLATOR_OPENAI_BASE_URL"),
		"https://api.openai.com/v1",
	)

	return &cfg, nil
}

// normalizeAPIKey treats template placeholders as empty (for local OpenAI-compatible servers).
func normalizeAPIKey(key string) string {
	switch strings.TrimSpace(key) {
	case "", "sk-your-key", "your-key", "changeme":
		return ""
	default:
		return strings.TrimSpace(key)
	}
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("chunk.size_paragraphs", 10)
	v.SetDefault("chunk.overlap_paragraphs", 2)
	v.SetDefault("context.strategy", "fixed_window")
	v.SetDefault("context.max_tokens", 2000)
	v.SetDefault("llm.model", "gpt-4o-mini")
	v.SetDefault("llm.temperature", 0.3)
	v.SetDefault("llm.max_tokens", 4096)
	v.SetDefault("request_delay_ms", 1000)
	v.SetDefault("allowed_languages", []string{"ru", "en"})
}

func firstNonEmpty(vals ...string) string {
	for _, s := range vals {
		if s != "" {
			return s
		}
	}
	return ""
}
