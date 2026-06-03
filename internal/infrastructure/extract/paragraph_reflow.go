package extract

import (
	"strings"
	"unicode"
)

// noiseLinePatterns drop common PDF watermark / garbage lines before reflow.
var noiseLinePatterns = []string{
	"OceanofPDF.com",
}

// ReflowPlainText joins PDF line wraps into paragraphs separated by blank lines.
func ReflowPlainText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")

	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || isNoiseLine(line) {
			continue
		}
		cleaned = append(cleaned, line)
	}
	if len(cleaned) == 0 {
		return ""
	}

	var paragraphs []string
	var cur strings.Builder
	for _, line := range cleaned {
		if cur.Len() == 0 {
			cur.WriteString(line)
			continue
		}
		prev := cur.String()
		if shouldJoinLines(prev, line) {
			joinLine(&cur, prev, line)
			continue
		}
		paragraphs = append(paragraphs, prev)
		cur.Reset()
		cur.WriteString(line)
	}
	if cur.Len() > 0 {
		paragraphs = append(paragraphs, cur.String())
	}

	return strings.Join(paragraphs, "\n\n")
}

func isNoiseLine(line string) bool {
	for _, noise := range noiseLinePatterns {
		if strings.EqualFold(line, noise) {
			return true
		}
	}
	// Drop lines that are only replacement glyphs / punctuation.
	letters, digits := 0, 0
	for _, r := range line {
		if unicode.IsLetter(r) {
			letters++
		}
		if unicode.IsDigit(r) {
			digits++
		}
	}
	return letters == 0 && digits == 0
}

func shouldJoinLines(prev, next string) bool {
	prev = strings.TrimSpace(prev)
	next = strings.TrimSpace(next)
	if prev == "" || next == "" {
		return false
	}

	// Figure captions are standalone; body text starts on the next line.
	if isFigureCaption(prev) && startsProseLine(next) {
		return false
	}

	// Hyphenated word split across lines.
	if strings.HasSuffix(prev, "-") {
		return true
	}

	// Wrapped ALL-CAPS headings (e.g. "THE FOO\nBAR BAZ").
	if isMostlyUppercase(prev) && isMostlyUppercase(next) {
		return true
	}
	// Heading title followed by subtitle line.
	if isMostlyUppercase(prev) && strings.HasSuffix(prev, ":") {
		return true
	}
	// Heading block finished; body text starts on the next line.
	if isMostlyUppercase(prev) && hasLowercase(next) {
		return false
	}

	// Short standalone lines (TOC entries, labels) stay separate.
	if len(prev) < 20 && !lineLooksContinued(prev) {
		return false
	}

	// Sentence continues on the next line.
	if lineLooksContinued(prev) {
		return true
	}

	// Next line starts lowercase — almost always a wrapped continuation.
	if firstRune(next) != 0 && unicode.IsLower(firstRune(next)) {
		return true
	}

	// Previous line does not end a sentence — join wrapped prose.
	if !endsSentence(prev) {
		return true
	}

	return false
}

func joinLine(cur *strings.Builder, prev, next string) {
	cur.Reset()
	prev = strings.TrimSpace(prev)
	if strings.HasSuffix(prev, "-") {
		cur.WriteString(strings.TrimSuffix(prev, "-"))
		cur.WriteString(next)
		return
	}
	cur.WriteString(prev)
	cur.WriteByte(' ')
	cur.WriteString(next)
}

func lineLooksContinued(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	if strings.HasSuffix(line, "-") || strings.HasSuffix(line, ",") || strings.HasSuffix(line, ";") {
		return true
	}
	// Long lines ending in lowercase are usually mid-sentence wraps.
	if len(line) >= 25 && unicode.IsLower(lastRune(line)) {
		return true
	}
	return false
}

func endsSentence(line string) bool {
	line = strings.TrimRight(line, "\"')]}»”")
	if line == "" {
		return false
	}
	switch lastRune(line) {
	case '.', '!', '?':
		return true
	default:
		return false
	}
}

func isFigureCaption(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "Figure ") && len(line) < 120
}

func startsProseLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	r := firstRune(line)
	if r == 0 {
		return false
	}
	// Body sentences usually start with a capital and contain lowercase words.
	return unicode.IsUpper(r) && hasLowercase(line)
}

func isMostlyUppercase(line string) bool {
	upper, lower := 0, 0
	for _, r := range line {
		if r >= 'A' && r <= 'Z' {
			upper++
		}
		if r >= 'a' && r <= 'z' {
			lower++
		}
	}
	if upper == 0 {
		return false
	}
	return lower == 0 || upper > lower*2
}

func hasLowercase(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}

func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}

func lastRune(s string) rune {
	var last rune
	for _, r := range s {
		last = r
	}
	return last
}
