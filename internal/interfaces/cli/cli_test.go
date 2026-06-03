package cli_test

import (
	"testing"

	"github.com/tikhomirovv/book-translator/internal/interfaces/cli"
)

func TestRoot_hasExpectedSubcommands(t *testing.T) {
	root := cli.Root()
	names := map[string]bool{}
	for _, cmd := range root.Commands() {
		names[cmd.Name()] = true
	}

	for _, want := range []string{"translate", "extract", "resume", "status", "list", "version"} {
		if !names[want] {
			t.Errorf("missing subcommand %q", want)
		}
	}
}

func TestExecute_withoutAppReturnsError(t *testing.T) {
	cli.SetApp(nil)
	root := cli.Root()
	root.SetArgs([]string{"list"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error when app is not initialized")
	}
}
