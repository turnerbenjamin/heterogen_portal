// Package handlers contains HTTP handlers for the application.
//
// This file exposes handlers for the users resource
package handlers

import (
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/constants"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

type BusinessesHandler struct {
	tableColumns         []ColumnData[*model.BusinessesModel]
	defaultHeaderColumns []HeaderColumn[*model.BusinessesModel]
	templateStore        TemplateStore
}

func NewBusinessesHandler(templateStore TemplateStore) *BusinessesHandler {

	return &BusinessesHandler{
		templateStore: templateStore,
		tableColumns:  OrderedColumnData(businessTableColumns),
		defaultHeaderColumns: []HeaderColumn[*model.BusinessesModel]{
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
	c *PipelineContext[*TableState[*model.BusinessesModel]],
) *AppError {
	pageConfig := templates.PageConfig{
		ContentOnly: r.Header.Get(constants.HxRequestHeaderRequest) != "",
		Title:       "HETEROGEN - BUSINESSES",
	}

	c.state.TableName = "Businesses"
	c.state.SchemaVersion = 1
	c.state.ColumnDefinitions = h.tableColumns
	c.state.TableHeader = h.defaultHeaderColumns

	err := h.templateStore.Execute(
		templates.TmplPageUsers,
		w,
		templates.TemplateArgs{PageConfig: pageConfig, Data: c.state},
	)
	if err != nil {
		return NewServerError(err)
	}
	return nil
}
