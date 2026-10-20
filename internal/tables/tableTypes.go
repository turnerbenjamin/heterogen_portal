package tables

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

type FilterItem interface {
	isFilterItem()
}

type FilterGroup struct {
	GroupType querymodel.LogicalOperator
	Items     []FilterItem
}

func (g FilterGroup) isFilterItem() {}

type FilterRule struct {
	ColumnType model.DataType
	ColumnPath string

	Operator           *ComparisonOperator
	SupportedOperators []*ComparisonOperator

	Value querymodel.Value
}

func (r FilterRule) isFilterItem() {}

type TableData[T querymodel.TableModel] struct {
	TableName     string
	SchemaVersion uint64
	RequestTarget MainTableRequestTarget

	OrderedColumnDefinitions []ColumnData[T]

	ColumnDefinitions map[string]ColumnData[T]

	Query *TableQuery[T]

	Rows   []T
	Filter FilterItem
}
