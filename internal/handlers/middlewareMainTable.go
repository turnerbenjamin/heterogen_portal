package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
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

			tableQuery, err := parseTableQuery(r, d.ColumnDefinitions)
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
	columns map[string]templates.ColumnData[T],
) (*templates.TableQuery[T], error) {
	queryValues := r.URL.Query()
	selectQuery := queryValues.Get("select")
	if selectQuery == "" {
		return nil, nil
	}
	selectQueryLen := len(selectQuery)

	query := templates.TableQuery[T]{
		SelectCols:    []templates.ColumnData[T]{},
		SelectColsSet: map[string]struct{}{},
	}

	selectBuilder := strings.Builder{}
	selectCount := 0

	expandBuilder := strings.Builder{}
	expandCount := 0

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
			return nil, fmt.Errorf(
				"unable to parse select columns: invalid column path '%s'",
				colPath,
			)
		}

		if _, selected := query.SelectColsSet[colPath]; selected {
			return nil, fmt.Errorf(
				"unable to parse select columns: duplicate column found '%s'",
				colPath,
			)
		}

		query.SelectCols = append(query.SelectCols, col)
		query.SelectColsSet[colPath] = struct{}{}

		if col.Expansion != "" {
			if expandCount > 0 {
				expandBuilder.WriteByte(',')
			}
			expandBuilder.WriteString(col.Expansion)
			expandCount++
		} else {
			if selectCount > 0 {
				selectBuilder.WriteByte(',')
			}
			selectBuilder.WriteString(col.Path)
			selectCount++
		}

		ri += 2
		li = ri
	}

	query.QueryString = buildQueryString(selectBuilder, expandBuilder)
	return &query, nil
}

func parseRequestTarget(r *http.Request) (templates.MainTableRequestTarget, error) {
	target := r.Header.Get(constants.HxRequestHeaderTarget)

	switch target {
	case "":
		return templates.MainTableRequestTargetFullPage, nil
	case "div#page-content":
		return templates.MainTableRequestTargetPageContent, nil
	case "table#main-table":
		return templates.MainTableRequestTargetTable, nil
	case "tbody#main-table-body":
		return templates.MainTableRequestTargetTableBody, nil
	default:
		return templates.MainTableRequestTargetFullPage, fmt.Errorf(
			"unable to enrich main table request: unrecognised target: '%s'",
			target,
		)
	}
}

func newTableQuery[T querymodel.TableModel](selectCols []templates.ColumnData[T]) *templates.TableQuery[T] {
	selectBuilder := strings.Builder{}
	expandBuilder := strings.Builder{}
	selectColsSet := make(map[string]struct{}, len(selectCols))

	expandCount := 0
	selectCount := 0
	for _, c := range selectCols {
		selectColsSet[c.Path] = struct{}{}
		if c.Expansion != "" {
			if expandCount > 0 {
				expandBuilder.WriteByte(',')
			}
			expandBuilder.WriteString(c.Expansion)
			expandCount++
		} else {
			if selectCount > 0 {
				selectBuilder.WriteByte(',')
			}
			selectBuilder.WriteString(c.Path)
			selectCount++
		}
	}

	return &templates.TableQuery[T]{
		SelectCols:    selectCols,
		SelectColsSet: selectColsSet,
		QueryString:   buildQueryString(selectBuilder, expandBuilder),
	}
}

func buildQueryString(selectBuilder strings.Builder, expandBuilder strings.Builder) string {
	selectStmt := selectBuilder.String()
	expandStmt := expandBuilder.String()

	if selectStmt == "" && expandStmt == "" {
		return ""
	}

	if selectStmt != "" && expandStmt == "" {
		return fmt.Sprintf("select=%s", selectStmt)
	}

	if expandStmt != "" && selectStmt == "" {
		return fmt.Sprintf("expand=%s", expandStmt)
	}

	return fmt.Sprintf("select=%s&expand=%s", selectStmt, expandStmt)
}
