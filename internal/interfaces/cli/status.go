package cli

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var statusID string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show translation progress",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().StringVar(&statusID, "id", "", "Translation ID")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	a, err := requireApp()
	if err != nil {
		return err
	}

	id, err := promptTranslationID(statusID)
	if err != nil {
		return err
	}
	if err := validateTranslationID(id); err != nil {
		return err
	}

	view, err := a.Status.Execute(context.Background(), id)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID\t%s\n", view.ID)
	fmt.Fprintf(w, "Source\t%s\n", view.SourcePath)
	fmt.Fprintf(w, "Target lang\t%s\n", view.TargetLang)
	fmt.Fprintf(w, "Status\t%s\n", view.Status)
	fmt.Fprintf(w, "Progress\t%d/%d\n", view.CompletedChunks, view.TotalChunks)
	fmt.Fprintf(w, "Tokens\t%d (input %d, output %d)\n",
		view.Usage.TotalTokens, view.Usage.PromptTokens, view.Usage.CompletionTokens)
	if view.LastError != "" {
		fmt.Fprintf(w, "Last error\t%s\n", view.LastError)
	}
	return w.Flush()
}
