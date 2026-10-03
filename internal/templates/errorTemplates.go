package templates

import (
	"io"

	"github.com/turnerbenjamin/heterogen_portal/internal/icons"
)

type ErrorTemplateData struct {
	Code       int
	Error      string
	SubErrors  []string
	InnerError error
}

func (s *Store) ExecutePageOutOfAppError(
	w io.Writer,
	config TemplateConfig[ErrorTemplateData],
) error {
	d := templateData[ErrorTemplateData]{
		PageConfig:   config.PageConfig,
		Data:         config.Data,
		WebResources: WebResourceDependencies{},
	}
	return s.templates.ExecuteTemplate(w, TmplPageOutOfAppErr, d)
}

func (s *Store) ExecuteComponentErrors(
	w io.Writer,
	config TemplateConfig[ErrorTemplateData],
) error {
	d := templateData[ErrorTemplateData]{
		PageConfig:   config.PageConfig,
		Data:         config.Data,
		WebResources: WebResourceDependencies{},
		Icons: icons.Set,
	}
	return s.templates.ExecuteTemplate(w, TmplComponentErrors, d)
}
