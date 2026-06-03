package chunk

import "github.com/tikhomirovv/book-translator/internal/domain"

// ParagraphRange selects source paragraphs by their Index (inclusive).
// Use From/To < 0 to leave that bound open. When both bounds are < 0, all paragraphs are kept.
type ParagraphRange struct {
	From int
	To   int
}

// Active reports whether any range limit is configured.
func (r ParagraphRange) Active() bool {
	return r.From >= 0 || r.To >= 0
}

// FilterParagraphs returns paragraphs whose Index falls within the configured range.
func FilterParagraphs(paragraphs []domain.Paragraph, r ParagraphRange) []domain.Paragraph {
	if !r.Active() || len(paragraphs) == 0 {
		return paragraphs
	}

	start := r.From
	if start < 0 {
		start = 0
	}
	end := r.To
	if end < 0 {
		end = maxParagraphIndex(paragraphs)
	}
	if end < start {
		return nil
	}

	out := make([]domain.Paragraph, 0, len(paragraphs))
	for _, p := range paragraphs {
		if p.Index >= start && p.Index <= end {
			out = append(out, p)
		}
	}
	return out
}

func maxParagraphIndex(paragraphs []domain.Paragraph) int {
	max := paragraphs[0].Index
	for _, p := range paragraphs[1:] {
		if p.Index > max {
			max = p.Index
		}
	}
	return max
}
