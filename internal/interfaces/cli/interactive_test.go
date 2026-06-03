package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/interfaces/cli"
)

func TestValidateTranslateArgs_missingFields(t *testing.T) {
	t.Parallel()

	err := cli.ValidateTranslateArgs("", "out.md", "")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "input") || !strings.Contains(msg, "target language") {
		t.Fatalf("error = %q", msg)
	}
	if strings.Contains(msg, "output") {
		t.Fatalf("output should not be listed as missing: %q", msg)
	}
}

func TestPromptTranslateFlags_rejectsEmptyTargetLangAfterRetries(t *testing.T) {
	input := strings.NewReader("\n\n\n") // three empty lines for target lang retries
	prev := cli.SetStdinForTest(input)
	defer prev()

	_, _, _, _, err := cli.PromptTranslateFlagsForTest(
		"book.pdf", "book.md", "", "nonfiction", []string{"ru", "en"},
	)
	if err == nil {
		t.Fatal("expected error after empty target language retries")
	}
	if !strings.Contains(err.Error(), "Target language is required") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestPromptTranslateFlags_acceptsInteractiveValues(t *testing.T) {
	input := bytes.NewBufferString("ru\n")
	prev := cli.SetStdinForTest(input)
	defer prev()

	in, out, to, pt, err := cli.PromptTranslateFlagsForTest("book.pdf", "book.md", "", "nonfiction", []string{"ru", "en", "de"})
	if err != nil {
		t.Fatalf("PromptTranslateFlags: %v", err)
	}
	if in != "book.pdf" || out != "book.md" {
		t.Fatalf("paths not preserved: %q %q", in, out)
	}
	if to != "ru" || pt != "nonfiction" {
		t.Fatalf("to=%q pt=%q", to, pt)
	}
}

func TestPromptTranslateFlags_defaultPromptType(t *testing.T) {
	input := bytes.NewBufferString("\n") // Enter for default nonfiction
	prev := cli.SetStdinForTest(input)
	defer prev()

	_, _, _, pt, err := cli.PromptTranslateFlagsForTest("a.pdf", "b.md", "ru", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if pt != "nonfiction" {
		t.Fatalf("prompt type = %q", pt)
	}
}
