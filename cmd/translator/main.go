// Command translator is the CLI entrypoint for book-translator.
package main

import (
	"fmt"
	"os"

	"github.com/tikhomirovv/book-translator/internal/interfaces/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
