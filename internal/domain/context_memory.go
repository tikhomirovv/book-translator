package domain

// Usage tracks token consumption from LLM responses.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	EstimatedCost    *float64
}

// Add merges another usage record into this one.
func (u *Usage) Add(other Usage) {
	u.PromptTokens += other.PromptTokens
	u.CompletionTokens += other.CompletionTokens
	u.TotalTokens += other.TotalTokens
	if other.EstimatedCost != nil {
		if u.EstimatedCost == nil {
			v := *other.EstimatedCost
			u.EstimatedCost = &v
		} else {
			*u.EstimatedCost += *other.EstimatedCost
		}
	}
}

// TranslationState is persisted progress for resume/status.
type TranslationState struct {
	TranslationID      string
	LastCompletedChunk int
	TotalChunks        int
	Glossary           map[string]any
	ContextSummary     string
	LastError          string
	Usage              Usage
}
