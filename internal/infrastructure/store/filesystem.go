package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
)

const defaultBaseDir = "translations"

// FilesystemStore persists translations under baseDir/<uuid>/.
type FilesystemStore struct {
	baseDir string
}

var _ ports.TranslationStore = (*FilesystemStore)(nil)

// NewFilesystemStore creates a store rooted at baseDir (defaults to "translations").
func NewFilesystemStore(baseDir string) *FilesystemStore {
	if baseDir == "" {
		baseDir = defaultBaseDir
	}
	return &FilesystemStore{baseDir: baseDir}
}

// NewTranslationID returns a new UUID v4 string.
func NewTranslationID() string {
	return uuid.NewString()
}

// Create initializes a new translation directory on disk.
func (s *FilesystemStore) Create(ctx context.Context, t *domain.Translation) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if t == nil {
		return domain.ErrInvalidInput
	}
	if t.ID == "" {
		t.ID = NewTranslationID()
	}
	if err := uuid.Validate(t.ID); err != nil {
		return fmt.Errorf("%w: invalid translation id", domain.ErrInvalidInput)
	}

	dir := s.translationDir(t.ID)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("%w: translation %s already exists", domain.ErrInvalidInput, t.ID)
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Join(dir, "chunks"), 0o755); err != nil {
		return err
	}

	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = now
	}
	if t.Status == "" {
		t.Status = domain.StatusPending
	}

	if err := writeJSON(filepath.Join(dir, "source.meta.json"), sourceMetaFromTranslation(t)); err != nil {
		return err
	}

	state := &domain.TranslationState{
		TranslationID:      t.ID,
		LastCompletedChunk: t.LastCompletedChunk,
		TotalChunks:        t.TotalChunks,
		Glossary:           map[string]any{},
	}
	return writeJSON(filepath.Join(dir, "state.json"), stateFileFromDomain(state))
}

// Load reads translation metadata and persisted state.
func (s *FilesystemStore) Load(ctx context.Context, id string) (*domain.TranslationState, *domain.Translation, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if id == "" {
		return nil, nil, domain.ErrInvalidInput
	}

	dir := s.translationDir(id)
	metaPath := filepath.Join(dir, "source.meta.json")
	statePath := filepath.Join(dir, "state.json")

	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return nil, nil, domain.ErrNotFound
	} else if err != nil {
		return nil, nil, err
	}

	var meta sourceMeta
	if err := readJSON(metaPath, &meta); err != nil {
		return nil, nil, err
	}
	t := meta.toTranslation()

	var sf stateFile
	if err := readJSON(statePath, &sf); err != nil {
		if !os.IsNotExist(err) {
			return nil, nil, err
		}
		// Allow missing state.json for partially created dirs.
		sf = stateFile{TranslationID: id, Glossary: map[string]any{}}
	}

	state := sf.toDomain()
	// Keep translation counters aligned with state when state is authoritative.
	t.LastCompletedChunk = state.LastCompletedChunk
	t.TotalChunks = state.TotalChunks
	return state, t, nil
}

// SaveState writes state.json.
func (s *FilesystemStore) SaveState(ctx context.Context, id string, state *domain.TranslationState) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if state == nil {
		return domain.ErrInvalidInput
	}
	if id == "" {
		id = state.TranslationID
	}
	if id == "" {
		return domain.ErrInvalidInput
	}
	dir := s.translationDir(id)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return domain.ErrNotFound
	} else if err != nil {
		return err
	}
	state.TranslationID = id
	if state.Glossary == nil {
		state.Glossary = map[string]any{}
	}
	return writeJSON(filepath.Join(dir, "state.json"), stateFileFromDomain(state))
}

// SaveChunk writes chunks/NNNN.md with YAML frontmatter and translated body.
func (s *FilesystemStore) SaveChunk(ctx context.Context, id string, chunk domain.Chunk) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id == "" {
		return domain.ErrInvalidInput
	}
	dir := s.translationDir(id)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return domain.ErrNotFound
	} else if err != nil {
		return err
	}

	chunksDir := filepath.Join(dir, "chunks")
	if err := os.MkdirAll(chunksDir, 0o755); err != nil {
		return err
	}

	content, err := encodeChunkMarkdown(chunk)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%04d.md", chunk.Index)
	return os.WriteFile(filepath.Join(chunksDir, name), []byte(content), 0o644)
}

