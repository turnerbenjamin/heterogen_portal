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
	businessesService    BusinessesService
	tableColumns         []templates.ColumnData[*model.BusinessesModel]
	defaultHeaderColumns []templates.HeaderColumn[*model.BusinessesModel]
	templateStore        TemplateStore
}

func NewBusinessesHandler(
	templateStore TemplateStore,
	businessesService BusinessesService,
) *BusinessesHandler {

	return &BusinessesHandler{
		templateStore:     templateStore,
		businessesService: businessesService,
		tableColumns:      OrderedColumnData(businessTableColumns),
		defaultHeaderColumns: []templates.HeaderColumn[*model.BusinessesModel]{
			{Column: businessTableColumns[model.ColBusinessesTradingName]},
			{Column: businessTableColumns[model.ColBusinessesReference]},
			{Column: businessTableColumns[model.ColBusinessesBusinessType]},
			{Column: businessTableColumns[model.ColBusinessesCounty]},
		},
	}
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
		TableName:         "Businesses",
		SchemaVersion:     1,
		ColumnDefinitions: h.tableColumns,
		TableHeader:       h.defaultHeaderColumns,
		Rows:              result.Data,
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
