package extract

import (
	"strings"

	"github.com/tikhomirovv/book-translater/internal/domain"
)

// NormalizeParagraphs splits raw text on blank lines and returns indexed paragraphs.
func NormalizeParagraphs(text string) []domain.Paragraph {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	parts := strings.Split(text, "\n\n")

	var out []domain.Paragraph
	idx := 0
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, domain.Paragraph{
			Index: idx,
			Text:  part,
		})
		idx++
	}
	return out
}
