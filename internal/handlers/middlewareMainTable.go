package handlers

import (
	"fmt"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"github.com/turnerbenjamin/querystack/querymodel"
)

func MainTableRequestEnrichment[MT querymodel.TableModel, PT TableState[MT]]() Middleware[PT] {
	return func(next AppHandler[PT]) AppHandler[PT] {
		return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[PT]) *AppError {
			d := c.state.GetTableData()

			requestTarget, error := parseRequestTarget(r)
			if error != nil {
				return NewServerError(error)
			}
			d.RequestTarget = requestTarget

			selectCols, selectColsSet, err := parseSelectedColumns(r, d.ColumnDefinitions)
			if err != nil {
				return NewServerError(err)
			}
			d.HeaderColumns = selectCols
			d.HeaderColumnSet = selectColsSet

			return next(w, r, c)
		}
	}
}

func parseSelectedColumns[T querymodel.TableModel](
	r *http.Request,
	columns map[string]templates.ColumnData[T],
) ([]templates.ColumnData[T], map[string]struct{}, error) {
	queryValues := r.URL.Query()

	selectQuery := queryValues.Get("select")
	if selectQuery == "" {
		return nil, nil, nil
	}
	selectQueryLen := len(selectQuery)

	selectCols := []templates.ColumnData[T]{}
	selectColsSet := map[string]struct{}{}

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
		col, exists := columns[colPath]
		if !exists {
			return nil, nil, fmt.Errorf(
				"unable to parse select columns: invalid column path '%s'",
				colPath,
			)
		}

		if _, selected := selectColsSet[colPath]; selected {
			return nil, nil, fmt.Errorf(
				"unable to parse select columns: duplicate column found '%s'",
				colPath,
			)
		}

		selectCols = append(selectCols, col)
		selectColsSet[colPath] = struct{}{}
		ri += 2
		li = ri
	}
	return selectCols, selectColsSet, nil
}

func parseRequestTarget(r *http.Request) (templates.MainTableRequestTarget, error) {
	target := r.Header.Get(constants.HxRequestHeaderTarget)

	switch target {
	case "":
		return templates.MainTableRequestTargetFullPage, nil
	case "page-content":
		return templates.MainTableRequestTargetPageContent, nil
	case "main-table":
		return templates.MainTableRequestTargetTable, nil
	case "main-table-body":
		return templates.MainTableRequestTargetTableBody, nil
	default:
		return templates.MainTableRequestTargetFullPage, fmt.Errorf(
			"unable to enrich main table request: unrecognised target: '%s'",
			target,
		)
	}
}
