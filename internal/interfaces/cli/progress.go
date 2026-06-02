package cli

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"

	"github.com/tikhomirovv/book-translator/internal/infrastructure/logging"
)

// newProgressReporter renders chunk progress to stderr (compatible with zerolog on stderr).
func newProgressReporter(total int, label string) func(completed, total int) {
	if total <= 0 {
		return func(int, int) {}
	}
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetWriter(logging.Writer()),
		progressbar.OptionSetDescription(label),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(30),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetRenderBlankState(true),
	)
	return func(completed, total int) {
		_ = bar.Set(completed)
		if completed >= total {
			_, _ = fmt.Fprintln(logging.Writer())
		}
	}
}

// finishProgress clears the progress line before normal log output continues.
func finishProgress() {
	_, _ = fmt.Fprint(os.Stderr, "\r\033[K")
}
