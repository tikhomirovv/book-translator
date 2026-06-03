// Package cli provides Cobra commands for the translator binary.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "translator",
	Short: "Translate books (PDF) to Markdown using LLM",
	Long:  "book-translator: resumable book translation with context memory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("use a subcommand (translate, extract, resume, status, list)")
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// Root returns the root cobra command (for tests).
func Root() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}
