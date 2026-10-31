package handlers

import (
	"fmt"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/tables"
	"github.com/turnerbenjamin/querystack/querymodel"
)

func MainTableRequestEnrichment[MT querymodel.TableModel, PT TableState[MT]]() Middleware[PT] {
	return func(next AppHandler[PT]) AppHandler[PT] {
		return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[PT]) *etc.AppError {
			d := c.state.GetTableData()

			requestTarget, error := parseRequestTarget(r)
			if error != nil {
				return etc.NewServerError(error)
			}
			d.RequestTarget = requestTarget

			tableQuery, err := parseTableQuery(r, d.TableDefinition)
			if err != nil {
				return etc.NewServerError(err)
			}
			d.Query = tableQuery

			return next(w, r, c)
		}
	}
}

func parseTableQuery[T querymodel.TableModel](
	r *http.Request,
	tableDefinition model.TableDefinition[T],
) (*tables.TableQuery[T], error) {
	queryValues := r.URL.Query()

	filterString := queryValues.Get("filter")

	selectQueryString := queryValues.Get("select")
	selectQuery, err := parseSelectQuery(selectQueryString, tableDefinition)
	if err != nil {
		return nil, err
	}

	return tables.NewTableQuery(selectQuery, filterString), nil
}

func parseSelectQuery[T querymodel.TableModel](
	selectQuery string,
	tableDefinition model.TableDefinition[T],
) (tables.SelectQuery[T], error) {
	if selectQuery == "" {
		return tables.SelectQuery[T]{
			Columns:   make([]model.TableColumn[T], 0),
			ColumnSet: map[string]struct{}{},
		}, nil
	}
	selectQueryLen := len(selectQuery)

	query := tables.SelectQuery[T]{
		Columns:   []model.TableColumn[T]{},
		ColumnSet: map[string]struct{}{},
	}

	li := 0
	ri := 0

	for ri < selectQueryLen {
		nextC := byte(0)
		if (ri + 1) < selectQueryLen {
			nextC = selectQuery[ri+1]
		}

		if nextC != ',' && nextC != byte(0) {
			ri++
			continue
		}

		colPath := string(selectQuery[li : ri+1])
		col, exists := tableDefinition.MainGridColumnsMap[colPath]
		if !exists {
			return query, fmt.Errorf(
				"unable to parse select columns: invalid column path '%s'",
				colPath,
			)
		}

		if _, selected := query.ColumnSet[colPath]; selected {
			return query, fmt.Errorf(
				"unable to parse select columns: duplicate column found '%s'",
				colPath,
			)
		}

		query.Columns = append(query.Columns, col)
		query.ColumnSet[colPath] = struct{}{}

		ri += 2
		li = ri
	}

	return query, nil
}

func parseRequestTarget(r *http.Request) (tables.MainTableRequestTarget, error) {
	target := r.Header.Get(constants.HxRequestHeaderTarget)

	switch target {
	case "":
		return tables.MainTableRequestTargetFullPage, nil
	case "div#page-content":
		return tables.MainTableRequestTargetPageContent, nil
	case "table#main-table":
		return tables.MainTableRequestTargetTable, nil
	case "tbody#main-table-body":
		return tables.MainTableRequestTargetTableBody, nil
	default:
		return tables.MainTableRequestTargetFullPage, fmt.Errorf(
			"unable to enrich main table request: unrecognised target: '%s'",
			target,
		)
	}
}
