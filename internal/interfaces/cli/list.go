package cli

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List translations",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	a, err := requireApp()
	if err != nil {
		return err
	}

	items, err := a.List.Execute(context.Background())
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No translations yet.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSOURCE\tLANG\tPROGRESS\tSTATUS\tUPDATED")
	for _, item := range items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.ID, item.SourcePath, item.TargetLang, item.Progress, item.Status, item.UpdatedAt)
	}
	return w.Flush()
}
