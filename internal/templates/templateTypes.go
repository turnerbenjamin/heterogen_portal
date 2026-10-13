package templates

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack/querymodel"
)

type MainTableRequestTarget string

const (
	MainTableRequestTargetFullPage    = MainTableRequestTarget("full-page")
	MainTableRequestTargetPageContent = MainTableRequestTarget("page-content")
	MainTableRequestTargetTable       = MainTableRequestTarget("main-table")
	MainTableRequestTargetTableBody   = MainTableRequestTarget("main-table-body")
)

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
	ContentOnly  bool
	Title        string
	ToastSuccess string
}

type HeaderColumn[T querymodel.TableModel] struct {
	Column ColumnData[T]
	Width  string
}

type ColumnData[T querymodel.TableModel] struct {
	Path        string
	DisplayName string
	DataType    model.DataType
	Expansion   string
	GetData     func(T) string
}

type TableQuery[T querymodel.TableModel] struct {
	SelectCols    []ColumnData[T]
	SelectColsSet map[string]struct{}
	QueryString   string
}

type TableData[T querymodel.TableModel] struct {
	TableName     string
	SchemaVersion uint64
	RequestTarget MainTableRequestTarget

	OrderedColumnDefinitions []ColumnData[T]

	ColumnDefinitions map[string]ColumnData[T]

	Query *TableQuery[T]

	Rows []T
}
