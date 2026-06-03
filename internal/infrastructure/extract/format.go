package extract

import (
	"fmt"
	"strings"

	"github.com/tikhomirovv/book-translator/internal/domain"
)

// FormatParagraphs renders normalized paragraphs for inspection (same layout as translation store).
func FormatParagraphs(paragraphs []domain.Paragraph) string {
	var b strings.Builder
	for _, p := range paragraphs {
		fmt.Fprintf(&b, "# paragraph %d\n%s\n\n", p.Index, p.Text)
	}
	return b.String()
}
