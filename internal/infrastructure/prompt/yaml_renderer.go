package prompt

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/tikhomirovv/book-translator/internal/domain"
	"github.com/tikhomirovv/book-translator/internal/domain/ports"
	"github.com/tikhomirovv/book-translator/internal/infrastructure/config"
)

// Supported template keys loaded from config PromptSet fields.
const (
	TemplateSystem            = "system"
	TemplateTranslation       = "translation"
	TemplateContextExtraction = "context_extraction"
)

// YAMLRenderer renders prompts from config PromptSet entries using text/template.
type YAMLRenderer struct {
	// templates[promptType][templateKey] holds a parsed template.
	templates map[string]map[string]*template.Template
}

// NewYAMLRenderer builds a renderer from config prompts. Templates are parsed once at startup.
func NewYAMLRenderer(prompts map[string]config.PromptSet) (*YAMLRenderer, error) {
	r := &YAMLRenderer{
		templates: make(map[string]map[string]*template.Template, len(prompts)),
	}

	for promptType, set := range prompts {
		entries := map[string]string{
			TemplateSystem:            set.System,
			TemplateTranslation:       set.Translation,
			TemplateContextExtraction: set.ContextExtraction,
		}

		parsed := make(map[string]*template.Template, len(entries))
		for key, body := range entries {
			if body == "" {
				continue
			}
			name := fmt.Sprintf("%s/%s", promptType, key)
			tpl, err := template.New(name).Parse(body)
			if err != nil {
				return nil, fmt.Errorf("parse prompt template %q: %w", name, err)
			}
			parsed[key] = tpl
		}

		if len(parsed) == 0 {
			return nil, fmt.Errorf("%w: prompt type %q has no templates", domain.ErrInvalidInput, promptType)
		}

		r.templates[promptType] = parsed
	}

	return r, nil
}

// Render executes the template for promptType and templateKey with data.
func (r *YAMLRenderer) Render(promptType, templateKey string, data ports.PromptData) (string, error) {
	byType, ok := r.templates[promptType]
	if !ok {
		return "", fmt.Errorf("%w: unknown prompt type %q", domain.ErrInvalidInput, promptType)
	}

	tpl, ok := byType[templateKey]
	if !ok {
		return "", fmt.Errorf("%w: unknown template key %q for prompt type %q", domain.ErrInvalidInput, templateKey, promptType)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render prompt %q/%q: %w", promptType, templateKey, err)
	}

	return buf.String(), nil
}
