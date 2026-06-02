package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tikhomirovv/book-translator/internal/application/resume"
)

var resumeID string

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume an interrupted translation",
	RunE:  runResume,
}

func init() {
	resumeCmd.Flags().StringVar(&resumeID, "id", "", "Translation ID")
	rootCmd.AddCommand(resumeCmd)
}

func runResume(cmd *cobra.Command, args []string) error {
	a, err := requireApp()
	if err != nil {
		return err
	}

	id, err := promptTranslationID(resumeID)
	if err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("translation id is required")
	}

	a.Logger.Info().Str("translation_id", id).Msg("resuming translation")

	var reporter func(completed, total int)
	a.Resume.OnProgress = func(completed, total int) {
		if reporter == nil {
			reporter = newProgressReporter(total, "Resuming")
		}
		reporter(completed, total)
	}
	defer finishProgress()

	if err := a.Resume.Execute(context.Background(), resume.ResumeTranslationRequest{
		TranslationID: id,
	}); err != nil {
		return err
	}

	a.Logger.Info().Str("translation_id", id).Msg("resume completed")
	fmt.Fprintf(cmd.OutOrStdout(), "Translation %s completed.\n", id)
	return nil
}
