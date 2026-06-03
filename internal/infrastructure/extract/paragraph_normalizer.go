package extract

import (
	"strings"

	"github.com/tikhomirovv/book-translator/internal/domain"
)

const maxParagraphRunes = 6000

// NormalizeParagraphs splits raw text into indexed paragraphs.
// PDF plain text is reflowed first; Markdown-style blank lines are preferred.
func NormalizeParagraphs(text string) []domain.Paragraph {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if strings.TrimSpace(text) == "" {
		return nil
	}

	text = ReflowPlainText(text)

	parts := splitDoubleNewline(text)
	if len(parts) > 1 {
		return indexParagraphs(parts)
	}

	blocks := splitEmptyLineBlocks(text)
	if len(blocks) == 0 {
		return nil
	}

	var expanded []string
	for _, block := range blocks {
		if len([]rune(block)) <= maxParagraphRunes {
			expanded = append(expanded, block)
			continue
		}
		expanded = append(expanded, splitLines(block)...)
	}

	if len(expanded) <= 1 {
		return indexParagraphs(splitLines(text))
	}

	return indexParagraphs(expanded)
}

func splitDoubleNewline(text string) []string {
	raw := strings.Split(text, "\n\n")
	var parts []string
	for _, part := range raw {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func splitEmptyLineBlocks(text string) []string {
	lines := strings.Split(text, "\n")
	var blocks []string
	var cur []string
	flush := func() {
		if len(cur) == 0 {
			return
		}
		blocks = append(blocks, strings.Join(cur, "\n"))
		cur = nil
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		cur = append(cur, strings.TrimSpace(line))
	}
	flush()
	return blocks
}

func splitLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func indexParagraphs(parts []string) []domain.Paragraph {
	out := make([]domain.Paragraph, 0, len(parts))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, domain.Paragraph{
			Index: i,
			Text:  part,
		})
	}
	for i := range out {
		out[i].Index = i
	}
	return out
}
