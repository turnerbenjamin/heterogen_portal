// Package handlers contains HTTP handlers for the application.
//
// This file exposes handlers for the users resource
package handlers

import (
	"context"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"github.com/turnerbenjamin/querystack/querymodel"
)

type BusinessesService interface {
	ExecuteQuery(
		ctx context.Context,
		queryString string,
	) (*querymodel.TypedExecuteResult[*model.BusinessesModel], error)
}

type BusinessesHandler struct {
	templateStore       TemplateStore
	businessesService   BusinessesService
	orderedTableColumns []templates.ColumnData[*model.BusinessesModel]
	tableColumns        map[string]templates.ColumnData[*model.BusinessesModel]

	defaultHeaderColumns    []templates.ColumnData[*model.BusinessesModel]
	defaultHeaderColumnsSet map[string]struct{}
}

func NewBusinessesHandler(
	templateStore TemplateStore,
	businessesService BusinessesService,
) (*BusinessesHandler, error) {
	defaultHeaderColumns := []templates.ColumnData[*model.BusinessesModel]{
		businessTableColumns[model.ColBusinessesTradingName],
		businessTableColumns[model.ColBusinessesReference],
		businessTableColumns[model.ColBusinessesBusinessType],
		businessTableColumns[model.ColBusinessesCounty],
	}

	h := &BusinessesHandler{
		templateStore:     templateStore,
		businessesService: businessesService,

		orderedTableColumns: OrderedColumnData(businessTableColumns),
		tableColumns:        businessTableColumns,

		defaultHeaderColumns:    defaultHeaderColumns,
		defaultHeaderColumnsSet: make(map[string]struct{}, len(defaultHeaderColumns)),
	}

	for _, c := range h.defaultHeaderColumns {
		h.defaultHeaderColumnsSet[c.Path] = struct{}{}
	}

	return h, nil
}

// GetRootHandler returns a handler for the application root.
func (h *BusinessesHandler) GetRoot(
	w http.ResponseWriter,
	r *http.Request,
	c *PipelineContext[*BusinessesTableState],
) *AppError {
	pageConfig := templates.PageConfig{
		ContentOnly: DoSendContentOnly(r),
		Title:       "HETEROGEN - BUSINESSES",
	}

	result, err := h.businessesService.ExecuteQuery(
		r.Context(),
		"",
	)
	if err != nil {
		return NewServerError(err)
	}

	c.state.TableData = templates.TableData[*model.BusinessesModel]{
		TableName:     "Businesses",
		SchemaVersion: 1,

		OrderedColumnDefinitions: h.orderedTableColumns,
		ColumnDefinitions:        h.tableColumns,

		HeaderColumns:   h.defaultHeaderColumns,
		HeaderColumnSet: h.defaultHeaderColumnsSet,

		Rows: result.Data,
	}

	err = h.templateStore.ExecutePageBusinesses(
		r.Context(),
		w,
		pageConfig,
		c.state.TableData,
	)
	if err != nil {
		return NewServerError(err)
	}
	return nil
}
