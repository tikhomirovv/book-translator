package domain

import "time"

// TranslationStatus represents lifecycle of a translation job.
type TranslationStatus string

const (
	StatusPending   TranslationStatus = "pending"
	StatusRunning   TranslationStatus = "running"
	StatusPaused    TranslationStatus = "paused"
	StatusCompleted TranslationStatus = "completed"
	StatusFailed    TranslationStatus = "failed"
)

// Translation is the main aggregate for a single translation run.
type Translation struct {
	ID                 string
	SourcePath         string
	OutputPath         string
	TargetLang         string
	PromptType         string
	Status             TranslationStatus
	LastCompletedChunk int
	TotalChunks        int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewTranslation creates a translation in pending state.
func NewTranslation(id, sourcePath, outputPath, targetLang, promptType string) *Translation {
	now := time.Now().UTC()
	return &Translation{
		ID:         id,
		SourcePath: sourcePath,
		OutputPath: outputPath,
		TargetLang: targetLang,
		PromptType: promptType,
		Status:     StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
