package ports

// TokenCounter estimates token counts (tiktoken fallback in v0.2).
type TokenCounter interface {
	Count(text string) (int, error)
}

// NoopTokenCounter returns zero; used until tiktoken is implemented.
type NoopTokenCounter struct{}

// Count implements TokenCounter.
func (NoopTokenCounter) Count(string) (int, error) {
	return 0, nil
}
