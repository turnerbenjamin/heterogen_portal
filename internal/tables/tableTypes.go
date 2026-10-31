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
	Column model.TableColumn[T]
	Width  string
}

type SelectQuery[T querymodel.TableModel] struct {
	Columns   []model.TableColumn[T]
	ColumnSet map[string]struct{}
}

type FilterItem interface {
	isFilterItem()
}

type FilterGroup struct {
	GroupType querymodel.LogicalOperator
	Items     []FilterItem
}

func (g FilterGroup) isFilterItem() {}

type FilterRule[T querymodel.TableModel] struct {
	Column   *model.TableColumn[T]
	Operator *ComparisonOperator
	Value    FilterValue
}

func (r FilterRule[T]) isFilterItem() {}

type TableData[T querymodel.TableModel] struct {
	RequestTarget   MainTableRequestTarget
	TableDefinition model.TableDefinition[T]

	Query *TableQuery[T]

	Rows   []T
	Filter FilterItem
}
