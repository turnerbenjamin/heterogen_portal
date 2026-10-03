package templates

import (
	"io"

	"github.com/turnerbenjamin/heterogen_portal/internal/icons"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
)

func (s *Store) ExecutePageUsers(
	w io.Writer,
	config TemplateConfig[TableData[*model.UsersModel]],
) error {
	d := templateData[TableData[*model.UsersModel]]{
		PageConfig: config.PageConfig,
		Data:       config.Data,
		WebResources: WebResourceDependencies{
			HTMX:      true,
			HG_COMMON: true,
		},
		Icons: icons.Set,
	}
	return s.templates.ExecuteTemplate(w, TmplPageUsers, d)
}
