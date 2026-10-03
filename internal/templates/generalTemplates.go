package templates

import (
	"io"

	"github.com/turnerbenjamin/heterogen_portal/internal/icons"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
)

type PageAppData struct {
	User *model.UsersModel
}

func (s *Store) ExecutePageApp(
	w io.Writer,
	config TemplateConfig[PageAppData],
) error {
	d := templateData[PageAppData]{
		PageConfig: config.PageConfig,
		Data:       config.Data,
		WebResources: WebResourceDependencies{
			HTMX:      true,
			HG_COMMON: true,
		},
		Icons: icons.Set,
	}
	return s.templates.ExecuteTemplate(w, TmplPageApp, d)
}

type PageUserSignedOutData struct{}

func (s *Store) ExecutePageUserSignedOut(
	w io.Writer,
	config TemplateConfig[PageUserSignedOutData],
) error {
	d := templateData[PageUserSignedOutData]{
		PageConfig:   config.PageConfig,
		WebResources: WebResourceDependencies{},
	}
	return s.templates.ExecuteTemplate(w, TmplPageUserSignedOut, d)
}
