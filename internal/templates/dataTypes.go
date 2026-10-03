package templates

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack/querymodel"
)

type WebResourceDependencies struct {
	HTMX      bool
	HG_COMMON bool
}

type ToastData struct {
	ToastSuccess string
}

type ErrorTemplateData struct {
	Code       int
	Error      string
	SubErrors  []string
	InnerError error
}

type PageConfig struct {
	ContentOnly             bool
	Title                   string
	ToastSuccess            string
	WebResourceDependencies WebResourceDependencies
}

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
