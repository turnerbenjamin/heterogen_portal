// Package handlers contains HTTP handlers for the application.
//
// This file exposes handlers for the users resource
package handlers

import (
	"context"
	"net/http"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/tables"
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
	orderedTableColumns []tables.ColumnData[*model.BusinessesModel]
	tableColumns        map[string]tables.ColumnData[*model.BusinessesModel]

	defaultQuery *tables.TableQuery[*model.BusinessesModel]
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

		defaultQuery: newTableQuery([]tables.ColumnData[*model.BusinessesModel]{
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

	// DUMMY FILTER
	d.Query.QueryString +=
		"&filter=reference contains '0' and (county eq 'Dorset' or county " +
			"contains 'shire' or business_type eq 1) and created_at lt " +
			"datetime(2026-10-15T07:30:44.958590557+01:00) and " +
			"(created_by_id/display_name startswith 'p') and (trading_name ne null " +
			" or (trading_name startswith 'T' and trading_name endswith 't'))"
	// d.Query.QueryString += "&filter=reference contains '0' or reference startswith '0'"
	// d.Query.QueryString += "&filter=reference contains '0'"

	result, err := h.businessesService.ExecuteQuery(
		r.Context(),
		d.Query.QueryString,
	)
	if err != nil {
		return etc.NewServerError(err)
	}

	tableFilter, err := tables.BuildTableFilter(
		d.ColumnDefinitions,
		result.Filter,
	)
	if err != nil {
		return etc.NewServerError(err)
	}

	d.Filter = tableFilter
	d.Rows = result.Data

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
