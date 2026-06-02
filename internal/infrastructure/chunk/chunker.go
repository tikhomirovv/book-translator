package chunk

import (
	"strings"

	"github.com/tikhomirovv/book-translator/internal/domain"
)

// BuildChunks groups paragraphs into chunks with sliding overlap.
// Chunk indices are 1-based to match persisted chunk filenames.
func BuildChunks(paragraphs []domain.Paragraph, size, overlap int) []domain.Chunk {
	if size <= 0 {
		size = 10
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= size {
		overlap = size - 1
	}
	if len(paragraphs) == 0 {
		return nil
	}

	step := size - overlap
	if step <= 0 {
		step = 1
	}

	var chunks []domain.Chunk
	chunkIndex := 1
	for start := 0; start < len(paragraphs); start += step {
		end := start + size
		if end > len(paragraphs) {
			end = len(paragraphs)
		}

		var texts []string
		for i := start; i < end; i++ {
			texts = append(texts, paragraphs[i].Text)
		}

		chunk := domain.Chunk{
			Index:          chunkIndex,
			ParagraphStart: paragraphs[start].Index,
			ParagraphEnd:   paragraphs[end-1].Index + 1,
			SourceText:     strings.Join(texts, "\n\n"),
		}

		if start > 0 && overlap > 0 {
			overlapEnd := start + overlap
			if overlapEnd > end {
				overlapEnd = end
			}
			var overlapTexts []string
			for i := start; i < overlapEnd; i++ {
				overlapTexts = append(overlapTexts, paragraphs[i].Text)
			}
			chunk.OverlapFromPrev = strings.Join(overlapTexts, "\n\n")
		}

		chunks = append(chunks, chunk)
		chunkIndex++
		if end == len(paragraphs) {
			break
		}
	}
	return chunks
}
