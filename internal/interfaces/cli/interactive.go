package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var stdinReader = bufio.NewReader(os.Stdin)

// promptLine asks the user for a single line of input.
func promptLine(label string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", label)
	line, err := stdinReader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// promptRequiredLine re-prompts until the user enters a non-empty value.
func promptRequiredLine(label, fieldName string) (string, error) {
	for attempt := 0; attempt < 3; attempt++ {
		fmt.Fprintf(os.Stderr, "%s: ", label)
		line, err := stdinReader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}
		value := strings.TrimSpace(line)
		if value != "" {
			return value, nil
		}
		if err == io.EOF {
			return "", fmt.Errorf("%s is required", fieldName)
		}
		fmt.Fprintf(os.Stderr, "%s is required.\n", fieldName)
	}
	return "", fmt.Errorf("%s is required", fieldName)
}

// promptTranslateFlags collects missing translate flags interactively.
func promptTranslateFlags(input, output, to, promptType string, allowedLangs []string) (string, string, string, string, error) {
	var err error

	if input == "" {
		input, err = promptRequiredLine("Input PDF path", "Input PDF path")
		if err != nil {
			return "", "", "", "", err
		}
	}
	if output == "" {
		output, err = promptRequiredLine("Output Markdown path", "Output Markdown path")
		if err != nil {
			return "", "", "", "", err
		}
	}
	if to == "" {
		label := formatTargetLangPrompt(allowedLangs)
		to, err = promptRequiredLine(label, "Target language")
		if err != nil {
			return "", "", "", "", err
		}
	}
	if promptType == "" {
		promptType, err = promptLine("Prompt type (nonfiction/fiction, Enter=nonfiction)")
		if err != nil && err != io.EOF {
			return "", "", "", "", err
		}
		if promptType == "" {
			promptType = "nonfiction"
		}
	}
	return input, output, to, promptType, nil
}

func formatTargetLangPrompt(allowedLangs []string) string {
	if len(allowedLangs) == 0 {
		return "Target language code (e.g. ru)"
	}
	return fmt.Sprintf("Target language code (allowed: %s)", strings.Join(allowedLangs, ", "))
}

// validateTranslateArgs returns an error naming every missing required field.
func validateTranslateArgs(input, output, to string) error {
	var missing []string
	if input == "" {
		missing = append(missing, "input (--input / -i)")
	}
	if output == "" {
		missing = append(missing, "output (--output / -o)")
	}
	if to == "" {
		missing = append(missing, "target language (--to)")
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("missing required: %s", strings.Join(missing, ", "))
}

// promptTranslationID asks for a translation id when the flag is omitted.
func promptTranslationID(id string) (string, error) {
	if id != "" {
		return id, nil
	}
	return promptRequiredLine("Translation ID", "Translation ID")
}

func validateTranslationID(id string) error {
	if id == "" {
		return fmt.Errorf("missing required: translation id (--id)")
	}
	return nil
}