// LoadTranslatedChunks reads saved chunk files in index order.
func (s *FilesystemStore) LoadTranslatedChunks(ctx context.Context, id string) ([]domain.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, domain.ErrInvalidInput
	}

	chunksDir := filepath.Join(s.translationDir(id), "chunks")
	entries, err := os.ReadDir(chunksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var chunks []domain.Chunk
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(chunksDir, ent.Name()))
		if err != nil {
			return nil, err
		}
		chunk, err := decodeChunkMarkdown(string(data))
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", ent.Name(), err)
		}
		chunks = append(chunks, chunk)
	}

	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Index < chunks[j].Index
	})
	return chunks, nil
}

// WriteOutput writes final Markdown to the user path and translations/<id>/output.md.
func (s *FilesystemStore) WriteOutput(ctx context.Context, t *domain.Translation, markdown string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if t == nil || t.ID == "" {
		return domain.ErrInvalidInput
	}

	dir := s.translationDir(t.ID)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return domain.ErrNotFound
	} else if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "output.md"), []byte(markdown), 0o644); err != nil {
		return err
	}
	if t.OutputPath == "" {
		return nil
	}

	outDir := filepath.Dir(t.OutputPath)
	if outDir != "." && outDir != "" {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(t.OutputPath, []byte(markdown), 0o644)
}

// UpdateTranslation updates source.meta.json.
func (s *FilesystemStore) UpdateTranslation(ctx context.Context, t *domain.Translation) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if t == nil || t.ID == "" {
		return domain.ErrInvalidInput
	}
	dir := s.translationDir(t.ID)
	metaPath := filepath.Join(dir, "source.meta.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return domain.ErrNotFound
	} else if err != nil {
		return err
	}

	var existing sourceMeta
	if err := readJSON(metaPath, &existing); err != nil {
		return err
	}
	updated := sourceMetaFromTranslation(t)
	if updated.CreatedAt.IsZero() {
		updated.CreatedAt = existing.CreatedAt
	}
	updated.UpdatedAt = time.Now().UTC()
	return writeJSON(metaPath, updated)
}

