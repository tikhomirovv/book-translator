package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// promptLine asks the user for a single line of input.
func promptLine(label string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", label)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// promptTranslateFlags collects missing translate flags interactively.
func promptTranslateFlags(input, output, to, promptType string) (string, string, string, string, error) {
	var err error
	if input == "" {
		input, err = promptLine("Input PDF path")
		if err != nil {
			return "", "", "", "", err
		}
	}
	if output == "" {
		output, err = promptLine("Output Markdown path")
		if err != nil {
			return "", "", "", "", err
		}
	}
	if to == "" {
		to, err = promptLine("Target language code (e.g. ru)")
		if err != nil {
			return "", "", "", "", err
		}
	}
	if promptType == "" {
		promptType, err = promptLine("Prompt type (nonfiction/fiction, Enter=nonfiction)")
		if err != nil {
			return "", "", "", "", err
		}
		if promptType == "" {
			promptType = "nonfiction"
		}
	}
	return input, output, to, promptType, nil
}

// promptTranslationID asks for a translation id when the flag is omitted.
func promptTranslationID(id string) (string, error) {
	if id != "" {
		return id, nil
	}
	return promptLine("Translation ID")
}
