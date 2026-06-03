package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tikhomirovv/book-translator/internal/application/translate"
)

var (
	translateInput      string
	translateOutput     string
	translateTo         string
	translatePromptType string
)

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "Start a new book translation",
	RunE:  runTranslate,
}

func init() {
	translateCmd.Flags().StringVarP(&translateInput, "input", "i", "", "Source PDF path")
	translateCmd.Flags().StringVarP(&translateOutput, "output", "o", "", "Output Markdown path")
	translateCmd.Flags().StringVar(&translateTo, "to", "", "Target language code")
	translateCmd.Flags().StringVar(&translatePromptType, "prompt-type", "", "Prompt profile (nonfiction, fiction)")
	rootCmd.AddCommand(translateCmd)
}

func runTranslate(cmd *cobra.Command, args []string) error {
	a, err := requireApp()
	if err != nil {
		return err
	}

	input, output, to, promptType, err := promptTranslateFlags(
		translateInput, translateOutput, translateTo, translatePromptType, a.AllowedLanguages,
	)
	if err != nil {
		return err
	}
	if err := validateTranslateArgs(input, output, to); err != nil {
		return err
	}

	a.Logger.Info().Str("input", input).Str("output", output).Str("to", to).Msg("starting translation")

	var reporter func(completed, total int)
	a.Start.OnProgress = func(completed, total int) {
		if reporter == nil {
			reporter = newProgressReporter(total, "Translating")
		}
		reporter(completed, total)
	}
	defer finishProgress()

	result, err := a.Start.Execute(context.Background(), translate.StartTranslationRequest{
		SourcePath: input,
		OutputPath: output,
		TargetLang: to,
		PromptType: promptType,
	})
	if err != nil {
		return err
	}

	a.Logger.Info().Str("translation_id", result.TranslationID).Msg("translation completed")
	fmt.Fprintf(cmd.OutOrStdout(), "Translation complete.\nID: %s\nOutput: %s\n", result.TranslationID, output)
	return nil
}