// List returns summaries for all translation directories.
func (s *FilesystemStore) List(ctx context.Context) ([]ports.TranslationSummary, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var out []ports.TranslationSummary
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		id := ent.Name()
		if err := uuid.Validate(id); err != nil {
			continue
		}

		state, tr, err := s.Load(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			return nil, fmt.Errorf("load %s: %w", id, err)
		}

		out = append(out, ports.TranslationSummary{
			ID:                 tr.ID,
			SourcePath:         tr.SourcePath,
			TargetLang:         tr.TargetLang,
			Status:             tr.Status,
			LastCompletedChunk: state.LastCompletedChunk,
			TotalChunks:        state.TotalChunks,
			UpdatedAt:          tr.UpdatedAt,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

func (s *FilesystemStore) translationDir(id string) string {
	return filepath.Join(s.baseDir, id)
}

// sourceMeta is persisted in source.meta.json.
type sourceMeta struct {
	ID                 string                    `json:"id"`
	SourcePath         string                    `json:"source_path"`
	OutputPath         string                    `json:"output_path"`
	TargetLang         string                    `json:"target_lang"`
	PromptType         string                    `json:"prompt_type"`
	Status             domain.TranslationStatus  `json:"status"`
	LastCompletedChunk int                       `json:"last_completed_chunk"`
	TotalChunks        int                       `json:"total_chunks"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
}

func sourceMetaFromTranslation(t *domain.Translation) sourceMeta {
	return sourceMeta{
		ID:                 t.ID,
		SourcePath:         t.SourcePath,
		OutputPath:         t.OutputPath,
		TargetLang:         t.TargetLang,
		PromptType:         t.PromptType,
		Status:             t.Status,
		LastCompletedChunk: t.LastCompletedChunk,
		TotalChunks:        t.TotalChunks,
		CreatedAt:          t.CreatedAt.UTC(),
		UpdatedAt:          t.UpdatedAt.UTC(),
	}
}

func (m sourceMeta) toTranslation() *domain.Translation {
	return &domain.Translation{
		ID:                 m.ID,
		SourcePath:         m.SourcePath,
		OutputPath:         m.OutputPath,
		TargetLang:         m.TargetLang,
		PromptType:         m.PromptType,
		Status:             m.Status,
		LastCompletedChunk: m.LastCompletedChunk,
		TotalChunks:        m.TotalChunks,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

// stateFile mirrors domain.TranslationState on disk.
type stateFile struct {
	TranslationID      string         `json:"translation_id"`
	LastCompletedChunk int            `json:"last_completed_chunk"`
	TotalChunks        int            `json:"total_chunks"`
	Glossary           map[string]any `json:"glossary"`
	ContextSummary     string         `json:"context_summary"`
	Usage              usageFile      `json:"usage"`
}

type usageFile struct {
	PromptTokens     int      `json:"prompt_tokens"`
	CompletionTokens int      `json:"completion_tokens"`
	TotalTokens      int      `json:"total_tokens"`
	EstimatedCost    *float64 `json:"estimated_cost,omitempty"`
}

func stateFileFromDomain(state *domain.TranslationState) stateFile {
	if state.Glossary == nil {
		state.Glossary = map[string]any{}
	}
	return stateFile{
		TranslationID:      state.TranslationID,
		LastCompletedChunk: state.LastCompletedChunk,
		TotalChunks:        state.TotalChunks,
		Glossary:           state.Glossary,
		ContextSummary:     state.ContextSummary,
		Usage: usageFile{
			PromptTokens:     state.Usage.PromptTokens,
			CompletionTokens: state.Usage.CompletionTokens,
			TotalTokens:      state.Usage.TotalTokens,
			EstimatedCost:    state.Usage.EstimatedCost,
		},
	}
}

func (sf stateFile) toDomain() *domain.TranslationState {
	glossary := sf.Glossary
	if glossary == nil {
		glossary = map[string]any{}
	}
	return &domain.TranslationState{
		TranslationID:      sf.TranslationID,
		LastCompletedChunk: sf.LastCompletedChunk,
		TotalChunks:        sf.TotalChunks,
		Glossary:           glossary,
		ContextSummary:     sf.ContextSummary,
		Usage: domain.Usage{
			PromptTokens:     sf.Usage.PromptTokens,
			CompletionTokens: sf.Usage.CompletionTokens,
			TotalTokens:      sf.Usage.TotalTokens,
			EstimatedCost:    sf.Usage.EstimatedCost,
		},
	}
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func encodeChunkMarkdown(chunk domain.Chunk) (string, error) {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("index: %d\n", chunk.Index))
	b.WriteString(fmt.Sprintf("paragraph_start: %d\n", chunk.ParagraphStart))
	b.WriteString(fmt.Sprintf("paragraph_end: %d\n", chunk.ParagraphEnd))
	if chunk.OverlapFromPrev != "" {
		// Escape quotes minimally for a single-line YAML string.
		escaped := strings.ReplaceAll(chunk.OverlapFromPrev, "\"", "\\\"")
		b.WriteString(fmt.Sprintf("overlap_from_prev: \"%s\"\n", escaped))
	}
	b.WriteString("---\n")
	b.WriteString(chunk.TranslatedText)
	if !strings.HasSuffix(chunk.TranslatedText, "\n") && chunk.TranslatedText != "" {
		b.WriteByte('\n')
	}
	return b.String(), nil
}

func decodeChunkMarkdown(content string) (domain.Chunk, error) {
	const delim = "---"
	if !strings.HasPrefix(content, delim) {
		return domain.Chunk{}, fmt.Errorf("%w: missing frontmatter", domain.ErrInvalidInput)
	}

	rest := content[len(delim):]
	end := strings.Index(rest, "\n"+delim+"\n")
	if end < 0 {
		return domain.Chunk{}, fmt.Errorf("%w: missing frontmatter end", domain.ErrInvalidInput)
	}

	body := rest[end+len("\n"+delim+"\n"):]
	meta := rest[:end]
	var chunk domain.Chunk
	for _, line := range strings.Split(meta, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "index":
			fmt.Sscanf(val, "%d", &chunk.Index)
		case "paragraph_start":
			fmt.Sscanf(val, "%d", &chunk.ParagraphStart)
		case "paragraph_end":
			fmt.Sscanf(val, "%d", &chunk.ParagraphEnd)
		case "overlap_from_prev":
			chunk.OverlapFromPrev = strings.Trim(val, `"`)
		}
	}
	chunk.TranslatedText = body
	return chunk, nil
}
