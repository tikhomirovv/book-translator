package logging

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NewLogger configures zerolog to stderr with console-friendly output.
func NewLogger(level string) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(strings.ToLower(strings.TrimSpace(level)))
	if err != nil || level == "" {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	writer := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		NoColor:    true,
		TimeFormat: "15:04:05",
	}
	logger := zerolog.New(writer).With().Timestamp().Logger()
	log.Logger = logger
	return logger
}

// Writer returns stderr for progress bars and other CLI output that must not mix with logs on stdout.
func Writer() io.Writer {
	return os.Stderr
}
