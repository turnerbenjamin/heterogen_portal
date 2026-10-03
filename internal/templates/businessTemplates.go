package templates

import (
	"io"

	"github.com/turnerbenjamin/heterogen_portal/internal/icons"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack/querymodel"
)

type HeaderColumn[T querymodel.TableModel] struct {
	Column ColumnData[T]
	Width  string
}

type ColumnData[T querymodel.TableModel] struct {
	ColumnName  string
	DisplayName string
	DataType    model.DataType
	Path        string
	Expansion   string
	GetData     func(T) string
}

type TableData[T querymodel.TableModel] struct {
	TableName         string
	SchemaVersion     uint64
	ColumnDefinitions []ColumnData[T]
	TableHeader       []HeaderColumn[T]
	Rows              []T
}

func (s *Store) ExecutePageBusinesses(
	w io.Writer,
	config TemplateConfig[TableData[*model.BusinessesModel]],
) error {
	d := templateData[TableData[*model.BusinessesModel]]{
		PageConfig: config.PageConfig,
		Data:       config.Data,
		WebResources: WebResourceDependencies{
			HTMX:      true,
			HG_COMMON: true,
		},
		Icons: icons.Set,
	}
	return s.templates.ExecuteTemplate(w, TmplPageBusinesses, d)
}
