package cli

import (
	"bufio"
	"io"
)

// ValidateTranslateArgs exposes validation for tests.
func ValidateTranslateArgs(input, output, to string) error {
	return validateTranslateArgs(input, output, to)
}

// PromptTranslateFlagsForTest runs interactive flag collection with injectable stdin.
func PromptTranslateFlagsForTest(input, output, to, promptType string, allowedLangs []string) (string, string, string, string, error) {
	return promptTranslateFlags(input, output, to, promptType, allowedLangs)
}

// SetStdinForTest replaces stdin for prompt helpers; returns restore func.
func SetStdinForTest(r io.Reader) func() {
	prev := stdinReader
	stdinReader = bufio.NewReader(r)
	return func() { stdinReader = prev }
}
