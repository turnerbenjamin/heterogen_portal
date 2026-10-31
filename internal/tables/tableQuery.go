package tables

import (
	"fmt"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack/querymodel"
)

type TableQuery[T querymodel.TableModel] struct {
	sb           strings.Builder
	SelectQuery  SelectQuery[T]
	QueryString  string
	FilterString string
}

func NewTableQuery[T querymodel.TableModel](
	selectQuery SelectQuery[T],
	filterString string,
) *TableQuery[T] {
	q := &TableQuery[T]{
		SelectQuery:  selectQuery,
		FilterString: filterString,
	}

	q.buildQueryString()
	return q
}

func BuildSelectQuery[T querymodel.TableModel](
	cols []model.TableColumn[T],
) SelectQuery[T] {
	selectQuery := SelectQuery[T]{
		Columns:   cols,
		ColumnSet: make(map[string]struct{}, len(cols)),
	}

	for _, c := range selectQuery.Columns {
		selectQuery.ColumnSet[c.Path] = struct{}{}
	}

	return selectQuery
}

func (q *TableQuery[T]) SetSelect(selectQuery SelectQuery[T]) {
	q.SelectQuery = selectQuery
	q.buildQueryString()
}

func (q *TableQuery[T]) buildQueryString() {
	q.sb.Reset()

	operandCount := 0
	if q.FilterString != "" {
		fmt.Fprintf(&q.sb, "filter=%s", q.FilterString)
		operandCount++
	}

	expands := map[string][]string{}
	selectCount := 0
	for _, col := range q.SelectQuery.Columns {
		if col.Expansion != nil {
			if expands[col.Expansion.Path] == nil {
				expands[col.Expansion.Path] = []string{}
			}
			expands[col.Expansion.Path] = append(expands[col.Expansion.Path], col.Expansion.Select)
		} else {
			if selectCount == 0 {
				if operandCount > 0 {
					fmt.Fprintf(&q.sb, "&select=%s", col.Path)
				} else {
					fmt.Fprintf(&q.sb, "select=%s", col.Path)
				}
				operandCount++
			} else {
				fmt.Fprintf(&q.sb, ",%s", col.Path)
			}
			selectCount++
		}
	}

	if len(expands) == 0 {
		q.QueryString = q.sb.String()
	}

	expandCount := 0
	for path, selects := range expands {
		if expandCount == 0 {
			if operandCount > 0 {
				fmt.Fprintf(&q.sb, "&expand=%s(select=%s)", path, strings.Join(selects, ","))
			} else {
				fmt.Fprintf(&q.sb, "expand=%s(select=%s)", path, strings.Join(selects, ","))
			}
			operandCount++
		} else {
			fmt.Fprintf(&q.sb, ",%s(select=%s)", path, strings.Join(selects, ","))
		}
		expandCount++
	}
	q.QueryString = q.sb.String()
}
