package ports

// PromptData is passed into prompt templates.
type PromptData struct {
	TargetLang     string
	ContextBlock   string
	ChunkText      string
	OverlapText    string
}

// PromptRenderer renders prompts by prompt-type from configuration.
type PromptRenderer interface {
	Render(promptType, templateKey string, data PromptData) (string, error)
}
