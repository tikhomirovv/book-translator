package domain

// Chunk is a group of paragraphs processed together by the LLM.
type Chunk struct {
	Index            int
	ParagraphStart   int
	ParagraphEnd     int
	SourceText       string
	TranslatedText   string
	OverlapFromPrev  string
}
