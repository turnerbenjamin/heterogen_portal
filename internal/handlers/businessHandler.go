// Package handlers contains HTTP handlers for the application.
//
// This file exposes handlers for the users resource
package handlers

import (
	"context"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
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

	defaultQuery *templates.TableQuery[*model.BusinessesModel]
}

func NewBusinessesHandler(
	templateStore TemplateStore,
	businessesService BusinessesService,
) (*BusinessesHandler, error) {
	h := &BusinessesHandler{
		templateStore:     templateStore,
		businessesService: businessesService,

		orderedTableColumns: OrderedColumnData(businessTableColumns),
		tableColumns:        businessTableColumns,

		defaultQuery: newTableQuery([]templates.ColumnData[*model.BusinessesModel]{
			columnDataBusinessesTradingName,
			columnDataBusinessesReference,
			columnDataBusinessesBusinessType,
			columnDataBusinessesCounty,
		}),
	}

	return h, nil
}

// GetRootHandler returns a handler for the application root.
func (h *BusinessesHandler) GetRoot(
	w http.ResponseWriter,
	r *http.Request,
	c *PipelineContext[*BusinessesTableState],
) *etc.AppError {
	pageConfig := templates.PageConfig{
		ContentOnly: DoSendContentOnly(r),
		Title:       "HETEROGEN - BUSINESSES",
	}

	d := c.state.TableData

	d.TableName = "Businesses"
	d.SchemaVersion = 1

	d.OrderedColumnDefinitions = h.orderedTableColumns
	d.ColumnDefinitions = h.tableColumns

	if d.Query == nil {
		d.Query = h.defaultQuery
	}

	result, err := h.businessesService.ExecuteQuery(
		r.Context(),
		d.Query.QueryString,
	)
	if err != nil {
		return etc.NewServerError(err)
	}

	d.Rows = result.Data
	c.state.TableData = d

	err = h.templateStore.ExecutePageBusinesses(
		r.Context(),
		w,
		pageConfig,
		c.state.TableData,
		r.URL.Path,
	)
	if err != nil {
		return etc.NewServerError(err)
	}
	return nil
}
