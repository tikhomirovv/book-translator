package cli

import (
	"github.com/rs/zerolog"

	extracttext "github.com/tikhomirovv/book-translator/internal/application/extract"
	"github.com/tikhomirovv/book-translator/internal/application/query"
	"github.com/tikhomirovv/book-translator/internal/application/resume"
	"github.com/tikhomirovv/book-translator/internal/application/translate"
)

// App holds wired use cases for CLI commands.
type App struct {
	Start            *translate.StartTranslation
	Resume           *resume.ResumeTranslation
	Extract          *extracttext.ExtractSource
	Status           *query.GetStatus
	List             *query.ListTranslations
	Logger           zerolog.Logger
	AllowedLanguages []string
}

var app *App

// SetApp injects dependencies from the composition root.
func SetApp(a *App) {
	app = a
}

// AppInstance returns the wired app (for tests).
func AppInstance() *App {
	return app
}

func requireApp() (*App, error) {
	if app == nil {
		return nil, errAppNotInitialized
	}
	return app, nil
}
