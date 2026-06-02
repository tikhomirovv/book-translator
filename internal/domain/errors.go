package domain

import "errors"

var (
	// ErrNotFound is returned when a translation or resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidLanguage is returned when target language is not allowed.
	ErrInvalidLanguage = errors.New("invalid target language")
	// ErrInvalidInput is returned for bad user input.
	ErrInvalidInput = errors.New("invalid input")
)
