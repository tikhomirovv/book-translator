package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	extracttext "github.com/tikhomirovv/book-translator/internal/application/extract"
)

var (
	extractInput  string
	extractOutput string
)

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract text from a source file without translation",
	Long:  "Run the PDF/text extraction step only and write normalized paragraphs to a file. Useful for debugging extraction before calling the LLM.",
	RunE:  runExtract,
}

func init() {
	extractCmd.Flags().StringVarP(&extractInput, "input", "i", "", "Source PDF path")
	extractCmd.Flags().StringVarP(&extractOutput, "output", "o", "", "Output text file path")
	rootCmd.AddCommand(extractCmd)
}

func runExtract(cmd *cobra.Command, args []string) error {
	a, err := requireApp()
	if err != nil {
		return err
	}

	input, output, err := promptExtractFlags(extractInput, extractOutput)
	if err != nil {
		return err
	}
	if err := validateExtractArgs(input, output); err != nil {
		return err
	}

	a.Logger.Info().Str("input", input).Str("output", output).Msg("extracting source text")

	result, err := a.Extract.Execute(context.Background(), extracttext.ExtractSourceRequest{
		InputPath:  input,
		OutputPath: output,
	})
	if err != nil {
		return err
	}

	a.Logger.Debug().
		Int("paragraphs", result.ParagraphCount).
		Str("output", result.OutputPath).
		Msg("extract complete")

	fmt.Fprintf(cmd.OutOrStdout(), "Extracted %d paragraphs.\nOutput: %s\n", result.ParagraphCount, result.OutputPath)
	return nil
}

func promptExtractFlags(input, output string) (string, string, error) {
	var err error
	if input == "" {
		input, err = promptRequiredLine("Input PDF path", "Input PDF path")
		if err != nil {
			return "", "", err
		}
	}
	if output == "" {
		output, err = promptRequiredLine("Output text path", "Output text path")
		if err != nil {
			return "", "", err
		}
	}
	return input, output, nil
}

func validateExtractArgs(input, output string) error {
	var missing []string
	if input == "" {
		missing = append(missing, "input (--input / -i)")
	}
	if output == "" {
		missing = append(missing, "output (--output / -o)")
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("missing required: %s", strings.Join(missing, ", "))
}
