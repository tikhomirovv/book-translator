// Command translator is the CLI entrypoint for book-translater.
package main

import (
	"fmt"
	"os"

	"github.com/tikhomirovv/book-translater/internal/interfaces/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
